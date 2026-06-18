package scheduler

import (
	"context"
	"cronjob-rpc/internal/model"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/zeromicro/go-zero/core/logx"
)

// Scheduler 定时任务调度器
// 只在 Leader 节点运行，扫描数据库中的任务，触发执行
type Scheduler struct {
	taskJobRepo  *model.TaskJobRepo
	executor     *Executor
	scanInterval time.Duration
	taskMap      map[string]cron.EntryID
	cron         *cron.Cron
	mu           sync.RWMutex
	ctx          context.Context
	cancel       context.CancelFunc
	running      bool
}

func NewScheduler(taskJobRepo *model.TaskJobRepo, executor *Executor, scanIntervalSec int) *Scheduler {
	return &Scheduler{
		taskJobRepo:  taskJobRepo,
		executor:     executor,
		scanInterval: time.Duration(scanIntervalSec) * time.Second,
		taskMap:      make(map[string]cron.EntryID),
	}
}

// Start 启动调度器
func (s *Scheduler) Start(ctx context.Context) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return
	}

	s.ctx, s.cancel = context.WithCancel(ctx)
	s.running = true

	// 初始化 cron 调度器（秒级精度）
	s.cron = cron.New(cron.WithSeconds())

	// 加载已有任务
	s.reloadTasks()

	// 启动 cron
	s.cron.Start()

	// 启动定时扫描协程（动态检测新增/更新/删除的任务）
	go s.scanLoop()

	logx.Info("Scheduler started")
}

// Stop 停止调度器
func (s *Scheduler) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return
	}

	if s.cancel != nil {
		s.cancel()
	}

	if s.cron != nil {
		<-s.cron.Stop().Done()
	}

	s.running = false
	s.taskMap = make(map[string]cron.EntryID)
	logx.Info("Scheduler stopped")
}

// scanLoop 定时扫描任务列表
func (s *Scheduler) scanLoop() {
	ticker := time.NewTicker(s.scanInterval)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			s.reloadTasks()
		}
	}
}

// reloadTasks 从数据库重新加载任务并同步到 cron 调度器
func (s *Scheduler) reloadTasks() {
	jobs, err := s.taskJobRepo.ListAllEnabled(s.ctx)
	if err != nil {
		logx.Errorf("Failed to list enabled tasks: %v", err)
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// 构建当前任务集合
	currentTasks := make(map[string]bool)
	for _, job := range jobs {
		currentTasks[job.Name] = true

		// 如果是新任务，注册到 cron
		if _, exists := s.taskMap[job.Name]; !exists {
			entryID, err := s.cron.AddFunc(job.CronExpr, s.makeTaskRunner(job))
			if err != nil {
				logx.Errorf("Failed to register cron task %s (expr: %s): %v", job.Name, job.CronExpr, err)
				continue
			}
			s.taskMap[job.Name] = entryID
			logx.Infof("Registered cron task: %s, expr: %s", job.Name, job.CronExpr)
		}
	}

	// 移除已删除或禁用的任务
	for name, entryID := range s.taskMap {
		if !currentTasks[name] {
			s.cron.Remove(entryID)
			delete(s.taskMap, name)
			logx.Infof("Removed cron task: %s", name)
		}
	}
}

// makeTaskRunner 创建任务执行闭包
func (s *Scheduler) makeTaskRunner(job model.TaskJob) func() {
	return func() {
		s.executor.Execute(s.ctx, job)
	}
}

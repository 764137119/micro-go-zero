package svc

import (
	"context"
	"cronjob-rpc/internal/config"
	"cronjob-rpc/internal/model"
	"cronjob-rpc/internal/scheduler"
	"time"

	"common/gormx"

	"github.com/zeromicro/go-zero/core/logx"
	clientv3 "go.etcd.io/etcd/client/v3"
	"gorm.io/gorm"
)

type ServiceContext struct {
	Config         config.Config
	DB             *gorm.DB
	TaskJobRepo    *model.TaskJobRepo
	TaskExecRepo   *model.TaskExecutionRepo
	EtcdClient     *clientv3.Client
	LeaderElection *scheduler.LeaderElection
	Scheduler      *scheduler.Scheduler
	Executor       *scheduler.Executor
}

func NewServiceContext(c config.Config) *ServiceContext {
	// 数据库连接
	db := gormx.MustNewDB(c.DB.DataSource)
	model.AutoMigrate(db)

	// etcd 客户端
	etcdClient, err := clientv3.New(clientv3.Config{
		Endpoints:   c.Etcd.Hosts,
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		logx.Must(err)
	}

	taskJobRepo := model.NewTaskJobRepo(db)
	taskExecRepo := model.NewTaskExecutionRepo(db)

	// 创建核心组件
	executor := scheduler.NewExecutor(etcdClient, taskJobRepo, taskExecRepo, c.Scheduler.TaskLockTTLSec)
	sched := scheduler.NewScheduler(taskJobRepo, executor, c.Scheduler.ScanIntervalSec)
	leaderElection := scheduler.NewLeaderElection(etcdClient, c.Etcd.Key, c.Scheduler.LeaderTTLSec, sched)

	svcCtx := &ServiceContext{
		Config:         c,
		DB:             db,
		TaskJobRepo:    taskJobRepo,
		TaskExecRepo:   taskExecRepo,
		EtcdClient:     etcdClient,
		LeaderElection: leaderElection,
		Scheduler:      sched,
		Executor:       executor,
	}

	return svcCtx
}

// Start 启动所有后台组件
func (s *ServiceContext) Start() {
	// 启动 Leader 选举（内部会启动调度器）
	go s.LeaderElection.Elect(context.Background())
	logx.Info("Leader election started")
}

// Stop 优雅关闭
func (s *ServiceContext) Stop() {
	logx.Info("Shutting down scheduler...")

	// 停止调度器
	if s.Scheduler != nil {
		s.Scheduler.Stop()
	}

	// 退出 Leader 选举
	if s.LeaderElection != nil {
		s.LeaderElection.Resign()
	}

	// 关闭 etcd 连接
	if s.EtcdClient != nil {
		_ = s.EtcdClient.Close()
	}

	// 关闭数据库连接
	if s.DB != nil {
		sqlDB, err := s.DB.DB()
		if err == nil {
			_ = sqlDB.Close()
		}
	}

	logx.Info("Scheduler shutdown complete")
}

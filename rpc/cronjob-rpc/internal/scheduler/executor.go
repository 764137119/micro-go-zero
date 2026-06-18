package scheduler

import (
	"bytes"
	"context"
	"cronjob-rpc/internal/model"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	clientv3 "go.etcd.io/etcd/client/v3"
)

// Executor 任务执行引擎
// 负责实际执行任务，通过 etcd 分布式锁保证全局唯一执行
type Executor struct {
	etcdClient   *clientv3.Client
	taskJobRepo  *model.TaskJobRepo
	taskExecRepo *model.TaskExecutionRepo
	lockTTLSec   int
	httpClient   *http.Client
}

func NewExecutor(etcdClient *clientv3.Client, taskJobRepo *model.TaskJobRepo, taskExecRepo *model.TaskExecutionRepo, lockTTLSec int) *Executor {
	return &Executor{
		etcdClient:   etcdClient,
		taskJobRepo:  taskJobRepo,
		taskExecRepo: taskExecRepo,
		lockTTLSec:   lockTTLSec,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Execute 执行单个任务（由 cron 调度触发）
func (e *Executor) Execute(ctx context.Context, job model.TaskJob) {
	now := time.Now()

	// 0. 检查上次执行是否还在运行中（双保险：DB 状态检查 + etcd 锁）
	running, err := e.taskExecRepo.FindLastRunning(ctx, job.Name)
	if err == nil && running != nil {
		logx.Infof("Task %s: previous execution(%d) still running(status=%s), skip this trigger",
			job.Name, running.ID, running.Status)
		return
	}

	// 1. 尝试获取任务级分布式锁
	lockKey := fmt.Sprintf("/cronjob/task/%s/lock", job.Name)
	locked, err := e.acquireLock(ctx, lockKey)
	if err != nil {
		logx.Errorf("Task %s: lock error: %v", job.Name, err)
		return
	}
	if !locked {
		logx.Infof("Task %s: lock held by another execution, skip", job.Name)
		return
	}

	// 2. 创建执行记录
	exec := &model.TaskExecution{
		TaskName:    job.Name,
		ScheduledAt: now,
		StartedAt:   now,
		Status:      model.ExecStatusDispatching,
		MaxRetries:  job.MaxRetries,
		ExecNode:    getNodeID(),
	}

	if err := e.taskExecRepo.Create(ctx, exec); err != nil {
		logx.Errorf("Failed to create execution record for task %s: %v", job.Name, err)
		e.releaseLock(lockKey)
		return
	}

	// 3. 通过 HTTP 调用业务方
	callErr := e.callTargetService(ctx, job, exec.ID)

	if callErr != nil {
		// 分发失败
		exec.Status = model.ExecStatusDispatchFailed
		exec.Result = fmt.Sprintf("dispatch error: %v", callErr)
		exec.FinishedAt = time.Now()
		_ = e.taskExecRepo.Update(ctx, exec)
		e.releaseLock(lockKey)
		logx.Errorf("Task %s dispatch failed: %v", job.Name, callErr)
		return
	}

	// 4. 分发成功，等待业务方回调
	exec.Status = model.ExecStatusDispatched
	exec.Result = fmt.Sprintf("dispatched to %s, executionId=%d", job.Target, exec.ID)
	if err := e.taskExecRepo.Update(ctx, exec); err != nil {
		logx.Errorf("Failed to update execution record: %v", err)
	}
	// 锁不释放！等待业务方 ReportExecution 时释放

	logx.Infof("Task %s dispatched successfully, executionId=%d, waiting for callback", job.Name, exec.ID)
}

// ExecuteOnce 手动触发一次执行
func (e *Executor) ExecuteOnce(ctx context.Context, job model.TaskJob) (int64, error) {
	now := time.Now()

	exec := &model.TaskExecution{
		TaskName:    job.Name,
		ScheduledAt: now,
		StartedAt:   now,
		Status:      model.ExecStatusDispatching,
		MaxRetries:  job.MaxRetries,
		ExecNode:    getNodeID(),
	}

	if err := e.taskExecRepo.Create(ctx, exec); err != nil {
		return 0, fmt.Errorf("create execution record: %w", err)
	}

	// 手动触发不通过分布式锁，直接执行
	callErr := e.callTargetService(ctx, job, exec.ID)

	if callErr != nil {
		exec.Status = model.ExecStatusDispatchFailed
		exec.Result = fmt.Sprintf("dispatch error: %v", callErr)
	} else {
		exec.Status = model.ExecStatusDispatched
		exec.Result = fmt.Sprintf("dispatched to %s", job.Target)
	}
	exec.FinishedAt = time.Now()
	_ = e.taskExecRepo.Update(ctx, exec)

	return exec.ID, nil
}

// ReportExecution 业务方回调，报告执行结果
func (e *Executor) ReportExecution(ctx context.Context, executionID int64, status string, result string) error {
	exec, err := e.taskExecRepo.FindByID(ctx, executionID)
	if err != nil {
		return fmt.Errorf("execution not found: %w", err)
	}

	if status == model.ExecStatusSuccess {
		exec.Status = model.ExecStatusSuccess
	} else {
		exec.Status = model.ExecStatusFailed
	}
	exec.Result = result
	exec.FinishedAt = time.Now()

	if err := e.taskExecRepo.Update(ctx, exec); err != nil {
		return fmt.Errorf("update execution: %w", err)
	}

	// 释放任务级锁，允许下次执行
	lockKey := fmt.Sprintf("/cronjob/task/%s/lock", exec.TaskName)
	e.releaseLock(lockKey)

	logx.Infof("Task %s execution %d reported: status=%s, result=%s",
		exec.TaskName, executionID, status, result)
	return nil
}

// Retry 重试指定执行记录
func (e *Executor) Retry(ctx context.Context, executionID int64) error {
	exec, err := e.taskExecRepo.FindByID(ctx, executionID)
	if err != nil {
		return fmt.Errorf("execution not found: %w", err)
	}

	job, err := e.taskJobRepo.FindByName(ctx, exec.TaskName)
	if err != nil {
		return fmt.Errorf("task not found: %w", err)
	}

	exec.RetryCount++
	exec.Status = model.ExecStatusDispatching
	exec.StartedAt = time.Now()
	_ = e.taskExecRepo.Update(ctx, exec)

	callErr := e.callTargetService(ctx, *job, exec.ID)
	exec.FinishedAt = time.Now()

	if callErr != nil {
		exec.Status = model.ExecStatusDispatchFailed
		exec.Result = fmt.Sprintf("retry dispatch error: %v", callErr)
	} else {
		exec.Status = model.ExecStatusDispatched
		exec.Result = fmt.Sprintf("retry dispatched to %s", job.Target)
	}

	return e.taskExecRepo.Update(ctx, exec)
}

// callTargetService 通过 HTTP 调用业务方的任务处理接口
// target 格式: 完整的 HTTP URL, 如 "http://order-rpc:8080/api/cron/check-timeout"
func (e *Executor) callTargetService(ctx context.Context, job model.TaskJob, executionID int64) error {
	if job.TargetType == 1 { // HTTP
		body := map[string]interface{}{
			"executionId": executionID,
			"taskName":    job.Name,
			"payload":     job.RequestBody,
		}
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal request body: %w", err)
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, job.Target, bytes.NewReader(bodyBytes))
		if err != nil {
			return fmt.Errorf("create http request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := e.httpClient.Do(req)
		if err != nil {
			return fmt.Errorf("http call %s: %w", job.Target, err)
		}
		defer resp.Body.Close()

		if resp.StatusCode >= 400 {
			return fmt.Errorf("http call %s returned status: %d", job.Target, resp.StatusCode)
		}

		return nil
	}

	return fmt.Errorf("unsupported target type: %d, only HTTP(1) supported", job.TargetType)
}

// acquireLock 尝试获取 etcd 分布式锁（任务级互斥，不带时间戳）
func (e *Executor) acquireLock(ctx context.Context, key string) (bool, error) {
	lease, err := e.etcdClient.Grant(ctx, int64(e.lockTTLSec))
	if err != nil {
		return false, fmt.Errorf("grant lease: %w", err)
	}

	txn := e.etcdClient.Txn(ctx).
		If(clientv3.Compare(clientv3.CreateRevision(key), "=", 0)).
		Then(clientv3.OpPut(key, getNodeID(), clientv3.WithLease(lease.ID))).
		Else()

	txnResp, err := txn.Commit()
	if err != nil {
		return false, fmt.Errorf("txn: %w", err)
	}

	return txnResp.Succeeded, nil
}

// releaseLock 释放锁
func (e *Executor) releaseLock(key string) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_, _ = e.etcdClient.Delete(ctx, key)
}

func getNodeID() string {
	return fmt.Sprintf("node-%d", time.Now().UnixNano())
}

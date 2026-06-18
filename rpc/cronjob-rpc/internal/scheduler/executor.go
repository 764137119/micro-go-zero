package scheduler

import (
	"context"
	"cronjob-rpc/internal/model"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Executor 任务执行引擎
// 负责实际执行任务，通过 etcd 分布式锁保证全局唯一执行
type Executor struct {
	etcdClient   *clientv3.Client
	taskJobRepo  *model.TaskJobRepo
	taskExecRepo *model.TaskExecutionRepo
	lockTTLSec   int
	httpClient   *http.Client
	grpcConnPool map[string]*grpc.ClientConn
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
		grpcConnPool: make(map[string]*grpc.ClientConn),
	}
}

// Execute 执行单个任务（由 cron 调度触发）
func (e *Executor) Execute(ctx context.Context, job model.TaskJob) {
	now := time.Now()

	// 1. 创建执行记录
	exec := &model.TaskExecution{
		TaskName:    job.Name,
		ScheduledAt: now,
		StartedAt:   now,
		Status:      "running",
		MaxRetries:  job.MaxRetries,
		ExecNode:    getHostname(),
	}

	if err := e.taskExecRepo.Create(ctx, exec); err != nil {
		logx.Errorf("Failed to create execution record for task %s: %v", job.Name, err)
		return
	}

	// 2. 尝试获取分布式锁
	lockKey := fmt.Sprintf("/cronjob/task/%s/exec/%d", job.Name, now.Unix())
	locked, err := e.acquireLock(ctx, lockKey)
	if err != nil || !locked {
		logx.Infof("Task %s execution skipped (lock not acquired): %v", job.Name, err)
		exec.Status = "failed"
		exec.Result = "lock not acquired"
		exec.FinishedAt = time.Now()
		_ = e.taskExecRepo.Update(ctx, exec)
		return
	}
	defer e.releaseLock(lockKey)

	// 3. 执行任务
	result, err := e.doExecute(ctx, job)
	exec.FinishedAt = time.Now()

	if err != nil {
		exec.Status = "failed"
		exec.Result = fmt.Sprintf("error: %v", err)
		logx.Errorf("Task %s execution failed: %v", job.Name, err)
	} else {
		exec.Status = "success"
		exec.Result = result
		logx.Infof("Task %s execution succeeded", job.Name)
	}

	// 4. 更新执行记录
	if err := e.taskExecRepo.Update(ctx, exec); err != nil {
		logx.Errorf("Failed to update execution record for task %s: %v", job.Name, err)
	}

	// 5. 如果失败且有重试策略，触发重试
	if err != nil && exec.RetryCount < exec.MaxRetries {
		e.scheduleRetry(ctx, exec, job)
	}
}

// ExecuteOnce 手动触发一次执行
func (e *Executor) ExecuteOnce(ctx context.Context, job model.TaskJob) (int64, error) {
	now := time.Now()

	exec := &model.TaskExecution{
		TaskName:    job.Name,
		ScheduledAt: now,
		StartedAt:   now,
		Status:      "running",
		MaxRetries:  job.MaxRetries,
		ExecNode:    getHostname(),
	}

	if err := e.taskExecRepo.Create(ctx, exec); err != nil {
		return 0, fmt.Errorf("create execution record: %w", err)
	}

	// 手动触发不通过分布式锁，直接执行
	result, err := e.doExecute(ctx, job)
	exec.FinishedAt = time.Now()

	if err != nil {
		exec.Status = "failed"
		exec.Result = fmt.Sprintf("error: %v", err)
	} else {
		exec.Status = "success"
		exec.Result = result
	}

	if err := e.taskExecRepo.Update(ctx, exec); err != nil {
		return exec.ID, fmt.Errorf("update execution record: %w", err)
	}

	return exec.ID, nil
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
	exec.Status = "retrying"
	exec.StartedAt = time.Now()
	_ = e.taskExecRepo.Update(ctx, exec)

	result, err := e.doExecute(ctx, *job)
	exec.FinishedAt = time.Now()

	if err != nil {
		exec.Status = "failed"
		exec.Result = fmt.Sprintf("retry error: %v", err)
	} else {
		exec.Status = "success"
		exec.Result = result
	}

	return e.taskExecRepo.Update(ctx, exec)
}

// scheduleRetry 安排重试
func (e *Executor) scheduleRetry(ctx context.Context, exec *model.TaskExecution, job model.TaskJob) {
	exec.RetryCount++
	exec.Status = "retrying"
	_ = e.taskExecRepo.Update(ctx, exec)

	// 等待重试间隔后重试
	time.Sleep(time.Duration(job.RetryInterval) * time.Second)

	// 重试执行
	result, err := e.doExecute(ctx, job)
	exec.FinishedAt = time.Now()

	if err != nil {
		exec.Status = "failed"
		exec.Result = fmt.Sprintf("retry(%d) error: %v", exec.RetryCount, err)
		logx.Errorf("Task %s retry(%d) failed: %v", job.Name, exec.RetryCount, err)

		// 如果还有重试次数，继续安排
		if exec.RetryCount < exec.MaxRetries {
			go e.scheduleRetry(ctx, exec, job)
			return
		}
	} else {
		exec.Status = "success"
		exec.Result = result
	}

	_ = e.taskExecRepo.Update(ctx, exec)
}

// doExecute 实际执行任务（根据目标类型调用不同方式）
func (e *Executor) doExecute(ctx context.Context, job model.TaskJob) (string, error) {
	switch job.TargetType {
	case 0: // gRPC
		return e.callGRPC(ctx, job)
	case 1: // HTTP
		return e.callHTTP(ctx, job)
	default:
		return "", fmt.Errorf("unsupported target type: %d", job.TargetType)
	}
}

// callGRPC 通过 gRPC 调用目标服务
func (e *Executor) callGRPC(ctx context.Context, job model.TaskJob) (string, error) {
	// 目标格式： "serviceAddress/methodName"
	parts := strings.SplitN(job.Target, "/", 2)
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid gRPC target format: %s (expected: address/method)", job.Target)
	}

	addr := parts[0]
	method := parts[1]

	conn, err := e.getGRPCConn(addr)
	if err != nil {
		return "", fmt.Errorf("connect to %s: %w", addr, err)
	}

	// 构造通用 gRPC 调用
	var resp interface{}
	err = conn.Invoke(ctx, method, json.RawMessage(job.RequestBody), &resp)
	if err != nil {
		return "", fmt.Errorf("gRPC call %s: %w", method, err)
	}

	respBytes, _ := json.Marshal(resp)
	return string(respBytes), nil
}

// callHTTP 通过 HTTP 调用目标服务
func (e *Executor) callHTTP(ctx context.Context, job model.TaskJob) (string, error) {
	req, err := http.NewRequestWithContext(ctx, "POST", job.Target, strings.NewReader(job.RequestBody))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := e.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("http call: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return string(body), fmt.Errorf("http status %d", resp.StatusCode)
	}

	return string(body), nil
}

// acquireLock 尝试获取 etcd 分布式锁
func (e *Executor) acquireLock(ctx context.Context, key string) (bool, error) {
	lease, err := e.etcdClient.Grant(ctx, int64(e.lockTTLSec))
	if err != nil {
		return false, fmt.Errorf("grant lease: %w", err)
	}

	txn := e.etcdClient.Txn(ctx).
		If(clientv3.Compare(clientv3.CreateRevision(key), "=", 0)).
		Then(clientv3.OpPut(key, "locked", clientv3.WithLease(lease.ID))).
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

// getGRPCConn 获取 gRPC 连接（带缓存）
func (e *Executor) getGRPCConn(addr string) (*grpc.ClientConn, error) {
	if conn, ok := e.grpcConnPool[addr]; ok {
		return conn, nil
	}

	conn, err := grpc.NewClient(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}

	e.grpcConnPool[addr] = conn
	return conn, nil
}

func getHostname() string {
	return fmt.Sprintf("node-%d", time.Now().UnixNano())
}

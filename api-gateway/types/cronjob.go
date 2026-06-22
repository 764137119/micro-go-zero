package types

import (
	cronjobpb "cronjob-rpc/cronjob"
)

// RegisterTaskReq 注册定时任务请求
type RegisterTaskReq struct {
	Name        string       `json:"name" binding:"required"`
	CronExpr    string       `json:"cron_expr" binding:"required"`
	TargetType  int32        `json:"target_type" binding:"required,min=0,max=1"`
	Target      string       `json:"target" binding:"required"`
	RequestBody string       `json:"request_body"`
	RetryPolicy *RetryPolicy `json:"retry_policy"`
	Description string       `json:"description"`
}

// RetryPolicy 重试策略
type RetryPolicy struct {
	MaxRetries  int32 `json:"max_retries" binding:"required,min=0"`
	IntervalSec int32 `json:"interval_sec" binding:"required,min=1"`
}

func (r *RegisterTaskReq) ToRPC() *cronjobpb.RegisterTaskReq {
	req := &cronjobpb.RegisterTaskReq{
		Name:        r.Name,
		CronExpr:    r.CronExpr,
		TargetType:  cronjobpb.TargetType(r.TargetType),
		Target:      r.Target,
		RequestBody: r.RequestBody,
		Description: r.Description,
	}
	if r.RetryPolicy != nil {
		req.RetryPolicy = &cronjobpb.RetryPolicy{
			MaxRetries:  r.RetryPolicy.MaxRetries,
			IntervalSec: r.RetryPolicy.IntervalSec,
		}
	}
	return req
}

// UnregisterTaskReq 注销定时任务请求
type UnregisterTaskReq struct {
	Name string `json:"name" binding:"required"`
}

func (r *UnregisterTaskReq) ToRPC() *cronjobpb.UnregisterTaskReq {
	return &cronjobpb.UnregisterTaskReq{
		Name: r.Name,
	}
}

// ListTasksReq 列举任务请求
type ListTasksReq struct {
	Page     int32 `json:"page" binding:"required,min=1"`
	PageSize int32 `json:"page_size" binding:"required,min=1,max=100"`
}

func (r *ListTasksReq) ToRPC() *cronjobpb.ListTasksReq {
	return &cronjobpb.ListTasksReq{
		Page:     r.Page,
		PageSize: r.PageSize,
	}
}

// SetTaskEnabledReq 启用/禁用任务请求
type SetTaskEnabledReq struct {
	TaskID  int64  `json:"task_id"`
	Name    string `json:"name"`
	Enabled bool   `json:"enabled" binding:"required"`
}

func (r *SetTaskEnabledReq) ToRPC() *cronjobpb.SetTaskEnabledReq {
	return &cronjobpb.SetTaskEnabledReq{
		TaskId:  r.TaskID,
		Name:    r.Name,
		Enabled: r.Enabled,
	}
}

// ListExecutionsReq 列举执行历史请求
type ListExecutionsReq struct {
	TaskName string `json:"task_name" binding:"required"`
	Status   int32  `json:"status"`
	Page     int32  `json:"page" binding:"required,min=1"`
	PageSize int32  `json:"page_size" binding:"required,min=1,max=100"`
}

func (r *ListExecutionsReq) ToRPC() *cronjobpb.ListExecutionsReq {
	return &cronjobpb.ListExecutionsReq{
		TaskName: r.TaskName,
		Status:   cronjobpb.ExecStatus(r.Status),
		Page:     r.Page,
		PageSize: r.PageSize,
	}
}

// TriggerOnceReq 手动触发任务请求
type TriggerOnceReq struct {
	TaskID   int64  `json:"task_id"`
	TaskName string `json:"task_name"`
}

func (r *TriggerOnceReq) ToRPC() *cronjobpb.TriggerOnceReq {
	return &cronjobpb.TriggerOnceReq{
		TaskId:   r.TaskID,
		TaskName: r.TaskName,
	}
}

// RetryTaskReq 重试任务请求
type RetryTaskReq struct {
	ExecutionID int64 `json:"execution_id" binding:"required,min=1"`
}

func (r *RetryTaskReq) ToRPC() *cronjobpb.RetryTaskReq {
	return &cronjobpb.RetryTaskReq{
		ExecutionId: r.ExecutionID,
	}
}

// GetTaskStatsReq 任务统计请求
type GetTaskStatsReq struct {
	TaskName string `json:"task_name" binding:"required"`
}

func (r *GetTaskStatsReq) ToRPC() *cronjobpb.TaskStatsReq {
	return &cronjobpb.TaskStatsReq{
		TaskName: r.TaskName,
	}
}

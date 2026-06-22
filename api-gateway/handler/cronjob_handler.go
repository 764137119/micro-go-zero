package handler

import (
	"context"

	"api-gateway/svc"
	"api-gateway/types"
	cronjobpb "cronjob-rpc/cronjob"

	"github.com/gin-gonic/gin"
)

// RegisterTask 注册定时任务
func RegisterTask(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return HandleJSON(
		func(ctx context.Context, req *types.RegisterTaskReq) (*cronjobpb.RegisterTaskResp, error) {
			return svcCtx.CronJobRpc.RegisterTask(ctx, req.ToRPC())
		},
	)
}

// UnregisterTask 注销定时任务
func UnregisterTask(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return HandleJSON(
		func(ctx context.Context, req *types.UnregisterTaskReq) (*cronjobpb.UnregisterTaskResp, error) {
			return svcCtx.CronJobRpc.UnregisterTask(ctx, req.ToRPC())
		},
	)
}

// ListTasks 列举任务
func ListTasks(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return HandleJSON(
		func(ctx context.Context, req *types.ListTasksReq) (*cronjobpb.ListTasksResp, error) {
			return svcCtx.CronJobRpc.ListTasks(ctx, req.ToRPC())
		},
	)
}

// SetTaskEnabled 启用/禁用任务
func SetTaskEnabled(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return HandleJSON(
		func(ctx context.Context, req *types.SetTaskEnabledReq) (*cronjobpb.SetTaskEnabledResp, error) {
			return svcCtx.CronJobRpc.SetTaskEnabled(ctx, req.ToRPC())
		},
	)
}

// ListExecutions 列举执行历史
func ListExecutions(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return HandleJSON(
		func(ctx context.Context, req *types.ListExecutionsReq) (*cronjobpb.ListExecutionsResp, error) {
			return svcCtx.CronJobRpc.ListExecutions(ctx, req.ToRPC())
		},
	)
}

// TriggerOnce 手动触发任务
func TriggerOnce(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return HandleJSON(
		func(ctx context.Context, req *types.TriggerOnceReq) (*cronjobpb.TriggerOnceResp, error) {
			return svcCtx.CronJobRpc.TriggerOnce(ctx, req.ToRPC())
		},
	)
}

// RetryTask 重试任务
func RetryTask(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return HandleJSON(
		func(ctx context.Context, req *types.RetryTaskReq) (*cronjobpb.RetryTaskResp, error) {
			return svcCtx.CronJobRpc.RetryTask(ctx, req.ToRPC())
		},
	)
}

// GetTaskStats 获取任务统计
func GetTaskStats(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return HandleJSON(
		func(ctx context.Context, req *types.GetTaskStatsReq) (*cronjobpb.TaskStatsResp, error) {
			return svcCtx.CronJobRpc.GetTaskStats(ctx, req.ToRPC())
		},
	)
}

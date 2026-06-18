package logic

import (
	"context"
	"cronjob-rpc/cronjob"
	"cronjob-rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetTaskStatsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetTaskStatsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTaskStatsLogic {
	return &GetTaskStatsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetTaskStatsLogic) GetTaskStats(in *cronjob.TaskStatsReq) (*cronjob.TaskStatsResp, error) {
	total, success, failed, running, lastExecutedAt, lastStatus, err := l.svcCtx.TaskExecRepo.GetTaskStats(l.ctx, in.TaskName)
	if err != nil {
		logx.Errorf("Failed to get task stats for %s: %v", in.TaskName, err)
		return &cronjob.TaskStatsResp{}, nil
	}

	lastStatusEnum := cronjob.ExecStatus_PENDING
	switch lastStatus {
	case "running":
		lastStatusEnum = cronjob.ExecStatus_RUNNING
	case "success":
		lastStatusEnum = cronjob.ExecStatus_SUCCESS
	case "failed":
		lastStatusEnum = cronjob.ExecStatus_FAILED
	case "retrying":
		lastStatusEnum = cronjob.ExecStatus_RETRYING
	}

	lastExecAt := int64(0)
	if lastExecutedAt != nil {
		lastExecAt = lastExecutedAt.Unix()
	}

	return &cronjob.TaskStatsResp{
		TaskName:        in.TaskName,
		TotalExecutions: total,
		SuccessCount:    success,
		FailedCount:     failed,
		RunningCount:    running,
		LastExecutedAt:  lastExecAt,
		LastStatus:      lastStatusEnum,
	}, nil
}

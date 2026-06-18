package logic

import (
	"context"
	"cronjob-rpc/cronjob"
	"cronjob-rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type TriggerOnceLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewTriggerOnceLogic(ctx context.Context, svcCtx *svc.ServiceContext) *TriggerOnceLogic {
	return &TriggerOnceLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *TriggerOnceLogic) TriggerOnce(in *cronjob.TriggerOnceReq) (*cronjob.TriggerOnceResp, error) {
	// 查找任务
	job, err := l.svcCtx.TaskJobRepo.FindByName(l.ctx, in.TaskName)
	if err != nil {
		logx.Errorf("Task not found: %s, %v", in.TaskName, err)
		return &cronjob.TriggerOnceResp{}, nil
	}

	// 通过执行器手动触发一次
	executionID, err := l.svcCtx.Executor.ExecuteOnce(l.ctx, *job)
	if err != nil {
		logx.Errorf("Failed to trigger task %s: %v", in.TaskName, err)
		return &cronjob.TriggerOnceResp{}, nil
	}

	logx.Infof("Task %s triggered manually, executionID: %d", in.TaskName, executionID)
	return &cronjob.TriggerOnceResp{ExecutionId: executionID}, nil
}

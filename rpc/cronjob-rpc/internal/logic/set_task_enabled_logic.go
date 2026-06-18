package logic

import (
	"context"
	"cronjob-rpc/cronjob"
	"cronjob-rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type SetTaskEnabledLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSetTaskEnabledLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SetTaskEnabledLogic {
	return &SetTaskEnabledLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *SetTaskEnabledLogic) SetTaskEnabled(in *cronjob.SetTaskEnabledReq) (*cronjob.SetTaskEnabledResp, error) {
	job, err := l.svcCtx.TaskJobRepo.FindByName(l.ctx, in.Name)
	if err != nil {
		logx.Errorf("Task not found: %s, %v", in.Name, err)
		return &cronjob.SetTaskEnabledResp{Ok: false}, nil
	}

	job.Enabled = in.Enabled
	if err := l.svcCtx.TaskJobRepo.Update(l.ctx, job); err != nil {
		logx.Errorf("Failed to update task %s: %v", in.Name, err)
		return &cronjob.SetTaskEnabledResp{Ok: false}, nil
	}

	logx.Infof("Task %s enabled=%v", in.Name, in.Enabled)
	return &cronjob.SetTaskEnabledResp{Ok: true}, nil
}

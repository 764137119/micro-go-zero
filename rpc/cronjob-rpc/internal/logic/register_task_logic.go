package logic

import (
	"context"
	"cronjob-rpc/cronjob"
	"cronjob-rpc/internal/model"
	"cronjob-rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type RegisterTaskLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRegisterTaskLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegisterTaskLogic {
	return &RegisterTaskLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *RegisterTaskLogic) RegisterTask(in *cronjob.RegisterTaskReq) (*cronjob.RegisterTaskResp, error) {
	job := &model.TaskJob{
		Name:        in.Name,
		CronExpr:    in.CronExpr,
		TargetType:  int32(in.TargetType),
		Target:      in.Target,
		RequestBody: in.RequestBody,
		Description: in.Description,
		Enabled:     true,
	}

	if in.RetryPolicy != nil {
		job.MaxRetries = in.RetryPolicy.MaxRetries
		job.RetryInterval = in.RetryPolicy.IntervalSec
	} else {
		job.MaxRetries = 3
		job.RetryInterval = 30
	}

	if err := l.svcCtx.TaskJobRepo.Create(l.ctx, job); err != nil {
		logx.Errorf("Failed to register task %s: %v", in.Name, err)
		return &cronjob.RegisterTaskResp{Ok: false}, nil
	}

	logx.Infof("Task registered: %s, cron: %s", in.Name, in.CronExpr)
	return &cronjob.RegisterTaskResp{Ok: true}, nil
}

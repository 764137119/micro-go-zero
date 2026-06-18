package logic

import (
	"context"
	"cronjob-rpc/cronjob"
	"cronjob-rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type RetryTaskLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRetryTaskLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RetryTaskLogic {
	return &RetryTaskLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *RetryTaskLogic) RetryTask(in *cronjob.RetryTaskReq) (*cronjob.RetryTaskResp, error) {
	if err := l.svcCtx.Executor.Retry(l.ctx, in.ExecutionId); err != nil {
		logx.Errorf("Failed to retry execution %d: %v", in.ExecutionId, err)
		return &cronjob.RetryTaskResp{Ok: false}, nil
	}

	logx.Infof("Retry triggered for execution: %d", in.ExecutionId)
	return &cronjob.RetryTaskResp{Ok: true}, nil
}

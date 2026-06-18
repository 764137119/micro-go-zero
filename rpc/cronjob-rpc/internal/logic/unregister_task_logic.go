package logic

import (
	"context"

	"cronjob-rpc/cronjob"
	"cronjob-rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type UnregisterTaskLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUnregisterTaskLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UnregisterTaskLogic {
	return &UnregisterTaskLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UnregisterTaskLogic) UnregisterTask(in *cronjob.UnregisterTaskReq) (*cronjob.UnregisterTaskResp, error) {
	// todo: add your logic here and delete this line

	return &cronjob.UnregisterTaskResp{}, nil
}

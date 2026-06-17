package logic

import (
	"context"

	"order-rpc/internal/svc"
	"order-rpc/order"

	"github.com/zeromicro/go-zero/core/logx"
)

type CancelOrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCancelOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CancelOrderLogic {
	return &CancelOrderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 取消订单
func (l *CancelOrderLogic) CancelOrder(in *order.CancelOrderReq) (*order.CancelOrderRsp, error) {
	// todo: add your logic here and delete this line

	return &order.CancelOrderRsp{}, nil
}

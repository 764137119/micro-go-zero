package logic

import (
	"context"

	"order-rpc/internal/svc"
	"order-rpc/order"

	"github.com/zeromicro/go-zero/core/logx"
)

type OrderStateCheckLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewOrderStateCheckLogic(ctx context.Context, svcCtx *svc.ServiceContext) *OrderStateCheckLogic {
	return &OrderStateCheckLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 订单状态检测
func (l *OrderStateCheckLogic) OrderStateCheck(in *order.OrderStateCheckReq) (*order.OrderStateCheckResp, error) {
	// todo: add your logic here and delete this line

	return &order.OrderStateCheckResp{}, nil
}

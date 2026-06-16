package logic

import (
	"context"

	"order-rpc/internal/svc"
	"order-rpc/order"

	"github.com/zeromicro/go-zero/core/logx"
)

type OrderCommitPayLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewOrderCommitPayLogic(ctx context.Context, svcCtx *svc.ServiceContext) *OrderCommitPayLogic {
	return &OrderCommitPayLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 订单支付提交
func (l *OrderCommitPayLogic) OrderCommitPay(in *order.OrderCommitPayReq) (*order.OrderCommitPayResp, error) {
	// todo: add your logic here and delete this line

	return &order.OrderCommitPayResp{}, nil
}

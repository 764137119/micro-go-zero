package logic

import (
	"context"
	"errors"

	"order-rpc/internal/svc"
	"order-rpc/order"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
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

// OrderStateCheck 订单状态检测
// 查询订单当前状态，与传入的期望状态比对。如果匹配则返回订单信息，否则返回错误。
func (l *OrderStateCheckLogic) OrderStateCheck(in *order.OrderStateCheckReq) (*order.OrderStateCheckResp, error) {
	orderInfo, err := l.svcCtx.OrderRepo.FindByOrderId(l.ctx, in.OrderId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("order not found")
		}
		return nil, err
	}

	// 如果传入了期望状态，校验是否匹配
	if in.OrderState != 0 && orderInfo.OrderState != in.OrderState {
		return nil, errors.New("order state mismatch")
	}

	return &order.OrderStateCheckResp{
		OrderCommon: &order.OrderCommon{
			OrderId:    orderInfo.OrderId,
			OrderNo:    orderInfo.OrderNo,
			OrderState: orderInfo.OrderState,
		},
	}, nil
}

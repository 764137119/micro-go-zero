package logic

import (
	"context"
	"errors"

	"order-rpc/internal/model"
	"order-rpc/internal/svc"
	"order-rpc/order"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
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

// OrderCommitPay 订单支付提交（用户支付成功后回调）
// 将订单从"待支付(0)"推进到"已支付(1)"
// 要求订单必须处于待支付状态，否则返回错误
func (l *OrderCommitPayLogic) OrderCommitPay(in *order.OrderCommitPayReq) (*order.OrderCommitPayResp, error) {
	err := l.svcCtx.OrderRepo.Transaction(func(tx *gorm.DB) error {
		// 1. 查询订单
		orderInfo, err := l.svcCtx.OrderRepo.FindByOrderIdTx(tx, in.OrderId)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("order not found")
			}
			return err
		}

		// 2. 状态机校验：只有"待支付(0)"状态的订单可以支付
		//    TCC 创建的订单状态为待支付，Confirm 阶段已推送到已支付
		//    此处做幂等：如果已经是已支付状态，直接返回成功
		if orderInfo.OrderState == model.OrderStatePaid {
			return nil // 幂等：已支付，直接成功
		}
		if orderInfo.OrderState != model.OrderStatePending {
			return errors.New("order state not pending, cannot commit pay")
		}

		// 3. 更新订单状态为已支付
		if err := tx.Model(&model.Order{}).
			Where("order_id = ? AND order_state = ?", in.OrderId, model.OrderStatePending).
			Update("order_state", model.OrderStatePaid).Error; err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &order.OrderCommitPayResp{Ok: true}, nil
}

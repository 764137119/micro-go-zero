package logic

import (
	"context"
	"errors"
	"strings"
	"time"

	"order-rpc/internal/model"
	"order-rpc/internal/svc"
	"order-rpc/order"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type TccConfirmOrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewTccConfirmOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *TccConfirmOrderLogic {
	return &TccConfirmOrderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// TCC Confirm：确认订单并将订单状态从"待支付"推进到"已支付"
// 由 DTM 在 Confirm 阶段回调，DTM 自带重试机制，业务层保证幂等
func (l *TccConfirmOrderLogic) TccConfirmOrder(in *order.TccConfirmOrderReq) (*order.TccConfirmOrderResp, error) {
	err := l.svcCtx.OrderRepo.Transaction(func(tx *gorm.DB) error {
		// 1. 【原子占坑】尝试插入控制表，状态为 CONFIRMED
		control := &model.OrderTccControl{
			Xid:       in.Xid,
			Status:    model.TCCCONFIRMED,
			CreatedAt: time.Now().Unix(),
		}
		err := tx.Create(control).Error
		if err != nil && !strings.Contains(err.Error(), "Duplicate") {
			return err
		}

		// 2. 插入成功（我是第一个到的），Confirm 正常提交
		if err == nil {
			// 将订单状态从"待支付(0)"推进到"已支付(1)"
			if err := tx.Model(&model.Order{}).
				Where("xid = ?", in.Xid).
				Update("order_state", model.OrderStatePaid).Error; err != nil {
				return err
			}
			return nil
		}

		// 3. 主键冲突，说明控制表已存在，查询真实状态
		var existing model.OrderTccControl
		if err := tx.Where("xid = ?", in.Xid).First(&existing).Error; err != nil {
			return err
		}

		if existing.Status == model.TCCCONFIRMED {
			// 【幂等】已经 Confirm 过了，直接返回成功
			return nil
		}

		if existing.Status == model.TCCCANCELLED {
			// 【严重异常】Cancel 比 Confirm 先到了，业务上不应该发生
			// DTM 保证同一分支只会有一个终态回调（Confirm 或 Cancel）
			return errors.New("confirm conflict: branch already cancelled")
		}

		if existing.Status == model.TCCTRYING {
			// 【正常推进】Try 已完成，推进订单状态并更新控制表
			if err := tx.Model(&model.Order{}).
				Where("xid = ?", in.Xid).
				Update("order_state", model.OrderStatePaid).Error; err != nil {
				return err
			}

			if err := tx.Model(&model.OrderTccControl{}).Where("xid = ?", in.Xid).
				Update("status", model.TCCCONFIRMED).Error; err != nil {
				return err
			}
			return nil
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &order.TccConfirmOrderResp{}, nil
}

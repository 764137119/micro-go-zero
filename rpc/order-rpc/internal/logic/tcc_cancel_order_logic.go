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

type TccCancelOrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewTccCancelOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *TccCancelOrderLogic {
	return &TccCancelOrderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// TCC Cancel：回滚订单（由 DTM 在 Cancel 阶段回调）
func (l *TccCancelOrderLogic) TccCancelOrder(in *order.TccCancelOrderReq) (*order.TccCancelOrderResp, error) {
	// todo: add your logic here and delete this line

	err := l.svcCtx.OrderRepo.Transaction(func(tx *gorm.DB) error {
		// 1. 【原子占坑】尝试插入控制表，状态为 CANCELLED
		control := &model.OrderTccControl{
			Xid:       in.Xid,
			OrderNo:   in.OrderNo,
			Status:    model.TCCCANCELLED,
			CreatedAt: time.Now().Unix(),
		}
		err := tx.Create(control).Error
		if err != nil && !strings.Contains(err.Error(), "Duplicate") {
			return err
		}

		// 2. 如果插入成功（affected=1），说明我是第一个到的（空回滚场景）
		if err == nil {
			// 【空回滚】Try还没执行，直接返回成功，控制表记录下 CANCELLED，拦截后续Try
			return nil
		}

		// 3. 主键冲突，说明控制表已存在，查询真实状态
		var existing model.OrderTccControl
		tx.Where("xid = ?", in.Xid).First(&existing)

		if existing.Status == model.TCCCANCELLED {
			// 幂等：已经回滚过了，直接返回成功
			return nil
		}

		if existing.Status == model.TCCCONFIRMED {
			// 严重错误：业务上 Confirm 比 Cancel 先到且成功了，Cancel 绝对不能执行！
			// 抛出异常或人工介入，绝不能回滚已确认的业务
			return errors.New("cannot cancel confirmed transaction")
		}

		if existing.Status == model.TCCTRYING {
			// 【正常回滚】Try已执行，现在执行业务补偿
			// 3.1 更新订单状态：待支付(0) -> 已取消(-1)
			if err := tx.Model(&model.Order{}).Where("order_no = ? AND order_state = 0", in.OrderNo).
				Update("order_state", -1).Error; err != nil {
				return err
			}

			// 3.2 释放库存冻结（可用+，锁定-）
			// ... 你的库存回滚逻辑（恢复可用库存）

			// 3.3 将控制表状态更新为 CANCELLED（标记终态）
			if err := tx.Model(&model.OrderTccControl{}).Where("xid = ?", in.Xid).
				Update("status", model.TCCCANCELLED).Error; err != nil {
				return err
			}
			return nil
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	var resp = order.TccCancelOrderResp{
		OrderId: 0,
	}
	return &resp, nil
}

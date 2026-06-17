package logic

import (
	"context"
	"strings"
	"time"

	"order-rpc/internal/model"
	"order-rpc/internal/svc"
	"order-rpc/order"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type TccTryOrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewTccTryOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *TccTryOrderLogic {
	return &TccTryOrderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// TCC Try：预留订单资源（由业务方/编排层在 Try 阶段调用）
func (l *TccTryOrderLogic) TccTryOrder(in *order.TccTryOrderReq) (*order.TccTryOrderResp, error) {
	// todo: add your logic here and delete this line

	// 开启数据库事务
	err := l.svcCtx.OrderRepo.Transaction(func(tx *gorm.DB) error {
		// 1. 【原子占坑】尝试插入控制表，状态为 TRYING
		control := &model.OrderTccControl{
			Xid:       in.Xid,
			Status:    "TRYING",
			OrderNo:   in.OrderNo,
			CreatedAt: time.Now().Unix(),
		}
		err := tx.Create(control).Error
		if err != nil && !strings.Contains(err.Error(), "Duplicate") {
			return err // 其他DB错误，回滚
		}

		// 2. 如果插入失败（主键冲突），说明记录已存在，查询当前状态
		if err != nil { // Duplicate key
			var existing model.OrderTccControl
			tx.Where("xid = ?", in.Xid).First(&existing)

			if existing.Status == "CANCELLED" {
				// 【防悬挂】Cancel 比 Try 先到了，必须放弃本次Try，返回成功给协调者
				// 注意：绝对不能创建订单和冻结库存，直接返回nil让事务提交（空事务）
				return nil
			}
			if existing.Status == "CONFIRMED" || existing.Status == "TRYING" {
				// 幂等处理：已处理过，直接返回成功
				return nil
			}
		}

		// 3. 插入成功（我是第一个到的），执行正常业务逻辑
		// 3.1 创建订单（状态=待支付）
		order := &model.Order{
			OrderNo:        in.OrderNo,
			OrderState:     model.OrderStatePending, // 待支付
			SkuId:          in.SkuId,
			Quantity:       in.Quantity,
			OrderBeginTime: in.OrderBeginTime,
			OrderEndTime:   in.OrderEndTime,
			UserId:         in.UserId,
			OrderPrice:     in.OrderPrice,
			OrderDes:       in.OrderDes,
			// ... 其他字段
			Xid: in.Xid, // 注意：你的Order表有Xid字段，可以用来追溯，但不做控制用
		}
		if err := tx.Create(order).Error; err != nil {
			return err // 失败则回滚，控制表记录也会回滚（因为在一个事务里）
		}
		return nil // 提交事务，控制表状态为 TRYING 持久化
	})
	if err != nil {
		return nil, err
	}

	existin, err := l.svcCtx.OrderRepo.FindByXid(l.ctx, in.Xid)
	if err != nil {
		return nil, err
	}
	var resp = &order.TccTryOrderResp{
		OrderId: existin.OrderId,
	}
	return resp, nil
}

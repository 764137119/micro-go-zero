package logic

import (
	"context"
	"errors"
	"strings"
	"time"

	"stock-rpc/internal/model"
	"stock-rpc/internal/svc"
	"stock-rpc/stock"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/protobuf/types/known/emptypb"
	"gorm.io/gorm"
)

type TccCancelDeductStockLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewTccCancelDeductStockLogic(ctx context.Context, svcCtx *svc.ServiceContext) *TccCancelDeductStockLogic {
	return &TccCancelDeductStockLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// TCC Cancel：释放库存（dtm TCC 全局事务 Cancel 阶段回调）
// 使用控制表（stock_tcc_control）实现防悬挂、幂等、空回滚保护
// 参考 order-rpc TccCancelOrder 的实现风格
func (l *TccCancelDeductStockLogic) TccCancelDeductStock(in *stock.TccCancelDeductStockReq) (*emptypb.Empty, error) {
	err := l.svcCtx.DB.Transaction(func(tx *gorm.DB) error {
		// 1. 【原子占坑】尝试插入控制表，状态为 CANCELLED
		control := &model.StockTccControl{
			Xid:       in.Xid,
			Status:    model.TCCCANCELLED,
			SkuId:     in.SkuId,
			OrderNo:   in.OrderNo,
			Quantity:  in.Quantity,
			CreatedAt: time.Now().UnixMilli(),
		}
		err := tx.Create(control).Error
		if err != nil && !strings.Contains(err.Error(), "Duplicate") {
			return err
		}

		// 2. 如果插入成功（我是第一个到的），空回滚场景
		if err == nil {
			// 【空回滚】Try 还没执行，直接返回成功
			// 控制表已记录 CANCELLED 状态，后续 Try 到达时会因防悬挂检查而放弃
			return nil
		}

		// 3. 主键冲突，说明控制表已存在，查询真实状态
		var existing model.StockTccControl
		if err := tx.Where("xid = ?", in.Xid).First(&existing).Error; err != nil {
			return err
		}

		if existing.Status == model.TCCCANCELLED {
			// 【幂等】已经回滚过了，直接返回成功
			return nil
		}

		if existing.Status == model.TCCCONFIRMED {
			// 【严重错误】Confirm 比 Cancel 先到且成功了，Cancel 绝对不能执行！
			// 抛出异常或人工介入，绝不能回滚已确认的业务
			return errors.New("cannot cancel confirmed transaction")
		}

		if existing.Status == model.TCCTRYING {
			// 【正常回滚】Try 已执行，现在执行业务补偿
			// 3.1 释放冻结库存：available_stock += qty, locked_stock -= qty
			if err := l.svcCtx.StockRepo.CancelDeductStockTx(tx, in.SkuId, in.Quantity); err != nil {
				return err
			}

			// 3.2 查询释放前后的库存快照，用于流水记录
			stockInfo, err := l.svcCtx.StockRepo.FindBySkuIdTx(tx, in.SkuId)
			if err != nil {
				return err
			}

			// 3.3 写入库存流水
			beforeAvailable := stockInfo.AvailableStock - in.Quantity
			afterAvailable := stockInfo.AvailableStock
			beforeLocked := stockInfo.LockedStock + in.Quantity
			afterLocked := stockInfo.LockedStock

			flowLog := &model.StockFlowLog{
				SkuId:           in.SkuId,
				OrderNo:         in.OrderNo,
				ChangeType:      model.ChangeTypeTccCancel,
				Quantity:        in.Quantity,
				BeforeAvailable: beforeAvailable,
				AfterAvailable:  afterAvailable,
				BeforeLocked:    beforeLocked,
				AfterLocked:     afterLocked,
				Xid:             in.Xid,
				Gid:             in.Xid,
				CreatedAt:       time.Now().UnixMilli(),
			}
			if err := tx.Create(flowLog).Error; err != nil {
				return err
			}

			// 3.4 将控制表状态更新为 CANCELLED（标记终态）
			if err := tx.Model(&model.StockTccControl{}).Where("xid = ?", in.Xid).
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

	return &emptypb.Empty{}, nil
}

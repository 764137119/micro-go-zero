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

type TccConfirmDeductStockLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewTccConfirmDeductStockLogic(ctx context.Context, svcCtx *svc.ServiceContext) *TccConfirmDeductStockLogic {
	return &TccConfirmDeductStockLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// TCC Confirm：确认扣减库存（dtm TCC 全局事务 Confirm 阶段回调）
// DTM 协调器自带重试机制，业务层只需保证幂等即可
// 使用控制表（stock_tcc_control）实现幂等保护
// 参考 order-rpc TccConfirmOrder 的实现风格
func (l *TccConfirmDeductStockLogic) TccConfirmDeductStock(in *stock.TccConfirmDeductStockReq) (*emptypb.Empty, error) {
	err := l.svcCtx.DB.Transaction(func(tx *gorm.DB) error {
		// 1. 【原子占坑】尝试插入控制表，状态为 CONFIRMED
		control := &model.StockTccControl{
			Xid:      in.Xid,
			Status:   model.TCCCONFIRMED,
			SkuId:    in.SkuId,
			OrderNo:  in.OrderNo,
			Quantity: in.Quantity,
		}
		err := tx.Create(control).Error
		if err != nil && !strings.Contains(err.Error(), "Duplicate") {
			return err
		}

		// 2. 插入成功（我是第一个到的），Confirm 正常提交
		if err == nil {
			// 执行确认扣减：locked_stock -= qty
			if err := l.svcCtx.StockRepo.ConfirmDeductStockTx(tx, in.SkuId, in.Quantity); err != nil {
				return err
			}

			// 查询确认扣减前后的库存快照，用于流水记录
			stockInfo, err := l.svcCtx.StockRepo.FindBySkuIdTx(tx, in.SkuId)
			if err != nil {
				return err
			}

			// 写入库存流水
			beforeLocked := stockInfo.LockedStock + in.Quantity
			flowLog := &model.StockFlowLog{
				SkuId:           in.SkuId,
				OrderNo:         in.OrderNo,
				ChangeType:      model.ChangeTypeConfirm,
				Quantity:        in.Quantity,
				BeforeAvailable: stockInfo.AvailableStock,
				AfterAvailable:  stockInfo.AvailableStock,
				BeforeLocked:    beforeLocked,
				AfterLocked:     stockInfo.LockedStock,
				Xid:             in.Xid,
				Gid:             in.Xid,
				CreatedAt:       time.Now().UnixMilli(),
			}
			if err := tx.Create(flowLog).Error; err != nil {
				return err
			}

			return nil
		}

		// 3. 主键冲突，说明控制表已存在，查询真实状态
		var existing model.StockTccControl
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
			// 【正常推进】Try 已完成，推进状态到 CONFIRMED
			if err := tx.Model(&model.StockTccControl{}).Where("xid = ?", in.Xid).
				Update("status", model.TCCCONFIRMED).Error; err != nil {
				return err
			}

			// 执行确认扣减：locked_stock -= qty
			if err := l.svcCtx.StockRepo.ConfirmDeductStockTx(tx, in.SkuId, in.Quantity); err != nil {
				return err
			}

			// 查询确认扣减前后的库存快照，用于流水记录
			stockInfo, err := l.svcCtx.StockRepo.FindBySkuIdTx(tx, in.SkuId)
			if err != nil {
				return err
			}

			// 写入库存流水
			beforeLocked := stockInfo.LockedStock + in.Quantity
			flowLog := &model.StockFlowLog{
				SkuId:           in.SkuId,
				OrderNo:         in.OrderNo,
				ChangeType:      model.ChangeTypeConfirm,
				Quantity:        in.Quantity,
				BeforeAvailable: stockInfo.AvailableStock,
				AfterAvailable:  stockInfo.AvailableStock,
				BeforeLocked:    beforeLocked,
				AfterLocked:     stockInfo.LockedStock,
				Xid:             in.Xid,
				Gid:             in.Xid,
				CreatedAt:       time.Now().UnixMilli(),
			}
			if err := tx.Create(flowLog).Error; err != nil {
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

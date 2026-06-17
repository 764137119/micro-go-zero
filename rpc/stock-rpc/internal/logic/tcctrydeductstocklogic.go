package logic

import (
	"context"
	"strings"
	"time"

	"stock-rpc/internal/model"
	"stock-rpc/internal/svc"
	"stock-rpc/stock"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/protobuf/types/known/emptypb"
	"gorm.io/gorm"
)

type TccTryDeductStockLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewTccTryDeductStockLogic(ctx context.Context, svcCtx *svc.ServiceContext) *TccTryDeductStockLogic {
	return &TccTryDeductStockLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// TCC Try：冻结库存（dtm TCC 全局事务 Try 阶段调用，预留资源）
// 使用控制表（stock_tcc_control）实现防悬挂、幂等、空回滚保护
// 参考 order-rpc TccTryOrder 的实现风格
func (l *TccTryDeductStockLogic) TccTryDeductStock(in *stock.TccTryDeductStockReq) (*emptypb.Empty, error) {
	err := l.svcCtx.DB.Transaction(func(tx *gorm.DB) error {
		// 1. 【原子占坑】尝试插入控制表，状态为 TRYING
		control := &model.StockTccControl{
			Xid:       in.Xid,
			Status:    model.TCCTRYING,
			SkuId:     in.SkuId,
			OrderNo:   in.OrderNo,
			Quantity:  in.Quantity,
			CreatedAt: time.Now().UnixMilli(),
		}
		err := tx.Create(control).Error
		if err != nil && !strings.Contains(err.Error(), "Duplicate") {
			return err // 其他DB错误，回滚
		}

		// 2. 如果插入失败（主键冲突），说明记录已存在，查询当前状态
		if err != nil { // Duplicate key
			var existing model.StockTccControl
			if err := tx.Where("xid = ?", in.Xid).First(&existing).Error; err != nil {
				return err
			}

			if existing.Status == model.TCCCANCELLED {
				// 【防悬挂】Cancel 比 Try 先到了，必须放弃本次Try，返回成功给协调者
				// 注意：绝对不能冻结库存，直接返回nil让事务提交（空事务）
				return nil
			}
			if existing.Status == model.TCCCONFIRMED || existing.Status == model.TCCTRYING {
				// 【幂等】已处理过，直接返回成功
				return nil
			}
			return nil
		}

		// 3. 插入成功（我是第一个到的），执行正常业务逻辑
		// 3.1 冻结库存：available_stock -= qty, locked_stock += qty
		if err := l.svcCtx.StockRepo.TryDeductStockTx(tx, in.SkuId, in.Quantity); err != nil {
			return err
		}

		// 3.2 查询冻结前后的库存快照，用于流水记录
		stockInfo, err := l.svcCtx.StockRepo.FindBySkuIdTx(tx, in.SkuId)
		if err != nil {
			return err
		}

		// 3.3 写入库存流水
		beforeAvailable := stockInfo.AvailableStock + in.Quantity
		afterAvailable := stockInfo.AvailableStock
		beforeLocked := stockInfo.LockedStock - in.Quantity
		afterLocked := stockInfo.LockedStock

		flowLog := &model.StockFlowLog{
			SkuId:           in.SkuId,
			OrderNo:         in.OrderNo,
			ChangeType:      model.ChangeTypeTry,
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

		return nil // 提交事务，控制表状态为 TRYING 持久化
	})
	if err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

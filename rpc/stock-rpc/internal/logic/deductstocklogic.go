package logic

import (
	"context"

	"stock-rpc/internal/svc"
	"stock-rpc/stock"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/protobuf/types/known/emptypb"
)

type DeductStockLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeductStockLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeductStockLogic {
	return &DeductStockLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// Saga 正向：扣减库存（dtm 在全局事务中调用）
// 使用 dtm Barrier 机制保证幂等和子事务屏障
func (l *DeductStockLogic) DeductStock(in *stock.DeductStockReq) (*emptypb.Empty, error) {
	// TODO: 使用 dtmcli barrier 实现幂等扣减库存
	// 1. 开启数据库事务
	// 2. Barrier.CallWithDB() 保证子事务屏障
	// 3. 扣减 available_stock，增加 locked_stock
	// 4. 写入库存流水表 stock_flow_log

	return &emptypb.Empty{}, nil
}

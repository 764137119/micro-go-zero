package logic

import (
	"context"

	"stock-rpc/internal/svc"
	"stock-rpc/stock"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/protobuf/types/known/emptypb"
)

type RollbackStockLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRollbackStockLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RollbackStockLogic {
	return &RollbackStockLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// Saga 补偿：回滚库存（dtm 在全局事务失败时调用）
// 使用 dtm Barrier 机制保证幂等和子事务屏障
func (l *RollbackStockLogic) RollbackStock(in *stock.RollbackStockReq) (*emptypb.Empty, error) {
	// TODO: 使用 dtmcli barrier 实现幂等回滚库存
	// 1. 开启数据库事务
	// 2. Barrier.CallWithDB() 保证子事务屏障
	// 3. 增加 available_stock，减少 locked_stock
	// 4. 写入库存流水表 stock_flow_log

	return &emptypb.Empty{}, nil
}

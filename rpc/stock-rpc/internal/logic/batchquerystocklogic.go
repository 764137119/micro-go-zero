package logic

import (
	"context"

	"stock-rpc/internal/svc"
	"stock-rpc/stock"

	"github.com/zeromicro/go-zero/core/logx"
)

type BatchQueryStockLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewBatchQueryStockLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BatchQueryStockLogic {
	return &BatchQueryStockLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 批量查询商品库存
func (l *BatchQueryStockLogic) BatchQueryStock(in *stock.BatchQueryStockReq) (*stock.BatchQueryStockResp, error) {
	// TODO: 从数据库批量查询库存信息

	return &stock.BatchQueryStockResp{}, nil
}

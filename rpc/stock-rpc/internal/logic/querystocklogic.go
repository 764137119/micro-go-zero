package logic

import (
	"context"

	"stock-rpc/internal/svc"
	"stock-rpc/stock"

	"github.com/zeromicro/go-zero/core/logx"
)

type QueryStockLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewQueryStockLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryStockLogic {
	return &QueryStockLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 查询单个商品库存
func (l *QueryStockLogic) QueryStock(in *stock.QueryStockReq) (*stock.QueryStockResp, error) {
	// TODO: 从数据库查询库存信息

	return &stock.QueryStockResp{}, nil
}

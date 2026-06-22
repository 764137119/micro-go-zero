package handler

import (
	"context"

	"api-gateway/svc"
	"api-gateway/types"
	stockpb "stock-rpc/stock"

	"github.com/gin-gonic/gin"
)

// BatchQueryStock 批量查询库存
func BatchQueryStock(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return HandleJSON(
		func(ctx context.Context, req *types.BatchQueryStockReq) (*stockpb.BatchQueryStockResp, error) {
			return svcCtx.StockRpc.BatchQueryStock(ctx, req.ToRPC())
		},
	)
}

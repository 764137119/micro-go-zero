package types

import (
	stockpb "stock-rpc/stock"
)

// BatchQueryStockReq 批量查询库存请求
type BatchQueryStockReq struct {
	SkuIDs []int64 `json:"sku_ids" binding:"required,min=1"`
}

func (r *BatchQueryStockReq) ToRPC() *stockpb.BatchQueryStockReq {
	return &stockpb.BatchQueryStockReq{
		SkuIds: r.SkuIDs,
	}
}

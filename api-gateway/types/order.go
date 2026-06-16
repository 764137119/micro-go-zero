package types

import (
	orderpb "order-rpc/order"
)

// OrderCommitPayReq 订单支付提交请求
type OrderCommitPayReq struct {
	OrderID int64 `json:"order_id" binding:"required,min=1"`
	PayID   int64 `json:"pay_id" binding:"required,min=1"`
}

func (r *OrderCommitPayReq) ToRPC() *orderpb.OrderCommitPayReq {
	return &orderpb.OrderCommitPayReq{
		OrderId: r.OrderID,
		PayId:   r.PayID,
	}
}

// OrderStateCheckReq 订单状态检测请求
type OrderStateCheckReq struct {
	OrderID    int64 `form:"order_id" binding:"required,min=1"`
	OrderState int32 `form:"order_state"`
}

func (r *OrderStateCheckReq) ToRPC() *orderpb.OrderStateCheckReq {
	return &orderpb.OrderStateCheckReq{
		OrderId:    r.OrderID,
		OrderState: r.OrderState,
	}
}

// CreateOrderReq 创建订单请求
type CreateOrderReq struct {
	UserID         int64  `json:"user_id" binding:"required,min=1"`
	OrderNo        string `json:"order_no" binding:"required"`
	OrderPrice     int64  `json:"order_price" binding:"required"`
	OrderDes       string `json:"order_des"`
	OrderBeginTime int64  `json:"order_begin_time" binding:"required"`
	OrderEndTime   int64  `json:"order_end_time" binding:"required"`
	SkuId          int64  `json:"sku_id" binding:"required,min=1"`   // 商品 SKU ID（用于 Saga 扣减库存）
	Quantity       int64  `json:"quantity" binding:"required,min=1"` // 商品数量（用于 Saga 扣减库存）
}

func (r *CreateOrderReq) ToRPC() *orderpb.CreateOrderReq {
	return &orderpb.CreateOrderReq{
		UserId:         r.UserID,
		OrderNo:        r.OrderNo,
		OrderPrice:     r.OrderPrice,
		OrderDes:       r.OrderDes,
		OrderBeginTime: r.OrderBeginTime,
		OrderEndTime:   r.OrderEndTime,
		SkuId:          r.SkuId,
		Quantity:       r.Quantity,
	}
}

// CancelOrderReq 取消订单请求
type CancelOrderReq struct {
	OrderID int64 `json:"order_id" binding:"required,min=1"`
}

func (r *CancelOrderReq) ToRPC() *orderpb.CancelOrderReq {
	return &orderpb.CancelOrderReq{
		OrderId: r.OrderID,
	}
}

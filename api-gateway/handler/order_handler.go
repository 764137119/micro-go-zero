package handler

import (
	"context"
	"fmt"

	"api-gateway/svc"
	"api-gateway/types"
	orderpb "order-rpc/order"
	stockpb "stock-rpc/stock"

	"github.com/dtm-labs/dtmgrpc"
	"github.com/gin-gonic/gin"
)

// OrderCommitPay 订单支付提交
func OrderCommitPay(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return HandleJSON(
		func(ctx context.Context, req *types.OrderCommitPayReq) (*orderpb.OrderCommitPayResp, error) {
			return svcCtx.OrderRpc.OrderCommitPay(ctx, req.ToRPC())
		},
	)
}

// OrderStateCheck 订单状态检测
func OrderStateCheck(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return HandleQuery(
		func(ctx context.Context, req *types.OrderStateCheckReq) (*orderpb.OrderStateCheckResp, error) {
			return svcCtx.OrderRpc.OrderStateCheck(ctx, req.ToRPC())
		},
	)
}

// CreateOrder 创建订单（DTM Saga 分布式事务）
//
// 编排流程：
//
//  1. 扣减库存（stock-rpc DeductStock — 正向）
//  2. 创建订单（order-rpc CreateOrder — 正向）
//  3. 任意分支失败 → 自动执行对应补偿
//     - stock-rpc RollbackStock（恢复库存）
//     - order-rpc CancelOrder（取消订单）
func CreateOrder(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return HandleJSON(
		func(ctx context.Context, req *types.CreateOrderReq) (*orderpb.CreateOrderRsp, error) {
			// 1. 生成 DTM 全局事务 ID（通过 DTM gRPC 端口）
			gid := dtmgrpc.MustGenGid(svcCtx.Config.DTMEndpoint)

			// 2. 构建 Saga gRPC 事务
			//    Add 的 action/compensate 使用 grpc:// 协议的 URL
			//    格式: grpc://{host}:{port}/{package}.{service}/{method}
			saga := dtmgrpc.NewSagaGrpc(svcCtx.Config.DTMEndpoint, gid).
				// 分支1: 扣减库存（正向）+ 回滚库存（补偿）
				Add(
					fmt.Sprintf("grpc://%s/stock.Stock/DeductStock", svcCtx.Config.StockRpcTarget),
					fmt.Sprintf("grpc://%s/stock.Stock/RollbackStock", svcCtx.Config.StockRpcTarget),
					&stockpb.DeductStockReq{
						Gid:       gid,
						TransType: "saga",
						SkuId:     req.SkuId,
						Quantity:  req.Quantity,
						OrderNo:   req.OrderNo,
					},
				).
				// 分支2: 创建订单（正向）+ 取消订单（补偿）
				Add(
					fmt.Sprintf("grpc://%s/order.Order/CreateOrder", svcCtx.Config.OrderRpcTarget),
					fmt.Sprintf("grpc://%s/order.Order/CancelOrder", svcCtx.Config.OrderRpcTarget),
					&orderpb.CreateOrderReq{
						UserId:         req.UserID,
						OrderNo:        req.OrderNo,
						OrderPrice:     req.OrderPrice,
						OrderDes:       req.OrderDes,
						OrderBeginTime: req.OrderBeginTime,
						OrderEndTime:   req.OrderEndTime,
						SkuId:          req.SkuId,
						Quantity:       req.Quantity,
						Gid:            gid,
						TransType:      "saga",
					},
				)

			// 3. 提交 Saga 事务（同步等待全部完成或失败）
			if err := saga.Submit(); err != nil {
				return nil, fmt.Errorf("dtm saga submit failed: %w", err)
			}

			// Saga 执行成功，订单已由 order-rpc 创建
			return &orderpb.CreateOrderRsp{
				// orderId 由 order-rpc 本地生成并写入 DB，此处无法直接获取
				// 客户端可通过 orderNo 后续查询订单信息
				OrderId: 0,
			}, nil
		},
	)
}

// CancelOrder 取消订单
func CancelOrder(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return HandleJSON(
		func(ctx context.Context, req *types.CancelOrderReq) (*orderpb.CancelOrderRsp, error) {
			return svcCtx.OrderRpc.CancelOrder(ctx, req.ToRPC())
		},
	)
}

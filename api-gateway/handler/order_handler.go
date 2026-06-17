package handler

import (
	"context"
	"fmt"

	"api-gateway/svc"
	"api-gateway/types"
	common "common/tools"
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

// CreateOrder 创建订单（DTM TCC 分布式事务）
//
// 编排流程（TCC 三阶段）：
//
//	Phase 1 — Try（资源预留）：
//	  - stock-rpc TccTryDeductStock：冻结库存（available_stock -= qty, locked_stock += qty）
//	  - order-rpc TccTryOrder：创建订单（状态=待支付）
//	Phase 2 — Confirm（全部 Try 成功后由 DTM 自动调用）：
//	  - stock-rpc TccConfirmDeductStock：确认扣减（locked_stock -= qty）
//	  - order-rpc TccConfirmOrder：确认订单
//	Phase 2 — Cancel（任意 Try 失败后由 DTM 自动调用）：
//	  - stock-rpc TccCancelDeductStock：释放库存（available_stock += qty, locked_stock -= qty）
//	  - order-rpc TccCancelOrder：取消订单（状态→已取消）
func CreateOrder(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return HandleJSON(
		func(ctx context.Context, req *types.CreateOrderReq) (*orderpb.CreateOrderRsp, error) {
			// 1. 后端生成唯一订单号（雪花算法），如前端已传入则覆盖
			req.OrderNo = common.GenOrderNoStr()

			// 2. 生成 DTM 全局事务 ID
			gid := dtmgrpc.MustGenGid(svcCtx.Config.DTMEndpoint)

			// 3. 定义 TCC 分支的业务数据（用于 CallBranch 注册）
			//    注意：orderReply 定义在闭包外，闭包内赋值、闭包外读取
			stockReq := &stockpb.TccTryDeductStockReq{
				Xid:       gid,
				TransType: "tcc",
				SkuId:     req.SkuId,
				Quantity:  req.Quantity,
				OrderNo:   req.OrderNo,
			}
			orderReq := &orderpb.TccTryOrderReq{
				UserId:         req.UserID,
				OrderNo:        req.OrderNo,
				OrderPrice:     req.OrderPrice,
				OrderDes:       req.OrderDes,
				OrderBeginTime: req.OrderBeginTime,
				OrderEndTime:   req.OrderEndTime,
				SkuId:          req.SkuId,
				Quantity:       req.Quantity,
				Xid:            gid,
				TransType:      "tcc",
			}

			// 定义在闭包外，以便闭包内赋值后外部能读取
			var orderReply orderpb.TccTryOrderResp

			// 4. 执行 TCC 全局事务（Prepare + Try 分支 + Submit / Abort）
			//    - Prepare：向 DTM 注册全局事务
			//    - CallBranch：注册分支并立即执行 Try
			//    - 全部 Try 成功 → Submit（DTM 异步调用 Confirm）
			//    - 任一 Try 失败   → Abort（DTM 异步调用 Cancel）
			err := dtmgrpc.TccGlobalTransaction(svcCtx.Config.DTMEndpoint, gid,
				func(tcc *dtmgrpc.TccGrpc) error {
					// 分支1：冻结库存（Try）
					if err := tcc.CallBranch(
						stockReq,
						fmt.Sprintf("grpc://%s/stock.Stock/TccTryDeductStock", svcCtx.Config.StockRpcTarget),
						fmt.Sprintf("grpc://%s/stock.Stock/TccConfirmDeductStock", svcCtx.Config.StockRpcTarget),
						fmt.Sprintf("grpc://%s/stock.Stock/TccCancelDeductStock", svcCtx.Config.StockRpcTarget),
						nil, // 不需要 reply
					); err != nil {
						return err
					}

					// 分支2：创建订单（Try），捕获返回的 orderId
					if err := tcc.CallBranch(
						orderReq,
						fmt.Sprintf("grpc://%s/order.Order/TccTryOrder", svcCtx.Config.OrderRpcTarget),
						fmt.Sprintf("grpc://%s/order.Order/TccConfirmOrder", svcCtx.Config.OrderRpcTarget),
						fmt.Sprintf("grpc://%s/order.Order/TccCancelOrder", svcCtx.Config.OrderRpcTarget),
						&orderReply,
					); err != nil {
						return err
					}

					return nil // 全部 Try 成功，DTM 将 Submit 全局事务
				},
			)
			if err != nil {
				return nil, fmt.Errorf("dtm tcc transaction failed: %w", err)
			}

			// TCC Try 阶段成功，订单已由 order-rpc TccTryOrder 创建（状态=待支付）
			return &orderpb.CreateOrderRsp{
				OrderId: orderReply.OrderId, // 从 CallBranch reply 中获取 orderId
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

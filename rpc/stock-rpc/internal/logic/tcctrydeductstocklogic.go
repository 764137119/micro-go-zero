package logic

import (
	"context"

	"google.golang.org/protobuf/types/known/emptypb"
	"stock-rpc/internal/svc"
	"stock-rpc/stock"

	"github.com/zeromicro/go-zero/core/logx"
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
func (l *TccTryDeductStockLogic) TccTryDeductStock(in *stock.TccTryDeductStockReq) (*emptypb.Empty, error) {
	// todo: add your logic here and delete this line

	return &emptypb.Empty{}, nil
}

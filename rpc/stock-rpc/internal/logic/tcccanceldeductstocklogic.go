package logic

import (
	"context"

	"google.golang.org/protobuf/types/known/emptypb"
	"stock-rpc/internal/svc"
	"stock-rpc/stock"

	"github.com/zeromicro/go-zero/core/logx"
)

type TccCancelDeductStockLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewTccCancelDeductStockLogic(ctx context.Context, svcCtx *svc.ServiceContext) *TccCancelDeductStockLogic {
	return &TccCancelDeductStockLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// TCC Cancel：释放库存（dtm TCC 全局事务 Cancel 阶段回调）
func (l *TccCancelDeductStockLogic) TccCancelDeductStock(in *stock.TccCancelDeductStockReq) (*emptypb.Empty, error) {
	// todo: add your logic here and delete this line

	return &emptypb.Empty{}, nil
}

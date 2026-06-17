package logic

import (
	"context"

	"google.golang.org/protobuf/types/known/emptypb"
	"stock-rpc/internal/svc"
	"stock-rpc/stock"

	"github.com/zeromicro/go-zero/core/logx"
)

type TccConfirmDeductStockLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewTccConfirmDeductStockLogic(ctx context.Context, svcCtx *svc.ServiceContext) *TccConfirmDeductStockLogic {
	return &TccConfirmDeductStockLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// TCC Confirm：确认扣减库存（dtm TCC 全局事务 Confirm 阶段回调）
func (l *TccConfirmDeductStockLogic) TccConfirmDeductStock(in *stock.TccConfirmDeductStockReq) (*emptypb.Empty, error) {
	// todo: add your logic here and delete this line

	return &emptypb.Empty{}, nil
}

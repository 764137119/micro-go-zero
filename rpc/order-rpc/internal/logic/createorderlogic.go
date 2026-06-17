package logic

import (
	"context"
	"errors"
	"time"

	commodel "common/model"
	"order-rpc/internal/svc"
	"order-rpc/order"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type CreateOrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateOrderLogic {
	return &CreateOrderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 创建订单
func (l *CreateOrderLogic) CreateOrder(in *order.CreateOrderReq) (*order.CreateOrderRsp, error) {
	// todo: add your logic here and delete this line
	var orderRsp = &order.CreateOrderRsp{}
	//更具全局事务ID查询是否有重复
	sage, err := l.svcCtx.SagaGlobalTransactionRepo.FindByXid(l.ctx, in.Gid)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			l.Errorf("查询分布式事务失败", err)
			return nil, err
		}
		//写日志
		var newSage = commodel.SagaGlobalTransaction{}
		newSage.Xid = in.Gid
		newSage.Status = commodel.SagaGlobalTransactionStatusRunning
		newSage.GmtCreate = time.Now()
		newSage.TransactionName = in.TransType
		newSage.StartTime = newSage.GmtCreate
		newSage.EndTime = nil
		newSage.Timeout = int32(time.Second)
		newSage.Version = 1
		if err = l.svcCtx.SagaGlobalTransactionRepo.Create(l.ctx, &newSage); err != nil {
			l.Errorf("写入日志失败", err)
			return nil, err
		}
		sage = &newSage
	}
	if sage.Status == commodel.SagaGlobalTransactionStatusSucceed {
		//查询订单
		order, err := l.svcCtx.OrderRepo.FindByXid(l.ctx, sage.Xid)
		if err != nil {
			l.Errorf("根据事务ID查询订单失败", err)
			return nil, err
		}
		orderRsp.OrderId = order.OrderId
		return orderRsp, nil
	}
	if sage.Status == commodel.SagaGlobalTransactionStatusRunning {
		//查询订单并创建
		order, err := l.svcCtx.OrderRepo.FindByXid(l.ctx, sage.Xid)
		if err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				l.Errorf("根据事务ID查询订单失败", err)
				return nil, err
			}
			//创建订单
			if order.OrderId == 0 {
				order.CreatedAt = time.Now().Unix()
				order.OrderBeginTime = order.CreatedAt
				order.OrderEndTime = time.Now().AddDate(0, 0, 1).Unix()
				order.OrderNo = ""
			}
		}
	}

	return orderRsp, nil
}

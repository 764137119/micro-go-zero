package svc

import (
	"api-gateway/config"
	"order-rpc/orderclient"
	"user-rpc/userclient"

	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config      config.Config
	UserRpc     userclient.User
	OrderRpc    orderclient.Order
	DTMEndpoint string // dtm gRPC 地址
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config:      c,
		UserRpc:     userclient.NewUser(zrpc.MustNewClient(c.UserRpc)),
		OrderRpc:    orderclient.NewOrder(zrpc.MustNewClient(c.OrderRpc)),
		DTMEndpoint: c.DTMEndpoint,
	}
}

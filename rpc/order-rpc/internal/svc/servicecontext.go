package svc

import (
	"order-rpc/internal/config"
	"stock-rpc/stockclient"

	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config      config.Config
	StockRpc    stockclient.Stock
	DTMEndpoint string // dtm 协调器地址
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config:      c,
		StockRpc:    stockclient.NewStock(zrpc.MustNewClient(c.StockRpc)),
		DTMEndpoint: c.DTM,
	}
}

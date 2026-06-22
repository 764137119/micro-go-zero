package svc

import (
	"api-gateway/config"
	"cronjob-rpc/cron_job_client"
	"order-rpc/orderclient"
	"stock-rpc/stock_client"
	"user-rpc/userclient"

	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config      config.Config
	UserRpc     userclient.User
	OrderRpc    orderclient.Order
	StockRpc    stock_client.Stock
	CronJobRpc  cron_job_client.CronJob
	DTMEndpoint string // dtm gRPC 地址
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config:      c,
		UserRpc:     userclient.NewUser(zrpc.MustNewClient(c.UserRpc)),
		OrderRpc:    orderclient.NewOrder(zrpc.MustNewClient(c.OrderRpc)),
		StockRpc:    stock_client.NewStock(zrpc.MustNewClient(c.StockRpc)),
		CronJobRpc:  cron_job_client.NewCronJob(zrpc.MustNewClient(c.CronJobRpc)),
		DTMEndpoint: c.DTMEndpoint,
	}
}

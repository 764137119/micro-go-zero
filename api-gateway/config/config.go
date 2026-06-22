package config

import "github.com/zeromicro/go-zero/zrpc"

type Config struct {
	Name string `yaml:"Name"`
	Host string `yaml:"Host"`
	Port int    `yaml:"Port"`

	UserRpc    zrpc.RpcClientConf `yaml:"UserRpc"`
	OrderRpc   zrpc.RpcClientConf `yaml:"OrderRpc"`
	StockRpc   zrpc.RpcClientConf `yaml:"StockRpc"`
	CronJobRpc zrpc.RpcClientConf `yaml:"CronJobRpc"`

	// DTM HTTP 地址（用于生成 gid），如 "http://dtm:36789"
	DTM string `yaml:"DTM"`
	// DTM gRPC 地址（用于提交 Saga），如 "dtm:36790"
	DTMEndpoint string `yaml:"DTMEndpoint"`

	// Saga 分支的 gRPC 目标地址（供 dtm 回调调用）
	StockRpcTarget string `yaml:"StockRpcTarget"` // "stock-rpc:8082"
	OrderRpcTarget string `yaml:"OrderRpcTarget"` // "order-rpc:8081"
}

package config

import "github.com/zeromicro/go-zero/zrpc"

type Config struct {
	Name string `json:"Name" yaml:"Name"`
	Host string `json:"Host" yaml:"Host"`
	Port int    `json:"Port" yaml:"Port"`

	UserRpc    zrpc.RpcClientConf `json:"UserRpc" yaml:"UserRpc"`
	OrderRpc   zrpc.RpcClientConf `json:"OrderRpc" yaml:"OrderRpc"`
	StockRpc   zrpc.RpcClientConf `json:"StockRpc" yaml:"StockRpc"`
	CronJobRpc zrpc.RpcClientConf `json:"CronJobRpc" yaml:"CronJobRpc"`

	// DTM HTTP 地址（用于生成 gid），如 "http://dtm:36789"
	DTM string `json:"DTM" yaml:"DTM"`
	// DTM gRPC 地址（用于提交 Saga），如 "dtm:36790"
	DTMEndpoint string `json:"DTMEndpoint" yaml:"DTMEndpoint"`

	// Saga 分支的 gRPC 目标地址（供 dtm 回调调用）
	StockRpcTarget string `json:"StockRpcTarget" yaml:"StockRpcTarget"` // "stock-rpc:8082"
	OrderRpcTarget string `json:"OrderRpcTarget" yaml:"OrderRpcTarget"` // "order-rpc:8081"
}

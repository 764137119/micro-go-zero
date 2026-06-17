package config

import (
	commconfig "common/config"

	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	zrpc.RpcServerConf
	DTM      string             // dtm 服务地址，如 "http://dtm:36789"
	StockRpc zrpc.RpcClientConf // stock-rpc 客户端配置
	DB       commconfig.DBConfig
}

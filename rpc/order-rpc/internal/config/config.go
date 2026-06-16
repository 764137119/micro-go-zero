package config

import "github.com/zeromicro/go-zero/zrpc"

type Config struct {
	zrpc.RpcServerConf
	DTM      string             // dtm 服务地址，如 "http://dtm:36789"
	StockRpc zrpc.RpcClientConf // stock-rpc 客户端配置
	DB       struct {
		DataSource string // MySQL 连接串，如 "root:root123@tcp(mysql:3306)/order_db?charset=utf8mb4&parseTime=True&loc=Local"
	}
}

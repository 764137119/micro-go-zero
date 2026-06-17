package config

import (
	commconfig "common/config"

	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	zrpc.RpcServerConf
	DB commconfig.DBConfig
}

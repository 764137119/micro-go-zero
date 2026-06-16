package config

import "github.com/zeromicro/go-zero/zrpc"

type Config struct {
	Name string `yaml:"Name"`
	Host string `yaml:"Host"`
	Port int    `yaml:"Port"`

	UserRpc  zrpc.RpcClientConf `yaml:"UserRpc"`
	OrderRpc zrpc.RpcClientConf `yaml:"OrderRpc"`
}

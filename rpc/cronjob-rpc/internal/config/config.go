package config

import (
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	zrpc.RpcServerConf

	DB struct {
		DataSource string
	}

	Scheduler struct {
		ScanIntervalSec int `yaml:"ScanIntervalSec"`
		LeaderTTLSec    int `yaml:"LeaderTTLSec"`
		TaskLockTTLSec  int `yaml:"TaskLockTTLSec"`
	}
}

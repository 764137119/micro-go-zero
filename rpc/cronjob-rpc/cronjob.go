package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"cronjob-rpc/cronjob"
	"cronjob-rpc/internal/config"
	"cronjob-rpc/internal/server"
	"cronjob-rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var configFile = flag.String("f", "etc/cronjob.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)
	ctx := svc.NewServiceContext(c)

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		cronjob.RegisterCronJobServer(grpcServer, server.NewCronJobServer(ctx))

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	defer s.Stop()

	// 启动后台调度组件（Leader 选举 + 定时调度）
	ctx.Start()
	defer ctx.Stop()

	logx.Info("CronJob RPC server started")
	fmt.Printf("Starting cronjob rpc server at %s...\n", c.ListenOn)

	// 等待优雅退出信号
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch
	logx.Info("Shutting down cronjob server...")
}

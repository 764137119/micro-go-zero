package main

import (
	"flag"
	"fmt"

	"api-gateway/config"
	"api-gateway/svc"

	"github.com/gin-gonic/gin"
	"github.com/zeromicro/go-zero/core/conf"
)

var configFile = flag.String("f", "etc/config.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	svcCtx := svc.NewServiceContext(c)

	r := gin.Default()

	RegisterRoutes(r, svcCtx)

	addr := fmt.Sprintf("%s:%d", c.Host, c.Port)
	fmt.Printf("Starting api-gateway server at %s...\n", addr)
	if err := r.Run(addr); err != nil {
		panic(err)
	}
}

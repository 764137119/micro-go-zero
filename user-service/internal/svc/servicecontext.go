// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package svc

import (
	"user-rpc/userclient"
	"user-service/internal/config"
	"user-service/internal/middleware"

	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config         config.Config
	AuthMiddleware rest.Middleware
	UserRpc        userclient.User
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config:         c,
		AuthMiddleware: middleware.NewAuthMiddleware().Handle,
		UserRpc:        userclient.NewUser(zrpc.MustNewClient(c.UserRpc)),
	}
}

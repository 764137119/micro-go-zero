package svc

import (
	"common/gormx"
	"user-rpc/internal/config"
	"user-rpc/internal/model"
)

type ServiceContext struct {
	Config   config.Config
	UserRepo *model.UserRepo
}

func NewServiceContext(c config.Config) *ServiceContext {
	db := gormx.MustNewDB(c.DB.DataSource)

	return &ServiceContext{
		Config:   c,
		UserRepo: model.NewUserRepo(db),
	}
}

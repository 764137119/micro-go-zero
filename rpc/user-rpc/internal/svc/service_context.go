package svc

import (
	"log"

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

	// 自动迁移公共基础设施表（Saga 等）
	gormx.MustMigrateCommon(db)

	// 自动迁移业务表
	if err := db.AutoMigrate(&model.User{}); err != nil {
		log.Fatalf("auto migrate user table failed: %v", err)
	}

	return &ServiceContext{
		Config:   c,
		UserRepo: model.NewUserRepo(db),
	}
}

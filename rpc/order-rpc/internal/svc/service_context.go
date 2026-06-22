package svc

import (
	"log"

	"common/gormx"
	commodel "common/model"
	"order-rpc/internal/config"
	"order-rpc/internal/model"
)

type ServiceContext struct {
	Config                    config.Config
	DTMEndpoint               string // dtm 协调器地址
	OrderRepo                 *model.OrderRepo
	SagaGlobalTransactionRepo *commodel.SagaGlobalTransactionRepo
}

func NewServiceContext(c config.Config) *ServiceContext {
	db := gormx.MustNewDB(c.DB.DataSource)

	// 自动迁移公共基础设施表（Saga 等）
	gormx.MustMigrateCommon(db)

	// 自动迁移业务表
	if err := db.AutoMigrate(
		&model.Order{},
		&model.OrderTccControl{},
	); err != nil {
		log.Fatalf("auto migrate order table failed: %v", err)
	}

	return &ServiceContext{
		Config:                    c,
		DTMEndpoint:               c.DTM,
		OrderRepo:                 model.NewOrderRepo(db),
		SagaGlobalTransactionRepo: commodel.NewSagaGlobalTransactionRepo(db),
	}
}

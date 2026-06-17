package svc

import (
	"log"

	"common/gormx"
	"stock-rpc/internal/config"
	"stock-rpc/internal/model"
)

type ServiceContext struct {
	Config    config.Config
	StockRepo *model.StockRepo
}

func NewServiceContext(c config.Config) *ServiceContext {
	db := gormx.MustNewDB(c.DB.DataSource)

	// 自动迁移公共基础设施表（Saga 等）
	gormx.MustMigrateCommon(db)

	// 自动迁移业务表
	if err := db.AutoMigrate(&model.Stock{}, &model.StockFlowLog{}); err != nil {
		log.Fatalf("auto migrate stock table failed: %v", err)
	}

	return &ServiceContext{
		Config:    c,
		StockRepo: model.NewStockRepo(db),
	}
}

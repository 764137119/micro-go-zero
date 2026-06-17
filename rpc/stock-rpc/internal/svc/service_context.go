package svc

import (
	"log"

	"common/gormx"
	"stock-rpc/internal/config"
	"stock-rpc/internal/model"

	"gorm.io/gorm"
)

type ServiceContext struct {
	Config    config.Config
	DB        *gorm.DB
	StockRepo *model.StockRepo
}

func NewServiceContext(c config.Config) *ServiceContext {
	db := gormx.MustNewDB(c.DB.DataSource)

	// 自动迁移公共基础设施表（Saga 等）
	gormx.MustMigrateCommon(db)

	// 自动迁移业务表
	if err := db.AutoMigrate(
		&model.Stock{},
		&model.StockFlowLog{},
		&model.StockTccControl{},
	); err != nil {
		log.Fatalf("auto migrate stock table failed: %v", err)
	}

	return &ServiceContext{
		Config:    c,
		DB:        db,
		StockRepo: model.NewStockRepo(db),
	}
}

package svc

import (
	"log"

	"order-rpc/internal/config"
	"order-rpc/internal/model"
	"stock-rpc/stockclient"

	"github.com/zeromicro/go-zero/zrpc"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type ServiceContext struct {
	Config      config.Config
	StockRpc    stockclient.Stock
	DTMEndpoint string // dtm 协调器地址
	OrderRepo   *model.OrderRepo
}

func NewServiceContext(c config.Config) *ServiceContext {
	db := mustNewDB(c.DB.DataSource)

	// 自动迁移建表
	if err := db.AutoMigrate(&model.Order{}); err != nil {
		log.Fatalf("auto migrate order table failed: %v", err)
	}

	return &ServiceContext{
		Config:      c,
		StockRpc:    stockclient.NewStock(zrpc.MustNewClient(c.StockRpc)),
		DTMEndpoint: c.DTM,
		OrderRepo:   model.NewOrderRepo(db),
	}
}

// mustNewDB 创建 GORM 数据库连接
func mustNewDB(dsn string) *gorm.DB {
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		SkipDefaultTransaction: true,
	})
	if err != nil {
		log.Fatalf("failed to connect mysql: %v, dsn: %s", err, dsn)
	}
	return db
}

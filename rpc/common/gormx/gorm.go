package gormx

import (
	"log"

	"common/model"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// MustNewDB 创建 GORM 数据库连接，失败则 Fatal
func MustNewDB(dsn string) *gorm.DB {
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		SkipDefaultTransaction: true,
	})
	if err != nil {
		log.Fatalf("failed to connect mysql: %v, dsn: %s", err, dsn)
	}
	return db
}

// NewDB 创建 GORM 数据库连接，失败返回 error
func NewDB(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		SkipDefaultTransaction: true,
	})
	if err != nil {
		return nil, err
	}
	return db, nil
}

// commonModels 公共基础设施表，各服务自动迁移时统一创建
var commonModels = []interface{}{
	&model.SagaGlobalTransaction{},
	&model.SagaBranchTransaction{},
}

// MustMigrateCommon 自动迁移公共基础设施表（Saga 等）
// 参与分布式事务的服务在启动时调用此函数
func MustMigrateCommon(db *gorm.DB) {
	for _, m := range commonModels {
		if err := db.AutoMigrate(m); err != nil {
			log.Fatalf("auto migrate %T failed: %v", m, err)
		}
	}
}

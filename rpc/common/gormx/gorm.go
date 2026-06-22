package gormx

import (
	"log"

	"common/model"

	mysqlDriver "github.com/go-sql-driver/mysql"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// MustNewDB 创建 GORM 数据库连接，失败则 Fatal
// 自动创建不存在的数据库
func MustNewDB(dsn string) *gorm.DB {
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		SkipDefaultTransaction: true,
	})
	if err != nil {
		// 数据库可能不存在，尝试创建
		if err := ensureDatabaseExists(dsn); err != nil {
			log.Fatalf("failed to create database: %v, dsn: %s", err, dsn)
		}
		db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
			SkipDefaultTransaction: true,
		})
		if err != nil {
			log.Fatalf("failed to connect mysql: %v, dsn: %s", err, dsn)
		}
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

// ensureDatabaseExists 确保数据库存在，不存在则创建
func ensureDatabaseExists(dsn string) error {
	cfg, err := mysqlDriver.ParseDSN(dsn)
	if err != nil {
		return err
	}
	dbName := cfg.DBName
	cfg.DBName = "" // 先不指定数据库连接

	tempDB, err := gorm.Open(mysql.Open(cfg.FormatDSN()), &gorm.Config{})
	if err != nil {
		return err
	}
	sqlDB, _ := tempDB.DB()
	defer sqlDB.Close()

	return tempDB.Exec("CREATE DATABASE IF NOT EXISTS `" + dbName + "` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci").Error
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

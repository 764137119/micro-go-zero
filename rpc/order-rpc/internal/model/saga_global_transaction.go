package model

import (
	"context"
	"time"

	"gorm.io/gorm"
)

// SagaGlobalTransactionStatus Saga全局事务状态枚举
type SagaGlobalTransactionStatus int32

const (
	SagaGlobalTransactionStatusRunning      SagaGlobalTransactionStatus = 0 // 运行中
	SagaGlobalTransactionStatusSucceed      SagaGlobalTransactionStatus = 1 // 成功
	SagaGlobalTransactionStatusFailed       SagaGlobalTransactionStatus = 2 // 失败
	SagaGlobalTransactionStatusCompensating SagaGlobalTransactionStatus = 3 // 补偿中
	SagaGlobalTransactionStatusCompensated  SagaGlobalTransactionStatus = 4 // 已补偿
)

/*
CREATE TABLE `saga_global_transaction` (
    `xid` VARCHAR(64) NOT NULL COMMENT '全局事务ID，唯一标识一个Saga事务',
    `transaction_name` VARCHAR(256) DEFAULT NULL COMMENT '事务名称/业务标识',
    `status` TINYINT NOT NULL COMMENT '状态: 0-Running, 1-Succeed, 2-Failed, 3-Compensating, 4-Compensated',
    `start_time` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '事务开始时间',
    `end_time` DATETIME(3) DEFAULT NULL COMMENT '事务结束时间',
    `timeout` INT NOT NULL DEFAULT 3600 COMMENT '超时时间(秒)，超时自动触发补偿',
    `retry_count` INT NOT NULL DEFAULT 0 COMMENT '全局重试次数',
    `version` BIGINT NOT NULL DEFAULT 1 COMMENT '乐观锁版本号，防止并发更新',
    `gmt_create` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    `gmt_modified` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    PRIMARY KEY (`xid`),
    KEY `idx_status_gmt_modified` (`status`, `gmt_modified`) COMMENT '用于定时任务扫描超时/异常事务'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Saga全局事务状态表';
*/

// SagaGlobalTransaction Saga全局事务状态表模型
type SagaGlobalTransaction struct {
	Xid             string                      `gorm:"primaryKey;column:xid;type:varchar(64);not null"`                          // 全局事务ID，唯一标识一个Saga事务
	TransactionName string                      `gorm:"column:transaction_name;type:varchar(256)"`                                // 事务名称/业务标识
	Status          SagaGlobalTransactionStatus `gorm:"column:status;type:tinyint;not null;index:idx_status_gmt_modified"`        // 状态: 0-Running, 1-Succeed, 2-Failed, 3-Compensating, 4-Compensated
	StartTime       time.Time                   `gorm:"column:start_time;type:datetime(3);not null;default:CURRENT_TIMESTAMP(3)"` // 事务开始时间
	EndTime         *time.Time                  `gorm:"column:end_time;type:datetime(3)"`                                         // 事务结束时间
	Timeout         int32                       `gorm:"column:timeout;type:int;not null;default:3600"`                            // 超时时间(秒)，超时自动触发补偿
	RetryCount      int32                       `gorm:"column:retry_count;type:int;not null;default:0"`                           // 全局重试次数
	Version         int64                       `gorm:"column:version;type:bigint;not null;default:1"`                            // 乐观锁版本号，防止并发更新
	GmtCreate       time.Time                   `gorm:"column:gmt_create;type:datetime(3);not null;autoCreateTime"`               // 创建时间
	GmtModified     time.Time                   `gorm:"column:gmt_modified;type:datetime(3);not null;autoUpdateTime"`             // 修改时间
}

// TableName 自定义表名
func (SagaGlobalTransaction) TableName() string {
	return "saga_global_transaction"
}

// SagaGlobalTransactionRepo Saga全局事务仓储层
type SagaGlobalTransactionRepo struct {
	db *gorm.DB
}

// NewSagaGlobalTransactionRepo 创建Saga全局事务仓储
func NewSagaGlobalTransactionRepo(db *gorm.DB) *SagaGlobalTransactionRepo {
	return &SagaGlobalTransactionRepo{db: db}
}

// Create 创建全局事务
func (r *SagaGlobalTransactionRepo) Create(ctx context.Context, tx *SagaGlobalTransaction) error {
	return r.db.WithContext(ctx).Create(tx).Error
}

// FindByXid 根据XID查询全局事务
func (r *SagaGlobalTransactionRepo) FindByXid(ctx context.Context, xid string) (*SagaGlobalTransaction, error) {
	var tx SagaGlobalTransaction
	err := r.db.WithContext(ctx).Where("xid = ?", xid).First(&tx).Error
	if err != nil {
		return nil, err
	}
	return &tx, nil
}

// UpdateStatus 更新事务状态（带乐观锁）
func (r *SagaGlobalTransactionRepo) UpdateStatus(ctx context.Context, xid string, fromStatus, toStatus SagaGlobalTransactionStatus, version int64) error {
	now := time.Now()
	updates := map[string]interface{}{
		"status":       toStatus,
		"version":      gorm.Expr("version + 1"),
		"gmt_modified": now,
	}
	// 如果更新到终态（成功/已补偿/失败），设置结束时间
	if toStatus == SagaGlobalTransactionStatusSucceed ||
		toStatus == SagaGlobalTransactionStatusCompensated ||
		toStatus == SagaGlobalTransactionStatusFailed {
		updates["end_time"] = now
	}
	return r.db.WithContext(ctx).Model(&SagaGlobalTransaction{}).
		Where("xid = ? AND status = ? AND version = ?", xid, fromStatus, version).
		Updates(updates).Error
}

// UpdateStatusDirect 直接更新事务状态（无乐观锁，用于补偿等强制操作）
func (r *SagaGlobalTransactionRepo) UpdateStatusDirect(ctx context.Context, xid string, status SagaGlobalTransactionStatus) error {
	updates := map[string]interface{}{
		"status":       status,
		"gmt_modified": time.Now(),
	}
	if status == SagaGlobalTransactionStatusSucceed ||
		status == SagaGlobalTransactionStatusCompensated ||
		status == SagaGlobalTransactionStatusFailed {
		updates["end_time"] = time.Now()
	}
	return r.db.WithContext(ctx).Model(&SagaGlobalTransaction{}).
		Where("xid = ?", xid).
		Updates(updates).Error
}

// IncrementRetryCount 增加重试次数
func (r *SagaGlobalTransactionRepo) IncrementRetryCount(ctx context.Context, xid string) error {
	return r.db.WithContext(ctx).Model(&SagaGlobalTransaction{}).
		Where("xid = ?", xid).
		Updates(map[string]interface{}{
			"retry_count":  gorm.Expr("retry_count + 1"),
			"gmt_modified": time.Now(),
		}).Error
}

// FindRunningTimeout 查询超时的运行中事务（用于定时任务扫描补偿）
func (r *SagaGlobalTransactionRepo) FindRunningTimeout(ctx context.Context, limit int) ([]*SagaGlobalTransaction, error) {
	var list []*SagaGlobalTransaction
	err := r.db.WithContext(ctx).
		Where("status = ? AND TIMESTAMPDIFF(SECOND, start_time, NOW()) > timeout", SagaGlobalTransactionStatusRunning).
		Order("gmt_modified ASC").
		Limit(limit).
		Find(&list).Error
	if err != nil {
		return nil, err
	}
	return list, nil
}

// FindByStatus 根据状态查询事务列表
func (r *SagaGlobalTransactionRepo) FindByStatus(ctx context.Context, status SagaGlobalTransactionStatus, limit int) ([]*SagaGlobalTransaction, error) {
	var list []*SagaGlobalTransaction
	err := r.db.WithContext(ctx).
		Where("status = ?", status).
		Order("gmt_modified ASC").
		Limit(limit).
		Find(&list).Error
	if err != nil {
		return nil, err
	}
	return list, nil
}

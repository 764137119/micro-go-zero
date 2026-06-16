package model

import (
	"context"
	"time"

	"gorm.io/gorm"
)

// SagaBranchTransactionStatus Saga分支事务状态枚举
type SagaBranchTransactionStatus int32

const (
	SagaBranchTransactionStatusTrying    SagaBranchTransactionStatus = 0 // 执行中
	SagaBranchTransactionStatusSucceed   SagaBranchTransactionStatus = 1 // 成功
	SagaBranchTransactionStatusFailed    SagaBranchTransactionStatus = 2 // 失败
	SagaBranchTransactionStatusCancelled SagaBranchTransactionStatus = 3 // 已取消
)

// SagaOperationType Saga分支操作类型枚举
type SagaOperationType int32

const (
	SagaOperationTypeForward    SagaOperationType = 1 // 正向操作
	SagaOperationTypeCompensate SagaOperationType = 2 // 补偿操作
)

// SagaBranchTransaction Saga分支事务与补偿日志表模型
type SagaBranchTransaction struct {
	BranchId      string                      `gorm:"primaryKey;column:branch_id;type:varchar(64);not null"`
	Xid           string                      `gorm:"column:xid;type:varchar(64);not null;index:idx_xid;uniqueIndex:uk_xid_service_op"`
	ServiceName   string                      `gorm:"column:service_name;type:varchar(256);not null;uniqueIndex:uk_xid_service_op"`
	OperationType SagaOperationType           `gorm:"column:operation_type;type:tinyint;not null;uniqueIndex:uk_xid_service_op"`
	Status        SagaBranchTransactionStatus `gorm:"column:status;type:tinyint;not null"`
	RequestData   *string                     `gorm:"column:request_data;type:text"`
	ResponseData  *string                     `gorm:"column:response_data;type:text"`
	RetryCount    int32                       `gorm:"column:retry_count;type:int;not null;default:0"`
	NextRetryTime *time.Time                  `gorm:"column:next_retry_time;type:datetime(3)"`
	Version       int64                       `gorm:"column:version;type:bigint;not null;default:1"`
	GmtCreate     time.Time                   `gorm:"column:gmt_create;type:datetime(3);not null;autoCreateTime"`
	GmtModified   time.Time                   `gorm:"column:gmt_modified;type:datetime(3);not null;autoUpdateTime"`
}

// TableName 自定义表名
func (SagaBranchTransaction) TableName() string {
	return "saga_branch_transaction"
}

// SagaBranchTransactionRepo Saga分支事务仓储层
type SagaBranchTransactionRepo struct {
	db *gorm.DB
}

// NewSagaBranchTransactionRepo 创建Saga分支事务仓储
func NewSagaBranchTransactionRepo(db *gorm.DB) *SagaBranchTransactionRepo {
	return &SagaBranchTransactionRepo{db: db}
}

// Create 创建分支事务
func (r *SagaBranchTransactionRepo) Create(ctx context.Context, tx *SagaBranchTransaction) error {
	return r.db.WithContext(ctx).Create(tx).Error
}

// FindByBranchId 根据分支ID查询
func (r *SagaBranchTransactionRepo) FindByBranchId(ctx context.Context, branchId string) (*SagaBranchTransaction, error) {
	var tx SagaBranchTransaction
	err := r.db.WithContext(ctx).Where("branch_id = ?", branchId).First(&tx).Error
	if err != nil {
		return nil, err
	}
	return &tx, nil
}

// FindByXid 根据全局事务ID查询所有分支
func (r *SagaBranchTransactionRepo) FindByXid(ctx context.Context, xid string) ([]*SagaBranchTransaction, error) {
	var list []*SagaBranchTransaction
	err := r.db.WithContext(ctx).Where("xid = ?", xid).
		Order("gmt_create ASC").
		Find(&list).Error
	if err != nil {
		return nil, err
	}
	return list, nil
}

// FindByXidAndServiceAndOp 根据幂等键查询（xid + service_name + operation_type）
func (r *SagaBranchTransactionRepo) FindByXidAndServiceAndOp(ctx context.Context, xid string, serviceName string, opType SagaOperationType) (*SagaBranchTransaction, error) {
	var tx SagaBranchTransaction
	err := r.db.WithContext(ctx).
		Where("xid = ? AND service_name = ? AND operation_type = ?", xid, serviceName, opType).
		First(&tx).Error
	if err != nil {
		return nil, err
	}
	return &tx, nil
}

// UpdateStatus 更新分支事务状态（带乐观锁）
func (r *SagaBranchTransactionRepo) UpdateStatus(ctx context.Context, branchId string, fromStatus, toStatus SagaBranchTransactionStatus, version int64) error {
	return r.db.WithContext(ctx).Model(&SagaBranchTransaction{}).
		Where("branch_id = ? AND status = ? AND version = ?", branchId, fromStatus, version).
		Updates(map[string]interface{}{
			"status":       toStatus,
			"version":      gorm.Expr("version + 1"),
			"gmt_modified": time.Now(),
		}).Error
}

// UpdateStatusDirect 直接更新分支事务状态（无乐观锁）
func (r *SagaBranchTransactionRepo) UpdateStatusDirect(ctx context.Context, branchId string, status SagaBranchTransactionStatus) error {
	return r.db.WithContext(ctx).Model(&SagaBranchTransaction{}).
		Where("branch_id = ?", branchId).
		Updates(map[string]interface{}{
			"status":       status,
			"gmt_modified": time.Now(),
		}).Error
}

// IncrementRetryCount 增加分支重试次数，并更新下次重试时间（指数退避）
func (r *SagaBranchTransactionRepo) IncrementRetryCount(ctx context.Context, branchId string) error {
	// 先查出当前重试次数
	var tx SagaBranchTransaction
	err := r.db.WithContext(ctx).Where("branch_id = ?", branchId).First(&tx).Error
	if err != nil {
		return err
	}
	// 指数退避：下次重试时间 = 当前时间 + 2^retry_count 秒
	nextRetry := time.Now().Add(time.Duration(1<<tx.RetryCount) * time.Second)
	return r.db.WithContext(ctx).Model(&SagaBranchTransaction{}).
		Where("branch_id = ?", branchId).
		Updates(map[string]interface{}{
			"retry_count":     gorm.Expr("retry_count + 1"),
			"next_retry_time": nextRetry,
			"gmt_modified":    time.Now(),
		}).Error
}

// UpdateNextRetryTime 更新下次重试时间
func (r *SagaBranchTransactionRepo) UpdateNextRetryTime(ctx context.Context, branchId string, nextRetryTime time.Time) error {
	return r.db.WithContext(ctx).Model(&SagaBranchTransaction{}).
		Where("branch_id = ?", branchId).
		Updates(map[string]interface{}{
			"next_retry_time": nextRetryTime,
			"gmt_modified":    time.Now(),
		}).Error
}

// FindByStatus 根据状态查询分支事务列表
func (r *SagaBranchTransactionRepo) FindByStatus(ctx context.Context, status SagaBranchTransactionStatus, limit int) ([]*SagaBranchTransaction, error) {
	var list []*SagaBranchTransaction
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

// FindRetryable 查询需要重试的分支事务（状态为失败且到达下次重试时间）
func (r *SagaBranchTransactionRepo) FindRetryable(ctx context.Context, limit int) ([]*SagaBranchTransaction, error) {
	var list []*SagaBranchTransaction
	err := r.db.WithContext(ctx).
		Where("status = ? AND next_retry_time IS NOT NULL AND next_retry_time <= NOW()", SagaBranchTransactionStatusFailed).
		Order("next_retry_time ASC").
		Limit(limit).
		Find(&list).Error
	if err != nil {
		return nil, err
	}
	return list, nil
}

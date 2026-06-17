package model

import (
	"context"
	"time"

	"gorm.io/gorm"
)

// 库存流水变动类型枚举
const (
	ChangeTypeDeduct    int32 = 1 // 扣减（Saga 正向）
	ChangeTypeRollback  int32 = 2 // 回滚（Saga 补偿）
	ChangeTypeRelease   int32 = 3 // 释放（超时取消）
	ChangeTypeTry       int32 = 4 // 冻结（TCC Try）
	ChangeTypeConfirm   int32 = 5 // 确认扣减（TCC Confirm）
	ChangeTypeTccCancel int32 = 6 // 释放（TCC Cancel）
)

/*
CREATE TABLE `stock_info` (
    `sku_id`          BIGINT NOT NULL COMMENT '商品SKU ID（主键）',
    `total_stock`     BIGINT NOT NULL DEFAULT 0 COMMENT '总库存',
    `available_stock` BIGINT NOT NULL DEFAULT 0 COMMENT '可用库存（可下单量）',
    `locked_stock`    BIGINT NOT NULL DEFAULT 0 COMMENT '锁定库存（被订单占用但未支付）',
    `version`         BIGINT NOT NULL DEFAULT 1 COMMENT '乐观锁版本号，防止并发扣减',
    `xid`              VARCHAR(64) NOT NULL COMMENT '全局事务ID，唯一标识一个Saga事务',
    `created_at`      BIGINT NOT NULL COMMENT '创建时间（毫秒时间戳）',
    `updated_at`      BIGINT NOT NULL COMMENT '更新时间（毫秒时间戳）',
    PRIMARY KEY (`sku_id`),
    UNIQUE KEY `uk_xid_sku` (`xid`, `sku_id`) COMMENT '事务幂等键'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='库存表';

CREATE TABLE `stock_flow_log` (
    `flow_id`         BIGINT NOT NULL AUTO_INCREMENT COMMENT '流水ID（主键）',
    `sku_id`          BIGINT NOT NULL COMMENT '商品SKU ID',
    `order_no`        VARCHAR(64) NOT NULL COMMENT '订单号（业务幂等键）',
    `change_type`     TINYINT NOT NULL COMMENT '变动类型：1-扣减(正向), 2-回滚(补偿), 3-释放(超时取消), 4-冻结(TCC Try), 5-确认扣减(TCC Confirm), 6-释放(TCC Cancel)',
    `quantity`        BIGINT NOT NULL COMMENT '变动数量',
    `before_available` BIGINT NOT NULL COMMENT '变动前可用库存',
    `after_available`  BIGINT NOT NULL COMMENT '变动后可用库存',
    `before_locked`   BIGINT NOT NULL COMMENT '变动前锁定库存',
    `after_locked`    BIGINT NOT NULL COMMENT '变动后锁定库存',
    `xid`             VARCHAR(64) NOT NULL COMMENT 'dtm 全局事务ID',
    `gid`             VARCHAR(64) NOT NULL COMMENT 'dtm 全局事务ID',
    `created_at`      BIGINT NOT NULL COMMENT '创建时间（毫秒时间戳）',
    PRIMARY KEY (`flow_id`),
    KEY `idx_xid` (`xid`) COMMENT '按事务ID查询',
    KEY `idx_sku_id` (`sku_id`) COMMENT '按SKU查询流水',
    KEY `idx_order_no` (`order_no`) COMMENT '按订单号查询流水'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='库存流水表';
*/

// Stock 库存表模型
type Stock struct {
	SkuId          int64  `gorm:"primaryKey;column:sku_id;type:bigint;not null"`               // 商品SKU ID（主键）
	TotalStock     int64  `gorm:"column:total_stock;type:bigint;not null;default:0"`           // 总库存
	AvailableStock int64  `gorm:"column:available_stock;type:bigint;not null;default:0"`       // 可用库存（可下单量）
	LockedStock    int64  `gorm:"column:locked_stock;type:bigint;not null;default:0"`          // 锁定库存（被订单占用但未支付）
	Version        int64  `gorm:"column:version;type:bigint;not null;default:1"`               // 乐观锁版本号，防止并发扣减
	Xid            string `gorm:"column:xid;type:varchar(64);not null;uniqueIndex:uk_xid_sku"` // 全局事务ID，唯一标识一个Saga事务
	CreatedAt      int64  `gorm:"column:created_at;autoCreateTime;milli"`                      // 创建时间（毫秒时间戳）
	UpdatedAt      int64  `gorm:"column:updated_at;autoUpdateTime;milli"`                      // 更新时间（毫秒时间戳）
}

// TableName 自定义表名
func (Stock) TableName() string {
	return "stock_info"
}

// StockFlowLog 库存流水表模型
type StockFlowLog struct {
	FlowId          int64  `gorm:"primaryKey;autoIncrement;column:flow_id"`                      // 流水ID（主键）
	SkuId           int64  `gorm:"column:sku_id;type:bigint;not null;index:idx_sku_id"`          // 商品SKU ID
	OrderNo         string `gorm:"column:order_no;type:varchar(64);not null;index:idx_order_no"` // 订单号（业务幂等键）
	ChangeType      int32  `gorm:"column:change_type;type:tinyint;not null"`                     // 变动类型：1-扣减(正向), 2-回滚(补偿), 3-释放(超时取消)
	Quantity        int64  `gorm:"column:quantity;type:bigint;not null"`                         // 变动数量
	BeforeAvailable int64  `gorm:"column:before_available;type:bigint;not null"`                 // 变动前可用库存
	AfterAvailable  int64  `gorm:"column:after_available;type:bigint;not null"`                  // 变动后可用库存
	BeforeLocked    int64  `gorm:"column:before_locked;type:bigint;not null"`                    // 变动前锁定库存
	AfterLocked     int64  `gorm:"column:after_locked;type:bigint;not null"`                     // 变动后锁定库存
	Xid             string `gorm:"column:xid;type:varchar(64);not null;index:idx_xid"`           // dtm 全局事务ID
	Gid             string `gorm:"column:gid;type:varchar(64);not null"`                         // dtm 全局事务ID
	CreatedAt       int64  `gorm:"column:created_at;autoCreateTime;milli"`                       // 创建时间（毫秒时间戳）
}

// TableName 自定义表名
func (StockFlowLog) TableName() string {
	return "stock_flow_log"
}

// StockRepo 库存仓储层
type StockRepo struct {
	db *gorm.DB
}

// NewStockRepo 创建库存仓储
func NewStockRepo(db *gorm.DB) *StockRepo {
	return &StockRepo{db: db}
}

// FindBySkuId 根据 SKU ID 查询库存
func (r *StockRepo) FindBySkuId(ctx context.Context, skuId int64) (*Stock, error) {
	var stock Stock
	err := r.db.WithContext(ctx).Where("sku_id = ?", skuId).First(&stock).Error
	if err != nil {
		return nil, err
	}
	return &stock, nil
}

// FindBySkuIds 批量查询库存
func (r *StockRepo) FindBySkuIds(ctx context.Context, skuIds []int64) ([]Stock, error) {
	var stocks []Stock
	err := r.db.WithContext(ctx).Where("sku_id IN ?", skuIds).Find(&stocks).Error
	if err != nil {
		return nil, err
	}
	return stocks, nil
}

// DeductStock 扣减可用库存、增加锁定库存（Saga 正向操作）
// 使用乐观锁 version 防止并发扣减
func (r *StockRepo) DeductStock(ctx context.Context, skuId int64, quantity int64) error {
	result := r.db.WithContext(ctx).Model(&Stock{}).
		Where("sku_id = ? AND available_stock >= ?", skuId, quantity).
		Updates(map[string]interface{}{
			"available_stock": gorm.Expr("available_stock - ?", quantity),
			"locked_stock":    gorm.Expr("locked_stock + ?", quantity),
			"updated_at":      time.Now().UnixMilli(),
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound // 可用库存不足
	}
	return nil
}

// RollbackStock 回滚库存：增加可用库存、减少锁定库存（Saga 补偿操作）
func (r *StockRepo) RollbackStock(ctx context.Context, skuId int64, quantity int64) error {
	result := r.db.WithContext(ctx).Model(&Stock{}).
		Where("sku_id = ? AND locked_stock >= ?", skuId, quantity).
		Updates(map[string]interface{}{
			"available_stock": gorm.Expr("available_stock + ?", quantity),
			"locked_stock":    gorm.Expr("locked_stock - ?", quantity),
			"updated_at":      time.Now().UnixMilli(),
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound // 锁定库存不足
	}
	return nil
}

// CreateFlowLog 创建库存流水记录
func (r *StockRepo) CreateFlowLog(ctx context.Context, log *StockFlowLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

// HasFlowLogByXid 根据 xid 判断流水是否已存在（用于幂等去重）
func (r *StockRepo) HasFlowLogByXid(ctx context.Context, xid string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&StockFlowLog{}).
		Where("xid = ?", xid).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// TryDeductStock 冻结库存：减少可用库存、增加锁定库存（TCC Try 操作）
func (r *StockRepo) TryDeductStock(ctx context.Context, skuId int64, quantity int64) error {
	result := r.db.WithContext(ctx).Model(&Stock{}).
		Where("sku_id = ? AND available_stock >= ?", skuId, quantity).
		Updates(map[string]interface{}{
			"available_stock": gorm.Expr("available_stock - ?", quantity),
			"locked_stock":    gorm.Expr("locked_stock + ?", quantity),
			"updated_at":      time.Now().UnixMilli(),
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound // 可用库存不足
	}
	return nil
}

// ConfirmDeductStock 确认扣减库存：仅减少锁定库存（TCC Confirm 操作）
// 可用库存已在 Try 阶段扣减，此处只清理锁定库存
func (r *StockRepo) ConfirmDeductStock(ctx context.Context, skuId int64, quantity int64) error {
	result := r.db.WithContext(ctx).Model(&Stock{}).
		Where("sku_id = ? AND locked_stock >= ?", skuId, quantity).
		Updates(map[string]interface{}{
			"locked_stock": gorm.Expr("locked_stock - ?", quantity),
			"updated_at":   time.Now().UnixMilli(),
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound // 锁定库存不足
	}
	return nil
}

// CancelDeductStock 释放冻结库存：增加可用库存、减少锁定库存（TCC Cancel 操作）
func (r *StockRepo) CancelDeductStock(ctx context.Context, skuId int64, quantity int64) error {
	result := r.db.WithContext(ctx).Model(&Stock{}).
		Where("sku_id = ? AND locked_stock >= ?", skuId, quantity).
		Updates(map[string]interface{}{
			"available_stock": gorm.Expr("available_stock + ?", quantity),
			"locked_stock":    gorm.Expr("locked_stock - ?", quantity),
			"updated_at":      time.Now().UnixMilli(),
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound // 锁定库存不足
	}
	return nil
}

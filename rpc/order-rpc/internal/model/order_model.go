package model

import (
	"context"
	"time"

	"gorm.io/gorm"
)

/*
CREATE TABLE `order_info` (
    `order_id`         BIGINT NOT NULL AUTO_INCREMENT COMMENT '订单ID（主键）',
    `order_no`         VARCHAR(64) NOT NULL COMMENT '订单号（唯一索引）',
    `order_state`      INT NOT NULL DEFAULT 0 COMMENT '订单状态：-1-已取消, 0-待支付, 1-已支付, 2-已完成',
    `user_id`          BIGINT NOT NULL COMMENT '用户ID',
    `order_price`      BIGINT NOT NULL DEFAULT 0 COMMENT '订单金额（单位：分）',
    `order_des`        VARCHAR(255) NOT NULL DEFAULT '' COMMENT '订单描述',
    `order_begin_time` BIGINT NOT NULL DEFAULT 0 COMMENT '订单开始时间（毫秒时间戳）',
    `order_end_time`   BIGINT NOT NULL DEFAULT 0 COMMENT '订单结束时间（毫秒时间戳）',
    `sku_id`           BIGINT NOT NULL DEFAULT 0 COMMENT '商品SKU ID',
    `quantity`         BIGINT NOT NULL DEFAULT 0 COMMENT '商品数量',
    `created_at`       BIGINT NOT NULL COMMENT '创建时间（毫秒时间戳）',
    `updated_at`       BIGINT NOT NULL COMMENT '更新时间（毫秒时间戳）',
    PRIMARY KEY (`order_id`),
    UNIQUE KEY `idx_order_no` (`order_no`),
    KEY `idx_user_id` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='订单表';
*/

// Order 订单表模型
type Order struct {
	OrderId        int64  `gorm:"primaryKey;autoIncrement;column:order_id"`                           // 订单ID（主键）
	OrderNo        string `gorm:"column:order_no;type:varchar(64);not null;uniqueIndex:idx_order_no"` // 订单号（唯一索引）
	OrderState     int32  `gorm:"column:order_state;type:int;not null;default:0"`                     // 订单状态：-1-已取消, 0-待支付, 1-已支付, 2-已完成
	UserId         int64  `gorm:"column:user_id;type:bigint;not null;index:idx_user_id"`              // 用户ID
	OrderPrice     int64  `gorm:"column:order_price;type:bigint;not null;default:0"`                  // 订单金额（单位：分）
	OrderDes       string `gorm:"column:order_des;type:varchar(255);not null;default:''"`             // 订单描述
	OrderBeginTime int64  `gorm:"column:order_begin_time;type:bigint;not null;default:0"`             // 订单开始时间（毫秒时间戳）
	OrderEndTime   int64  `gorm:"column:order_end_time;type:bigint;not null;default:0"`               // 订单结束时间（毫秒时间戳）
	SkuId          int64  `gorm:"column:sku_id;type:bigint;not null;default:0"`                       // 商品SKU ID
	Quantity       int64  `gorm:"column:quantity;type:bigint;not null;default:0"`                     // 商品数量
	CreatedAt      int64  `gorm:"column:created_at;autoCreateTime;milli"`                             // 创建时间（毫秒时间戳）
	UpdatedAt      int64  `gorm:"column:updated_at;autoUpdateTime;milli"`                             // 更新时间（毫秒时间戳）
}

// TableName 自定义表名
func (Order) TableName() string {
	return "order_info"
}

// OrderRepo 订单仓储层
type OrderRepo struct {
	db *gorm.DB
}

// NewOrderRepo 创建订单仓储
func NewOrderRepo(db *gorm.DB) *OrderRepo {
	return &OrderRepo{db: db}
}

// Create 创建订单
func (r *OrderRepo) Create(ctx context.Context, order *Order) error {
	return r.db.WithContext(ctx).Create(order).Error
}

// FindByOrderId 根据订单ID查询
func (r *OrderRepo) FindByOrderId(ctx context.Context, orderId int64) (*Order, error) {
	var order Order
	err := r.db.WithContext(ctx).Where("order_id = ?", orderId).First(&order).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

// FindByOrderNo 根据订单号查询
func (r *OrderRepo) FindByOrderNo(ctx context.Context, orderNo string) (*Order, error) {
	var order Order
	err := r.db.WithContext(ctx).Where("order_no = ?", orderNo).First(&order).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

// UpdateState 更新订单状态
func (r *OrderRepo) UpdateState(ctx context.Context, orderId int64, state int32) error {
	return r.db.WithContext(ctx).Model(&Order{}).
		Where("order_id = ?", orderId).
		Update("order_state", state).Error
}

// CancelOrder 取消订单（更新状态为已取消）
func (r *OrderRepo) CancelOrder(ctx context.Context, orderId int64) error {
	return r.db.WithContext(ctx).Model(&Order{}).
		Where("order_id = ?", orderId).
		Updates(map[string]interface{}{
			"order_state": -1,
			"updated_at":  time.Now().UnixMilli(),
		}).Error
}

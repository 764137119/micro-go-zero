package model

import "gorm.io/gorm"

// TccStatus TCC 状态类型
type TccStatus string

const (
	TCCINIT      TccStatus = "INIT"
	TCCTRYING    TccStatus = "TRYING"
	TCCCONFIRMED TccStatus = "CONFIRMED"
	TCCCANCELLED TccStatus = "CANCELLED"
)

/*
CREATE TABLE `stock_tcc_control` (
  `xid` varchar(64) NOT NULL COMMENT '全局事务ID',
  `status` varchar(20) NOT NULL DEFAULT 'INIT' COMMENT '状态: TRYING, CONFIRMED, CANCELLED',
  `sku_id` bigint NOT NULL COMMENT '商品SKU ID',
  `order_no` varchar(64) NOT NULL COMMENT '关联订单号',
  `quantity` bigint NOT NULL COMMENT '冻结数量',
  `created_at` bigint(13) DEFAULT NULL,
  `updated_at` bigint(13) DEFAULT NULL,
  PRIMARY KEY (`xid`),
  KEY `idx_sku_id` (`sku_id`),
  KEY `idx_order_no` (`order_no`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='库存服务TCC控制表';
*/

// 库存TCC控制表（防悬挂&空回滚）
type StockTccControl struct {
	Xid       string    `gorm:"primaryKey;column:xid;type:varchar(64);comment:全局事务ID"`
	Status    TccStatus `gorm:"column:status;type:varchar(20);not null;default:'INIT';comment:状态: TRYING, CONFIRMED, CANCELLED"`
	SkuId     int64     `gorm:"column:sku_id;type:bigint;not null;index:idx_sku_id;comment:商品SKU ID"`
	OrderNo   string    `gorm:"column:order_no;type:varchar(64);not null;index:idx_order_no;comment:关联订单号(用于对账)"`
	Quantity  int64     `gorm:"column:quantity;type:bigint;not null;comment:冻结数量(用于回滚时校验)"`
	CreatedAt int64     `gorm:"column:created_at;autoCreateTime;milli;comment:创建时间"`
	UpdatedAt int64     `gorm:"column:updated_at;autoUpdateTime;milli;comment:更新时间"`
}

func (StockTccControl) TableName() string {
	return "stock_tcc_control"
}

type StockTCCRepo struct {
	db *gorm.DB
}

// NewStockRepo 创建库存仓储
func NewStockTCCRepo(db *gorm.DB) *StockTCCRepo {
	return &StockTCCRepo{db: db}
}

package model

/*
CREATE TABLE `order_tcc_control` (

	`xid` varchar(64) NOT NULL COMMENT '全局事务ID',
	`status` varchar(20) NOT NULL DEFAULT 'INIT' COMMENT '状态: TRYING, CONFIRMED, CANCELLED',
	`order_no` varchar(64) NOT NULL COMMENT '订单号',
	`created_at` bigint(13) DEFAULT NULL,
	`updated_at` bigint(13) DEFAULT NULL,
	PRIMARY KEY (`xid`),
	UNIQUE KEY `idx_order_no` (`order_no`),  -- 业务幂等：一个订单号只能被一个TCC处理
	KEY `idx_status` (`status`)

) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='订单服务TCC控制表';
*/
// TccStatus TCC 状态类型
type TccStatus string

const (
	TCCINIT      TccStatus = "INIT"
	TCCTRYING    TccStatus = "TRYING"
	TCCCONFIRMED TccStatus = "CONFIRMED"
	TCCCANCELLED TccStatus = "CANCELLED"
)

// OrderTccControl 订单TCC控制表（防悬挂&空回滚）
type OrderTccControl struct {
	Xid       string    `gorm:"primaryKey;column:xid;type:varchar(64);comment:全局事务ID"`
	Status    TccStatus `gorm:"column:status;type:varchar(20);not null;default:'INIT';comment:状态: TRYING, CONFIRMED, CANCELLED"`
	OrderNo   string    `gorm:"column:order_no;type:varchar(64);not null;uniqueIndex:idx_order_no;comment:订单号(业务幂等键)"`
	CreatedAt int64     `gorm:"column:created_at;autoCreateTime;milli;comment:创建时间"`
	UpdatedAt int64     `gorm:"column:updated_at;autoUpdateTime;milli;comment:更新时间"`
}

func (OrderTccControl) TableName() string {
	return "order_tcc_control"
}

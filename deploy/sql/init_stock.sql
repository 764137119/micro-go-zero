-- ============================================================
-- 库存服务 DDL
-- 数据库：stock_db（已在 init.sql 中创建）
-- ============================================================

-- 库存表
CREATE TABLE IF NOT EXISTS `stock_db`.`stock_info` (
    `sku_id`           BIGINT NOT NULL COMMENT '商品SKU ID（主键）',
    `total_stock`      BIGINT NOT NULL DEFAULT 0 COMMENT '总库存',
    `available_stock`  BIGINT NOT NULL DEFAULT 0 COMMENT '可用库存（可下单量）',
    `locked_stock`     BIGINT NOT NULL DEFAULT 0 COMMENT '锁定库存（被订单占用但未支付）',
    `version`          BIGINT NOT NULL DEFAULT 1 COMMENT '乐观锁版本号，防止并发扣减',
    `xid`              VARCHAR(64) NOT NULL COMMENT '全局事务ID，唯一标识一个Saga事务',
    `created_at`       BIGINT NOT NULL COMMENT '创建时间（毫秒时间戳）',
    `updated_at`       BIGINT NOT NULL COMMENT '更新时间（毫秒时间戳）',
    PRIMARY KEY (`sku_id`),
    UNIQUE KEY `uk_xid_sku` (`xid`, `sku_id`) COMMENT '事务幂等键'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='库存表';

-- 库存流水表
CREATE TABLE IF NOT EXISTS `stock_db`.`stock_flow_log` (
    `flow_id`          BIGINT NOT NULL AUTO_INCREMENT COMMENT '流水ID（主键）',
    `sku_id`           BIGINT NOT NULL COMMENT '商品SKU ID',
    `order_no`         VARCHAR(64) NOT NULL COMMENT '订单号（业务幂等键）',
    `change_type`      TINYINT NOT NULL COMMENT '变动类型：1-扣减(正向), 2-回滚(补偿), 3-释放(超时取消)',
    `quantity`         BIGINT NOT NULL COMMENT '变动数量',
    `before_available` BIGINT NOT NULL COMMENT '变动前可用库存',
    `after_available`  BIGINT NOT NULL COMMENT '变动后可用库存',
    `before_locked`    BIGINT NOT NULL COMMENT '变动前锁定库存',
    `after_locked`     BIGINT NOT NULL COMMENT '变动后锁定库存',
    `xid`              VARCHAR(64) NOT NULL COMMENT 'dtm 全局事务ID',
    `gid`              VARCHAR(64) NOT NULL COMMENT 'dtm 全局事务ID',
    `created_at`       BIGINT NOT NULL COMMENT '创建时间（毫秒时间戳）',
    PRIMARY KEY (`flow_id`),
    KEY `idx_xid` (`xid`) COMMENT '按事务ID查询',
    KEY `idx_sku_id` (`sku_id`) COMMENT '按SKU查询流水',
    KEY `idx_order_no` (`order_no`) COMMENT '按订单号查询流水'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='库存流水表';

-- 健康检查：验证各表是否已成功创建
SELECT IF(
    EXISTS(
        SELECT 1 FROM information_schema.TABLES
        WHERE TABLE_SCHEMA = 'stock_db' AND TABLE_NAME = 'stock_info'
    ),
    'OK: stock_info 表已存在',
    'WARN: stock_info 表不存在，请确认 stock-rpc 服务是否已启动并完成自动建表'
) AS 'stock_db.stock_info 检查';

SELECT IF(
    EXISTS(
        SELECT 1 FROM information_schema.TABLES
        WHERE TABLE_SCHEMA = 'stock_db' AND TABLE_NAME = 'stock_flow_log'
    ),
    'OK: stock_flow_log 表已存在',
    'WARN: stock_flow_log 表不存在，请确认 stock-rpc 服务是否已启动并完成自动建表'
) AS 'stock_db.stock_flow_log 检查';
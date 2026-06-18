-- 初始化数据库（MySQL 首次启动时自动执行）
-- 已通过 MYSQL_DATABASE 环境变量自动创建的：dtm

CREATE DATABASE IF NOT EXISTS `order_db` DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE DATABASE IF NOT EXISTS `stock_db` DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE DATABASE IF NOT EXISTS `user_db`  DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE DATABASE IF NOT EXISTS `cronjob_db`  DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- 注：各业务表结构由对应服务启动时通过 gorm/migrate 自动维护
-- 或由开发者在各自服务目录下的 deploy/sql/ 中维护独立的 DDL

CREATE TABLE IF NOT EXISTS `stock_db`.`saga_branch_transaction` (
    `branch_id` VARCHAR(64) NOT NULL COMMENT '分支事务ID',
    `xid` VARCHAR(64) NOT NULL COMMENT '关联的全局事务ID',
    `service_name` VARCHAR(256) NOT NULL COMMENT '参与方服务名',
    `operation_type` TINYINT NOT NULL COMMENT '操作类型: 1-Forward(正向), 2-Compensate(补偿)',
    `status` TINYINT NOT NULL COMMENT '状态: 0-Trying, 1-Succeed, 2-Failed, 3-Cancelled',
    `request_data` TEXT DEFAULT NULL COMMENT '请求参数快照(JSON)，用于补偿时重放或审计',
    `response_data` TEXT DEFAULT NULL COMMENT '响应结果快照(JSON)',
    `retry_count` INT NOT NULL DEFAULT 0 COMMENT '该分支重试次数',
    `next_retry_time` DATETIME(3) DEFAULT NULL COMMENT '下次重试时间，配合指数退避',
    `version` BIGINT NOT NULL DEFAULT 1 COMMENT '乐观锁版本号',
    `gmt_create` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    `gmt_modified` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    PRIMARY KEY (`branch_id`),
    KEY `idx_xid` (`xid`),
    UNIQUE KEY `uk_xid_service_op` (`xid`, `service_name`, `operation_type`) COMMENT '幂等键：同一事务同一服务的同类型操作只能有一条记录'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Saga分支事务与补偿日志表';


CREATE TABLE IF NOT EXISTS `order_db`.`saga_global_transaction` (
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

CREATE TABLE IF NOT EXISTS `order_db`.`order_info` (
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

CREATE TABLE IF NOT EXISTS `user_db`.`user_info` (
    `user_id`    BIGINT NOT NULL AUTO_INCREMENT COMMENT '用户ID（主键）',
    `mobile`     VARCHAR(20) NOT NULL COMMENT '手机号',
    `nick_name`  VARCHAR(64) NOT NULL DEFAULT '' COMMENT '用户昵称',
    `sex`        INT NOT NULL DEFAULT 0 COMMENT '性别：0-未知, 1-男, 2-女',
    `avatar`     VARCHAR(255) NOT NULL DEFAULT '' COMMENT '头像URL',
    `email`      VARCHAR(128) NOT NULL DEFAULT '' COMMENT '邮箱',
    `password`   VARCHAR(128) NOT NULL DEFAULT '' COMMENT '密码（加密存储）',
    `created_at` BIGINT NOT NULL COMMENT '创建时间（毫秒时间戳）',
    `updated_at` BIGINT NOT NULL COMMENT '更新时间（毫秒时间戳）',
    PRIMARY KEY (`user_id`),
    UNIQUE KEY `idx_mobile` (`mobile`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户表';

-- 健康检查：验证各表是否已成功创建
SELECT IF(
    EXISTS(
        SELECT 1 FROM information_schema.TABLES
        WHERE TABLE_SCHEMA = 'user_db' AND TABLE_NAME = 'user_info'
    ),
    'OK: user_info 表已存在',
    'WARN: user_info 表不存在，请确认 user-rpc 服务是否已启动并完成自动建表'
) AS 'user_db.user_info 检查';

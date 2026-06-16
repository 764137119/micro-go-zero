-- 初始化数据库（MySQL 首次启动时自动执行）
-- 已通过 MYSQL_DATABASE 环境变量自动创建的：dtm

CREATE DATABASE IF NOT EXISTS `order_db` DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE DATABASE IF NOT EXISTS `stock_db` DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE DATABASE IF NOT EXISTS `user_db`  DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- 注：各业务表结构由对应服务启动时通过 gorm/migrate 自动维护
-- 或由开发者在各自服务目录下的 deploy/sql/ 中维护独立的 DDL

-- 检查 user_db 中是否存在 user_info 表
SELECT IF(
    EXISTS(
        SELECT 1 FROM information_schema.TABLES
        WHERE TABLE_SCHEMA = 'user_db' AND TABLE_NAME = 'user_info'
    ),
    'OK: user_info 表已存在',
    'WARN: user_info 表不存在，请确认 user-rpc 服务是否已启动并完成自动建表'
) AS 'user_db.user_info 检查';

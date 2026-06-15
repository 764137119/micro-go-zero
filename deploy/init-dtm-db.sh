#!/bin/bash
set -e

# 配置数据库连接
DB_HOST="your-mysql-host"
DB_PORT="3306"
DB_USER="dtm_user"
DB_PASS="secure_password"
DB_NAME="dtm"

echo "🔧 初始化 DTM 数据库表..."

# 创建数据库（如果不存在）
mysql -h$DB_HOST -P$DB_PORT -u$DB_USER -p$DB_PASS -e "CREATE DATABASE IF NOT EXISTS \`$DB_NAME\`;"

# 执行 DTM 官方 SQL（从 GitHub 获取最新版）
curl -fsSL https://raw.githubusercontent.com/dtm-labs/dtm/main/misc/sql/mysql/dtm.sql | \
  mysql -h$DB_HOST -P$DB_PORT -u$DB_USER -p$DB_PASS $DB_NAME

echo "✅ DTM 数据库初始化完成！"
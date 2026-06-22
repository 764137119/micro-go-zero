#!/bin/bash
set -e

# 从环境变量读取（默认值与 docker-compose 一致）
DB_HOST="${MYSQL_HOST:-mysql}"
DB_PORT="${MYSQL_PORT:-3306}"
DB_USER="${MYSQL_USER:-root}"
DB_PASS="${MYSQL_PASSWORD:-root123}"
DB_NAME="${MYSQL_DATABASE:-dtm}"

echo "⏳ 等待 MySQL 就绪（$DB_HOST:$DB_PORT）..."
for i in $(seq 1 30); do
  if mysqladmin ping -h"$DB_HOST" -P"$DB_PORT" -u"$DB_USER" -p"$DB_PASS" --silent 2>/dev/null; then
    echo "✅ MySQL 已就绪"
    break
  fi
  if [ "$i" -eq 30 ]; then
    echo "❌ MySQL 未在预期时间内就绪，退出"
    exit 1
  fi
  echo "  第 $i 次尝试..."
  sleep 2
done

echo "🔧 创建 DTM 数据库（如不存在）..."
mysql -h"$DB_HOST" -P"$DB_PORT" -u"$DB_USER" -p"$DB_PASS" -e "CREATE DATABASE IF NOT EXISTS \`$DB_NAME\` CHARACTER SET utf8mb4;"

echo "🔧 初始化 DTM 表结构..."
curl -fsSL https://raw.githubusercontent.com/dtm-labs/dtm/main/misc/sql/mysql/dtm.sql 2>/dev/null | \
  mysql -h"$DB_HOST" -P"$DB_PORT" -u"$DB_USER" -p"$DB_PASS" "$DB_NAME"

echo "✅ DTM 数据库初始化完成！"
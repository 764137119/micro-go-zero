#!/bin/bash
set -e

# 构建镜像
make build

# 启动 DTM 和微服务
docker-compose -f docker-compose.prod.yml up -d

echo "🚀 所有服务已启动！"
echo "DTM Dashboard: http://localhost:36789"
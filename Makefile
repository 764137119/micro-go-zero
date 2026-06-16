# 本地测试专用 Makefile
COMPOSE_FILE = docker-compose.local.yml
PROJECT_NAME = dtm-local-test

.PHONY: build up down restart logs init-db clean

# 构建所有镜像（不推送）
build:
	@echo "🔨 本地构建镜像..."
	docker compose -f $(COMPOSE_FILE) -p $(PROJECT_NAME) build

# 启动所有服务（后台运行）
up: build
	@echo "🚀 启动本地测试环境..."
	docker compose -f $(COMPOSE_FILE) -p $(PROJECT_NAME) up -d
	@echo "✅ DTM Dashboard: http://localhost:36789"

# 停止并删除容器
down:
	docker compose -f $(COMPOSE_FILE) -p $(PROJECT_NAME) down

# 重启指定服务（例如: make restart svc=order-service）
restart:
	docker compose -f $(COMPOSE_FILE) -p $(PROJECT_NAME) up -d --no-deps --build $(svc)

# 查看日志（支持跟踪: make logs svc=dtm）
logs:
	docker compose -f $(COMPOSE_FILE) -p $(PROJECT_NAME) logs -f $(svc)

# 初始化 DTM 数据库表
init-db:
	@chmod +x deploy/init-dtm-db.sh
	@./deploy/init-dtm-db.sh

# 清理本地构建产物
clean: down
	docker compose -f $(COMPOSE_FILE) -p $(PROJECT_NAME) down -v --rmi local

commit:
	@if [ -z "$(msg)" ]; then \
		read -p "请输入提交信息: " input_msg; \
	else \
		input_msg="$(msg)"; \
	fi; \
	git add . && git commit -m "$$input_msg" && git push 
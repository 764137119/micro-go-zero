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
	git add . && git commit -m "$$input_msg" && all_proxy=socks5://127.0.0.1:56666 git push

# 项目根目录的 Makefile

GOCTL_HOME := .goctl

# 代码生成（使用项目内自定义模板 + snake_case 文件名）
.PHONY: gen gen-user gen-order gen-stock

gen: gen-user gen-order gen-stock
	@echo "✅ 所有 RPC 代码生成完成"

gen-user:
	@echo "🔨 生成 user-rpc 代码..."
	cd rpc/user-rpc && goctl rpc protoc user.proto \
		--go_out=. --go-grpc_out=. \
		--zrpc_out=. \
		--home=$(abspath $(GOCTL_HOME)) \
		--style=go_zero
	@echo "✅ user-rpc 生成完成"

gen-order:
	@echo "🔨 生成 order-rpc 代码..."
	cd rpc/order-rpc && goctl rpc protoc order.proto \
		--go_out=. --go-grpc_out=. \
		--zrpc_out=. \
		--home=$(abspath $(GOCTL_HOME)) \
		--style=go_zero
	@echo "✅ order-rpc 生成完成"

gen-stock:
	@echo "🔨 生成 stock-rpc 代码..."
	cd rpc/stock-rpc && goctl rpc protoc stock.proto \
		--go_out=. --go-grpc_out=. \
		--zrpc_out=. \
		--home=$(abspath $(GOCTL_HOME)) \
		--style=go_zero
	@echo "✅ stock-rpc 生成完成"

# 构建所有服务
build-all: user-service order-service stock-service

# 构建 user-service
user-service:
	@cd user-service && docker build -t user-service:latest .

# 构建 order-service
order-service:
	@cd order-service && docker build -t order-service:latest .

# 构建 stock-service
stock-service:
	@cd stock-service && docker build -t stock-service:latest .


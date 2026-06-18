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

gen: gen-user gen-order gen-stock gen-cronjob
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

gen-cronjob:
	@echo "🔨 生成 cronjob-rpc 代码..."
	cd rpc/cronjob-rpc && goctl rpc protoc cronjob.proto \
		--go_out=. --go-grpc_out=. \
		--zrpc_out=. \
		--home=$(abspath $(GOCTL_HOME)) \
		--style=go_zero
	@echo "✅ cronjob-rpc 生成完成"

# 构建所有服务（先本地编译验证，再构建镜像）
build-all: user-service order-service stock-service cronjob-service api-gateway

# 先本地编译 user-rpc，通过后构建镜像
user-service:
	@echo "🔨 本地编译 user-rpc..."
	cd rpc/user-rpc && go build ./...
	@echo "✅ 本地编译通过，构建 user-rpc 镜像..."
	cd rpc/user-rpc && docker build -t user-rpc:latest .

# 先本地编译 order-rpc，通过后构建镜像
order-service:
	@echo "🔨 本地编译 order-rpc..."
	cd rpc/order-rpc && go build ./...
	@echo "✅ 本地编译通过，构建 order-rpc 镜像..."
	cd rpc/order-rpc && docker build -t order-rpc:latest .

# 先本地编译 stock-rpc，通过后构建镜像
stock-service:
	@echo "🔨 本地编译 stock-rpc..."
	cd rpc/stock-rpc && go build ./...
	@echo "✅ 本地编译通过，构建 stock-rpc 镜像..."
	cd rpc/stock-rpc && docker build -t stock-rpc:latest .

# 先本地编译 cronjob-rpc，通过后构建镜像
cronjob-service:
	@echo "🔨 本地编译 cronjob-rpc..."
	cd rpc/cronjob-rpc && go build ./...
	@echo "✅ 本地编译通过，构建 cronjob-rpc 镜像..."
	cd rpc/cronjob-rpc && docker build -t cronjob-rpc:latest .

# 先本地编译 api-gateway，通过后构建镜像
api-gateway:
	@echo "🔨 本地编译 api-gateway..."
	cd api-gateway && go build ./...
	@echo "✅ 本地编译通过，构建 api-gateway 镜像..."
	cd api-gateway && docker build -t api-gateway:latest .


本地使用docker-compose 测试分布式事务
dtm+go-zero
开发新功能步骤
是的，你现在的架构思路非常清晰：

## 当前架构总结



## 新功能开发流程

**第一步：加 Gin 对外接口**（在 api-gateway 中）
- `types/` 定义请求/响应 struct
- `handler/` 写 handler 函数，调用 RPC client
- `router.go` 注册路由

**第二步：写 RPC 逻辑**（在对应 rpc 服务中）
- `xxx.proto` 定义 protobuf 协议
- `internal/logic/` 实现业务逻辑
- 注册到 server，对外暴露 gRPC 接口

## 补充说明


你的开发流程确实简化了——api-gateway 只做 HTTP 协议转 gRPC 的薄层，不涉及业务逻辑，业务全部收敛在 RPC 服务中。符合微服务的单一职责原则。

## 项目分布式事务现状分析

经过全面阅读项目代码，以下是当前分布式事务实现的完整梳理：

---

### 一、技术选型

| 项目 | 方案 |
|------|------|
| 协调器 | [DTM](https://github.com/dtm-labs/dtm) v1.15.0 |
| 事务模式 | **Saga** |
| DTM 存储后端 | MySQL（`dtm` 数据库） |
| 通信协议 | gRPC（dtmgrpc） |
| 关键依赖库 | `dtmgrpc`（api-gateway）、`dtmcli`（间接依赖） |

---

### 二、整体架构



### 三、核心流程（Saga 编排）
![整体架构图](static/image/tongyi-mermaid-2026-06-16-175329.png)

### TCC 流程

![Saga 流程图](static/image/tongyi-mermaid-2026-06-16-175339.png)

代码位于 `api-gateway/handler/order_handler.go:CreateOrder`：

1. **生成全局事务 ID**：`dtmgrpc.MustGenGid(svcCtx.Config.DTMEndpoint)`
2. **构建 Saga**：`dtmgrpc.NewSagaGrpc(DTMEndpoint, gid)`
3. **添加分支**（按顺序）：
   - **分支1**：`stock.Stock/DeductStock` → 补偿 `stock.Stock/RollbackStock`
   - **分支2**：`order.Order/CreateOrder` → 补偿 `order.Order/CancelOrder`
4. **提交事务**：`saga.Submit()`（同步等待）

**执行顺序**：先扣库存 → 再创建订单。任一失败则逆序执行补偿。

---

### 四、项目现状（完成度评估）

#### ✅ 已完成部分

| 组件 | 状态 | 说明 |
|------|------|------|
| DTM 服务部署 | ✅ | docker-compose + Dockerfile + init-dtm-db.sh |
| 数据库初始化 | ✅ | `dtm` 数据库自动创建 |
| Proto 定义 | ✅ | stock-rpc 定义了 DeductStock/RollbackStock |
| Proto 定义 | ✅ | order-rpc 定义了 CreateOrder/CancelOrder |
| Saga 编排逻辑 | ✅ | api-gateway 侧编排代码完整 |
| 库存字段设计 | ✅ | DTO 含 `totalStock`/`availableStock`/`lockedStock` 三种状态 |
| 库存流水字段 | ✅ | proto 预留了 `orderNo` 作为幂等键 |

#### ⚠️ TODO / 未实现部分

| 组件 | 状态 | 问题 |
|------|------|------|
| **stock-rpc DeductStock** | ❌ TODO | 仅注释，未实现 barrier 幂等扣减逻辑 |
| **stock-rpc RollbackStock** | ❌ TODO | 仅注释，未实现 barrier 幂等回滚逻辑 |
| **order-rpc CreateOrder** | ❌ TODO | 空实现，未写入订单表 |
| **order-rpc CancelOrder** | ❌ TODO | 空实现，未取消订单 |
| **order-rpc OrderCommitPay** | ❌ TODO | 空实现，支付提交逻辑未完成 |
| **order-rpc OrderStateCheck** | ❌ TODO | 空实现，订单状态检测未完成 |
| **dtmcli/dtmgrpc 依赖** | ❌ 缺失 | order-rpc 和 stock-rpc 的 go.mod **未引入** `dtmgrpc`/`dtmcli`，无法使用 Barrier 机制 |
| **dtm 回调可达性** | ⚠️ 未验证 | DTM 需要能回调 stock-rpc:8082 和 order-rpc:8081，docker 内网是否可达需确认 |
| **业务数据库表** | ❌ 缺失 | `order_db`/`stock_db` 数据库已创建，但业务表结构未定义 |
| **DTM 自定义配置** | ⚠️ 可能缺失 | Dockerfile 引用了 `config.yml`，但 `/dtm/` 目录下未看到该文件 |

---

### 五、潜在问题分析

1. **order-rpc 的 `stock-rpc` replace 依赖**：order-rpc 的 go.mod 中 `replace stock-rpc => ../stock-rpc`，意味着 order-rpc 依赖了 stock-rpc。但 order-rpc 的 `CreateOrder` 中是否需要调用 stock-rpc？当前代码未体现。如果不需要，建议移除该依赖。

2. **DTM 回调地址配置**：当前 Saga 分支地址通过 `svcCtx.Config.StockRpcTarget` 和 `svcCtx.Config.OrderRpcTarget` 传给 DTM。这些地址需确保 DTM 容器能从内部网络访问到。当前 `api-gateway/etc/config.yaml` 中需要配置为 docker-compose 服务名（如 `stock-rpc:8082`）。

3. **库存模型设计**：proto 中设计了 `availableStock`（可用库存）和 `lockedStock`（锁定库存）两个字段，典型的两阶段库存设计。正向操作：`availableStock--`，`lockedStock++`；补偿操作：`availableStock++`，`lockedStock--`。但当前只是设计，未实现。

---

### 六、后续需要完善的工作

1. **补齐 dtmcli/dtmgrpc 依赖**到 order-rpc 和 stock-rpc
2. **实现 stock-rpc** 的 DeductStock（带 Barrier 幂等）和 RollbackStock
3. **实现 order-rpc** 的 CreateOrder 和 CancelOrder
4. **创建业务数据库表结构**（order 表、stock 表、stock_flow_log 流水表）
5. **验证 DTM 回调网络可达性**
6. **补充 OrderCommitPay / OrderStateCheck** 完整逻辑



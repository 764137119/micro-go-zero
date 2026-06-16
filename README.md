本地使用docker-compose 测试分布式事务
dtm+go-zero
开发新功能步骤
是的，你现在的架构思路非常清晰：

## 当前架构总结

```
客户端请求 → api-gateway (Gin, :8888) → RPC 服务 (user-rpc/order-rpc) → 业务逻辑
                      ↓
              通过 etcd 服务发现
```

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

```
客户端
  │
  ▼
api-gateway（编排层）
  │  使用 dtmgrpc.NewSagaGrpc() 编排 Saga
  │  Add(正向Action, 补偿Action, 请求参数)
  │
  ├──► stock-rpc（第1分支）
  │     ├── 正向: DeductStock  (扣减库存)
  │     └── 补偿: RollbackStock (回滚库存)
  │
  └──► order-rpc（第2分支）
        ├── 正向: CreateOrder   (创建订单)
        └── 补偿: CancelOrder   (取消订单)

协调器: DTM (docker:36789 HTTP, 36790 gRPC)
注册中心: etcd (2379)
```

### 三、核心流程（Saga 编排）

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

---


flowchart TD
    A[Saga 数据安全] --> B[幂等性]
    A --> C[持久化日志]
    A --> D[隔离性兜底]
    A --> E[补偿可靠性]
    
    B --> B1[全局事务ID去重]
    B --> B2[补偿状态标记]
    
    C --> C1[WAL写入原则]
    C --> C2[协调者故障恢复]
    
    D --> D1[状态机/语义锁]
    D --> D2[乐观锁版本号]
    D --> D3[操作步骤合并]
    
    E --> E1[无限重试+退避]
    E --> E2[死信队列+人工兜底]
    E --> E3[补偿本地事务化]





sequenceDiagram
    participant Client as 用户
    participant Order as order-service<br/>(order-db)
    participant Stock as stock-service<br/>(stock-db)
    participant TC as 事务协调器

    Client->>Order: 创建订单
    Order->>TC: 开启全局事务
    TC-->>Order: xid
    
    Note over Order: 本地事务:<br/>INSERT t_order(status=待支付)
    
    Order->>Stock: Try: 预占库存(xid)
    Note over Stock: 本地事务:<br/>UPDATE t_stock SET available=available-N<br/>INSERT t_stock_occupy(xid, qty)
    Stock-->>Order: 预占成功
    
    Order->>TC: 注册分支事务
    Order-->>Client: 下单成功(待支付)
    
    Note over Client: 用户支付...
    Client->>Order: 支付回调
    Order->>TC: 提交全局事务
    TC->>Stock: Confirm: 确认扣减(xid)
    Note over Stock: 本地事务:<br/>DELETE t_stock_occupy WHERE xid=?<br/>(库存已在Try时扣过,此处仅清理预占记录)
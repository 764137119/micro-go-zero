# micro-go-zero

本地使用 docker-compose 测试分布式事务 (DTM + go-zero)

## 整体架构

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                            Client / Frontend                                     │
└─────────────────────────────────┬───────────────────────────────────────────────┘
                                  │ HTTP (JSON)
                                  ▼
┌──────────────────────────────────────────────────────────────────────────────────┐
│                              api-gateway (Port 8888)                              │
│                              Gin HTTP Framework + dtmgrpc                         │
│  ┌──────────────┐  ┌───────────────────────────────────────────────────────────┐ │
│  │  中间件层     │  │  路由层 /api/v1                                          │ │
│  │  ├─ CORS()   │  │  ┌───────────────────┐  ┌──────────────────────────────┐ │ │
│  │  └─ Auth()   │  │  │ /user/*           │  │ /order/*                    │ │ │
│  │               │  │  │ Login/Logout      │  │ CreateOrder (DTM TCC)      │ │ │
│  │               │  │  │ CreateUser/Update │  │ CommitPay/StateCheck/Cancel│ │ │
│  │               │  │  │ WxLogin/Refresh   │  │                              │ │ │
│  │               │  │  │ UserList/Info     │  │                              │ │ │
│  │               │  │  └────────┬──────────┘  └──────────────┬───────────────┘ │ │
│  └──────────────┘  └────────────┼─────────────────────────────┼─────────────────┘ │
└─────────────────────────────────┼─────────────────────────────┼───────────────────┘
                                  │                             │
                    ┌─────────────┼─────────────────────────────┼──────────────┐
                    │             │                             │              │
                    ▼             ▼                             ▼              │
         ┌─────────────────┐  ┌──────────────────┐  ┌────────────────────┐  │
         │   user-rpc      │  │   order-rpc      │  │   stock-rpc        │  │
         │   (Port 8080)   │  │   (Port 8081)    │  │   (Port 8082)     │  │
         │                 │  │                  │  │                    │  │
         │ Login/Logout    │  │ TccTryOrder      │  │ TccTryDeductStock  │  │
         │ Create/Update   │  │ TccConfirmOrder  │  │ TccConfirmDeduct   │  │
         │ UserInfo/List   │  │ TccCancelOrder   │  │ TccCancelDeduct    │  │
         │ WxLogin/Refresh │  │ CommitPay        │  │ QueryStock         │  │
         │                 │  │ StateCheck       │  │ BatchQueryStock    │  │
         └────────┬────────┘  └────────┬─────────┘  └────────┬───────────┘  │
                  │                    │                      │              │
                  ▼                    ▼                      ▼              │
         ┌──────────────────────────────────────────────────────────────┐   │
         │                        etcd                                  │   │
         │                  服务注册与发现                                │   │
         └──────────────────────────────────────────────────────────────┘   │
                                                                           │
         ┌──────────────────────────────────────────────────────────────┐   │
         │                    DTM 协调器                                 │◄──┘
         │              TCC 分布式事务管理                                │
         │           API: 36789 (HTTP) / 36790 (gRPC)                    │
         └─────────────────────────┬────────────────────────────────────┘
                                   │
                                   ▼
         ┌──────────────────────────────────────────────────────────────┐
         │                     MySQL (单实例)                           │
         │  ┌───────────┐  ┌───────────┐  ┌───────────┐  ┌─────────┐ │
         │  │ user_db   │  │ order_db  │  │ stock_db  │  │ dtm     │ │
         │  │ user_info │  │ order_info│  │ stock_info│  │(DTM自用)│ │
         │  │           │  │ order_tcc_│  │ stock_tcc_│  │         │ │
         │  │           │  │ control   │  │ control   │  │         │ │
         │  │           │  │           │  │ stock_    │  │         │ │
         │  │           │  │           │  │ flow_log  │  │         │ │
         │  └───────────┘  └───────────┘  └───────────┘  └─────────┘ │
         └──────────────────────────────────────────────────────────────┘
```

### 核心架构说明

| 层次 | 组件 | 技术选型 | 说明 |
|------|------|---------|------|
| **接入层** | api-gateway | Gin + go-zero zrpc client + dtmgrpc | HTTP API 统一入口，路由分发，编排 TCC 事务 |
| **服务层** | user-rpc / order-rpc / stock-rpc | go-zero zrpc server | 三个独立微服务，通过 gRPC 通信 |
| **协调层** | DTM | dtm-labs/dtm | TCC 分布式事务协调器 |
| **注册层** | etcd | bitnami/etcd:3.5 | 服务注册与发现 |
| **持久层** | MySQL 8.0 | GORM (ORM) | 4个业务数据库（user_db, order_db, stock_db, dtm） |

## 请求流转示例（以创建订单为例）

```
客户端 → api-gateway(POST /api/v1/order/create)
  → 生成订单号（雪花算法）
  → DTM: GenGid() 生成全局事务ID
  → api-gateway 编排 TCC 全局事务:
      Phase 1 - Try（同步执行）:
        分支1: stock-rpc.TccTryDeductStock（冻结库存: available--, locked++）
        分支2: order-rpc.TccTryOrder（创建订单，状态=待支付）
      → 全部 Try 成功 → DTM 异步调用 Confirm
      → 任一 Try 失败   → DTM 异步调用 Cancel
  → 返回结果给客户端（包含 orderId）
```

### TCC 三阶段详解

| 阶段 | stock-rpc | order-rpc |
|------|-----------|-----------|
| **Try**（资源预留） | `TccTryDeductStock`：冻结库存（available -= qty, locked += qty） | `TccTryOrder`：创建订单（状态=待支付） |
| **Confirm**（提交） | `TccConfirmDeductStock`：确认扣减（locked -= qty） | `TccConfirmOrder`：确认订单（推进状态） |
| **Cancel**（回滚） | `TccCancelDeductStock`：释放库存（available += qty, locked -= qty） | `TccCancelOrder`：取消订单（状态→已取消） |

每个阶段均通过**控制表**（`stock_tcc_control` / `order_tcc_control`）实现幂等、防悬挂、空回滚保护。

## 优缺点分析

### ✅ 优点

1. **微服务架构清晰**：按业务边界划分 user/order/stock 三个独立服务，职责单一
2. **分布式事务完善**：采用 DTM TCC 模式，具备完整的控制表幂等机制（防悬挂、空回滚）、分支事务追踪日志
3. **通用 Handler 模式**：`HandleJSON/HandleQuery` 泛型封装消除了大量模板代码
4. **基础设施完善**：etcd 服务发现 + Docker Compose 一键部署，开发体验良好
5. **数据层设计规范**：Repository 模式 + 控制表并发控制 + 完整的库存流水日志
6. **使用 go-zero**：成熟的微服务框架，内置服务治理、自适应负载均衡等能力
7. **CORS 支持**：方便前后端分离开发
8. **微信生态集成**：预留了微信小程序登录接口

## 新功能开发流程

**第一步：加 Gin 对外接口**（在 api-gateway 中）
- `types/` 定义请求/响应 struct
- `handler/` 写 handler 函数，调用 RPC client
- `router.go` 注册路由

**第二步：写 RPC 逻辑**（在对应 rpc 服务中）
- `xxx.proto` 定义 protobuf 协议
- `internal/logic/` 实现业务逻辑
- 注册到 server，对外暴露 gRPC 接口

## 分布式事务实现现状

### 一、技术选型

| 项目 | 方案 |
|------|------|
| 协调器 | [DTM](https://github.com/dtm-labs/dtm) v1.15.0 |
| 事务模式 | **TCC**（Try-Confirm-Cancel） |
| DTM 存储后端 | MySQL（`dtm` 数据库） |
| 通信协议 | gRPC（dtmgrpc） |
| 幂等机制 | 控制表（`stock_tcc_control` / `order_tcc_control`），自实现无 dtmcli 依赖 |

### 二、核心流程

TCC 全局事务编排代码位于 `api-gateway/handler/order_handler.go:CreateOrder`：

1. **生成全局事务 ID**：`dtmgrpc.MustGenGid(svcCtx.Config.DTMEndpoint)`
2. **定义 TCC 分支参数**（stockReq / orderReq）
3. **执行 TCC 全局事务**：`dtmgrpc.TccGlobalTransaction()`
   - **分支1**：`stock.Stock/TccTryDeductStock` → Confirm `TccConfirmDeductStock` / Cancel `TccCancelDeductStock`
   - **分支2**：`order.Order/TccTryOrder` → Confirm `TccConfirmOrder` / Cancel `TccCancelOrder`
4. 全部 Try 成功 → DTM 异步 Submit（回调 Confirm）
5. 任一 Try 失败 → DTM 异步 Abort（回调 Cancel）

### 三、完成度评估

| 组件 | 状态 | 说明 |
|------|------|------|
| DTM 服务部署 | ✅ | docker-compose + Dockerfile + init-dtm-db.sh |
| 数据库初始化 | ✅ | `dtm` 数据库自动创建 |
| Proto 定义 | ✅ | stock-rpc / order-rpc 完整 TCC 三阶段接口 |
| TCC 编排逻辑 | ✅ | api-gateway 侧 TccGlobalTransaction 编排完整 |
| api-gateway dtmgrpc 依赖 | ✅ | go.mod 已引入 |
| **stock-rpc TccTryDeductStock** | ✅ | 控制表幂等 + 防悬挂 + 库存流水 |
| **stock-rpc TccConfirmDeductStock** | ✅ | 控制表幂等 + 确认扣减 + 库存流水 |
| **stock-rpc TccCancelDeductStock** | ✅ | 控制表幂等 + 空回滚 + 防悬挂 + 库存流水 |
| **order-rpc TccTryOrder** | ✅ | 控制表 + 创建订单（待支付） |
| **order-rpc TccConfirmOrder** | ⚠️ 部分实现 | 控制表幂等已实现，但**未推进订单状态**（未从待支付→已支付） |
| **order-rpc TccCancelOrder** | ✅ | 控制表幂等 + 空回滚 + 取消订单 |
| **库存表（stock_info）** | ✅ | 字段完整（total/available/locked + version） |
| **库存流水表（stock_flow_log）** | ✅ | 完整记录每次库存变动 |
| **订单表（order_info）** | ✅ | 字段完整（含 xid 事务追溯） |
| **控制表（stock_tcc_control / order_tcc_control）** | ✅ | TCC 三阶段状态完整 |
| **stock-rpc go.mod dtmcli** | ⚠️ 不需要 | 使用自实现控制表，未依赖 dtmcli |
| **order-rpc go.mod dtmcli** | ⚠️ 不需要 | 使用自实现控制表，未依赖 dtmcli |
| **stock-rpc 遗留 DeductStock / RollbackStock** | ⚠️ 未使用 | 留作 Saga 接口，当前 TCC 流程不调用 |
| **order-rpc 遗留 CreateOrder / CancelOrder** | ⚠️ 空实现 | 留作 Saga 接口，当前 TCC 流程不调用 |
| **order-rpc OrderCommitPay** | ❌ 空实现 | 支付提交逻辑未完成 |
| **order-rpc OrderStateCheck** | ❌ 空实现 | 订单状态检测未完成 |
| **DTM 自定义配置** | ⚠️ 可能缺失 | Dockerfile 引用了 `config.yml`，但 `/dtm/` 目录下未看到该文件 |

### 四、控制表机制说明

项目使用**控制表**（Branch Control Table）实现 DTM TCC 的标准四防：

| 保护机制 | 说明 | 实现方式 |
|---------|------|---------|
| **幂等** | 同一分支反复调用不重复执行 | 控制表主键冲突后查询状态，已处理则直接返回成功 |
| **防悬挂** | Cancel 先于 Try 到达时，Try 必须空执行 | Try 检测到控制表状态为 CANCELLED 则放弃执行 |
| **空回滚** | Cancel 在 Try 之前到达，不需要执行业务补偿 | Cancel 插入成功（Try 未执行）直接返回成功 |
| **防重复提交** | Confirm 或 Cancel 不会重复提交 | 控制表终态检查，已终态则幂等拒绝 |

### 五、潜在问题

1. **order-rpc TccConfirmOrder 未推进订单状态**：Confirm 回调仅更新控制表状态为 CONFIRMED，但未将订单从"待支付"推进到"已支付"状态。

2. **order-rpc 的 `stock-rpc` replace 依赖**：order-rpc 的 go.mod 中 `replace stock-rpc => ../stock-rpc`，但 order-rpc 实际未调用 stock-rpc。这个依赖可以移除。

3. **DTM 回调地址配置**：TCC 分支地址通过 `svcCtx.Config.StockRpcTarget` 和 `svcCtx.Config.OrderRpcTarget` 传给 DTM。这些地址需确保 DTM 容器能从内部网络访问到。`api-gateway/etc/config.yaml` 中需配置为 docker-compose 服务名（如 `stock-rpc:8082`）。

4. **order-rpc TccCancelOrder 调用库存回滚**：TccCancelOrder 中有 `// 3.2 释放库存冻结` 的注释但未实现。理论上库存回滚应由 stock-rpc TccCancelDeductStock 完成，order-rpc 无需重复回滚。

### 六、后续需要完善的工作

1. **完善 order-rpc TccConfirmOrder**：Confirm 阶段将订单状态从"待支付(0)"推进到"已支付(1)"
2. **补充 OrderCommitPay / OrderStateCheck** 完整逻辑
3. **移除或确认 order-rpc 对 stock-rpc 的依赖**（go.mod replace）
4. **验证 DTM 回调网络可达性**（docker-compose 内网）
5. **确认 DTM config.yml 配置是否存在**

## 补充说明

api-gateway 只做 HTTP 协议转 gRPC 的薄层 + TCC 编排，不涉及业务逻辑，业务全部收敛在 RPC 服务中。符合微服务的单一职责原则。

## 关于分布式事务的思考

**当业务跨越了数据库实例边界（无论是因为公司、地区、国家，还是仅仅因为微服务拆分），无法用单个 MySQL 事务保证一致性时，才被迫采用分布式事务。它不是"升级方案"，而是在一致性与可用性之间做出妥协的无奈之选。**

### 微服务解决的是什么问题？

微服务从来不是为了技术而技术，它是 **康威定律（Conway's Law）** 的工程化体现。它解决的核心问题是：

- **单体应用随团队规模扩大，代码合并冲突呈指数级增长。**
- **发布耦合：** 改一行文案需要重启整个核心交易链路，导致发布窗口极短、风险极高。
- **技术栈锁定：** 老系统无法重构，新业务无法使用新技术，团队陷入泥潭。
- **弹性隔离失效：** 一个非核心的导出功能 OOM，拖垮了整个电商下单服务。

**一句话总结：微服务解决的是"人与组织的协作效率问题"，而不是"技术问题"。**

### RPC / 分布式事务 / 服务发现... 全是"税"

当为了解决上述组织问题而把单体拆成多个服务后，原本进程内的函数调用变成了跨网络调用，原本单库的事务变成了跨库事务。于是你被迫缴纳以下"分布式税"：

| 原本免费的东西 | 拆分后被迫引入的"税" | 本质 |
| :--- | :--- | :--- |
| `func Call()` | gRPC / Dubbo / HTTP | 跨进程通讯补偿 |
| `BEGIN...COMMIT` | TCC / Saga / 事务消息 | 跨数据源一致性补偿 |
| 内存变量共享 | Redis / 配置中心 | 跨进程状态同步补偿 |
| 堆栈追踪 | OpenTelemetry / SkyWalking | 跨进程可观测性补偿 |
| 本地单元测试 | 契约测试 / 集成测试环境 | 跨服务验证补偿 |

**RPC 的选择（gRPC vs Dubbo vs HTTP）和分布式事务的选择（TCC vs Saga），确实没有本质区别——它们都是在不同的约束条件下，挑选一种"交税姿势"而已。**

### 终极结论

> **微服务是组织架构问题的解药，却是技术复杂度的毒药。**
> **RPC、分布式事务、服务网格……全都是这剂毒药的副作用缓解剂。**
> **在决定吃这剂药之前，永远先问自己：我的组织真的病到了必须吃药的程度吗？**

## ✅ 分布式定时任务调度服务 `cronjob-rpc` 已创建完成

### 架构核心设计

```
┌───────────────────────────────────────────────────────────┐
│                   cronjob-rpc (3副本)                      │
│                                                           │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐      │
│  │  Replica-1  │  │  Replica-2  │  │  Replica-3  │      │
│  │  (Leader)   │  │ (Follower)  │  │ (Follower)  │      │
│  └──────┬──────┘  └──────┬──────┘  └──────┬──────┘      │
│         │ etcd 选主       │                │             │
│         ▼                ▼                ▼             │
│  ┌──────────────────────────────────────────┐           │
│  │              etcd 集群                    │           │
│  │  • 选主: /cronjob/leader/{key}           │           │
│  │  • 锁:  /cronjob/task/{name}/exec/{ts}   │           │
│  └──────────────────────────────────────────┘           │
│                                                           │
│  Leader 节点内部:                                          │
│  ┌─────────────┐  ┌──────────────┐  ┌────────────────┐  │
│  │  Scheduler  │─▶│   Executor   │─▶│  TaskJobRepo   │  │
│  │  (cron 库)  │  │ (etcd 锁)    │  │  (MySQL)       │  │
│  └─────────────┘  └──────────────┘  └────────────────┘  │
│                                                           │
│  gRPC Server ──▶ Register/List/Trigger/Retry/Stats       │
└───────────────────────────────────────────────────────────┘
```


### 分布式特性

| 要求 | 实现 |
|------|------|
| ✅ **etcd 选主** | `concurrency.Election` + Lease TTL=15s |
| ✅ **只执行一次** | etcd Txn 分布式锁 (CreateRevision=0) |
| ✅ **失败可重试** | 自动重试（指数退避）+ 手动 `RetryTask` API |
| ✅ **全程可观测** | `ListTasks` / `ListExecutions` / `GetTaskStats` |
| ✅ **动态管理** | `RegisterTask` / `UnregisterTask` / `SetTaskEnabled` |
| ✅ **高可用** | 3副本部署，Leader 宕机秒级切换 |
| ✅ **部署集成** | Docker Compose + Makefile |

### gRPC 接口一览

```
RegisterTask     — 注册定时任务
UnregisterTask   — 注销任务
ListTasks        — 列举所有任务
SetTaskEnabled   — 启用/禁用任务
ListExecutions   — 查询执行历史
TriggerOnce      — 手动触发一次执行
RetryTask        — 手动重试失败任务
GetTaskStats     — 获取任务统计
```
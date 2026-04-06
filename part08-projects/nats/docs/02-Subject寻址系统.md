# 02 · Subject 寻址系统

> NATS 系列指南第 2 篇 — 深入理解消息寻址的基石

---

## 目录

1. [Subject 是什么](#1-subject-是什么)
2. [Subject 语法规则](#2-subject-语法规则)
3. [通配符详解](#3-通配符详解)
4. [Subject 命名最佳实践](#4-subject-命名最佳实践)
5. [Subject 在集群中的传播机制](#5-subject-在集群中的传播机制)
6. [Special Subjects：系统保留地址](#6-special-subjects系统保留地址)
7. [Subject 与 JetStream Filter 结合使用](#7-subject-与-jetstream-filter-结合使用)
8. [ASCII 图解 Subject 层级结构](#8-ascii-图解-subject-层级结构)
9. [小结](#9-小结)

---

## 1. Subject 是什么

在 NATS 中，**Subject（主题）** 是所有通信的基础地址。它是一个纯文本字符串，发布者（Publisher）在发布消息时指定目标 Subject，服务器根据 Subject 将消息路由到所有匹配该 Subject 的订阅者（Subscriber）。

你可以把 Subject 理解为一个**信道名称**或**地址标签**，它不需要预先注册或创建，发布者和订阅者只需约定好同一个字符串即可实现通信。

```
发布者                  NATS Server               订阅者
   │                        │                        │
   │  PUB "orders.created"  │                        │
   │───────────────────────▶│                        │
   │                        │  订阅 "orders.created" │
   │                        │◀───────────────────────│
   │                        │                        │
   │                        │  推送消息到订阅者       │
   │                        │───────────────────────▶│
```

### Subject 的核心特性

| 特性 | 说明 |
|------|------|
| 轻量 | 只是一个字符串，无需服务端预创建 |
| 即时生效 | 订阅立即生效，发布立即路由 |
| 多对多 | 一个 Subject 可有任意数量的发布者和订阅者 |
| 大小写敏感 | `Orders.Created` 与 `orders.created` 是不同的 Subject |
| 无状态 | Core NATS 的 Subject 本身不持久化消息（JetStream 另说） |

---

## 2. Subject 语法规则

### 2.1 合法字符集

Subject 由若干 **token（词元）** 组成，token 之间用英文句点 `.` 分隔。每个 token 可以包含：

- **字母**：`a-z`、`A-Z`（大小写均可，但有区分）
- **数字**：`0-9`
- **连字符**：`-`（hyphen）
- **下划线**：`_`（技术上合法，但不推荐，见命名实践章节）

**不可包含的字符**：

| 字符 | 原因 |
|------|------|
| 空格 ` ` | 用于协议分隔，会截断 Subject |
| 换行 `\r\n` | NATS 协议行结束符 |
| `>` | 通配符，只能作为独立末尾 token |
| `*` | 通配符，只能作为独立 token |
| `$` 开头 | 系统保留前缀（`$JS.`、`$KV.` 等） |

### 2.2 基本结构

```
<token1>.<token2>.<token3>...<tokenN>
```

示例：

```
orders.created
orders.us-east.created
iot.factory-a.sensor.temperature
$JS.API.STREAM.CREATE
```

### 2.3 长度限制

NATS 服务器对 Subject 的字节长度没有硬性限制，但实际使用中建议：
- **不超过 256 字节**（超长 Subject 会增加路由表开销）
- **层级不超过 8 层**（过深会影响可读性和维护性）

### 2.4 大小写敏感性

这是初学者最常踩的坑之一：

```
orders.created    ≠    Orders.Created    ≠    ORDERS.CREATED
```

三者是完全不同的 Subject。建议团队统一使用**全小写 + 连字符**风格。

---

## 3. 通配符详解

NATS 提供两种通配符，专门用于**订阅端**（发布时必须使用精确 Subject）。

### 3.1 单层通配符 `*`

`*` 匹配某一个层级中的**任意单个 token**，不能跨越 `.` 分隔符。

**规则：**
- `*` 必须单独占据一个层级，即两侧都是 `.` 或位于首尾
- 可以在 Subject 的任意位置出现（包括多次）

**示例：**

```
订阅模式: orders.*.created

匹配:
  orders.us.created        ✓  （us 匹配 *）
  orders.eu.created        ✓  （eu 匹配 *）
  orders.apac.created      ✓  （apac 匹配 *）

不匹配:
  orders.created           ✗  （缺少中间层级）
  orders.us.east.created   ✗  （us.east 是两个 token，* 只匹配一个）
  orders.us.updated        ✗  （末尾不是 created）
```

多个 `*` 的使用：

```
订阅模式: iot.*.sensor.*

匹配:
  iot.factory-a.sensor.temp      ✓
  iot.factory-b.sensor.pressure  ✓

不匹配:
  iot.factory-a.actuator.motor   ✗  （第三级不是 sensor）
  iot.sensor.temp                ✗  （层级数不对）
```

### 3.2 多层通配符 `>`

`>` 匹配当前位置**及之后所有层级**的任意内容，类似"剩余全部"的含义。

**规则：**
- `>` **只能出现在 Subject 的最末尾**，作为独立 token
- 匹配一层或多层（至少一个 token）

**示例：**

```
订阅模式: orders.>

匹配:
  orders.created              ✓  （匹配 1 层）
  orders.us.created           ✓  （匹配 2 层）
  orders.us.east.created      ✓  （匹配 3 层）
  orders.us.east.2024.created ✓  （匹配 4 层）

不匹配:
  orders                      ✗  （> 至少要匹配一层）
  payments.created            ✗  （第一级不是 orders）
```

### 3.3 通配符对比表格

下面以几个典型 Subject 为例，测试不同订阅模式的匹配结果：

| 发布的 Subject | `orders.*` | `orders.>` | `orders.*.created` | `*.created` | `>` |
|---|:---:|:---:|:---:|:---:|:---:|
| `orders.created` | ✓ | ✓ | ✗ | ✓ | ✓ |
| `orders.us` | ✓ | ✓ | ✗ | ✗ | ✓ |
| `orders.us.created` | ✗ | ✓ | ✓ | ✗ | ✓ |
| `orders.us.east.created` | ✗ | ✓ | ✗ | ✗ | ✓ |
| `payments.created` | ✗ | ✗ | ✗ | ✓ | ✓ |
| `created` | ✗ | ✗ | ✗ | ✗ | ✓ |

**关键结论：**
- `>` 是"catch-all"模式，只有 Subject 本身只有一个 token 时才不匹配（比如裸 token `orders` 不匹配 `orders.>`）
- `*` 是精确的单层替换，层级数必须完全一致

### 3.4 通配符的性能考虑

通配符订阅不会比精确订阅慢很多，NATS 服务器内部使用**树形路由表（subject trie）**进行匹配，时间复杂度与 Subject 深度成线性关系而非与订阅数量成正比。不过：

- `>` 订阅会匹配非常多的消息，慎用于高吞吐场景
- 避免在同一进程中注册大量重叠的通配符订阅（增加路由计算开销）

---

## 4. Subject 命名最佳实践

### 4.1 层级设计原则

推荐遵循从**宏观到微观**的层级顺序：

```
<domain>.<service>.<action>.<entity>
```

或针对事件驱动场景：

```
<domain>.<entity>.<event>
```

**每一层的含义：**

| 层级 | 说明 | 示例 |
|------|------|------|
| domain | 业务领域或系统名称 | `orders`, `iot`, `auth` |
| service/region | 服务实例或地理分区 | `us-east`, `payment-svc` |
| entity | 操作的业务实体 | `invoice`, `user`, `sensor` |
| action/event | 动作或事件类型 | `created`, `updated`, `deleted` |

### 4.2 实际项目命名示例

**电商订单系统：**

```
orders.created                  # 新订单创建
orders.payment.completed        # 订单支付完成
orders.payment.failed           # 订单支付失败
orders.shipment.dispatched      # 发货
orders.shipment.delivered       # 签收
orders.cancelled                # 订单取消
orders.refund.requested         # 退款申请
orders.refund.approved          # 退款审批通过
```

**用户认证系统：**

```
auth.user.registered            # 用户注册
auth.user.login.success         # 登录成功
auth.user.login.failed          # 登录失败
auth.user.password.reset        # 密码重置
auth.session.expired            # 会话过期
auth.token.refreshed            # Token 刷新
```

**IoT 设备管理：**

```
iot.device.{device-id}.telemetry       # 设备遥测数据
iot.device.{device-id}.command         # 下发命令
iot.device.{device-id}.status          # 设备状态变化
iot.factory.line-a.sensor.temperature  # 工厂产线传感器
iot.factory.line-a.sensor.humidity     # 工厂产线湿度
iot.alert.critical                     # 严重告警
iot.alert.warning                      # 普通告警
```

**微服务内部通信：**

```
svc.inventory.stock.check              # 库存检查 Request
svc.inventory.stock.reserve            # 库存预占
svc.notification.email.send            # 发送邮件
svc.notification.sms.send             # 发送短信
```

### 4.3 反模式（Anti-Patterns）

**1. 不要使用下划线作为层级内分隔**

```
# 错误 - 使用下划线模拟层级
orders_payment_completed

# 正确 - 使用 . 分隔层级
orders.payment.completed
```

原因：下划线无法被通配符单独匹配，丧失了 Subject 层级系统的灵活性。

**2. 不要层级过深（超过 6-8 层）**

```
# 错误 - 层级过深，难以维护
com.company.product.region.datacenter.service.module.entity.action.version

# 正确 - 精简到关键层级
product.region.service.entity.action
```

**3. 不要第一层过于宽泛**

```
# 错误 - 顶层太宽泛，所有消息都在同一命名空间
events.orders.created
events.users.registered
events.payments.completed

# 正确 - 顶层直接用业务域
orders.created
auth.user.registered
payments.completed
```

**4. 不要在 Subject 中嵌入可变 ID 作为前缀层级**

```
# 危险 - 设备 ID 作为顶层，导致路由表膨胀
{device-id}.telemetry          # 百万设备 = 百万个不同顶层

# 正确 - 将 ID 放在靠后的层级
iot.device.{device-id}.telemetry
```

**5. 不要大小写混用**

```
# 错误 - 混用大小写，订阅者难以判断正确写法
Orders.Created
orders.Created
ORDERS.created

# 正确 - 统一全小写
orders.created
```

**6. 不要把动词变成名词（命名不一致）**

```
# 错误 - 风格不一致
orders.create       # 动词
orders.deleted      # 过去式
orders.updating     # 进行式

# 正确 - 统一使用过去式表示已发生的事件
orders.created
orders.deleted
orders.updated
```

---

## 5. Subject 在集群中的传播机制

### 5.1 订阅信息的传播

当客户端向 NATS 集群中的某个节点（Server A）订阅一个 Subject 时，该订阅信息会通过**集群内部路由协议**传播给集群中所有其他节点。

```
集群拓扑（3节点）:

Client-1                    Client-2
   │                            │
   │ SUB "orders.>"             │
   ▼                            ▼
[Server A] ─────────── [Server B] ─────────── [Server C]
               RS+                    RS+
          (订阅广播)              (订阅广播)
```

**RS+ 和 RS-：**

| 控制消息 | 含义 |
|----------|------|
| `RS+ <account> <subject>` | Route Subscribe：告知邻居节点"我这里有订阅者关注此 Subject" |
| `RS- <account> <subject>` | Route Unsubscribe：告知邻居节点"此 Subject 已无订阅者" |

### 5.2 消息路由流程

假设 Client-1 连接 Server A，Client-2 连接 Server C，Client-2 订阅了 `orders.>`：

```
步骤 1: Client-2 订阅
  Client-2 ──SUB "orders.>"──▶ Server C
  Server C 将订阅记录到本地路由表
  Server C ──RS+ orders.>──▶ Server B ──RS+ orders.>──▶ Server A

步骤 2: Client-1 发布
  Client-1 ──PUB "orders.created" msg──▶ Server A
  Server A 查路由表：orders.created 匹配 orders.>（在 Server C 有订阅者）
  Server A ──route msg──▶ Server C
  Server C ──推送 msg──▶ Client-2

步骤 3: Client-2 取消订阅
  Client-2 ──UNSUB "orders.>"──▶ Server C
  Server C ──RS- orders.>──▶ Server B ──RS- orders.>──▶ Server A
  Server A 清除路由表中的该条目
```

### 5.3 路由表的优化

NATS 集群路由表经过了专门优化：

- **兴趣聚合**：如果同一 Subject 有多个本地订阅者，对外只传播一条 RS+ 消息（计数引用）
- **账户隔离**：不同 Account 的订阅信息互不干扰（多租户隔离）
- **超级集群（SuperCluster/Gateway）**：跨地域集群通过 Gateway 传播订阅兴趣，使用**惰性传播**减少跨机房流量

---

## 6. Special Subjects：系统保留地址

NATS 系统内部使用了几类特殊 Subject，开发者需要了解以避免冲突，也可以直接利用它们。

### 6.1 `_INBOX.*`：Request/Reply 临时回复地址

当使用 NATS 的 Request/Reply 模式时，客户端库会自动生成一个唯一的 `_INBOX.<随机token>` Subject 作为回复地址，放入消息的 `Reply-To` 字段。

```
结构: _INBOX.<nuid>

示例: _INBOX.4k9fVa3Km1G8TqzH2xW7pE
```

**工作机制：**

```
请求方                   NATS Server                  响应方
   │                         │                           │
   │  订阅 _INBOX.<token>     │                           │
   │────────────────────────▶│                           │
   │                         │                           │
   │  PUB "svc.query"        │                           │
   │  Reply-To: _INBOX.<token>│                          │
   │────────────────────────▶│──────────────────────────▶│
   │                         │                           │
   │                         │  PUB _INBOX.<token> resp  │
   │                         │◀──────────────────────────│
   │◀────────────────────────│                           │
   │    收到回复               │                           │
```

新版 NATS 客户端使用**统一 Inbox（Unified Inbox）**优化，所有 request 共用一个订阅 `_INBOX.<prefix>.*`，避免每次请求都创建新订阅。

### 6.2 `$JS.API.*`：JetStream API

JetStream 的所有管理操作都通过特殊 Subject 实现：

| Subject | 用途 |
|---------|------|
| `$JS.API.INFO` | 获取 JetStream 信息 |
| `$JS.API.STREAM.CREATE.<name>` | 创建 Stream |
| `$JS.API.STREAM.INFO.<name>` | 获取 Stream 信息 |
| `$JS.API.STREAM.LIST` | 列出所有 Stream |
| `$JS.API.STREAM.DELETE.<name>` | 删除 Stream |
| `$JS.API.CONSUMER.CREATE.<stream>` | 创建 Consumer |
| `$JS.API.CONSUMER.MSG.NEXT.<stream>.<consumer>` | 拉取下一条消息 |

这些 Subject 由服务器直接处理，客户端无需手动订阅，NATS 客户端库封装了这些调用。

### 6.3 `$KV.*`：Key-Value Store

NATS JetStream 提供的 KV Store 使用如下 Subject 前缀：

| Subject 模式 | 用途 |
|--------------|------|
| `$KV.<bucket>.<key>` | KV 操作的基础 Subject |
| `$KV.<bucket>.>` | 订阅某个 bucket 的所有变化 |

例如，一个名为 `configs` 的 KV bucket，存取 key `app.timeout` 时，底层 Subject 为 `$KV.configs.app.timeout`。

### 6.4 `$SYS.*`：系统监控 Subject

NATS 服务器通过 `$SYS` 前缀发布内部监控事件：

| Subject | 说明 |
|---------|------|
| `$SYS.ACCOUNT.<id>.CONNECT` | 客户端连接事件 |
| `$SYS.ACCOUNT.<id>.DISCONNECT` | 客户端断开事件 |
| `$SYS.SERVER.STATSZ` | 服务器统计数据 |
| `$SYS.REQ.SERVER.<id>.STATSZ` | 请求特定服务器的统计 |

### 6.5 `_STAN.*`：旧版 Streaming（已废弃）

NATS Streaming（stan）使用 `_STAN.` 前缀的 Subject 进行内部通信。该服务已于 2023 年正式废弃，新项目应迁移到 JetStream。列在此处仅为说明不要与旧系统冲突。

---

## 7. Subject 与 JetStream Filter 结合使用

JetStream 在 Subject 系统之上增加了**持久化和过滤**能力。理解 Subject 是正确配置 JetStream 的前提。

### 7.1 Stream 的 Subject 绑定

创建 Stream 时，需要指定它监听哪些 Subject：

```go
// 一个 Stream 可以监听多个 Subject
js.AddStream(&nats.StreamConfig{
    Name:     "ORDERS",
    Subjects: []string{
        "orders.>",           // 捕获所有 orders 相关消息
    },
})

// 或者更精细的绑定
js.AddStream(&nats.StreamConfig{
    Name:     "ORDERS_EVENTS",
    Subjects: []string{
        "orders.created",
        "orders.updated",
        "orders.cancelled",
    },
})
```

### 7.2 Consumer 的 Subject Filter

Consumer 可以在 Stream 绑定的 Subject 范围内进一步过滤：

```
Stream "ORDERS" 监听: orders.>

Consumer A: FilterSubject = "orders.created"     → 只消费新订单
Consumer B: FilterSubject = "orders.*.failed"    → 只消费各类失败事件
Consumer C: FilterSubject = "orders.>"           → 消费所有订单事件
```

```go
// 创建只消费 orders.created 的 Consumer
js.AddConsumer("ORDERS", &nats.ConsumerConfig{
    Durable:       "order-processor",
    FilterSubject: "orders.created",
    AckPolicy:     nats.AckExplicitPolicy,
})
```

### 7.3 多 Filter Consumer（NATS 2.10+）

NATS 2.10 引入了 `FilterSubjects`（复数），允许单个 Consumer 过滤多个不相交的 Subject：

```go
js.AddConsumer("ORDERS", &nats.ConsumerConfig{
    Durable: "critical-events",
    FilterSubjects: []string{
        "orders.payment.failed",
        "orders.cancelled",
        "orders.refund.requested",
    },
})
```

### 7.4 Subject 与消息重放

JetStream 支持按 Subject 重放历史消息，这在故障恢复场景中非常有用：

```go
// 从 Stream 的起始位置重放所有 orders.created 消息
sub, _ := js.SubscribeSync("orders.created",
    nats.StartSequence(1),  // 从第一条消息开始
    nats.BindStream("ORDERS"),
)
```

---

## 8. ASCII 图解 Subject 层级结构

### 8.1 Subject 树形结构图

```
Subject 命名空间（树形视图）

root
├── orders
│   ├── created                    → "orders.created"
│   ├── cancelled                  → "orders.cancelled"
│   ├── payment
│   │   ├── completed              → "orders.payment.completed"
│   │   └── failed                 → "orders.payment.failed"
│   └── shipment
│       ├── dispatched             → "orders.shipment.dispatched"
│       └── delivered              → "orders.shipment.delivered"
├── auth
│   ├── user
│   │   ├── registered             → "auth.user.registered"
│   │   └── login
│   │       ├── success            → "auth.user.login.success"
│   │       └── failed             → "auth.user.login.failed"
│   └── session
│       └── expired                → "auth.session.expired"
└── iot
    └── device
        └── {id}
            ├── telemetry          → "iot.device.abc123.telemetry"
            ├── command            → "iot.device.abc123.command"
            └── status             → "iot.device.abc123.status"
```

### 8.2 通配符匹配范围图

```
Subject 空间可视化（以 orders.* 为例）

orders
├── created         ← 被 "orders.*" 匹配
├── cancelled       ← 被 "orders.*" 匹配
├── payment
│   ├── completed   ← 不被 "orders.*" 匹配（层级过深）
│   └── failed      ← 不被 "orders.*" 匹配（层级过深）
└── shipment
    ├── dispatched  ← 不被 "orders.*" 匹配（层级过深）
    └── delivered   ← 不被 "orders.*" 匹配（层级过深）

使用 "orders.>" 则匹配以上全部（包括深层）

使用 "orders.*.completed" 则只匹配:
  orders.payment.completed   ✓
  orders.shipment.completed  ✓（如果存在的话）
```

### 8.3 集群中 Subject 路由图

```
多节点集群的 Subject 路由

                    ┌─────────────────────────────────────────┐
                    │           NATS Cluster                  │
                    │                                         │
  Publisher         │  ┌──────────┐    ┌──────────┐          │  Subscriber
  (Client A)        │  │ Server 1 │    │ Server 2 │          │  (Client B)
     │              │  │          │    │          │          │     │
     │  PUB         │  │  Route   │    │  Route   │          │  SUB│
     │  "orders.    │  │  Table:  │◄──►│  Table:  │          │  "orders.>"
     │   created"   │  │          │    │          │          │     │
     └─────────────►│  │ orders.> │    │ orders.> │◄─────────┘     │
                    │  │ → Srv2   │    │ → local  │                │
                    │  └──────────┘    └──────────┘                │
                    │       │               │                      │
                    │       │  route msg    │  deliver msg         │
                    │       └──────────────►└─────────────────────►│
                    │                                         │
                    └─────────────────────────────────────────┘
```

---

## 9. 小结

| 知识点 | 核心要点 |
|--------|----------|
| Subject 基础 | 纯文本字符串，`.` 分隔层级，大小写敏感，无需预创建 |
| `*` 通配符 | 匹配单层，可多次使用，层级数必须严格一致 |
| `>` 通配符 | 匹配剩余所有层级，只能在末尾，至少匹配一层 |
| 命名规范 | domain.service.entity.action，全小写，不超过 6-8 层 |
| 集群传播 | RS+/RS- 控制消息在节点间广播订阅兴趣 |
| 系统 Subject | `$JS.`、`$KV.`、`$SYS.`、`_INBOX.` 为系统保留，了解不冲突 |
| JetStream 集成 | Stream 绑定 Subject，Consumer 可进一步 Filter，支持多 Filter |

Subject 系统是 NATS 中最简单却也最重要的概念，掌握它是高效使用 NATS 所有功能的前提。下一篇将介绍基于 Subject 的 **Pub/Sub 发布订阅模式**，深入理解消息如何在发布者和订阅者之间流动。

---

*下一篇：[03 · Pub/Sub 发布订阅](./03-发布订阅模式.md)*
*上一篇：[01 · NATS 核心概念与架构](./01-core-concepts.md)*

---

## 10. Subject 设计决策树

当设计 Subject 命名规范时，可以参考以下决策流程：

```
开始设计 Subject
       │
       ▼
┌─────────────────────────────────────┐
│ 1. 确定顶层命名空间                   │
│    - 使用业务域或系统名               │
│    - 例: orders, iot, auth          │
└─────────────────────────────────────┘
       │
       ▼
┌─────────────────────────────────────┐
│ 2. 确定消息类型                       │
│    - 事件（过去式）: created, updated │
│    - 命令（动词）: create, update    │
│    - 查询（名词）: get, list         │
└─────────────────────────────────────┘
       │
       ▼
┌─────────────────────────────────────┐
│ 3. 是否需要按实体 ID 分区？           │
│    - 是: 将 ID 放在中间层级          │
│    - 例: device.{id}.command        │
│    - 否: 使用通配符订阅              │
└─────────────────────────────────────┘
       │
       ▼
┌─────────────────────────────────────┐
│ 4. 是否需要多租户隔离？               │
│    - 是: 添加租户层级                │
│    - 例: {tenant}.orders.created    │
│    - 或使用 NATS Account 隔离       │
└─────────────────────────────────────┘
       │
       ▼
┌─────────────────────────────────────┐
│ 5. 验证设计                          │
│    - 层级是否合理（< 8 层）？         │
│    - 通配符订阅是否方便？             │
│    - 是否与现有 Subject 冲突？        │
└─────────────────────────────────────┘
```

---

## 11. 行业命名规范参考

### 11.1 电商系统

```
# 订单域
orders.created                    # 订单创建事件
orders.updated                    # 订单更新事件
orders.cancelled                  # 订单取消事件
orders.paid                       # 订单支付完成
orders.shipped                    # 订单发货
orders.delivered                  # 订单签收
orders.refund.requested           # 退款申请
orders.refund.approved            # 退款审批通过
orders.refund.rejected            # 退款审批拒绝

# 库存域
inventory.reserved                # 库存预占
inventory.released                # 库存释放
inventory.stock.low               # 库存预警
inventory.stock.out               # 库存缺货

# 支付域
payment.initiated                 # 支付发起
payment.completed                 # 支付完成
payment.failed                    # 支付失败
payment.refunded                  # 支付退款
```

### 11.2 IoT 平台

```
# 设备管理
device.{device-id}.connected      # 设备上线
device.{device-id}.disconnected   # 设备下线
device.{device-id}.heartbeat      # 设备心跳
device.{device-id}.status         # 设备状态变化

# 遥测数据
telemetry.{device-id}.temperature # 温度数据
telemetry.{device-id}.humidity    # 湿度数据
telemetry.{device-id}.location    # 位置数据
telemetry.{device-id}.custom.{metric-name}  # 自定义指标

# 命令下发
command.{device-id}.reboot        # 重启命令
command.{device-id}.config        # 配置下发
command.{device-id}.firmware      # 固件升级
command.{device-id}.rpc           # RPC 调用

# 告警
alert.{device-id}.critical        # 严重告警
alert.{device-id}.warning         # 警告
alert.{device-id}.info            # 信息
```

### 11.3 微服务通信

```
# 服务发现
service.discovery.register        # 服务注册
service.discovery.deregister      # 服务注销
service.discovery.heartbeat       # 服务心跳

# 服务间 RPC
rpc.{service}.{method}            # RPC 调用
rpc.user.get                      # 用户服务 - 获取用户
rpc.order.create                  # 订单服务 - 创建订单
rpc.payment.process               # 支付服务 - 处理支付

# 事件总线
event.{domain}.{entity}.{action}  # 领域事件
event.user.created                # 用户创建
event.order.completed             # 订单完成
event.notification.sent           # 通知发送
```

### 11.4 游戏服务

```
# 玩家事件
player.{player-id}.login          # 玩家登录
player.{player-id}.logout         # 玩家登出
player.{player-id}.levelup        # 玩家升级
player.{player-id}.achievement    # 成就达成

# 房间/匹配
room.{room-id}.created            # 房间创建
room.{room-id}.joined             # 玩家加入
room.{room-id}.left               # 玩家离开
room.{room-id}.started            # 游戏开始
room.{room-id}.ended              # 游戏结束

# 实时同步
sync.{room-id}.position           # 位置同步
sync.{room-id}.action             # 动作同步
sync.{room-id}.chat               # 聊天消息
```

---

## 12. Subject 与微服务架构

### 12.1 事件驱动架构（EDA）

```
                    ┌─────────────────────────────────────────┐
                    │              NATS Event Bus              │
                    │                                         │
订单服务            │  orders.created   payment.completed     │
   │                │       │                  │              │
   │── PUB orders.created ──▶│                  │              │
   │                │       │                  │              │
   │                │       ▼                  │              │
   │                │  ┌─────────┐             │              │
   │                │  │ 库存服务 │             │              │
   │                │  │ SUB     │             │              │
   │                │  └─────────┘             │              │
   │                │                          │              │
   │                │                          ▼              │
   │                │                     ┌─────────┐        │
   │                │                     │ 通知服务 │        │
   │                │                     │ SUB     │        │
   │                │                     └─────────┘        │
                    │                                         │
                    └─────────────────────────────────────────┘
```

### 12.2 CQRS 模式

```
Command Side                    Query Side
    │                               │
    │  PUB orders.create            │
    │──────────────────▶            │
    │                               │
    │  NATS 处理命令                 │
    │  更新写模型                     │
    │  PUB orders.created           │
    │──────────────────▶            │
    │                               │
    │                          SUB orders.created
    │                               │
    │                          更新读模型
    │                               │
    │                          查询服务读取
    │                               │
```

### 12.3 Saga 编排模式

```
订单创建 Saga 流程：

1. 订单服务发布: orders.create
2. 库存服务订阅: orders.create → 预占库存 → PUB inventory.reserved
3. 支付服务订阅: inventory.reserved → 处理支付 → PUB payment.completed
4. 订单服务订阅: payment.completed → 确认订单 → PUB orders.confirmed

补偿流程（支付失败）：
1. 支付服务发布: payment.failed
2. 库存服务订阅: payment.failed → 释放库存 → PUB inventory.released
3. 订单服务订阅: payment.failed → 取消订单 → PUB orders.cancelled
```

---

## 13. Subject 性能考量

### 13.1 路由表大小

```
Subject 数量对路由表的影响：

少量 Subject（< 1000）：
  - 路由表完全在内存中
  - 匹配延迟 < 1µs
  - 无性能问题

大量 Subject（> 100,000）：
  - 路由表占用内存增加
  - 建议使用通配符减少订阅数
  - 避免为每个实体 ID 创建独立订阅

错误示例：
  # 百万设备，每个设备一个订阅
  for _, deviceID := range deviceIDs {
      nc.Subscribe(fmt.Sprintf("device.%s.data", deviceID), handler)
  }
  # 结果：百万订阅，内存爆炸

正确示例：
  # 使用通配符订阅，一个订阅覆盖所有设备
  nc.Subscribe("device.*.data", handler)
  # 在 handler 中通过 msg.Subject 解析设备 ID
```

### 13.2 通配符匹配性能

```
NATS 使用前缀树（Trie）进行 Subject 匹配：

精确匹配：O(L)，L = Subject 长度
通配符匹配：O(L)，与订阅数量无关

性能对比（100 万订阅）：
  精确 Subject 订阅：匹配延迟 ~1µs
  通配符订阅：匹配延迟 ~1-2µs

结论：通配符不会显著影响性能，可以放心使用
```

---

## 14. 常见问题

### Q1: Subject 中可以使用中文吗？

**不建议**。虽然技术上可行，但可能导致编码问题、日志乱码、跨语言兼容性问题。建议使用英文和 ASCII 字符。

### Q2: Subject 区分大小写吗？

**区分**。`Orders.Created` 和 `orders.created` 是不同的 Subject。建议统一使用小写。

### Q3: Subject 最大长度是多少？

NATS 没有硬性限制，但建议不超过 256 字节。过长的 Subject 会增加网络开销和路由表内存占用。

### Q4: 可以动态创建 Subject 吗？

**可以**。Subject 不需要预先定义，发布时直接使用即可。但建议在项目初期规划好命名规范。

### Q5: 如何处理 Subject 冲突？

使用命名空间前缀或 NATS Account 隔离。例如：
- `service-a.events.created`
- `service-b.events.created`
- 或使用不同的 Account 完全隔离

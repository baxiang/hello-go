# 01 - NATS 是什么

> **系列导航**: [← 返回目录](./README.md) | [下一篇：Subject 寻址系统 →](./02-Subject寻址系统.md)

---

## 目录

1. [起源与历史](#1-起源与历史)
2. [设计哲学](#2-设计哲学)
3. [核心特性详解](#3-核心特性详解)
4. [NATS 解决什么问题](#4-nats-解决什么问题)
5. [四层架构全景图](#5-四层架构全景图)
6. [与 Kafka、RabbitMQ 的对比](#6-与-kafka-rabbitmq-的对比)
7. [适用场景 vs 不适用场景](#7-适用场景-vs-不适用场景)
8. [单二进制架构优势](#8-单二进制架构优势)
9. [官方性能基准数据](#9-官方性能基准数据)
10. [小结](#10-小结)

---

## 1. 起源与历史

### 诞生背景

NATS 由 **Derek Collison** 于 **2010 年**创建，最初作为 **Cloud Foundry**（VMware 的 PaaS 平台）的内部消息总线使用。Derek 是 Google 任期内 TIBCO 消息系统的资深架构师，他深刻理解企业级消息中间件的复杂性与痛点，在设计 NATS 时刻意走向了另一个极端：**极致简单**。

```
时间线
──────
2010  Derek Collison 在 VMware/Cloud Foundry 项目中创建 NATS
      最初用 Ruby 实现（ruby-nats），作为 CF 内部组件通信总线

2012  用 Go 语言重写，性能大幅提升
      开源发布：github.com/nats-io/gnatsd

2016  成立 Synadia Communications 公司，专职维护 NATS
      发布 NATS Streaming（早期持久化方案，已被 JetStream 取代）

2020  NATS 2.0 发布
      引入去中心化安全模型（Decentralized Security / NKEY + JWT）
      引入 Account 隔离机制

2021  NATS 2.2 发布：JetStream 正式 GA（General Availability）
      这是 NATS 历史上最重要的里程碑之一
      JetStream 内置于 nats-server 主程序，零额外依赖

2023  NATS 2.10 发布
      JetStream 稳定性与性能大幅提升
      新增 Consumer filter subjects 等高级特性
```

### 名字的含义

NATS 是 **Neural Autonomic Transport System** 的缩写，灵感来源于神经系统中的自主信号传输机制 —— 快速、去中心化、不依赖中央协调者。

### 项目规模（2024 年数据）

- GitHub Stars：15,000+（nats-server）
- CNCF（Cloud Native Computing Foundation）毕业项目（Graduated，最高级别）
- 被 Apple、Mercedes-Benz、Netlify、Chick-fil-A、HTC 等公司用于生产环境

---

## 2. 设计哲学

NATS 的设计遵循三个核心原则，这三个原则相互制约，共同塑造了 NATS 的一切设计决策。

### 原则一：简单（Simplicity First）

NATS 的协议是文本协议，极其简单：

```
# 发布消息
PUB foo 11\r\n
hello world\r\n

# 订阅主题
SUB foo 1\r\n

# 收到消息
MSG foo 1 11\r\n
hello world\r\n
```

就这些。整个 Core NATS 协议只有不到 10 条命令（`CONNECT`、`PUB`、`SUB`、`UNSUB`、`MSG`、`PING`、`PONG`、`+OK`、`-ERR`）。

**复杂性的代价**：RabbitMQ 的 AMQP 协议有 Exchange、Queue、Binding、Virtual Host、Channel 等多个抽象层次。学习成本高，运维复杂，出问题时排查困难。

NATS 的理念：**让 90% 的场景用 10% 的复杂度解决**。

### 原则二：高速（Performance Without Compromise）

NATS 的设计目标从一开始就是 **sub-millisecond latency** 和 **multi-million messages/second throughput**：

- 服务器内部使用无锁数据结构（lock-free subscription routing）
- 消息在内存中流转，不做持久化（Core NATS 层面）
- 协议解析开销极低（简单文本协议 + 高效 Go 实现）
- 连接数上限极高（单节点支持 100 万+ 并发连接）

### 原则三：始终在线（Always On）

NATS 服务器被设计为 **永不停止服务消息路由**。即使出现：

- 某个客户端连接断开
- 某个 Subscriber 崩溃
- 集群中某个节点不可达

服务器本身不会崩溃，不会拒绝服务其他客户端。这是 **At-Most-Once（最多一次）** 投递语义的代价换来的稳健性。

### At Most Once vs At Least Once

这是理解 NATS 的核心二分法：

```
Core NATS 层（At Most Once）
┌─────────────────────────────────────────────────────────────────┐
│ 发布者发送消息后，不关心是否有订阅者、是否被成功收到             │
│                                                                  │
│ Publisher ──PUB──▶ NATS Server ──MSG──▶ Subscriber（如果在线） │
│                         │                                        │
│                         └──▶ 消息丢弃（如果无人订阅/订阅者离线）│
└─────────────────────────────────────────────────────────────────┘

JetStream 层（At Least Once）
┌─────────────────────────────────────────────────────────────────┐
│ 消息写入持久化日志，Consumer 消费后需要显式 Ack                  │
│ 未 Ack 的消息会按策略重投                                        │
│                                                                  │
│ Publisher ──PUB──▶ NATS Server ──写入 Stream──▶ Consumer        │
│                         │                        │              │
│                         └──持久化存储             └──Ack/Nak    │
└─────────────────────────────────────────────────────────────────┘
```

**选择哪种语义**取决于业务需求：
- 实时遥测数据（丢几个点无关紧要）→ Core NATS，At Most Once
- 支付指令、订单创建 → JetStream，At Least Once
- 配置下发（需要幂等）→ JetStream KV，Exactly Once（通过 CAS）

---

## 3. 核心特性详解

### 3.1 Subject-Based Messaging（基于主题的消息路由）

NATS 中所有消息都以 **Subject（主题字符串）** 作为地址，而不是 Queue 名称或 Topic ID。

```
发布者                NATS Server             订阅者
   │                      │                      │
   │─── PUB sensor.room1.temp ──────────────────▶│ SUB sensor.*.temp
   │                      │                      │
   │                      │──── MSG ─────────────▶│
```

Subject 的灵活性远超传统队列名称：通配符 `*` 和 `>` 允许订阅者用一个订阅覆盖海量 Subject，详见 [02-Subject寻址系统.md](./02-Subject寻址系统.md)。

### 3.2 三种通信模式

```
模式一：Publish/Subscribe（广播）
┌──────────┐    PUB news.sports    ┌──────────┐
│Publisher │ ─────────────────────▶│ Server   │
└──────────┘                       └──────────┘
                                        │ MSG（fan-out）
                              ┌─────────┼─────────┐
                              ▼         ▼         ▼
                          Sub A       Sub B     Sub C
                       (all receive the same message)

模式二：Request/Reply（点对点 RPC）
┌──────────┐  REQ to "service.add"  ┌──────────┐
│ Requester│ ──────────────────────▶│  Server  │
└──────────┘                        └──────────┘
      ▲                                   │
      │                                   ▼
      │           reply to _INBOX.xxx  ┌──────────┐
      └────────────────────────────────│ Replier  │
                                       └──────────┘

模式三：Queue Subscribe（负载均衡）
┌──────────┐    PUB work.task      ┌──────────┐
│Publisher │ ─────────────────────▶│  Server  │
└──────────┘                       └──────────┘
                                        │ MSG（只发给一个）
                              ┌─────────┴─────────┐
                              ▼                   (待机)
                          Worker A              Worker B
                       (queue group "workers")
```

### 3.3 JetStream（持久化引擎）

JetStream 是 NATS 2.2+ 内置的持久化层，直接构建于 nats-server 内部：

- **Stream**：持久化消息日志，类似 Kafka 的 Partition Log
- **Consumer**：读取 Stream 的游标，支持 Push 和 Pull 两种模式
- **KV Store**：基于 Stream 的键值存储，支持版本历史和 Watch
- **Object Store**：基于 Stream 的大文件分块存储

### 3.4 集群与高可用

- **Cluster**：多个 nats-server 节点组成集群，通过 Route 协议互联
- **JetStream Cluster**：基于 Raft 算法实现 Stream 副本和 Leader 选举
- **Leaf Node**：轻量级边缘节点，通过单连接接入中心集群
- **Gateway**：连接多个独立集群，构成 Super Cluster

### 3.5 安全模型（NATS 2.0+）

```
传统中心化安全（大多数消息系统）：
  用户名/密码 → 中央 ACL 数据库 → 验证通过

NATS 去中心化安全（NKEY + JWT）：
  Operator（根证书）
    └── Account（租户）
          ├── User（JWT Token，含 ACL 策略）
          └── 私钥签名，服务器离线验证
```

不需要中央认证服务，JWT 本身携带权限声明，服务器用公钥验证签名即可。

---

## 4. NATS 解决什么问题

### 传统消息队列的痛点

#### 痛点一：运维复杂度高

以 Kafka 为例：

```
Kafka 生产环境最小部署：

  ZooKeeper Ensemble（3节点）
       ↑ 管理元数据
  Kafka Broker（3节点）
       ↑ 存储消息
  Schema Registry（可选）
       ↑ 管理消息格式
  Kafka Connect（可选）
       ↑ 数据集成

  至少需要 6 个进程 + JVM 调优 + ZooKeeper 调优
  （KRaft 模式虽然去掉了 ZooKeeper，但 Kafka 本身依然复杂）
```

NATS 的做法：

```
NATS 生产环境最小部署：

  nats-server（3节点集群）

  就这些。单二进制，零外部依赖，Go 原生运行。
```

#### 痛点二：延迟不够低

Kafka 的设计目标是高吞吐量批量处理，p99 延迟通常在几十到几百毫秒量级。对于需要实时响应的场景（如微服务间 RPC、IoT 实时控制）并不合适。

NATS Core 的 p99 延迟在 **< 1ms**（局域网内），适合低延迟场景。

#### 痛点三：扩展方式不灵活

RabbitMQ 的横向扩展（Clustering、Federation、Shovel）配置复杂，网络拓扑固定。

NATS 提供三种组网方式（Cluster / Leaf Node / Gateway），可以根据网络拓扑灵活选择。

#### 痛点四：协议/SDK 重

AMQP（RabbitMQ）、Kafka Wire Protocol 都是复杂的二进制协议，实现一个新语言的客户端成本极高。

NATS 协议极简，实现 Core NATS 客户端只需几百行代码，官方维护了 40+ 语言的客户端。

### NATS 填补的空白

```
消息系统能力矩阵

                    低延迟实时    持久化可靠    协议简单    运维轻量
                    ──────────   ──────────   ────────   ────────
Kafka              ○（批量）     ●            ○           ○
RabbitMQ           ●            ●            ○           ○（单节点）
Redis Pub/Sub      ●            ✗            ●           ●
NATS Core          ●            ✗（AtMost1） ●           ●
NATS + JetStream   ●            ●            ●           ●

● = 擅长  ○ = 一般  ✗ = 不支持/不适合
```

NATS + JetStream 是目前少数能同时在**低延迟**和**持久化可靠**两个维度都做得很好的系统。

---

## 5. 四层架构全景图

NATS 的能力被分为四个逐层递进的架构层次：

```
╔══════════════════════════════════════════════════════════════════════╗
║                    NATS 四层架构全景                                  ║
╠══════════════════════════════════════════════════════════════════════╣
║                                                                      ║
║  Layer 4: 网络拓扑层（Multi-Cluster Networking）                      ║
║  ┌────────────────────────────────────────────────────────────────┐  ║
║  │  Super Cluster（超级集群）                                      │  ║
║  │  ┌──────────────┐  Gateway  ┌──────────────┐                  │  ║
║  │  │  Cluster A   │◀─────────▶│  Cluster B   │  跨地域连接      │  ║
║  │  │  (us-east)   │           │  (eu-west)   │                  │  ║
║  │  └──────────────┘           └──────────────┘                  │  ║
║  │         ▲                                                      │  ║
║  │    Leaf Node（边缘接入，如 IoT 网关、边缘 DC）                   │  ║
║  └────────────────────────────────────────────────────────────────┘  ║
║                                                                      ║
║  Layer 3: 集群层（Horizontal Scaling）                                ║
║  ┌────────────────────────────────────────────────────────────────┐  ║
║  │  NATS Cluster（Route Protocol）                                │  ║
║  │                                                                │  ║
║  │  ┌─────────┐    Route    ┌─────────┐    Route    ┌─────────┐  │  ║
║  │  │ nats-1  │◀──────────▶│ nats-2  │◀──────────▶│ nats-3  │  │  ║
║  │  └─────────┘            └─────────┘            └─────────┘  │  ║
║  │                                                                │  ║
║  │  • 订阅兴趣在节点间同步传播                                      │  ║
║  │  • 消息按 Subject 兴趣路由到正确节点                             │  ║
║  │  • JetStream 在集群间通过 Raft 选举 Stream Leader               │  ║
║  └────────────────────────────────────────────────────────────────┘  ║
║                                                                      ║
║  Layer 2: JetStream 层（Persistence & HA）                           ║
║  ┌────────────────────────────────────────────────────────────────┐  ║
║  │  JetStream（内置于 nats-server，无需独立进程）                   │  ║
║  │                                                                │  ║
║  │  ┌──────────────┐  ┌──────────────┐  ┌──────────────────────┐ │  ║
║  │  │   Streams    │  │  Consumers   │  │  KV / Object Store   │ │  ║
║  │  │  持久化日志   │  │  消费游标    │  │  高级存储原语        │ │  ║
║  │  │  AtLeastOnce │  │  Push / Pull │  │  基于 Stream 实现    │ │  ║
║  │  └──────────────┘  └──────────────┘  └──────────────────────┘ │  ║
║  └────────────────────────────────────────────────────────────────┘  ║
║                                                                      ║
║  Layer 1: Core NATS 层（基础消息路由）                                ║
║  ┌────────────────────────────────────────────────────────────────┐  ║
║  │  Core NATS（纯内存，At Most Once）                              │  ║
║  │                                                                │  ║
║  │  ┌──────────────┐  ┌──────────────┐  ┌──────────────────────┐ │  ║
║  │  │  Pub / Sub   │  │    Request   │  │   Queue Subscribe    │ │  ║
║  │  │  发布订阅     │  │   / Reply    │  │    队列订阅（LB）     │ │  ║
║  │  │  Fan-out     │  │   同步 RPC   │  │   负载均衡投递        │ │  ║
║  │  └──────────────┘  └──────────────┘  └──────────────────────┘ │  ║
║  └────────────────────────────────────────────────────────────────┘  ║
║                                                                      ║
╚══════════════════════════════════════════════════════════════════════╝
```

### 各层的独立性

这四层是**累加而非替代**的关系：

- 只需要 **实时广播/RPC** → 只用 Layer 1（Core NATS）
- 需要 **消息持久化** → 启用 JetStream（Layer 2）
- 需要 **水平扩展** → 组建 Cluster（Layer 3）
- 需要 **跨数据中心** → 配置 Gateway/Leaf Node（Layer 4）

每一层的启用都是配置文件中的一个开关，不需要部署额外的进程。

---

## 6. 与 Kafka、RabbitMQ 的对比

> 重要前提：**NATS 不是 Kafka 或 RabbitMQ 的替代品，它们面向不同的核心场景**。

### 定位对比

```
三种系统的核心定位：

Kafka:      高吞吐量事件流处理平台
            核心场景：日志聚合、流式计算、数据管道、事件溯源
            核心优势：超高吞吐（百万级 msg/s + 批量压缩）、长期消息保留

RabbitMQ:   传统企业消息队列
            核心场景：任务队列、工作流、AMQP 互操作性
            核心优势：灵活路由（Exchange/Binding）、协议生态丰富

NATS:       云原生连接层（Connective Tissue）
            核心场景：微服务通信、IoT、实时数据分发、边缘计算
            核心优势：极低延迟、极简运维、灵活组网
```

### 详细对比表

| 维度 | NATS + JetStream | Kafka | RabbitMQ |
|------|-----------------|-------|----------|
| **消息延迟** | < 1ms (p99) | 数十 ms（批量优化） | < 5ms |
| **吞吐量** | 18M msg/s（单节点） | 极高（批量压缩） | 中等 |
| **消息持久化** | JetStream（可选） | 默认持久化 | 可选（持久化队列） |
| **消息回溯** | JetStream Consumer（支持） | Consumer Group Offset | 不支持 |
| **依赖项** | 零依赖（单二进制） | JVM + ZooKeeper/KRaft | Erlang VM |
| **协议** | 自研简单文本协议 | Kafka Wire Protocol | AMQP 0-9-1 |
| **部署复杂度** | 极低 | 高 | 中等 |
| **多租户** | Account 隔离 | 无原生支持 | Virtual Host |
| **跨数据中心** | Gateway + Leaf Node | MirrorMaker 2 | Federation |
| **消息顺序** | Stream 内有序 | Partition 内有序 | Queue 内有序 |
| **流式处理** | 有限支持 | Kafka Streams / Flink | 不适合 |
| **学习曲线** | 低 | 高 | 中等 |

### 什么时候选 Kafka？

- 需要**超大规模日志持久化**（TB 级别，长期保留）
- 需要与 **Kafka Streams / Flink** 集成做流式计算
- 已有大量 Kafka 生态工具（Debezium、Kafka Connect 等）
- 团队已熟悉 Kafka 运维

### 什么时候选 RabbitMQ？

- 遗留系统需要 **AMQP 协议兼容**
- 需要**复杂的消息路由逻辑**（Topic Exchange、Headers Exchange）
- 需要 **STOMP、MQTT** 等多协议支持

### 什么时候选 NATS？

- **微服务间通信**（替代 gRPC 或 HTTP，更适合内部总线）
- **IoT 场景**（设备数量大，连接数多，延迟要求低）
- **实时数据分发**（行情数据、监控遥测）
- **边缘计算**（Leaf Node 接入边缘节点）
- **希望简化运维**（不想维护 Kafka 集群的复杂性）
- **事件驱动微服务**（需要持久化但吞吐量不是极端量级）

---

## 7. 适用场景 vs 不适用场景

### 适用场景

```
✓ 微服务内部通信总线
  服务 A、B、C 通过 NATS Subject 解耦，不需要两两建立 gRPC 连接

✓ IoT 设备遥测数据上报
  数万台设备同时连接，每台每秒上报传感器数据
  NATS 单节点支持 100 万+ 连接

✓ 实时命令下发（如本项目 livis-claw-hub）
  客户端发 Request，设备通过 Reply 流式返回结果
  Core NATS At-Most-Once 语义 + 低延迟，完全满足实时控制场景

✓ 配置中心 / Feature Flag（用 KV Store）
  KV Store 支持 Watch，配置变更实时推送给所有服务实例
  不需要额外部署 etcd 或 Consul

✓ 事件驱动微服务（用 JetStream）
  订单创建事件 → JetStream Stream → 库存服务 / 通知服务各自消费
  At-Least-Once，支持消费者重启后继续消费

✓ 多数据中心同步
  通过 Gateway 连接两个地域的 NATS 集群
  本地 Subject 消息自动路由到对端集群
```

### 不适用场景

```
✗ 超大规模日志持久化（>TB 级别）
  JetStream 的设计目标不是替代 Kafka 做日志存储平台
  极端吞吐量场景（百万 msg/s 持久化）建议用 Kafka

✗ 需要精确 Exactly-Once 语义的金融交易
  JetStream 的 Exactly-Once 依赖客户端 MsgID 幂等，
  实现不如 Kafka Transactions 那么完整

✗ 需要复杂流式计算（聚合、窗口函数）
  NATS 没有内置的流处理引擎
  这类需求用 Kafka + Flink/Kafka Streams 更合适

✗ 遗留系统 AMQP/STOMP 协议互操作
  NATS 协议与 AMQP 不兼容，迁移成本高

✗ 需要消息优先级队列
  NATS 不支持消息优先级，所有消息按 FIFO 处理
```

---

## 8. 单二进制架构优势

这是 NATS 在运维层面最被低估的优势之一。

### 什么是单二进制

```
下载 nats-server 二进制文件（约 20MB）

nats-server 包含：
  ├── Core NATS 路由引擎
  ├── JetStream 持久化引擎（内置 BoltDB / 文件存储）
  ├── HTTP 监控 API（:8222）
  ├── Cluster Route 协议处理
  ├── Leaf Node / Gateway 协议处理
  ├── TLS 支持（内置，无需额外配置 nginx）
  └── NKEY / JWT 鉴权引擎

零外部依赖：不需要 JVM、不需要 Erlang、不需要 Python
```

### 对比的启示

```
Kafka 集群的最小生产部署（传统 ZooKeeper 模式）：
  3x ZooKeeper 节点（JVM 进程，需要独立运维）
  3x Kafka Broker 节点（JVM 进程）
  1x Schema Registry（可选，JVM 进程）
  1x Kafka Connect（可选，JVM 进程）
  = 至少 6 个进程，JVM 调优，ZooKeeper 调优

NATS 集群的最小生产部署：
  3x nats-server 节点（Go 单二进制）
  = 3 个进程，无需额外调优
```

### 资源占用极低

NATS Server 的典型资源消耗：

```
空载状态（无连接）：
  内存占用：~10MB
  CPU：<0.1%

中等负载（10 万连接，10 万 msg/s）：
  内存占用：约 1-2GB（主要是连接缓冲区）
  CPU：约 20-40%（4 核）

对比：空载 Kafka Broker 通常需要 1-2GB JVM Heap
```

### Docker 镜像大小

```
nats:latest Docker 镜像：约 10-15MB（基于 scratch 或 alpine）
vs
confluentinc/cp-kafka：约 700MB+（基于 JVM 镜像）
```

---

## 9. 官方性能基准数据

> 数据来源：NATS 官方文档和 Synadia 公开基准测试（硬件：8 核 CPU，高速网络）

### Core NATS 吞吐量

```
测试场景：单 Publisher → 单 Subscriber，局域网内，消息大小 128 bytes

单节点 NATS Server：
  吞吐量：18,000,000 msg/s（1800 万条/秒）
  延迟（中位数）：0.1ms
  延迟（p99）：0.3ms

对比：Redis Pub/Sub 约 1-2M msg/s，RabbitMQ 约 200K-500K msg/s
```

### JetStream 持久化吞吐量

```
测试场景：Publish 到 JetStream Stream，R1（单副本），FileStore

单节点 JetStream：
  吞吐量：约 3,000,000 msg/s（持久化写入）
  SSD 持久化 + 同步刷盘：约 300,000 msg/s

R3 副本（3 节点集群，Raft 多数确认）：
  吞吐量：约 500,000 msg/s（受 Raft Quorum 限制）
```

### 连接数

```
单节点 NATS Server：
  支持并发连接数：1,000,000+（100 万）
  每个连接的内存开销：约 1-2KB（不含消息缓冲区）

对比：每个 Kafka Consumer 需要一个 TCP 连接，
  Broker 侧的每连接开销更重（JVM 对象）
```

### 延迟基准

```
测试条件：同一机房（RTT < 0.1ms），消息大小 128 bytes

Core NATS Pub/Sub：
  p50: 0.1ms
  p99: 0.3ms
  p999: 0.5ms

JetStream Publish (R1, MemoryStore)：
  p50: 0.2ms
  p99: 0.5ms

JetStream Publish (R3, FileStore)：
  p50: 1ms
  p99: 3ms
```

### 网络带宽效率

```
NATS 协议的消息 overhead：
  固定头部：约 15-30 bytes（subject + msg_id + 换行符）
  对比 Kafka：每条消息约 96 bytes 固定 overhead（批量时摊薄）
  对比 AMQP：每条消息约 100-200 bytes overhead

结论：对小消息场景（< 1KB），NATS 的带宽利用率更高
```

---

## 10. 小结

NATS 是一个**为云原生时代设计的连接层**，其核心价值在于：

```
┌─────────────────────────────────────────────────────┐
│                  NATS 的核心价值                      │
│                                                     │
│  1. 极致简单       零依赖单二进制，协议极简           │
│  2. 超低延迟       < 1ms p99（Core），局域网内        │
│  3. 高并发连接     单节点百万级并发连接               │
│  4. 灵活语义       AtMostOnce ↔ AtLeastOnce 按需选择 │
│  5. 灵活组网       Cluster + Leaf Node + Gateway     │
│  6. 内置持久化     JetStream 无需额外进程             │
│  7. 内置安全       NKEY + JWT 去中心化鉴权            │
└─────────────────────────────────────────────────────┘
```

NATS 不是"更好的 Kafka"，也不是"更快的 RabbitMQ"。它是一个不同定位的工具：**在微服务互联、IoT 接入、实时控制、边缘计算等场景中，NATS 是目前最简单、最高效的选择**。

---

## 11. NATS 2.0+ 新特性详解

### 11.1 去中心化安全模型（Decentralized Security）

NATS 2.0 引入了全新的安全架构，无需中央认证服务：

```
传统模型：
  Client → 中央认证服务 → 验证 → NATS Server

NATS 2.0 模型：
  Operator（根信任锚）
    └── Account（租户隔离）
          └── User（JWT Token，自包含权限）

  验证流程：
  1. 客户端持有 User JWT + NKey 私钥
  2. 连接时发送 JWT（包含公钥、权限声明）
  3. NATS Server 用 Operator 公钥验证 JWT 签名
  4. 客户端用 NKey 私钥签名挑战，证明身份
```

**关键概念：**

| 概念 | 说明 |
|------|------|
| Operator | 信任根，拥有根密钥对，签发 Account JWT |
| Account | 租户/命名空间，资源隔离单位，可配置导出/导入 Subject |
| User | 最终用户，JWT 包含权限声明（发布/订阅 Subject 列表） |
| NKey | 基于 Ed25519 的高性能密钥对，用于身份验证 |

**Account 隔离示例：**

```yaml
# 账户配置示例
accounts: {
  # 订单服务账户
  ORDER_SERVICE: {
    users: [{user: "order-user", password: "secret"}]
    # 可发布的 Subject
    exports: [
      {service: "orders.>"}
    ]
    # 可订阅的 Subject（来自其他账户）
    imports: [
      {service: {account: "PAYMENT_SERVICE", subject: "payment.events.>"}}
    ]
    # JetStream 配额
    jetstream: {
      max_memory: 1GB
      max_storage: 10GB
      max_streams: 10
      max_consumers: 100
    }
  }
}
```

### 11.2 JetStream 高级特性（2.2+）

**消息压缩（NATS 2.9+）：**

```go
// 启用 S2 压缩（基于 Snappy，高速低开销）
stream, _ := js.CreateOrUpdateStream(ctx, jetstream.StreamConfig{
    Name:        "LOGS",
    Subjects:    []string{"logs.>"},
    Compression: jetstream.S2Compression, // 启用压缩
    MaxAge:      30 * 24 * time.Hour,
})
```

**多 Subject 过滤（NATS 2.8+）：**

```go
// Consumer 可以同时过滤多个不相交的 Subject
cons, _ := js.CreateOrUpdateConsumer(ctx, "ORDERS", jetstream.ConsumerConfig{
    Name: "multi-filter",
    FilterSubjects: []string{
        "orders.created",
        "orders.cancelled",
        "orders.refunded",
    },
})
```

**Stream Sources 增强（NATS 2.10+）：**

```go
// 从多个 Stream 聚合消息，支持 Subject 转换
stream, _ := js.CreateOrUpdateStream(ctx, jetstream.StreamConfig{
    Name: "GLOBAL_EVENTS",
    Sources: []*jetstream.StreamSource{
        {
            Name:          "US_EVENTS",
            FilterSubject: "events.>",
            SubjectTransform: &jetstream.SubjectTransformConfig{
                Source:      "events.>",
                Destination: "us.events.>",
            },
        },
        {
            Name:          "EU_EVENTS",
            FilterSubject: "events.>",
            SubjectTransform: &jetstream.SubjectTransformConfig{
                Source:      "events.>",
                Destination: "eu.events.>",
            },
        },
    },
})
```

### 11.3 性能优化特性

**批量发布（PublishAsync）：**

```go
// 高吞吐场景：异步批量发布
for i := 0; i < 10000; i++ {
    paf, _ := js.PublishAsync("events", data)
    pendingAcks = append(pendingAcks, paf)
    
    // 每 1000 条等待一次确认
    if len(pendingAcks) >= 1000 {
        for _, paf := range pendingAcks {
            select {
            case <-paf.Ok():
                // 成功
            case err := <-paf.Err():
                log.Printf("发布失败: %v", err)
            }
        }
        pendingAcks = pendingAcks[:0]
    }
}
```

**Pull Consumer 批量拉取：**

```go
// 一次拉取多条消息，减少网络往返
msgs, _ := consumer.Fetch(100, jetstream.FetchMaxWait(5*time.Second))
for msg := range msgs.Messages() {
    process(msg)
    msg.Ack()
}
```

---

## 12. 与本项目（livis-claw-hub）的关系

本项目是一个**远程控制中继服务**，NATS 可以在以下场景发挥作用：

### 12.1 设备消息分发

```
当前架构：
  Client → HTTP SSE → livis-claw-hub → WebSocket → Device

NATS 增强架构：
  Client → HTTP API → livis-claw-hub → NATS → Device Worker
                              ↓
                         JetStream（持久化）
                              ↓
                         审计日志 / 重放
```

### 12.2 建议集成点

| 场景 | NATS 模式 | 说明 |
|------|----------|------|
| 设备状态广播 | Pub/Sub | 设备上线/下线通知 |
| 命令下发 | Request/Reply | 同步等待设备响应 |
| 命令队列 | JetStream WorkQueue | 命令持久化，设备离线不丢失 |
| 设备状态存储 | KV Store | 设备在线状态、元数据 |
| 审计日志 | JetStream Stream | 所有命令执行记录 |

### 12.3 迁移建议

```go
// 示例：将设备命令改为通过 JetStream 下发
func (s *Service) SendCommand(ctx context.Context, deviceID, command string) error {
    // 发布到 JetStream，确保命令持久化
    ack, err := s.js.Publish(ctx, 
        fmt.Sprintf("device.%s.command", deviceID),
        []byte(command),
        jetstream.WithMsgID(uuid.New().String()), // 去重
    )
    if err != nil {
        return fmt.Errorf("发布命令失败: %w", err)
    }
    
    log.Printf("命令已持久化: seq=%d", ack.Sequence)
    return nil
}
```

---

> **下一篇**: [02 - Subject 寻址系统](./02-Subject寻址系统.md) — 深入理解 NATS 中一切通信的基础：Subject 命名规范、通配符语义与层级设计模式。


# JetStream Streams 详解

> 系列导航：[01-概念与架构](#) | [02-集群与高可用](#) | [03-安全认证](#) | [04-JetStream 基础](#) | **05-Streams** | [06-Consumers](#)

---

## 目录

1. [Stream 是什么](#1-stream-是什么)
2. [Stream vs Core NATS 的本质区别](#2-stream-vs-core-nats-的本质区别)
3. [内部存储结构](#3-内部存储结构)
4. [StreamConfig 配置项全解](#4-streamconfig-配置项全解)
5. [三种 Retention 策略详解](#5-三种-retention-策略详解)
6. [消息去重机制](#6-消息去重机制)
7. [Stream 操作](#7-stream-操作)
8. [Stream Mirror 与 Source](#8-stream-mirror-与-source)
9. [实际配置示例](#9-实际配置示例)
10. [Go 代码示例](#10-go-代码示例)

---

## 1. Stream 是什么

### 1.1 核心定义

Stream 是 JetStream 的基础存储单元，本质是一个**持久化的、有序的、可重放的消息日志**。你可以把它想象成一个订阅了若干 NATS subject 的"消息仓库"——所有发布到这些 subject 的消息都会被捕获并持久化存储，不论当时是否有消费者在线。

```
                    ┌─────────────────────────────────────────────┐
                    │              NATS Stream: ORDERS             │
  Publisher         │                                              │
  ──────────────    │  seq=1  seq=2  seq=3  seq=4  seq=5  seq=6   │
  orders.created ──▶│  [msg]  [msg]  [msg]  [msg]  [msg]  [msg]  │
  orders.updated ──▶│                                              │
  orders.deleted ──▶│  ← 按时间顺序，全局递增序号，持久化到磁盘 →  │
                    │                                              │
                    └─────────────────────────────────────────────┘
                          ▲              ▲              ▲
                    Consumer-A     Consumer-B     Consumer-C
                    (从头消费)    (从最新消费)   (按时间点消费)
```

### 1.2 类比 Kafka Partition

如果你熟悉 Kafka，可以这样理解：

| 概念        | Kafka              | NATS JetStream           |
|-------------|--------------------|--------------------------|
| 消息日志    | Partition          | Stream                   |
| 主题        | Topic              | Subject（支持通配符）    |
| 偏移量      | Offset             | Sequence Number          |
| 消费者组    | Consumer Group     | Consumer（Durable）      |
| 消费位点    | Committed Offset   | Consumer Sequence        |
| 副本        | Replication Factor | Replicas                 |
| 消息保留    | Retention Policy   | Retention + Limits       |

**关键差异**：NATS Stream 的 Subject 支持通配符，一个 Stream 可以捕获多个 subject 的消息；Kafka 的 Partition 只属于一个 Topic。

### 1.3 Stream 的核心特性

- **持久化**：消息写入磁盘（或内存），服务重启后不丢失
- **有序性**：同一 Stream 内的消息严格按写入顺序排列，全局递增的 sequence number
- **可重放**：消费者可以从任意位置（起点、时间点、序号）开始消费
- **多消费者**：同一 Stream 可以有多个独立的 Consumer，各自维护消费进度
- **消息保留策略**：按时间、数量、大小自动清理旧消息
- **去重**：通过 `Nats-Msg-Id` header 在指定时间窗口内去重

---

## 2. Stream vs Core NATS 的本质区别

### 2.1 消息生命周期对比

```
Core NATS（发布/订阅）：
──────────────────────────────────────────────
  Publisher          NATS Server         Subscriber
     │                   │                   │
     │── publish(msg) ──▶│                   │
     │                   │── deliver(msg) ──▶│  (如果有订阅者)
     │                   │                   │
     │                   │  消息立即丢弃 ✗    │
     │                   │  (无持久化)        │
     │                   │                   │

JetStream Stream：
──────────────────────────────────────────────
  Publisher          NATS Server         Consumer
     │                   │                   │
     │── publish(msg) ──▶│                   │
     │                   │ 写入 Stream ✓     │
     │◀── ACK ──────────│ (持久化磁盘)      │
     │                   │                   │
     │                   │  (Consumer 稍后)  │
     │                   │── deliver(msg) ──▶│
     │                   │◀── ACK ──────────│
     │                   │                   │
```

### 2.2 主要区别总结

| 维度               | Core NATS                  | JetStream Stream              |
|--------------------|----------------------------|-------------------------------|
| 消息持久化         | 否（内存中转）             | 是（文件或内存存储）          |
| 消费者离线         | 消息丢失                   | 消息保留，上线后可消费        |
| 消息顺序           | 尽力保证                   | 严格有序（sequence number）   |
| 消息重放           | 不支持                     | 支持，可从任意位置            |
| 发布确认           | 无                         | 有（PubAck with sequence）    |
| 消费确认           | 无                         | 有（Ack/Nak/Term）            |
| 消息重试           | 不支持                     | 支持（MaxDeliver + AckWait）  |
| 去重               | 不支持                     | 支持（Nats-Msg-Id）           |
| 资源消耗           | 极低                       | 较高（磁盘 I/O、内存）        |
| 适用场景           | 实时通知、低延迟事件       | 任务队列、事件溯源、审计日志  |

### 2.3 何时选择 Stream

**应该使用 Stream 的场景**：
- 消费者可能离线，不能丢消息
- 需要消息重放（如：事件溯源、调试）
- 需要保证消息被处理（任务队列）
- 需要多个独立消费者各自消费完整的消息集
- 需要审计日志

**应该使用 Core NATS 的场景**：
- 实时性要求极高，延迟敏感（如：游戏状态同步）
- 消费者始终在线，丢消息可接受
- 消息量极大，不需要持久化
- 简单的请求/响应模式

---

## 3. 内部存储结构

理解 JetStream 的存储结构有助于调优和故障排查。

### 3.1 File Storage 目录结构

JetStream 数据默认存储在 `$NATS_JETSTREAM_STORE_DIR`（通常配置为 `/data/jetstream` 或 `/tmp/nats/jetstream`）：

```
/data/jetstream/
└── $SERVER_NAME/                    # 服务器标识目录
    ├── meta.inf                     # JetStream 元数据（版本、配置）
    └── streams/
        ├── ORDERS/                  # Stream 名称
        │   ├── meta.inf             # Stream 配置快照
        │   ├── meta.sum             # 配置校验和
        │   └── msgs/                # 消息存储目录
        │       ├── 1.blk            # 第 1 个消息块（message block）
        │       ├── 1.idx            # 第 1 个块的索引文件
        │       ├── 1.fss            # 第 1 个块的 subject 状态
        │       ├── 2.blk
        │       ├── 2.idx
        │       ├── 2.fss
        │       └── ...
        ├── EVENTS/
        │   ├── meta.inf
        │   └── msgs/
        │       └── ...
        └── consumers/
            └── ORDERS/              # 属于 ORDERS Stream 的 Consumer
                └── my-consumer/
                    ├── meta.inf     # Consumer 配置
                    └── o.dat        # Consumer 状态（已 Ack 序号等）
```

### 3.2 Message Block 文件格式（.blk）

每个 `.blk` 文件是一个消息块，包含若干条消息，按追加方式写入：

```
Block 文件内部结构（.blk）：
┌─────────────────────────────────────────────────────────┐
│  Record 1                                               │
│  ┌──────────┬──────────┬──────────┬──────────────────┐  │
│  │ length   │ sequence │timestamp │ headers + payload │  │
│  │ (4 bytes)│ (8 bytes)│ (8 bytes)│  (variable len)  │  │
│  └──────────┴──────────┴──────────┴──────────────────┘  │
│  Record 2                                               │
│  ┌──────────┬──────────┬──────────┬──────────────────┐  │
│  │ length   │ sequence │timestamp │ headers + payload │  │
│  └──────────┴──────────┴──────────┴──────────────────┘  │
│  ...                                                    │
└─────────────────────────────────────────────────────────┘
```

每个 block 默认最大 **64MB**（可通过 `max_file_store` 配置），达到上限后滚动生成新 block。消息数据使用 LZ4 压缩（可配置）。

### 3.3 Index 文件（.idx）

`.idx` 文件记录对应 block 中每条消息的偏移量，用于按 sequence 快速定位：

```
Index 文件内部结构（.idx）：
┌─────────────────────────────────────────────────┐
│  Header                                         │
│  ┌───────────┬──────────┬──────────┬──────────┐ │
│  │ magic     │ version  │ first_seq│ last_seq │ │
│  └───────────┴──────────┴──────────┴──────────┘ │
│                                                 │
│  Entries（每条消息一个条目）                    │
│  ┌──────────┬────────────────────────────────┐  │
│  │ sequence │ offset in .blk file            │  │
│  └──────────┴────────────────────────────────┘  │
│  ┌──────────┬────────────────────────────────┐  │
│  │ sequence │ offset                         │  │
│  └──────────┴────────────────────────────────┘  │
│  ...                                            │
└─────────────────────────────────────────────────┘
```

### 3.4 Subject State 文件（.fss）

`.fss`（File Subject State）文件记录每个 block 内各 subject 的消息分布，用于支持 `MaxMsgsPerSubject` 限制和 `DeliverLastPerSubject` 策略：

```
FSS 文件内部结构（.fss）：
┌─────────────────────────────────────────────────────┐
│  subject "orders.created"                           │
│  ┌──────────────┬───────────────┬────────────────┐  │
│  │ first_seq=1  │ last_seq=150  │  total_msgs=89 │  │
│  └──────────────┴───────────────┴────────────────┘  │
│  subject "orders.updated"                           │
│  ┌──────────────┬───────────────┬────────────────┐  │
│  │ first_seq=3  │ last_seq=148  │  total_msgs=45 │  │
│  └──────────────┴───────────────┴────────────────┘  │
│  ...                                                │
└─────────────────────────────────────────────────────┘
```

### 3.5 Memory Storage

当 `Storage: nats.MemoryStorage` 时，消息完全存储在进程内存中：

**优点**：
- 读写延迟极低（无磁盘 I/O）
- 适合临时数据、缓存场景

**缺点**：
- 服务重启后消息丢失
- 受限于可用内存

**适用场景**：
- 毫秒级延迟要求的临时队列
- 测试环境
- 消息量小且可接受丢失的实时计算场景
- 配合 `MaxAge` 短期保留（如：最近 5 分钟的指标数据）

```
Memory Storage 内部结构：

  msgs map[uint64]*storedMsg    // sequence -> 消息数据
  ├── 1 -> {subj, hdr, msg, ts}
  ├── 2 -> {subj, hdr, msg, ts}
  └── ...

  subjects map[string]uint64    // subject -> 最新 sequence
```

### 3.6 存储性能参考

```
典型硬件（NVMe SSD）下的写入吞吐量参考：

  File Storage（单节点）：
  ┌─────────────────────────────────────────────┐
  │  消息大小  │  吞吐量         │  延迟 p99   │
  │  1 KB      │  ~200,000 msg/s │  < 1 ms     │
  │  4 KB      │  ~100,000 msg/s │  < 2 ms     │
  │  64 KB     │  ~20,000 msg/s  │  < 5 ms     │
  └─────────────────────────────────────────────┘

  Memory Storage（单节点）：
  ┌─────────────────────────────────────────────┐
  │  消息大小  │  吞吐量         │  延迟 p99   │
  │  1 KB      │  ~1,000,000/s   │  < 0.1 ms   │
  │  4 KB      │  ~500,000/s     │  < 0.1 ms   │
  └─────────────────────────────────────────────┘
```

---

## 4. StreamConfig 配置项全解

```go
// nats.go / jetstream.go 中的 StreamConfig
type StreamConfig struct {
    Name         string          // Stream 名称（不可变）
    Description  string          // 描述（可热更新）
    Subjects     []string        // 捕获的 subject 列表（可热更新）
    Retention    RetentionPolicy // 保留策略（不可变）
    MaxConsumers int             // 最大 Consumer 数量，-1 为不限制
    MaxMsgs      int64           // 最大消息数量，-1 为不限制
    MaxBytes     int64           // 最大存储字节数，-1 为不限制
    MaxAge       time.Duration   // 消息最大保留时间，0 为不限制
    MaxMsgSize   int32           // 单条消息最大字节，-1 为不限制
    MaxMsgsPerSubject int64      // 每个 subject 最大消息数，0 为不限制
    Storage      StorageType     // File 或 Memory
    Replicas     int             // 副本数（1、3、5）
    NoAck        bool            // 发布时不需要 ACK
    Duplicates   time.Duration   // 去重窗口时长
    Discard      DiscardPolicy   // 超限时的丢弃策略
    DiscardNewPerSubject bool    // 按 subject 粒度执行 Discard New
    AllowRollup  bool            // 允许 KV Rollup 操作
    DenyDelete   bool            // 禁止按 sequence 删除消息
    DenyPurge    bool            // 禁止 Purge Stream
    AllowDirect  bool            // 允许直接消息获取（绕过 Consumer）
    MirrorDirect bool            // Mirror 是否允许直接获取
    Mirror       *StreamSource   // 镜像配置
    Sources      []*StreamSource // 聚合来源配置
    Sealed       bool            // 封存（不再接受新消息）
    Compression  StoreCompression // 压缩算法（None/S2）
    NumReplicas  int             // 同 Replicas（别名）
    Placement    *Placement      // 集群节点约束
    SubjectTransform *SubjectTransformConfig // subject 转换
}
```

### 4.1 Name（命名规范）

Stream 名称是全局唯一标识符，创建后**不可修改**。

**命名规则**：
- 只能包含字母、数字、`-`、`_`
- 不能包含空格或 `.`（点）
- 大小写敏感
- 最大长度 256 字符

**推荐命名规范**：
```
{DOMAIN}-{ENTITY}-{VERSION}

示例：
  ORDERS-V1           # 订单事件流，v1 版本
  DEVICE-TELEMETRY    # 设备遥测数据
  USER-AUDIT-LOG      # 用户审计日志
  PAYMENT-EVENTS      # 支付事件
```

### 4.2 Subjects（subject 捕获规则）

`Subjects` 定义哪些 NATS subject 的消息会被写入这个 Stream。支持通配符。

```
通配符规则：
  *  匹配单层（一个 token）
  >  匹配多层（一个或多个 token，只能在末尾）

示例：
  "orders.>"         匹配 orders.created, orders.updated, orders.v2.created
  "orders.*"         匹配 orders.created, orders.updated（不匹配 orders.v2.created）
  "device.*.status"  匹配 device.abc123.status, device.xyz789.status
```

**重要约束**：
- 一个 subject 只能被一个 Stream 捕获（NATS Server 会拒绝冲突配置）
- 同一 Stream 可以配置多个 subjects

```go
// 多 subject 示例
cfg := jetstream.StreamConfig{
    Name: "ORDERS",
    Subjects: []string{
        "orders.created",
        "orders.updated",
        "orders.deleted",
        "orders.*.cancelled",  // 支持通配符
    },
}
```

### 4.3 Retention

| 值                        | 含义                                           |
|---------------------------|------------------------------------------------|
| `LimitsPolicy`（默认）    | 按 MaxAge/MaxMsgs/MaxBytes 限制保留            |
| `WorkQueuePolicy`         | 消息被 Consumer Ack 后自动删除                 |
| `InterestPolicy`          | 所有 Consumer 都 Ack 后才删除                  |

详见 [第 5 节](#5-三种-retention-策略详解)。

### 4.4 MaxAge / MaxMsgs / MaxBytes / MaxMsgSize

这四个字段共同定义 Limits 策略下的保留上限，**任何一个条件触发都会清理旧消息**：

```go
cfg := jetstream.StreamConfig{
    Name:       "ORDERS",
    Subjects:   []string{"orders.>"},
    MaxAge:     7 * 24 * time.Hour, // 保留 7 天
    MaxMsgs:    10_000_000,         // 最多 1000 万条
    MaxBytes:   10 * 1024 * 1024 * 1024, // 最多 10 GB
    MaxMsgSize: 1 * 1024 * 1024,   // 单条最大 1 MB
}
```

**清理行为**：
- 超出限制时，默认删除**最旧的消息**（Discard: Old）
- `MaxMsgSize` 超限时，发布操作直接返回错误，消息不写入

### 4.5 MaxMsgsPerSubject

限制每个 subject 保留的最新消息数量，是实现"每个 key 只保留最新 N 条"的关键配置：

```go
cfg := jetstream.StreamConfig{
    Name:               "DEVICE-STATUS",
    Subjects:           []string{"device.*.status"},
    MaxMsgsPerSubject:  1,  // 每个设备只保留最新状态
}
```

这个配置常用于：
- **KV 语义**：每个 key 只保留最新值（`MaxMsgsPerSubject: 1`）
- **滑动窗口**：每个 subject 保留最新 N 条记录

### 4.6 Storage

```go
// File Storage（默认，推荐生产环境）
Storage: jetstream.FileStorage

// Memory Storage（高性能，重启丢失）
Storage: jetstream.MemoryStorage
```

### 4.7 Replicas（副本数）

副本数决定数据的高可用性，**必须为奇数**（1、3、5），以满足 Raft 共识算法的多数派要求：

```
副本数与可用性：

  Replicas=1: 无副本，节点宕机消息丢失（开发/测试）
  Replicas=3: 可容忍 1 节点故障（推荐生产）
  Replicas=5: 可容忍 2 节点故障（高可用场景）

Raft 共识需要 (N/2)+1 个节点存活才能继续写入：
  3 副本 → 需要 2 节点存活
  5 副本 → 需要 3 节点存活
```

**注意**：`Replicas` 不能大于集群中的服务器数量。

### 4.8 NoAck

```go
NoAck: false  // 默认：发布后等待 Server ACK，确认消息已持久化
NoAck: true   // 发布后不等待 ACK（类似 Core NATS），吞吐量更高但不保证持久化
```

通常保持默认 `false`，只有在对延迟极其敏感且可以接受少量丢失的场景下才设为 `true`。

### 4.9 Duplicates（去重窗口）

```go
Duplicates: 5 * time.Minute  // 5 分钟内相同 Nats-Msg-Id 的消息只接受一次
```

详见 [第 6 节](#6-消息去重机制)。

### 4.10 Discard（超限时的行为）

```go
// DiscardOld（默认）：超限时删除最旧的消息，新消息写入成功
Discard: jetstream.DiscardOld

// DiscardNew：超限时拒绝新消息，返回错误给发布者
Discard: jetstream.DiscardNew
```

`DiscardNew` 适用于需要保护旧数据的场景，例如审计日志——不允许因新消息挤占而丢失历史记录。

`DiscardNewPerSubject: true` 可以在 `DiscardNew` 模式下，按 subject 粒度执行（配合 `MaxMsgsPerSubject`）。

### 4.11 AllowRollup / DenyDelete / DenyPurge

```go
// AllowRollup：允许发布特殊的 Rollup 消息（用于 KV 压缩历史）
AllowRollup: true

// DenyDelete：禁止通过 API 按 sequence 删除单条消息
DenyDelete: true

// DenyPurge：禁止清空整个 Stream
DenyPurge: true
```

这三个字段通常用于**合规性要求**场景，防止数据被意外或恶意删除。

### 4.12 Compression

```go
// 无压缩（默认）
Compression: jetstream.NoCompression

// S2 压缩（基于 Snappy，高速压缩）
Compression: jetstream.S2Compression
```

S2 压缩可以显著减少磁盘占用（文本类消息可达 3-10x 压缩比），CPU 开销极低，推荐在存储受限的环境中开启。

---

## 5. 三种 Retention 策略详解

### 5.1 Limits Policy（最常用）

**行为**：消息根据 MaxAge/MaxMsgs/MaxBytes 自动清理，类似滚动日志文件。

```
Limits Policy 消息生命周期：

  写入 → 存储 → [达到 MaxAge 或 MaxMsgs 或 MaxBytes] → 自动删除

  时间轴：
  ┌──────────────────────────────────────────────────────────┐
  │  t=0h  t=24h  t=48h  t=72h  t=96h  t=120h  t=144h      │
  │  [msg1][msg2][msg3][msg4][msg5][ msg6 ][ msg7 ]          │
  │                                                          │
  │  MaxAge = 72h，则 t=72h 时 msg1 被删除：                 │
  │         [msg2][msg3][msg4][msg5][ msg6 ][ msg7 ]         │
  └──────────────────────────────────────────────────────────┘
```

**适用场景**：
- 事件日志（保留 30 天）
- 时序数据（保留最近 100 万条）
- 审计记录（保留 1 年）

```go
cfg := jetstream.StreamConfig{
    Name:     "AUDIT-LOG",
    Subjects: []string{"audit.>"},
    Retention: jetstream.LimitsPolicy,  // 默认，可省略
    MaxAge:   30 * 24 * time.Hour,
    MaxMsgs:  -1,   // 不限数量
    MaxBytes: 50 * 1024 * 1024 * 1024, // 50 GB
    Storage:  jetstream.FileStorage,
    Replicas: 3,
}
```

### 5.2 WorkQueue Policy（任务队列）

**行为**：消息被 **任意一个** Consumer Ack 后立即从 Stream 中删除。每条消息只处理一次。

```
WorkQueue Policy 消息流转：

  Publisher      Stream (TASKS)        Consumer A       Consumer B
      │               │                    │                │
      │── task1 ─────▶│                    │                │
      │── task2 ─────▶│                    │                │
      │── task3 ─────▶│                    │                │
      │               │                    │                │
      │               │── deliver task1 ──▶│                │
      │               │── deliver task2 ──────────────────▶│
      │               │── deliver task3 ──▶│                │
      │               │                    │                │
      │               │◀── ACK(task1) ─────│                │
      │               │  task1 被删除 ✓    │                │
      │               │◀── ACK(task2) ──────────────────────│
      │               │  task2 被删除 ✓    │                │
```

**关键限制**：
- WorkQueue Stream **只允许有一个** Consumer（NATS Server 强制限制）
- Consumer 必须有 `FilterSubject` 或覆盖所有消息

**适用场景**：
- 任务分发队列（Worker Pool 模式）
- 幂等性任务处理
- 邮件发送、图片处理等后台任务

```go
cfg := jetstream.StreamConfig{
    Name:      "TASKS",
    Subjects:  []string{"tasks.>"},
    Retention: jetstream.WorkQueuePolicy,
    Storage:   jetstream.FileStorage,
    Replicas:  3,
}
```

### 5.3 Interest Policy（按兴趣保留）

**行为**：只有当 Stream 上存在至少一个 Consumer 时才保留消息。所有 Consumer 都 Ack 后，消息被删除。如果没有 Consumer，消息直接丢弃（类似 Core NATS）。

```
Interest Policy 消息保留逻辑：

  场景 1：有 Consumer，所有 Consumer 都 Ack → 消息删除
  ┌─────────────────────────────────────────────────────────┐
  │  Consumer A (关注 orders.*)                             │
  │  Consumer B (关注 orders.*)                             │
  │                                                         │
  │  msg1 写入 → A 收到并 Ack + B 收到并 Ack → 删除 ✓      │
  └─────────────────────────────────────────────────────────┘

  场景 2：没有 Consumer → 消息不保留
  ┌─────────────────────────────────────────────────────────┐
  │  （无 Consumer 时）                                     │
  │  msg1 写入 → 没有订阅者 → 直接丢弃 ✗                   │
  └─────────────────────────────────────────────────────────┘
```

**适用场景**：
- 广播场景：需要确保所有在线消费者都收到，收到后即可清理
- 数据扇出（Fan-out）：同一消息需要被多个服务各自处理一次

```go
cfg := jetstream.StreamConfig{
    Name:      "NOTIFICATIONS",
    Subjects:  []string{"notify.>"},
    Retention: jetstream.InterestPolicy,
    Storage:   jetstream.FileStorage,
    Replicas:  3,
}
```

---

## 6. 消息去重机制

### 6.1 为什么需要去重

在分布式系统中，网络故障可能导致发布者收不到 ACK，进而重试发布，造成消息重复。JetStream 通过 `Nats-Msg-Id` header 提供幂等发布：

```
无去重的问题：

  Publisher          NATS Server
      │                   │
      │── publish(msg) ──▶│  (Server 已持久化)
      │                   │── ACK ──X  (网络中断，Publisher 未收到)
      │                   │
      │── retry(msg) ────▶│  (重复写入！)
      │◀── ACK ───────────│
      │                   │
  结果：消息被存储了两次 ✗
```

### 6.2 Nats-Msg-Id 工作原理

```
有去重的流程：

  Publisher          NATS Server (Duplicates: 5min)
      │                   │
      │── publish(msg,    │
      │    Nats-Msg-Id:   │
      │    "order-123") ─▶│  写入，记录 ID 到去重缓存
      │                   │── ACK(seq=1) ──X  (网络中断)
      │                   │
      │── retry(msg,      │
      │    Nats-Msg-Id:   │
      │    "order-123") ─▶│  检测到 ID 已存在！
      │◀── ACK(seq=1) ────│  返回原始 seq，不重复写入 ✓
```

去重状态存储在内存中，时间窗口由 `Duplicates` 配置（默认 2 分钟，建议根据重试间隔设置）。

### 6.3 生成消息 ID 的最佳实践

```go
import (
    "fmt"
    "github.com/google/uuid"
    "github.com/nats-io/nats.go/jetstream"
)

// 方式 1：UUID（最常用）
msgID := uuid.New().String()

// 方式 2：业务语义 ID（推荐，便于追踪）
msgID := fmt.Sprintf("order-%s-%d", orderID, version)

// 方式 3：内容哈希（适合幂等更新）
msgID := fmt.Sprintf("%x", sha256.Sum256(payload))

// 发布时携带 Nats-Msg-Id
ack, err := js.Publish(ctx, "orders.created", payload,
    jetstream.WithMsgID(msgID),
)
```

### 6.4 去重的限制

- 去重缓存存储在**内存**中，Server 重启后清空
- 去重窗口外的重复消息**无法检测**
- 集群模式下，去重状态在 Stream Leader 节点维护
- `Duplicates` 窗口建议设置为重试超时时间的 2 倍

---

## 7. Stream 操作

### 7.1 创建 Stream

```go
js, _ := jetstream.New(nc)

stream, err := js.CreateStream(ctx, jetstream.StreamConfig{
    Name:       "ORDERS",
    Subjects:   []string{"orders.>"},
    MaxAge:     7 * 24 * time.Hour,
    Storage:    jetstream.FileStorage,
    Replicas:   3,
    Duplicates: 5 * time.Minute,
})
```

### 7.2 Update（热更新）

不是所有字段都支持热更新，下表列出哪些字段可以在不重建 Stream 的情况下修改：

| 字段                  | 可热更新 | 说明                           |
|-----------------------|----------|--------------------------------|
| `Description`         | 是       |                                |
| `Subjects`            | 是       | 可增减 subject（注意冲突检查） |
| `MaxAge`              | 是       |                                |
| `MaxMsgs`             | 是       |                                |
| `MaxBytes`            | 是       |                                |
| `MaxMsgSize`          | 是       |                                |
| `MaxMsgsPerSubject`   | 是       |                                |
| `MaxConsumers`        | 是       |                                |
| `Duplicates`          | 是       |                                |
| `Discard`             | 是       |                                |
| `AllowRollup`         | 是       |                                |
| `Compression`         | 是       |                                |
| `Name`                | **否**   | 不可变                         |
| `Storage`             | **否**   | 不可变（需重建）               |
| `Retention`           | **否**   | 不可变（需重建）               |
| `Replicas`            | 是*      | 需要集群支持                   |

```go
// 热更新示例：扩展保留时间
stream, err := js.UpdateStream(ctx, jetstream.StreamConfig{
    Name:   "ORDERS",
    MaxAge: 30 * 24 * time.Hour,  // 从 7 天改为 30 天
    // 其他字段保持不变...
})
```

### 7.3 Delete Stream

```go
err := js.DeleteStream(ctx, "ORDERS")
// 注意：这会删除 Stream 及其所有消息和 Consumer 配置！
```

### 7.4 Purge Stream

Purge 清空 Stream 中的所有消息，但保留 Stream 配置和 Consumer 配置：

```go
// 清空所有消息
err := stream.Purge(ctx)

// 清空特定 subject 的消息
err := stream.Purge(ctx, jetstream.WithPurgeSubject("orders.cancelled"))

// 清空 sequence 之前的消息（保留 seq > 1000 的消息）
err := stream.Purge(ctx, jetstream.WithPurgeSequence(1000))

// 保留最新 N 条
err := stream.Purge(ctx, jetstream.WithPurgeKeep(100))
```

### 7.5 查询 Stream 信息

```go
// 获取 Stream 对象
stream, err := js.Stream(ctx, "ORDERS")

// 获取运行时信息
info, err := stream.Info(ctx)
fmt.Printf("消息总数: %d\n", info.State.Msgs)
fmt.Printf("占用字节: %d\n", info.State.Bytes)
fmt.Printf("第一条序号: %d\n", info.State.FirstSeq)
fmt.Printf("最后序号: %d\n", info.State.LastSeq)
fmt.Printf("Consumer 数量: %d\n", info.State.Consumers)

// 列出所有 Stream
for name := range js.StreamNames(ctx) {
    fmt.Println(name)
}
```

---

## 8. Stream Mirror 与 Source

### 8.1 Mirror（镜像）

Mirror 创建一个只读的 Stream 副本，实时同步源 Stream 的所有消息：

```
Mirror 架构：

  DC-A (主数据中心)              DC-B (备份数据中心)
  ┌─────────────────────┐        ┌─────────────────────┐
  │  Stream: ORDERS     │        │  Stream: ORDERS-BKP │
  │  (源 Stream)        │───────▶│  (Mirror)           │
  │  seq: 1..1000000    │  同步  │  seq: 1..1000000    │
  └─────────────────────┘        └─────────────────────┘
        │                                   │
   Consumer (写入)               Consumer (只读，容灾)
```

**Mirror 的特性**：
- Mirror Stream 是**只读**的，不能直接写入
- 自动从源 Stream 复制消息，保持序号一致
- 支持跨 NATS Account 和跨集群（需要 LeafNode 或 NGS）
- 可以配置 `FilterSubject` 只镜像部分消息

```go
// 创建 Mirror Stream
mirrorCfg := jetstream.StreamConfig{
    Name: "ORDERS-BACKUP",
    Mirror: &jetstream.StreamSource{
        Name: "ORDERS",
        // 可选：从特定时间点开始同步
        OptStartTime: &startTime,
        // 可选：只镜像特定 subject
        FilterSubject: "orders.created",
    },
    Storage:  jetstream.FileStorage,
    Replicas: 1,
}
stream, err := js.CreateStream(ctx, mirrorCfg)
```

### 8.2 Source（聚合多个 Stream）

Source 允许将多个 Stream 的消息聚合到一个 Stream 中：

```
Source 架构（多数据源聚合）：

  Stream: ORDERS-US       ─────┐
  (美国区订单)                 │
                               ▼
  Stream: ORDERS-EU       ────▶  Stream: ORDERS-GLOBAL
  (欧洲区订单)                 │  (全球聚合 Stream)
                               │
  Stream: ORDERS-APAC     ─────┘
  (亚太区订单)
```

```go
// 创建聚合 Stream
globalCfg := jetstream.StreamConfig{
    Name: "ORDERS-GLOBAL",
    Sources: []*jetstream.StreamSource{
        {
            Name:          "ORDERS-US",
            FilterSubject: "orders.>",
        },
        {
            Name:          "ORDERS-EU",
            FilterSubject: "orders.>",
        },
        {
            Name:          "ORDERS-APAC",
            FilterSubject: "orders.>",
        },
    },
    Storage:  jetstream.FileStorage,
    Replicas: 3,
}
```

**Mirror vs Source 对比**：

| 维度           | Mirror                   | Source                     |
|----------------|--------------------------|----------------------------|
| 来源数量       | 只能一个                 | 可以多个                   |
| 写入权限       | 只读                     | 可以直接写入               |
| 序号           | 与源一致                 | 重新编号                   |
| 主要用途       | 灾备、异地副本           | 多流聚合、数据汇总         |

---

## 9. 实际配置示例

### 9.1 事件日志流（保留 30 天）

```go
eventLogCfg := jetstream.StreamConfig{
    Name:        "AUDIT-EVENTS",
    Description: "系统审计事件日志，保留 30 天",
    Subjects:    []string{"audit.>"},
    Retention:   jetstream.LimitsPolicy,
    MaxAge:      30 * 24 * time.Hour,
    MaxBytes:    100 * 1024 * 1024 * 1024, // 100 GB 上限
    MaxMsgSize:  64 * 1024,                // 单条最大 64 KB
    Storage:     jetstream.FileStorage,
    Replicas:    3,
    Compression: jetstream.S2Compression,
    DenyDelete:  true,  // 禁止删除单条记录（合规要求）
    DenyPurge:   true,  // 禁止清空
    Duplicates:  10 * time.Minute,
}
```

### 9.2 任务队列流（Work Queue）

```go
taskQueueCfg := jetstream.StreamConfig{
    Name:        "BACKGROUND-TASKS",
    Description: "后台任务队列",
    Subjects:    []string{"tasks.email", "tasks.sms", "tasks.push"},
    Retention:   jetstream.WorkQueuePolicy,
    MaxAge:      24 * time.Hour,   // 未处理任务最多保留 24 小时
    MaxMsgs:     1_000_000,
    MaxMsgSize:  1 * 1024 * 1024,  // 1 MB
    Storage:     jetstream.FileStorage,
    Replicas:    3,
    Discard:     jetstream.DiscardNew, // 队列满时拒绝新任务，不丢弃旧任务
}
```

### 9.3 设备状态流（每个设备只保留最新状态）

```go
deviceStatusCfg := jetstream.StreamConfig{
    Name:               "DEVICE-STATUS",
    Description:        "IoT 设备状态，每个设备保留最新一条",
    Subjects:           []string{"device.*.status"},
    Retention:          jetstream.LimitsPolicy,
    MaxMsgsPerSubject:  1,             // 每个设备只保留最新状态
    MaxAge:             7 * 24 * time.Hour,
    Storage:            jetstream.FileStorage,
    Replicas:           3,
    AllowDirect:        true,          // 允许直接按 subject 查询最新值
    Duplicates:         30 * time.Second,
}
```

### 9.4 高吞吐遥测数据流（Memory + 短期保留）

```go
telemetryCfg := jetstream.StreamConfig{
    Name:        "TELEMETRY-REALTIME",
    Description: "实时遥测数据，内存存储，保留 5 分钟",
    Subjects:    []string{"telemetry.>"},
    Retention:   jetstream.LimitsPolicy,
    MaxAge:      5 * time.Minute,
    MaxBytes:    512 * 1024 * 1024,  // 512 MB 内存上限
    Storage:     jetstream.MemoryStorage,
    Replicas:    1,  // 内存存储不适合多副本
    NoAck:       false,
}
```

### 9.5 合规性消息流（防篡改）

```go
complianceCfg := jetstream.StreamConfig{
    Name:        "COMPLIANCE-LOG",
    Description: "合规日志，防删除防清空，保留 7 年",
    Subjects:    []string{"compliance.>"},
    Retention:   jetstream.LimitsPolicy,
    MaxAge:      7 * 365 * 24 * time.Hour,  // 7 年
    MaxBytes:    1024 * 1024 * 1024 * 1024, // 1 TB
    Storage:     jetstream.FileStorage,
    Replicas:    5,            // 最高可用性
    DenyDelete:  true,
    DenyPurge:   true,
    Compression: jetstream.S2Compression,
    Duplicates:  1 * time.Hour,
}
```

---

## 10. Go 代码示例

### 10.1 完整的 Stream 管理示例

```go
package main

import (
    "context"
    "encoding/json"
    "fmt"
    "log"
    "time"

    "github.com/nats-io/nats.go"
    "github.com/nats-io/nats.go/jetstream"
)

func main() {
    // 连接 NATS Server
    nc, err := nats.Connect("nats://localhost:4222")
    if err != nil {
        log.Fatalf("连接失败: %v", err)
    }
    defer nc.Close()

    // 创建 JetStream 上下文
    js, err := jetstream.New(nc)
    if err != nil {
        log.Fatalf("JetStream 初始化失败: %v", err)
    }

    ctx := context.Background()

    // 创建 Stream
    stream, err := createOrderStream(ctx, js)
    if err != nil {
        log.Fatalf("创建 Stream 失败: %v", err)
    }

    // 发布消息
    if err := publishMessages(ctx, js); err != nil {
        log.Fatalf("发布消息失败: %v", err)
    }

    // 查询 Stream 信息
    if err := queryStreamInfo(ctx, stream); err != nil {
        log.Fatalf("查询信息失败: %v", err)
    }

    // 按 subject 直接获取消息（需要 AllowDirect: true）
    if err := directGet(ctx, stream); err != nil {
        log.Printf("直接获取失败: %v", err)
    }
}

// createOrderStream 创建订单 Stream
func createOrderStream(ctx context.Context, js jetstream.JetStream) (jetstream.Stream, error) {
    cfg := jetstream.StreamConfig{
        Name:        "ORDERS",
        Description: "订单事件流",
        Subjects:    []string{"orders.>"},
        Retention:   jetstream.LimitsPolicy,
        MaxAge:      7 * 24 * time.Hour,
        MaxMsgs:     10_000_000,
        MaxBytes:    10 * 1024 * 1024 * 1024,
        MaxMsgSize:  1 * 1024 * 1024,
        Storage:     jetstream.FileStorage,
        Replicas:    1, // 单节点开发环境
        Duplicates:  5 * time.Minute,
        AllowDirect: true,
        Compression: jetstream.S2Compression,
    }

    // CreateOrUpdateStream：存在则更新，不存在则创建（幂等）
    stream, err := js.CreateOrUpdateStream(ctx, cfg)
    if err != nil {
        return nil, fmt.Errorf("CreateOrUpdateStream: %w", err)
    }

    info, _ := stream.Info(ctx)
    fmt.Printf("Stream 创建成功: %s (已有 %d 条消息)\n",
        info.Config.Name, info.State.Msgs)

    return stream, nil
}

// OrderEvent 订单事件结构
type OrderEvent struct {
    OrderID   string    `json:"order_id"`
    Event     string    `json:"event"`
    Amount    float64   `json:"amount"`
    CreatedAt time.Time `json:"created_at"`
}

// publishMessages 发布带去重 ID 的消息
func publishMessages(ctx context.Context, js jetstream.JetStream) error {
    events := []struct {
        subject string
        event   OrderEvent
    }{
        {
            subject: "orders.created",
            event: OrderEvent{
                OrderID:   "ORD-001",
                Event:     "created",
                Amount:    299.99,
                CreatedAt: time.Now(),
            },
        },
        {
            subject: "orders.updated",
            event: OrderEvent{
                OrderID:   "ORD-001",
                Event:     "updated",
                Amount:    349.99,
                CreatedAt: time.Now(),
            },
        },
        {
            subject: "orders.created",
            event: OrderEvent{
                OrderID:   "ORD-002",
                Event:     "created",
                Amount:    199.50,
                CreatedAt: time.Now(),
            },
        },
    }

    for _, e := range events {
        payload, err := json.Marshal(e.event)
        if err != nil {
            return fmt.Errorf("序列化失败: %w", err)
        }

        // 使用业务 ID 作为去重 ID
        msgID := fmt.Sprintf("%s-%s", e.event.OrderID, e.event.Event)

        ack, err := js.Publish(ctx, e.subject, payload,
            jetstream.WithMsgID(msgID),
        )
        if err != nil {
            return fmt.Errorf("发布 %s 失败: %w", e.subject, err)
        }

        fmt.Printf("发布成功: subject=%s, seq=%d, duplicate=%v\n",
            e.subject, ack.Sequence, ack.Duplicate)
    }

    // 演示去重：重复发布相同 MsgID
    payload, _ := json.Marshal(events[0].event)
    ack, err := js.Publish(ctx, "orders.created", payload,
        jetstream.WithMsgID("ORD-001-created"), // 已存在的 ID
    )
    if err != nil {
        return fmt.Errorf("重复发布失败: %w", err)
    }
    fmt.Printf("重复发布（应被去重）: seq=%d, duplicate=%v\n",
        ack.Sequence, ack.Duplicate) // duplicate=true

    return nil
}

// queryStreamInfo 查询 Stream 运行时信息
func queryStreamInfo(ctx context.Context, stream jetstream.Stream) error {
    info, err := stream.Info(ctx)
    if err != nil {
        return fmt.Errorf("获取 Stream 信息失败: %w", err)
    }

    fmt.Printf("\n=== Stream 状态 ===\n")
    fmt.Printf("名称:         %s\n", info.Config.Name)
    fmt.Printf("消息总数:     %d\n", info.State.Msgs)
    fmt.Printf("存储字节:     %d bytes\n", info.State.Bytes)
    fmt.Printf("第一条序号:   %d\n", info.State.FirstSeq)
    fmt.Printf("最后序号:     %d\n", info.State.LastSeq)
    fmt.Printf("Consumer 数:  %d\n", info.State.Consumers)
    fmt.Printf("Subject 数:   %d\n", info.State.NumSubjects)
    fmt.Printf("Leader:       %s\n", info.Cluster.Leader)

    return nil
}

// directGet 直接按 subject 获取最新消息（无需 Consumer）
func directGet(ctx context.Context, stream jetstream.Stream) error {
    // 获取指定 subject 的最新消息
    msg, err := stream.GetLastMsgForSubject(ctx, "orders.created")
    if err != nil {
        return fmt.Errorf("GetLastMsg 失败: %w", err)
    }

    fmt.Printf("\n=== 直接获取最新消息 ===\n")
    fmt.Printf("Subject: %s\n", msg.Subject)
    fmt.Printf("Seq:     %d\n", msg.Sequence)
    fmt.Printf("Time:    %s\n", msg.Time.Format(time.RFC3339))
    fmt.Printf("Payload: %s\n", string(msg.Data))

    // 按序号获取消息
    msg2, err := stream.GetMsg(ctx, 1)
    if err != nil {
        return fmt.Errorf("GetMsg(seq=1) 失败: %w", err)
    }
    fmt.Printf("\n序号 1 的消息: %s\n", string(msg2.Data))

    return nil
}
```

### 10.2 批量发布高性能示例

```go
// asyncPublish 异步批量发布，高吞吐
func asyncPublish(ctx context.Context, js jetstream.JetStream, count int) error {
    type result struct {
        seq uint64
        err error
    }

    results := make(chan result, count)

    // 并发发布
    for i := 0; i < count; i++ {
        i := i
        go func() {
            payload, _ := json.Marshal(map[string]interface{}{
                "index":     i,
                "timestamp": time.Now().UnixNano(),
            })

            ack, err := js.PublishAsync("orders.created", payload,
                jetstream.WithMsgID(fmt.Sprintf("msg-%d", i)),
            )
            if err != nil {
                results <- result{err: err}
                return
            }

            // 等待异步 ACK
            select {
            case pubAck := <-ack.Ok():
                results <- result{seq: pubAck.Sequence}
            case err := <-ack.Err():
                results <- result{err: err}
            case <-ctx.Done():
                results <- result{err: ctx.Err()}
            }
        }()
    }

    // 收集结果
    successCount := 0
    for i := 0; i < count; i++ {
        r := <-results
        if r.err != nil {
            log.Printf("发布失败: %v", r.err)
        } else {
            successCount++
        }
    }

    fmt.Printf("成功发布 %d/%d 条消息\n", successCount, count)
    return nil
}
```

### 10.3 Stream 监控示例

```go
// monitorStream 持续监控 Stream 状态
func monitorStream(ctx context.Context, js jetstream.JetStream, streamName string) {
    ticker := time.NewTicker(10 * time.Second)
    defer ticker.Stop()

    var lastMsgs uint64

    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            stream, err := js.Stream(ctx, streamName)
            if err != nil {
                log.Printf("获取 Stream 失败: %v", err)
                continue
            }

            info, err := stream.Info(ctx)
            if err != nil {
                log.Printf("获取 Stream 信息失败: %v", err)
                continue
            }

            currentMsgs := info.State.Msgs
            rate := (currentMsgs - lastMsgs) / 10 // 每秒速率

            fmt.Printf("[%s] msgs=%d bytes=%s rate=%d/s consumers=%d\n",
                streamName,
                currentMsgs,
                formatBytes(info.State.Bytes),
                rate,
                info.State.Consumers,
            )

            lastMsgs = currentMsgs
        }
    }
}

func formatBytes(b uint64) string {
    const unit = 1024
    if b < unit {
        return fmt.Sprintf("%d B", b)
    }
    div, exp := uint64(unit), 0
    for n := b / unit; n >= unit; n /= unit {
        div *= unit
        exp++
    }
    return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}
```

---

## 总结

```
Stream 核心要点速览：

  存储         File（推荐生产）或 Memory（高性能临时）
  Retention    Limits（日志）/ WorkQueue（任务）/ Interest（广播）
  Replicas     1（开发）/ 3（生产）/ 5（高可用）  [必须奇数]
  去重         设置 Duplicates + 发布时携带 Nats-Msg-Id
  超限行为     DiscardOld（默认，保护新数据）/ DiscardNew（保护旧数据）

  常见配置组合：
  ┌──────────────┬──────────────┬──────────────┬──────────────┐
  │ 场景         │ Retention    │ MaxAge       │ 其他         │
  ├──────────────┼──────────────┼──────────────┼──────────────┤
  │ 事件日志     │ Limits       │ 30d          │ DenyDelete   │
  │ 任务队列     │ WorkQueue    │ 24h          │ DiscardNew   │
  │ 设备状态     │ Limits       │ 7d           │ MaxMsgsPerSub│
  │ 实时指标     │ Limits       │ 5m           │ MemoryStorage│
  └──────────────┴──────────────┴──────────────┴──────────────┘
```

下一篇：[06-JetStream Consumers 详解](./06-JetStream-Consumers.md) — 深入讲解 Pull/Push Consumer、Ack 策略、重试机制和高并发消费模式。

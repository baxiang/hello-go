# JetStream Consumers 详解

> 系列导航：[01-概念与架构](#) | [02-集群与高可用](#) | [03-安全认证](#) | [04-JetStream 基础](#) | [05-Streams](./05-JetStream-Streams.md) | **06-Consumers**

---

## 目录

1. [Consumer 是什么](#1-consumer-是什么)
2. [Durable vs Ephemeral Consumer](#2-durable-vs-ephemeral-consumer)
3. [Push Consumer 详解](#3-push-consumer-详解)
4. [Pull Consumer 详解](#4-pull-consumer-详解)
5. [Deliver Policy（消费起点）全解](#5-deliver-policy消费起点全解)
6. [Ack Policy 全解](#6-ack-policy-全解)
7. [Ack 操作全解](#7-ack-操作全解)
8. [MaxDeliver 和 AckWait（重试机制）](#8-maxdeliver-和-ackwait重试机制)
9. [Consumer Filter（按 Subject 过滤）](#9-consumer-filter按-subject-过滤)
10. [并发消费（多 Client 同一 Pull Consumer）](#10-并发消费多-client-同一-pull-consumer)
11. [Ordered Consumer（有序消费）](#11-ordered-consumer有序消费)
12. [Go 代码示例](#12-go-代码示例)

---

## 1. Consumer 是什么

### 1.1 核心定义

Consumer 是 JetStream Stream 上的一个**读取游标**（read cursor）。你可以把它理解为一个有状态的书签——它记录了"读到哪里了"以及"哪些消息已经被确认处理"。

```
Stream: ORDERS
┌────────────────────────────────────────────────────────────────┐
│  seq=1    seq=2    seq=3    seq=4    seq=5    seq=6    seq=7   │
│  [msg]    [msg]    [msg]    [msg]    [msg]    [msg]    [msg]   │
└────────────────────────────────────────────────────────────────┘
     ▲                  ▲                           ▲
     │                  │                           │
Consumer-A           Consumer-B                Consumer-C
(已读到 seq=1,       (已读到 seq=3,             (从最新开始，
 等待处理)            已 Ack seq=1,2)            还未读任何消息)
```

### 1.2 Consumer 的核心职责

- **记录消费进度**：知道下一条要投递的消息序号（`DeliverSeq`）
- **跟踪未确认消息**：记录已投递但未被 Ack 的消息（待重试）
- **实现重试逻辑**：超过 AckWait 未收到 Ack 则重新投递
- **过滤消息**：可以只消费 Stream 中特定 subject 的消息

### 1.3 ConsumerConfig 关键字段概览

```go
type ConsumerConfig struct {
    // 基础标识
    Name           string          // Consumer 名称（Durable 时必填）
    Durable        string          // 持久化名称（同 Name，历史遗留）
    Description    string          // 描述

    // 消费起点
    DeliverPolicy  DeliverPolicy   // 从哪里开始消费
    OptStartSeq    uint64          // DeliverByStartSequence 时指定序号
    OptStartTime   *time.Time      // DeliverByStartTime 时指定时间

    // 消息过滤
    FilterSubject  string          // 单 subject 过滤（旧 API）
    FilterSubjects []string        // 多 subject 过滤（新 API）

    // 确认策略
    AckPolicy      AckPolicy       // Explicit / None / All
    AckWait        time.Duration   // Ack 超时（超时后重投）
    MaxDeliver     int             // 最大投递次数（超出后 Term）

    // Push Consumer 专用
    DeliverSubject string          // Push 投递目标 subject
    DeliverGroup   string          // Push 队列组（实现负载均衡）
    FlowControl    bool            // 流量控制
    IdleHeartbeat  time.Duration   // 心跳间隔

    // Pull Consumer 专用
    MaxWaiting     int             // 最大等待中的 Fetch 请求数
    MaxRequestBatch int            // 单次 Fetch 最大消息数
    MaxRequestExpires time.Duration // Fetch 请求最大超时

    // 其他
    ReplayPolicy   ReplayPolicy    // Instant / Original
    RateLimit      uint64          // 投递速率限制（bits/s）
    MaxAckPending  int             // 最大待确认消息数（流控）
    HeadersOnly    bool            // 只投递消息头，不投递 payload
    Backoff        []time.Duration // 重试退避策略
    Metadata       map[string]string // 用户自定义元数据
}
```

---

## 2. Durable vs Ephemeral Consumer

### 2.1 Durable Consumer（持久化消费者）

Durable Consumer 有名称，消费状态持久化在 Server 端，客户端断开重连后可以继续从断点消费。

```
Durable Consumer 生命周期：

  时刻 T1：创建 Consumer "my-worker"
  ┌──────────────────────────────────────┐
  │  Consumer: my-worker                 │
  │  状态: DeliverSeq=1, AckSeq=0        │
  └──────────────────────────────────────┘

  时刻 T2：消费到 seq=50，全部 Ack
  ┌──────────────────────────────────────┐
  │  Consumer: my-worker                 │
  │  状态: DeliverSeq=51, AckSeq=50      │
  └──────────────────────────────────────┘

  时刻 T3：客户端断线（网络故障）
  ┌──────────────────────────────────────┐
  │  Consumer: my-worker（仍在 Server）  │
  │  状态持久化，不丢失                  │
  └──────────────────────────────────────┘

  时刻 T4：客户端重连，绑定同名 Consumer
  ┌──────────────────────────────────────┐
  │  Consumer: my-worker                 │
  │  从 seq=51 继续消费 ✓               │
  └──────────────────────────────────────┘
```

**适用场景**：生产环境所有需要可靠消费的场景。

### 2.2 Ephemeral Consumer（临时消费者）

Ephemeral Consumer 没有 Durable 名称，客户端断开一段时间（默认 5 秒）后 Server 自动删除。

```
Ephemeral Consumer 生命周期：

  客户端连接 ──▶ 自动创建 Consumer（随机名称）
                     │
               客户端断线
                     │
               等待 InactiveThreshold（默认 5s）
                     │
               自动删除 Consumer ✓
```

**适用场景**：
- 临时订阅，只需消费当前最新消息
- 开发调试
- 一次性数据导出

### 2.3 对比表

| 维度               | Durable Consumer       | Ephemeral Consumer       |
|--------------------|------------------------|--------------------------|
| 名称               | 必须有（唯一）         | 自动生成                 |
| 断线续传           | 支持                   | 不支持（断后删除）       |
| 服务器端持久化     | 是                     | 否                       |
| 多客户端共享       | 支持                   | 不推荐                   |
| 适用环境           | 生产                   | 开发/临时                |
| 资源占用           | 持续占用（需手动删除） | 自动回收                 |

---

## 3. Push Consumer 详解

### 3.1 Push Consumer 工作模式

Push Consumer 让 Server 主动将消息**推送**到指定的 NATS subject，客户端订阅该 subject 接收消息：

```
Push Consumer 数据流：

  Stream: ORDERS
  ┌──────────────────────────────────┐
  │  seq=1  seq=2  seq=3  seq=4     │
  └──────────────────────────────────┘
          │
          │ Server 主动推送
          ▼
  subject: "_INBOX.orders.push.abc"
          │
  ┌───────┴──────────┐
  │                  │
  Client A         Client B
  (订阅推送 subject)
```

### 3.2 DeliverSubject 配置

```go
// Push Consumer 配置
cons, err := js.CreateOrUpdateConsumer(ctx, "ORDERS",
    jetstream.ConsumerConfig{
        Name:           "orders-push-consumer",
        Durable:        "orders-push-consumer",
        DeliverSubject: "orders.push.delivery", // Server 推送到此 subject
        DeliverGroup:   "push-workers",         // 队列组（多客户端负载均衡）
        AckPolicy:      jetstream.AckExplicitPolicy,
        AckWait:        30 * time.Second,
        DeliverPolicy:  jetstream.DeliverNewPolicy,
    },
)
```

### 3.3 Push Consumer 适用场景

**优点**：
- 低延迟：消息到达 Stream 后立即推送，无需轮询
- 简单：订阅一个 subject 即可，无需管理 Fetch 逻辑

**缺点**：
- 背压控制困难：Server 持续推送，消费者来不及处理时缓冲区膨胀
- 多客户端时需要配置 `DeliverGroup` 才能实现负载均衡
- 客户端处理能力不足时容易 OOM

```
Push Consumer 适用场景：
✓ 消费速度 >= 生产速度的场景
✓ 实时通知（消息量不大）
✓ 监控告警推送
✗ 高吞吐批处理（改用 Pull Consumer）
✗ 消费速度不稳定的场景
```

### 3.4 流控（FlowControl 和 IdleHeartbeat）

```go
cons, err := js.CreateOrUpdateConsumer(ctx, "ORDERS",
    jetstream.ConsumerConfig{
        Name:           "orders-push-fc",
        DeliverSubject: "_push.orders",
        AckPolicy:      jetstream.AckExplicitPolicy,
        MaxAckPending:  1000,                   // 最多 1000 条未 Ack 消息在途
        FlowControl:    true,                   // 启用流控（Server 等待客户端应答）
        IdleHeartbeat:  10 * time.Second,       // 无消息时的心跳间隔
    },
)
```

`FlowControl` 启用后，Server 会定期发送流控消息，客户端必须响应，否则 Server 暂停推送。这防止了消费者被消息淹没。

`IdleHeartbeat` 在无新消息时让 Server 定期发送心跳，客户端可以检测 Consumer 是否健康。

### 3.5 Push Consumer 推送失败处理

```go
// 订阅推送 subject，手动处理消息
sub, err := nc.QueueSubscribeSync("_push.orders", "push-workers")
if err != nil {
    log.Fatal(err)
}

for {
    msg, err := sub.NextMsgWithContext(ctx)
    if err != nil {
        log.Printf("等待消息失败: %v", err)
        continue
    }

    // 判断是否为 JetStream 控制消息（心跳、流控）
    if msg.Header.Get("Status") != "" {
        status := msg.Header.Get("Status")
        if status == "100" {
            // 流控消息，需要响应
            msg.Respond(nil)
        }
        continue
    }

    // 处理业务消息
    if err := processMessage(msg.Data); err != nil {
        // Nak：立即重试
        msg.Nak()
        continue
    }
    msg.Ack()
}
```

---

## 4. Pull Consumer 详解

### 4.1 Pull Consumer 工作模式（推荐）

Pull Consumer 由客户端**主动拉取**消息，完全由消费者控制消费节奏：

```
Pull Consumer 数据流：

  Stream: ORDERS
  ┌──────────────────────────────────┐
  │  seq=1  seq=2  seq=3  ...       │
  └──────────────────────────────────┘
          ▲
          │ 客户端主动 Fetch
          │
  ┌───────────────────────┐
  │ Pull Consumer         │
  │ "orders-worker"       │◀── Fetch(batch=10) ── Client
  │ (Server 端状态)        │
  └───────────────────────┘
```

**Pull Consumer 的优势**：
- **背压天然可控**：消费者按自身处理能力拉取，不会被压垮
- **批量处理**：一次 Fetch 多条，减少网络往返
- **横向扩展**：多个客户端共享同一 Pull Consumer，实现负载均衡
- **灵活的超时控制**：每次 Fetch 可以指定等待时间

### 4.2 三种 Fetch 方式

#### 4.2.1 Fetch（最常用）

```go
cons, _ := js.Consumer(ctx, "ORDERS", "orders-worker")

// Fetch 最多 10 条，等待最多 5 秒
msgs, err := cons.Fetch(10, jetstream.FetchMaxWait(5*time.Second))
if err != nil {
    log.Printf("Fetch 失败: %v", err)
    return
}

for msg := range msgs.Messages() {
    processMsg(msg)
    msg.Ack()
}

if err := msgs.Error(); err != nil {
    log.Printf("消息流错误: %v", err)
}
```

Fetch 会阻塞等待，直到：
- 收到 `batch` 条消息
- 等待超时（`FetchMaxWait`）
- Stream 中没有更多消息（`ErrEndOfData`）

#### 4.2.2 FetchNoWait（非阻塞）

```go
// 立即返回当前可用消息，不等待
msgs, err := cons.FetchNoWait(10)
if err != nil {
    log.Printf("FetchNoWait 失败: %v", err)
    return
}

count := 0
for msg := range msgs.Messages() {
    count++
    msg.Ack()
}

if count == 0 {
    // 当前 Stream 为空，可以休眠一段时间再试
    time.Sleep(100 * time.Millisecond)
}
```

适用于：需要精确控制轮询时机，不希望阻塞的场景。

#### 4.2.3 FetchBytes（按字节数拉取）

```go
// 拉取最多 1MB 的消息
msgs, err := cons.FetchBytes(1*1024*1024,
    jetstream.FetchMaxWait(5*time.Second),
)
```

适用于：消息大小不均匀，希望按数据量而非条数控制批次的场景（如：批量写入数据库）。

### 4.3 Fetch Batch 拉取性能

```
Batch Size 对性能的影响：

  batch=1:
    每条消息一次网络往返
    吞吐量低，延迟低
    适合：低吞吐、处理耗时不稳定

  batch=50:
    每 50 条一次网络往返
    吞吐量中等，延迟适中（推荐起点）
    适合：大多数场景

  batch=500:
    每 500 条一次网络往返
    吞吐量高，单次处理耗时较长
    适合：批量写入 DB、聚合计算

  典型吞吐量参考（单 Consumer，1KB 消息）：
  ┌──────────┬──────────────────┬─────────────┐
  │ batch    │ 吞吐量           │ 延迟 p99    │
  ├──────────┼──────────────────┼─────────────┤
  │ 1        │ ~5,000 msg/s     │ 200 µs      │
  │ 10       │ ~30,000 msg/s    │ 500 µs      │
  │ 50       │ ~100,000 msg/s   │ 2 ms        │
  │ 200      │ ~200,000 msg/s   │ 8 ms        │
  └──────────┴──────────────────┴─────────────┘
```

### 4.4 MaxWaiting 配置

`MaxWaiting` 限制同时等待中的 Fetch 请求数量（默认 512）。当多个客户端并发 Fetch 且 Stream 暂时为空时，服务器会将这些请求排队等待，直到有新消息到来。

```go
cons, err := js.CreateOrUpdateConsumer(ctx, "TASKS",
    jetstream.ConsumerConfig{
        Name:            "task-worker",
        Durable:         "task-worker",
        AckPolicy:       jetstream.AckExplicitPolicy,
        MaxWaiting:      100,  // 最多 100 个并发 Fetch 在等待
        MaxRequestBatch: 50,   // 每次 Fetch 最多 50 条
    },
)
```

---

## 5. Deliver Policy（消费起点）全解

Deliver Policy 决定 Consumer **从 Stream 的哪个位置开始**消费消息。

```
Stream 消息序列：
  seq: 1  2  3  4  5  6  7  8  9  10  ...  100  101  102(最新)
       ▲                    ▲                         ▲
       │                    │                         │
  DeliverAll           某个时间点                DeliverNew
  (从头开始)          DeliverByStartTime         (仅新消息)
```

### 5.1 DeliverAll（从头开始，默认值）

```go
DeliverPolicy: jetstream.DeliverAllPolicy
```

从 Stream 中**第一条消息**开始消费。适用于：
- 数据回放/重处理
- 首次启动的批处理作业
- 消费历史全量数据

### 5.2 DeliverLast（最新一条）

```go
DeliverPolicy: jetstream.DeliverLastPolicy
```

跳过所有历史消息，只从**当前最新的一条**开始。适用于：
- 只关心当前状态的监控程序
- 启动时不想处理历史积压

### 5.3 DeliverLastPerSubject（每个 Subject 最新一条）

```go
DeliverPolicy: jetstream.DeliverLastPerSubjectPolicy
```

对 Stream 中每个 subject，各取**最新的一条**，然后继续消费新消息。这是实现"设备最新状态快照"的关键策略：

```
示例：Stream 包含 device.A.status 和 device.B.status

  历史消息：
  seq=1: device.A.status (online)
  seq=2: device.B.status (offline)
  seq=3: device.A.status (offline)   ← A 的最新
  seq=4: device.B.status (online)    ← B 的最新
  seq=5: device.A.status (online)    ← A 的最新（更新）

  DeliverLastPerSubject 会投递：
    seq=5: device.A.status (online)
    seq=4: device.B.status (online)
    然后继续投递 seq=6, 7, 8... (新消息)
```

### 5.4 DeliverNew（只消费新消息）

```go
DeliverPolicy: jetstream.DeliverNewPolicy
```

Consumer 创建后，**只消费新写入的消息**，完全忽略历史消息。适用于：
- 实时事件处理，不需要历史数据
- 日志实时监控

### 5.5 DeliverByStartTime（按时间点）

```go
startTime := time.Now().Add(-1 * time.Hour) // 从 1 小时前开始
DeliverPolicy: jetstream.DeliverByStartTimePolicy,
OptStartTime:  &startTime,
```

从指定时间点之后的第一条消息开始消费。适用于：
- 故障恢复（从故障发生时间点重新处理）
- 按时间段的数据导出

### 5.6 DeliverByStartSequence（按序号）

```go
DeliverPolicy: jetstream.DeliverByStartSequencePolicy,
OptStartSeq:   12345,  // 从 seq=12345 开始（含）
```

从精确的序号位置开始消费。适用于：
- 精确断点续传
- 已知序号的重试场景

---

## 6. Ack Policy 全解

Ack Policy 定义消息确认的方式，影响消息可靠性和吞吐量之间的权衡。

### 6.1 AckExplicit（推荐）

```go
AckPolicy: jetstream.AckExplicitPolicy
```

每条消息必须**单独显式确认**。这是最安全的策略，也是默认值。

```
AckExplicit 示例：

  Server 投递：  msg1  msg2  msg3  msg4  msg5
                  │     │     │     │     │
  处理结果：     Ack   Nak   Ack  (超时)  Ack
                  │     │     │     │     │
  Server 操作：  删除  重投  删除  重投   删除
```

### 6.2 AckNone（不确认）

```go
AckPolicy: jetstream.AckNonePolicy
```

消息投递后**不需要确认**，Server 认为投递即成功。

```
AckNone 示例：

  Server 投递：  msg1  msg2  msg3
                  │     │     │
  处理结果：     (任意，Server 不关心)
                  │     │     │
  Server 操作：  视为成功，不重试
```

**适用场景**：
- 可以接受消息丢失的指标推送
- 有 Ordered Consumer 的有序消费（见第 11 节）
- 超高吞吐场景，Ack 开销明显

### 6.3 AckAll（累积确认）

```go
AckPolicy: jetstream.AckAllPolicy
```

Ack 最新的消息 N，则隐式确认所有 seq <= N 的消息。类似 TCP 的累积确认：

```
AckAll 示例：

  Server 投递：  msg1  msg2  msg3  msg4  msg5
  序号：          1     2     3     4     5

  客户端处理完 msg5，调用 msg5.Ack()
  等价于同时 Ack: msg1, msg2, msg3, msg4, msg5 ✓

  适合：严格顺序处理的批量确认，减少 Ack 次数
  风险：如果 msg3 处理失败但没有 Nak，调用 msg5.Ack() 后 msg3 也被认为成功
```

---

## 7. Ack 操作全解

### 7.1 五种 Ack 操作总览

```
消息处理决策树：

                 收到消息
                    │
            ┌───────┴────────┐
            │  能否处理？    │
            └───────┬────────┘
                    │
          ┌─────────┼──────────┐
          │         │          │
         是        否        不确定
          │         │          │
         Ack       Nak        InProgress
       (成功)    (立即重试) (延长超时)
                    │
          ┌─────────┴──────────┐
          │  是否应该重试？    │
          └─────────┬──────────┘
                    │
          ┌─────────┼──────────┐
          │         │          │
        重试      延迟重试    放弃
          Nak   NakWithDelay  Term
```

### 7.2 Ack（处理成功）

```go
msg.Ack()
```

告诉 Server：**消息已成功处理**。Server 不再重试，从待确认列表中移除。

### 7.3 Nak（失败，立即重试）

```go
msg.Nak()
```

告诉 Server：**处理失败，请立即重新投递**。Server 会尽快（默认 1ms 延迟）重新投递该消息。

**注意**：频繁 Nak 会造成"重试风暴"。如果知道错误需要一段时间才能恢复，应使用 `NakWithDelay`。

### 7.4 NakWithDelay（延迟重试）

```go
// 等待 30 秒后再重试（适合下游服务暂时不可用）
msg.NakWithDelay(30 * time.Second)

// 等待 5 分钟后重试（适合需要人工干预的情况）
msg.NakWithDelay(5 * time.Minute)
```

**最佳实践**：使用指数退避延迟：

```go
func retryDelay(deliverCount int) time.Duration {
    base := time.Second
    max := 5 * time.Minute
    delay := base * time.Duration(1<<uint(deliverCount-1)) // 指数退避
    if delay > max {
        delay = max
    }
    return delay
}

// 根据投递次数计算延迟
meta, _ := msg.Metadata()
delay := retryDelay(int(meta.NumDelivered))
msg.NakWithDelay(delay)
```

### 7.5 InProgress（延长 Ack 超时）

```go
// 每次调用重置 AckWait 计时器
msg.InProgress()
```

告诉 Server：**消息正在处理，请延长超时时间**。适用于处理时间可能超过 `AckWait` 的长任务（如：视频转码、大文件处理）。

**最佳实践**：在长任务中定期发送心跳：

```go
func processLongTask(ctx context.Context, msg jetstream.Msg) error {
    // 启动心跳 goroutine
    heartbeatDone := make(chan struct{})
    go func() {
        ticker := time.NewTicker(25 * time.Second) // AckWait = 30s，每 25s 发一次
        defer ticker.Stop()
        for {
            select {
            case <-ticker.C:
                msg.InProgress()
            case <-heartbeatDone:
                return
            case <-ctx.Done():
                return
            }
        }
    }()

    defer close(heartbeatDone)

    // 执行长任务
    if err := doHeavyWork(msg.Data()); err != nil {
        msg.Nak()
        return err
    }

    msg.Ack()
    return nil
}
```

### 7.6 Term（放弃，不再重试）

```go
msg.Term()
// 或者带原因（会记录在 Advisory 事件中）
msg.TermWithReason("消息格式错误，无法解析")
```

告诉 Server：**此消息无法处理，永久放弃，不再重试**。消息会被移出待确认列表，即使还没达到 `MaxDeliver`。

**适用场景**：
- 消息格式错误，重试也无法成功
- 业务规则明确排除（如：订单已取消，不再处理发货任务）
- 发现消息是重复的（应对非 Server 端去重的场景）

Term 后 Server 会发布一条 Advisory 消息到 `$JS.EVENT.ADVISORY.CONSUMER.MSG_TERMINATED.*`，可以订阅用于告警和死信队列处理。

---

## 8. MaxDeliver 和 AckWait（重试机制）

### 8.1 重试流程全图

```
重试机制完整流程：

  Server                          Client
    │                               │
    │── deliver(msg, seq=5) ───────▶│
    │   (DeliverCount=1)            │  处理中...
    │                               │
    │   [等待 AckWait=30s]          │
    │                               │
    │── redeliver(msg, seq=5) ─────▶│  (超时，重试 #1)
    │   (DeliverCount=2)            │  处理中...
    │                               │
    │   [等待 AckWait=30s]          │
    │                               │
    │── redeliver(msg, seq=5) ─────▶│  (超时，重试 #2)
    │   (DeliverCount=3)            │  msg.Nak()
    │                               │
    │   [立即或 NakDelay 后]        │
    │── redeliver(msg, seq=5) ─────▶│  (重试 #3)
    │   (DeliverCount=4)            │  ...
    │                               │
    │   [达到 MaxDeliver=5]         │
    │── (不再投递，发布 Advisory)   │
```

### 8.2 配置示例

```go
cons, err := js.CreateOrUpdateConsumer(ctx, "ORDERS",
    jetstream.ConsumerConfig{
        Name:       "orders-processor",
        Durable:    "orders-processor",
        AckPolicy:  jetstream.AckExplicitPolicy,
        AckWait:    30 * time.Second,    // 30 秒内必须 Ack
        MaxDeliver: 5,                   // 最多投递 5 次（1 次原始 + 4 次重试）
        // 指数退避重试
        Backoff: []time.Duration{
            1 * time.Second,   // 第 1 次重试等待 1s
            10 * time.Second,  // 第 2 次重试等待 10s
            30 * time.Second,  // 第 3 次重试等待 30s
            2 * time.Minute,   // 第 4 次重试等待 2m
        },
    },
)
```

### 8.3 Backoff 退避策略

`Backoff` 字段指定每次重试前的等待时间（从第 2 次投递开始）：

```
Backoff: [1s, 10s, 30s, 2min]

  第 1 次投递（原始）：立即
  第 2 次投递（重试1）：等待 1s
  第 3 次投递（重试2）：等待 10s
  第 4 次投递（重试3）：等待 30s
  第 5 次投递（重试4）：等待 2min
  第 6 次以上：使用最后一个 Backoff 值（2min）
```

### 8.4 达到 MaxDeliver 后的处理

当消息达到最大投递次数后，Server 不再投递，并发布一条 Advisory 消息。可以订阅此事件实现死信队列（Dead Letter Queue）：

```go
// 订阅消息终止事件（Dead Letter Queue 效果）
nc.Subscribe("$JS.EVENT.ADVISORY.CONSUMER.MSG_TERMINATED.ORDERS.orders-processor",
    func(msg *nats.Msg) {
        var advisory struct {
            Stream   string `json:"stream"`
            Consumer string `json:"consumer"`
            Subject  string `json:"subject"`
            Seq      uint64 `json:"seq"`
        }
        json.Unmarshal(msg.Data, &advisory)
        log.Printf("消息达到最大重试次数，进入死信队列: stream=%s seq=%d",
            advisory.Stream, advisory.Seq)
        // 可以将消息写入死信 Stream 或发送告警
    },
)
```

---

## 9. Consumer Filter（按 Subject 过滤）

### 9.1 单 Subject 过滤

```go
// 只消费 orders.created 消息
cons, err := js.CreateOrUpdateConsumer(ctx, "ORDERS",
    jetstream.ConsumerConfig{
        Name:          "created-only",
        FilterSubject: "orders.created",
        AckPolicy:     jetstream.AckExplicitPolicy,
    },
)
```

### 9.2 多 Subject 过滤（新 API）

```go
// 消费 orders.created 和 orders.updated，跳过 orders.deleted
cons, err := js.CreateOrUpdateConsumer(ctx, "ORDERS",
    jetstream.ConsumerConfig{
        Name: "create-update-consumer",
        FilterSubjects: []string{
            "orders.created",
            "orders.updated",
            "orders.*.payment",  // 通配符
        },
        AckPolicy: jetstream.AckExplicitPolicy,
    },
)
```

**注意**：`FilterSubjects`（复数）是 NATS 2.10+ 新增的功能，旧版本只支持 `FilterSubject`（单数）。

### 9.3 使用过滤实现按类型分流

```
Stream: ORDERS（捕获 orders.>）
         │
    ┌────┴──────────────────────────────────┐
    │                                       │
Consumer A                          Consumer B
FilterSubject: "orders.created"     FilterSubject: "orders.cancelled"
（新订单处理器）                    （取消订单处理器）
```

---

## 10. 并发消费（多 Client 同一 Pull Consumer）

### 10.1 并发消费原理

多个客户端可以同时向同一个 Pull Consumer 发起 Fetch 请求，Server 会将不同的消息分发给不同的客户端，实现负载均衡：

```
并发消费架构：

  Stream: TASKS
  ┌─────────────────────────────────────────────────────┐
  │  task1  task2  task3  task4  task5  task6  task7   │
  └─────────────────────────────────────────────────────┘
                         │
              Consumer: task-worker
                         │
         ┌───────────────┼───────────────┐
         │               │               │
      Worker-1        Worker-2        Worker-3
    Fetch(task1,     Fetch(task3,    Fetch(task5,
           task2)           task4)          task6)
```

### 10.2 并发消费的关键点

```
并发消费注意事项：

  ✓ 同一条消息只会被投递给一个 Worker（不会重复）
  ✓ Worker 异常（未 Ack 超时）后消息会被重新投递给其他 Worker
  ✓ 不保证消息的全局有序性（各 Worker 并发处理）
  ✓ MaxAckPending 控制总的在途消息数（所有 Worker 合计）

  MaxAckPending 建议值：
    = workerCount × batchSize × 安全系数(2~3)
    例：10 个 Worker，每次 Fetch 50 条 → MaxAckPending = 1500
```

### 10.3 Worker Pool 实现

```go
func startWorkerPool(ctx context.Context, js jetstream.JetStream,
    streamName, consumerName string, workerCount int) {
    
    cons, err := js.Consumer(ctx, streamName, consumerName)
    if err != nil {
        log.Fatalf("获取 Consumer 失败: %v", err)
    }

    var wg sync.WaitGroup
    for i := 0; i < workerCount; i++ {
        wg.Add(1)
        go func(workerID int) {
            defer wg.Done()
            worker(ctx, cons, workerID)
        }(i)
    }

    wg.Wait()
    log.Println("所有 Worker 已停止")
}

func worker(ctx context.Context, cons jetstream.Consumer, id int) {
    log.Printf("Worker-%d 启动", id)
    for {
        select {
        case <-ctx.Done():
            log.Printf("Worker-%d 停止", id)
            return
        default:
        }

        msgs, err := cons.Fetch(50,
            jetstream.FetchMaxWait(5*time.Second),
        )
        if err != nil {
            if errors.Is(err, context.Canceled) {
                return
            }
            log.Printf("Worker-%d Fetch 失败: %v", id, err)
            time.Sleep(time.Second)
            continue
        }

        for msg := range msgs.Messages() {
            if err := processTask(ctx, msg); err != nil {
                meta, _ := msg.Metadata()
                delay := retryDelay(int(meta.NumDelivered))
                log.Printf("Worker-%d 处理失败，%v 后重试: %v", id, delay, err)
                msg.NakWithDelay(delay)
                continue
            }
            msg.Ack()
        }

        if err := msgs.Error(); err != nil {
            log.Printf("Worker-%d 消息流错误: %v", id, err)
        }
    }
}

func processTask(ctx context.Context, msg jetstream.Msg) error {
    // 模拟处理
    var task map[string]interface{}
    if err := json.Unmarshal(msg.Data(), &task); err != nil {
        // 格式错误，直接 Term 不重试
        msg.TermWithReason("invalid JSON")
        return nil // 返回 nil 避免上层再调用 Nak
    }
    // ... 业务处理
    return nil
}
```

---

## 11. Ordered Consumer（有序消费）

### 11.1 什么是 Ordered Consumer

Ordered Consumer 是一种特殊的 **Ephemeral Push Consumer**，专门用于需要严格按顺序、不丢消息的只读场景（如：数据导出、日志分析）。

**特性**：
- 严格按 sequence 顺序投递
- 自动处理重连和重传（客户端断线后自动恢复）
- `AckPolicy: AckNone`（不需要手动 Ack）
- 一个 Ordered Consumer 只能有一个消费者（不支持并发）

```
Ordered Consumer vs 普通 Consumer：

  普通 Consumer：
    消费者可以 Nak → 消息重新投递（可能打乱顺序）
    多消费者并发 → 消息乱序

  Ordered Consumer：
    检测到消息序号跳跃 → 自动重建 Consumer 从断点续传
    严格保证 seq=1, 2, 3, 4, 5... 的顺序交付
```

### 11.2 Ordered Consumer 使用示例

```go
// 创建 Ordered Consumer（简化 API）
cons, err := js.OrderedConsumer(ctx, "ORDERS",
    jetstream.OrderedConsumerConfig{
        FilterSubjects: []string{"orders.>"},
        DeliverPolicy:  jetstream.DeliverAllPolicy,
        // 无需设置 AckPolicy（强制 AckNone）
    },
)
if err != nil {
    log.Fatalf("创建 Ordered Consumer 失败: %v", err)
}

// 使用 Consume 方法（推送模式，但客户端库内部管理）
cc, err := cons.Consume(func(msg jetstream.Msg) {
    meta, _ := msg.Metadata()
    fmt.Printf("seq=%d subject=%s data=%s\n",
        meta.Sequence.Stream,
        msg.Subject(),
        string(msg.Data()),
    )
    // 不需要调用 msg.Ack()
})
if err != nil {
    log.Fatalf("启动消费失败: %v", err)
}
defer cc.Stop()

// 等待消费完成
<-ctx.Done()
```

---

## 12. Go 代码示例

### 12.1 Pull Consumer 基础使用

```go
package main

import (
    "context"
    "errors"
    "fmt"
    "log"
    "time"

    "github.com/nats-io/nats.go"
    "github.com/nats-io/nats.go/jetstream"
)

func setupPullConsumer(ctx context.Context, js jetstream.JetStream) (jetstream.Consumer, error) {
    cons, err := js.CreateOrUpdateConsumer(ctx, "ORDERS",
        jetstream.ConsumerConfig{
            Name:        "orders-pull-worker",
            Durable:     "orders-pull-worker",
            Description: "订单处理 Pull Consumer",
            AckPolicy:   jetstream.AckExplicitPolicy,
            AckWait:     30 * time.Second,
            MaxDeliver:  5,
            Backoff: []time.Duration{
                2 * time.Second,
                10 * time.Second,
                30 * time.Second,
                2 * time.Minute,
            },
            DeliverPolicy:  jetstream.DeliverAllPolicy,
            FilterSubject:  "orders.>",
            MaxAckPending:  1000,
            MaxWaiting:     50,
            MaxRequestBatch: 100,
        },
    )
    if err != nil {
        return nil, fmt.Errorf("创建 Consumer 失败: %w", err)
    }

    info, _ := cons.Info(ctx)
    fmt.Printf("Consumer 创建成功: %s (待处理: %d)\n",
        info.Name, info.NumPending)

    return cons, nil
}

// simplePullConsume 简单的单循环消费
func simplePullConsume(ctx context.Context, cons jetstream.Consumer) {
    for {
        select {
        case <-ctx.Done():
            return
        default:
        }

        msgs, err := cons.Fetch(10, jetstream.FetchMaxWait(5*time.Second))
        if err != nil {
            if errors.Is(err, jetstream.ErrNoMessages) {
                continue // Stream 为空，继续等待
            }
            log.Printf("Fetch 错误: %v", err)
            time.Sleep(time.Second)
            continue
        }

        for msg := range msgs.Messages() {
            meta, err := msg.Metadata()
            if err != nil {
                log.Printf("获取消息元数据失败: %v", err)
                msg.Nak()
                continue
            }

            fmt.Printf("处理消息: subject=%s seq=%d deliverCount=%d\n",
                msg.Subject(),
                meta.Sequence.Stream,
                meta.NumDelivered,
            )

            if err := handleOrderMessage(msg.Data()); err != nil {
                log.Printf("处理失败 (尝试 %d/%d): %v",
                    meta.NumDelivered, 5, err)
                msg.NakWithDelay(retryDelay(int(meta.NumDelivered)))
                continue
            }

            msg.Ack()
        }

        if err := msgs.Error(); err != nil {
            log.Printf("消息批次错误: %v", err)
        }
    }
}

func handleOrderMessage(data []byte) error {
    // 业务处理逻辑
    fmt.Printf("  处理订单数据: %s\n", string(data))
    return nil
}

func retryDelay(deliverCount int) time.Duration {
    delays := []time.Duration{2 * time.Second, 10 * time.Second, 30 * time.Second, 2 * time.Minute}
    idx := deliverCount - 1
    if idx >= len(delays) {
        return delays[len(delays)-1]
    }
    if idx < 0 {
        return delays[0]
    }
    return delays[idx]
}
```

### 12.2 批量 Fetch 高性能消费

```go
// highThroughputConsumer 高吞吐批量消费
func highThroughputConsumer(ctx context.Context, js jetstream.JetStream) error {
    cons, err := js.CreateOrUpdateConsumer(ctx, "ORDERS",
        jetstream.ConsumerConfig{
            Name:            "orders-batch-worker",
            Durable:         "orders-batch-worker",
            AckPolicy:       jetstream.AckAllPolicy, // 累积确认，减少 Ack 次数
            AckWait:         60 * time.Second,
            MaxDeliver:      3,
            MaxAckPending:   5000,
            MaxRequestBatch: 500,
        },
    )
    if err != nil {
        return fmt.Errorf("创建 Consumer 失败: %w", err)
    }

    batchSize := 200
    var totalProcessed int64
    var lastReport time.Time = time.Now()

    for {
        select {
        case <-ctx.Done():
            log.Printf("停止，已处理 %d 条消息", totalProcessed)
            return nil
        default:
        }

        msgs, err := cons.Fetch(batchSize,
            jetstream.FetchMaxWait(3*time.Second),
        )
        if err != nil {
            if errors.Is(err, jetstream.ErrNoMessages) {
                time.Sleep(100 * time.Millisecond)
                continue
            }
            log.Printf("Fetch 错误: %v", err)
            time.Sleep(time.Second)
            continue
        }

        var lastMsg jetstream.Msg
        batchCount := 0

        for msg := range msgs.Messages() {
            // 批量处理（可以一次性写入 DB）
            processBatch(msg.Data())
            lastMsg = msg
            batchCount++
        }

        if msgs.Error() != nil {
            log.Printf("批次错误: %v", msgs.Error())
            continue
        }

        // 使用 AckAll 只确认最后一条，隐式确认批次中所有消息
        if lastMsg != nil {
            lastMsg.Ack()
        }

        totalProcessed += int64(batchCount)

        // 每 10 秒打印一次吞吐量统计
        if time.Since(lastReport) >= 10*time.Second {
            elapsed := time.Since(lastReport).Seconds()
            rate := float64(totalProcessed) / elapsed
            fmt.Printf("吞吐量: %.0f msg/s (累计: %d)\n", rate, totalProcessed)
            lastReport = time.Now()
            totalProcessed = 0
        }
    }
}

func processBatch(data []byte) {
    // 模拟批量写入数据库
    _ = data
}
```

### 12.3 错误重试模式（完整示例）

```go
// robustConsumer 带完整错误处理的生产级消费者
type MessageProcessor struct {
    cons       jetstream.Consumer
    maxRetries int
    dlqSubject string // 死信队列 subject
    js         jetstream.JetStream
}

func NewMessageProcessor(ctx context.Context, js jetstream.JetStream,
    stream, consumerName, dlqSubject string) (*MessageProcessor, error) {
    
    cons, err := js.CreateOrUpdateConsumer(ctx, stream,
        jetstream.ConsumerConfig{
            Name:      consumerName,
            Durable:   consumerName,
            AckPolicy: jetstream.AckExplicitPolicy,
            AckWait:   30 * time.Second,
            MaxDeliver: 5,
            Backoff: []time.Duration{
                1 * time.Second,
                5 * time.Second,
                15 * time.Second,
                60 * time.Second,
            },
        },
    )
    if err != nil {
        return nil, err
    }

    return &MessageProcessor{
        cons:       cons,
        maxRetries: 5,
        dlqSubject: dlqSubject,
        js:         js,
    }, nil
}

func (p *MessageProcessor) Run(ctx context.Context, handler func([]byte) error) {
    for {
        select {
        case <-ctx.Done():
            return
        default:
        }

        msgs, err := p.cons.Fetch(50, jetstream.FetchMaxWait(5*time.Second))
        if err != nil {
            if !errors.Is(err, context.Canceled) {
                log.Printf("Fetch 错误: %v", err)
                time.Sleep(time.Second)
            }
            continue
        }

        for msg := range msgs.Messages() {
            p.processWithRetry(ctx, msg, handler)
        }
    }
}

func (p *MessageProcessor) processWithRetry(
    ctx context.Context,
    msg jetstream.Msg,
    handler func([]byte) error,
) {
    meta, err := msg.Metadata()
    if err != nil {
        log.Printf("获取消息元数据失败，直接 Nak: %v", err)
        msg.Nak()
        return
    }

    deliverCount := int(meta.NumDelivered)

    // 达到最大重试次数：发送到死信队列
    if deliverCount >= p.maxRetries {
        log.Printf("消息 seq=%d 达到最大重试次数，发送到 DLQ", meta.Sequence.Stream)
        p.sendToDLQ(ctx, msg, meta)
        msg.Term()
        return
    }

    // 执行业务处理
    err = handler(msg.Data())
    if err == nil {
        msg.Ack()
        return
    }

    // 判断错误类型
    var permanentErr *PermanentError
    if errors.As(err, &permanentErr) {
        // 永久性错误（如：数据格式错误）→ 直接 Term
        log.Printf("永久性错误，消息 seq=%d 进入 DLQ: %v",
            meta.Sequence.Stream, err)
        p.sendToDLQ(ctx, msg, meta)
        msg.TermWithReason(err.Error())
        return
    }

    // 临时性错误（如：下游服务不可用）→ 退避重试
    delay := retryDelay(deliverCount)
    log.Printf("临时错误，seq=%d 将在 %v 后重试 (第 %d/%d 次): %v",
        meta.Sequence.Stream, delay, deliverCount, p.maxRetries, err)
    msg.NakWithDelay(delay)
}

func (p *MessageProcessor) sendToDLQ(
    ctx context.Context,
    msg jetstream.Msg,
    meta *jetstream.MsgMetadata,
) {
    // 将原始消息发送到死信队列（携带失败信息）
    dlqPayload := map[string]interface{}{
        "original_subject":  msg.Subject(),
        "original_seq":      meta.Sequence.Stream,
        "original_data":     string(msg.Data()),
        "original_headers":  msg.Headers(),
        "deliver_count":     meta.NumDelivered,
        "failed_at":         time.Now().UTC(),
    }
    payload, _ := json.Marshal(dlqPayload)
    
    if _, err := p.js.Publish(ctx, p.dlqSubject, payload); err != nil {
        log.Printf("发送到 DLQ 失败: %v", err)
    }
}

// PermanentError 表示不可重试的永久性错误
type PermanentError struct {
    Reason string
}

func (e *PermanentError) Error() string {
    return fmt.Sprintf("permanent error: %s", e.Reason)
}
```

### 12.4 并发消费 Worker Pool（完整示例）

```go
package main

import (
    "context"
    "errors"
    "fmt"
    "log"
    "os"
    "os/signal"
    "sync"
    "syscall"
    "time"

    "github.com/nats-io/nats.go"
    "github.com/nats-io/nats.go/jetstream"
)

const (
    StreamName    = "ORDERS"
    ConsumerName  = "orders-concurrent-worker"
    WorkerCount   = 10
    FetchBatch    = 50
    FetchTimeout  = 5 * time.Second
)

func main() {
    nc, err := nats.Connect("nats://localhost:4222",
        nats.MaxReconnects(-1),
        nats.ReconnectWait(2*time.Second),
    )
    if err != nil {
        log.Fatalf("连接 NATS 失败: %v", err)
    }
    defer nc.Close()

    js, err := jetstream.New(nc)
    if err != nil {
        log.Fatalf("初始化 JetStream 失败: %v", err)
    }

    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    // 优雅退出
    go func() {
        sigCh := make(chan os.Signal, 1)
        signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
        <-sigCh
        log.Println("收到退出信号，正在优雅停止...")
        cancel()
    }()

    // 创建 Consumer
    cons, err := createConcurrentConsumer(ctx, js)
    if err != nil {
        log.Fatalf("创建 Consumer 失败: %v", err)
    }

    // 启动 Worker Pool
    log.Printf("启动 %d 个 Worker...", WorkerCount)
    runWorkerPool(ctx, cons, WorkerCount)
    log.Println("所有 Worker 已停止")
}

func createConcurrentConsumer(ctx context.Context, js jetstream.JetStream) (jetstream.Consumer, error) {
    return js.CreateOrUpdateConsumer(ctx, StreamName,
        jetstream.ConsumerConfig{
            Name:    ConsumerName,
            Durable: ConsumerName,
            AckPolicy:       jetstream.AckExplicitPolicy,
            AckWait:         30 * time.Second,
            MaxDeliver:      5,
            // MaxAckPending = workerCount * fetchBatch * 2
            MaxAckPending:   WorkerCount * FetchBatch * 2,
            MaxWaiting:      WorkerCount * 2,
            MaxRequestBatch: FetchBatch,
            Backoff: []time.Duration{
                2 * time.Second,
                10 * time.Second,
                30 * time.Second,
                2 * time.Minute,
            },
            DeliverPolicy: jetstream.DeliverAllPolicy,
        },
    )
}

func runWorkerPool(ctx context.Context, cons jetstream.Consumer, workerCount int) {
    var wg sync.WaitGroup

    // 统计 channel
    statsCh := make(chan workerStats, workerCount*10)

    // 启动 Worker
    for i := 0; i < workerCount; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            runSingleWorker(ctx, cons, id, statsCh)
        }(i)
    }

    // 启动统计 goroutine
    go func() {
        aggregateStats(ctx, statsCh, workerCount)
    }()

    wg.Wait()
    close(statsCh)
}

type workerStats struct {
    workerID  int
    processed int
    failed    int
    duration  time.Duration
}

func runSingleWorker(
    ctx context.Context,
    cons jetstream.Consumer,
    id int,
    statsCh chan<- workerStats,
) {
    log.Printf("[Worker-%02d] 启动", id)
    
    for {
        select {
        case <-ctx.Done():
            log.Printf("[Worker-%02d] 收到退出信号，停止", id)
            return
        default:
        }

        start := time.Now()
        processed, failed := fetchAndProcess(ctx, cons, id)

        if processed+failed > 0 {
            statsCh <- workerStats{
                workerID:  id,
                processed: processed,
                failed:    failed,
                duration:  time.Since(start),
            }
        }
    }
}

func fetchAndProcess(
    ctx context.Context,
    cons jetstream.Consumer,
    workerID int,
) (processed, failed int) {
    
    msgs, err := cons.Fetch(FetchBatch,
        jetstream.FetchMaxWait(FetchTimeout),
    )
    if err != nil {
        if !errors.Is(err, context.Canceled) &&
            !errors.Is(err, jetstream.ErrNoMessages) {
            log.Printf("[Worker-%02d] Fetch 错误: %v", workerID, err)
            time.Sleep(time.Second)
        }
        return 0, 0
    }

    for msg := range msgs.Messages() {
        if err := doWork(msg.Data()); err != nil {
            meta, _ := msg.Metadata()
            msg.NakWithDelay(retryDelay(int(meta.NumDelivered)))
            failed++
            continue
        }
        msg.Ack()
        processed++
    }

    return processed, failed
}

func doWork(data []byte) error {
    // 模拟业务处理（50ms）
    time.Sleep(50 * time.Millisecond)
    _ = data
    return nil
}

func aggregateStats(ctx context.Context, statsCh <-chan workerStats, workerCount int) {
    ticker := time.NewTicker(10 * time.Second)
    defer ticker.Stop()

    var totalProcessed, totalFailed int

    for {
        select {
        case <-ctx.Done():
            return
        case stat, ok := <-statsCh:
            if !ok {
                return
            }
            totalProcessed += stat.processed
            totalFailed += stat.failed
            _ = stat
        case <-ticker.C:
            fmt.Printf("\n=== 消费统计（最近 10 秒）===\n")
            fmt.Printf("  总处理:   %d 条\n", totalProcessed)
            fmt.Printf("  失败重试: %d 条\n", totalFailed)
            fmt.Printf("  吞吐量:   %.0f msg/s\n", float64(totalProcessed)/10.0)
            totalProcessed = 0
            totalFailed = 0
        }
    }
}
```

---

## 总结

```
Consumer 核心要点速览：

  类型选择：
    Durable    → 生产环境，断线续传
    Ephemeral  → 临时消费，调试
    Ordered    → 有序只读，数据导出

  拉取方式（Pull Consumer，推荐）：
    Fetch          → 等待直到有消息或超时（最常用）
    FetchNoWait    → 立即返回，不等待
    FetchBytes     → 按字节数控制批次

  消费起点（DeliverPolicy）：
    DeliverAll              → 从头消费（历史回放）
    DeliverNew              → 只消费新消息
    DeliverLast             → 从最新一条开始
    DeliverLastPerSubject   → 每个 Subject 的最新状态
    DeliverByStartTime      → 从指定时间点
    DeliverByStartSequence  → 从指定序号

  Ack 操作：
    Ack()              → 成功，不再投递
    Nak()              → 失败，立即重试
    NakWithDelay(d)    → 失败，延迟 d 后重试
    InProgress()       → 处理中，延长 AckWait
    Term()             → 放弃，进入 DLQ

  重试配置黄金组合：
    AckWait:   30s
    MaxDeliver: 5
    Backoff:   [1s, 10s, 30s, 2m]

  并发消费：
    MaxAckPending = workerCount × batchSize × 2~3
    MaxWaiting    = workerCount × 2
```

---

## 延伸阅读

- [NATS JetStream 官方文档](https://docs.nats.io/nats-concepts/jetstream)
- [nats.go 客户端库](https://github.com/nats-io/nats.go)
- [JetStream API 参考](https://docs.nats.io/reference/reference-protocols/nats_api_reference)
- 上一篇：[05-JetStream Streams 详解](./05-JetStream-Streams.md)

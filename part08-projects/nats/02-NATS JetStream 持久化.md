# NATS JetStream 持久化

JetStream 是 NATS 2.0 引入的持久化消息系统，提供消息存储和 Exactly-Once 语义。

## 2.1 启用 JetStream

```bash
# 启动带 JetStream 的 NATS 服务器
nats-server -js

# 或使用 Docker
docker run -d -p 4222:4222 -p 8222:8222 nats:latest -js
```

## 2.2 创建 Stream

### 基础 Stream

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/nats-io/nats.go"
    "github.com/nats-io/nats.go/jetstream"
)

func main() {
    nc, err := nats.Connect(nats.DefaultURL)
    if err != nil {
        log.Fatal(err)
    }
    defer nc.Close()

    // 创建 JetStream 上下文
    js, err := jetstream.New(nc)
    if err != nil {
        log.Fatal(err)
    }

    // 创建 Stream
    ctx := context.Background()
    
    stream, err := js.CreateStream(ctx, jetstream.StreamConfig{
        Name:      "ORDERS",
        Subjects: []string{"orders.>"},
        Storage:  jetstream.FileStorage,  // 文件存储
        // Storage: jetstream.MemoryStorage, // 内存存储
        Replicas: 1,  // 副本数
        MaxBytes: 1024 * 1024 * 100,  // 100MB
        MaxAge:   time.Hour * 24 * 7, // 保留 7 天
    })
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("创建 Stream: %s\n", stream.CachedInfo().Config.Name)
}
```

### Stream 配置选项

```go
stream, err := js.CreateStream(ctx, jetstream.StreamConfig{
    Name:     "EVENTS",
    Subjects: []string{"events.>"},
    
    // 存储配置
    Storage:  jetstream.FileStorage,
    Replicas: 3,  // 集群模式需要 3 副本
    
    // 保留策略
    Retention: jetstream.LimitsPolicy{
        MaxBytes:       1024 * 1024 * 1024,  // 1GB
        MaxAge:         time.Hour * 24 * 30, // 30 天
        MaxMsgs:        100000,
        MaxMsgSize:     1024 * 1024, // 1MB
    },
    
    // 丢弃策略
    Discard: jetstream.DiscardOld,  // 丢弃旧消息
    // Discard: jetstream.DiscardNew, // 丢弃新消息
    
    // 主题复制
    AllowRollup: true,
    AllowDirect: true,
})
```

## 2.3 发布持久化消息

### 基本发布

```go
// 发布到 JetStream
js, _ := jetstream.New(nc)
ctx := context.Background()

// 同步发布 - 等待确认
pubAck, err := js.Publish(ctx, "orders.created", []byte(`{"order_id": 1001}`))
if err != nil {
    log.Fatal(err)
}
fmt.Printf("消息已发布: %s, 序列号: %d\n", pubAck.Stream, pubAck.Sequence)
```

### 带选项的发布

```go
// 带消息头的发布
pubAck, err = js.Publish(ctx, "orders.created", []byte(`{"order_id": 1002}`),
    jetstream.WithHeader(nats.Header{
        "Content-Type": []string{"application/json"},
    }),
    jetstream.WithExpectStream("ORDERS"),  // 期望的 Stream
)

// 设置消息 ID (用于幂等)
pubAck, err = js.Publish(ctx, "orders.created", data,
    jetstream.WithMsgID("unique-id-123"),
)

// 等待确认
pubAck, err = js.Publish(ctx, "orders.created", data,
    jetstream.WithPublishWait(5*time.Second),
)
```

### 批量发布

```go
// 批量发布
for i := 0; i < 100; i++ {
    data, _ := json.Marshal(map[string]interface{}{
        "order_id": 1000 + i,
        "amount":   float64(i * 100),
    })
    js.Publish(ctx, "orders.created", data)
}
```

## 2.4 订阅持久化消息

### 基本消费

```go
// 消费消息
consume, err := js.Consume(
    "ORDERS",
    "order-processor",  // 消费者名称
    jetstream.ConsumeErrHandler(func(consumeCtx jetstream.ConsumeContext, err error) {
        fmt.Printf("消费错误: %v\n", err)
    }),
)
if err != nil {
    log.Fatal(err)
}

// 获取消息
msg, err := consume.Next()
if err != nil {
    log.Fatal(err)
}

fmt.Printf("收到消息: %s\n", string(msg.Data()))

// 确认消息
msg.Ack()
```

### 拉取消费

```go
// 创建消费者
consumer, err := js.CreateConsumer(ctx, "ORDERS", jetstream.ConsumerConfig{
    Name:           "order-processor",
    Durable:        true,
    FilterSubject:  "orders.created",
    DeliverPolicy:  jetstream.DeliverAll,
    AckPolicy:      jetstream.AckExplicit,
    AckWait:        30 * time.Second,
    MaxDeliver:     3,
    MaxAckPending:  100,
})

// 拉取消息
msgs, err := consumer.Fetch(10, jetstream.FetchMaxWait(5*time.Second))
if err != nil {
    log.Fatal(err)
}

for msg := range msgs {
    fmt.Printf("处理: %s\n", string(msg.Data()))
    msg.Ack()
}
```

### 推送消费

```go
// 创建推送消费者
consumer, err := js.CreateConsumer(ctx, "ORDERS", jetstream.ConsumerConfig{
    Name:          "order-pusher",
    Durable:       true,
    FilterSubject: "orders.created",
    DeliverSubject: "orders.delivered",  // 推送目标主题
})

// 消费消息
messages, err := consumer.Messages()
for msg := range messages {
    fmt.Printf("收到: %s\n", string(msg.Data()))
    msg.Ack()
}
```

## 2.5 消息确认机制

### 确认策略

```go
// 自动确认 (默认)
consume, _ := js.Consume("ORDERS", "consumer", jetstream.AckAuto())

// 显式确认
consume, _ := js.Consume("ORDERS", "consumer", jetstream.AckExplicit())

// 无确认
consume, _ := js.Consume("ORDERS", "consumer", jetstream.AckNone())
```

### 手动确认

```go
msg, _ := consume.Next()
fmt.Printf("处理中: %s\n", string(msg.Data()))

// 确认成功
msg.Ack()

// 确认失败 - 重新入队
msg.Nak()           // 重新入队
msg.Term()          // 终止，不再重试
msg.InProgress()    // 标记正在处理，延长超时
```

### 重试机制

```go
// 使用 JetStream 的重试机制
consumer, _ := js.CreateConsumer(ctx, "ORDERS", jetstream.ConsumerConfig{
    Name:          "order-processor",
    FilterSubject: "orders.created",
    AckPolicy:     jetstream.AckExplicit,
    MaxDeliver:    3,  // 最多重试 3 次
    BackOff: []time.Duration{  // 退避策略
        time.Second,
        time.Second * 5,
        time.Second * 30,
    },
})
```

## 2.6 消费者组

### 创建消费者组

```go
// 创建消费者组
consumer, err := js.CreateConsumer(ctx, "ORDERS", jetstream.ConsumerConfig{
    Name:           "order-processors",
    Durable:        true,  // 持久消费者
    FilterSubject:  "orders.created",
    DeliverPolicy:  jetstream.DeliverAll,  // 投递策略
    AckPolicy:      jetstream.AckExplicit,
    AckWait:        30 * time.Second,
    MaxDeliver:     3,
    MaxAckPending:  100,
})
if err != nil {
    log.Fatal(err)
}

// 消费消息
msgs, _ := consumer.Fetch(10, jetstream.FetchMaxWait(5*time.Second))
for msg := range msgs {
    fmt.Printf("处理: %s\n", string(msg.Data()))
    msg.Ack()
}
```

### 投递策略

```go
// 从最新消息开始
DeliverPolicy: jetstream.DeliverNew,

// 从最后确认的消息开始
DeliverPolicy: jetstream.DeliverLast,

// 从指定序列号开始
DeliverPolicy:  jetstream.DeliverBySequence,
StartSequence:  100,

// 从指定时间开始
DeliverPolicy:  jetstream.DeliverByStartTime,
StartTime:      time.Now().Add(-time.Hour),
```

## 2.7 消息持久化策略

### 保留策略

```go
// 创建带持久化策略的 Stream
stream, _ := js.CreateStream(ctx, jetstream.StreamConfig{
    Name:     "EVENTS",
    Subjects: []string{"events.>"},
    Retention: jetstream.LimitsPolicy{
        MaxBytes:  1024 * 1024 * 1024,  // 1GB
        MaxAge:    time.Hour * 24 * 30, // 30 天
        MaxMsgs:   100000,
    },
    Discard: jetstream.DiscardOld,
})
```

### 工作队列模式

```go
// 工作队列 Stream
stream, _ := js.CreateStream(ctx, jetstream.StreamConfig{
    Name:      "TASKS",
    Subjects:  []string{"tasks.>"},
    Retention: jetstream.WorkQueuePolicy,  // 工作队列策略
    MaxBytes:  1024 * 1024 * 100,
})
```

## 2.8 消息去重

### 使用消息 ID

```go
// 发布时设置消息 ID
js.Publish(ctx, subject, data, jetstream.WithMsgID("unique-id-123"))

// 消费者检查重复
nc.Subscribe(subject, func(m *nats.Msg) {
    msgID := m.Header.Get("Nats-Msg-Id")
    if isProcessed(msgID) {
        return  // 跳过重复
    }
    processMessage(m.Data)
    markProcessed(msgID)
})
```

### 消费者端去重

```go
// 创建带去重的消费者
consumer, _ := js.CreateConsumer(ctx, "ORDERS", jetstream.ConsumerConfig{
    Name:           "order-dedup",
    FilterSubject:  "orders.created",
    ReplayPolicy:   jetstream.ReplayInstant,
    HeadersOnly:    false,
})
```

## 2.9 Stream 管理

### 列出所有 Stream

```go
// 列出所有 Stream
streams, err := js.ListStreams(ctx)
for _, s := range streams {
    info, _ := s.CachedInfo()
    fmt.Printf("Stream: %s, 消息数: %d\n", info.Config.Name, info.State.Msgs)
}
```

### 获取 Stream 信息

```go
// 获取 Stream 信息
stream, _ := js.Stream(ctx, "ORDERS")
info, _ := stream.CachedInfo()
fmt.Printf("Stream: %s\n", info.Config.Name)
fmt.Printf("消息数: %d\n", info.State.Msgs)
fmt.Printf("字节数: %d\n", info.State.Bytes)
```

### 删除消息

```go
// 删除消息
stream, _ := js.Stream(ctx, "ORDERS")
err := stream.DeleteMsg(ctx, 100)  // 删除序列号 100 的消息

// 清空 Stream
err = stream.Purge(ctx)  // 清空所有消息
err = stream.Purge(ctx, jetstream.WithSequence(100))  // 清空到指定序列号
```

### 删除 Stream

```go
// 删除 Stream
stream, _ := js.Stream(ctx, "ORDERS")
err := stream.Delete(ctx)
```

## 2.10 监控 JetStream

### 获取 JetStream 状态

```go
// 获取 JetStream 状态
js, _ := jetstream.New(nc)
ctx := context.Background()

// 列出所有 Stream
streams, _ := js.ListStreams(ctx)
for _, s := range streams {
    info, _ := s.CachedInfo()
    fmt.Printf("Stream: %s, 消息数: %d\n", info.Config.Name, info.State.Msgs)
}

// 获取消费者信息
consumers, _ := js.ListConsumers(ctx, "ORDERS")
for _, c := range consumers {
    info, _ := c.CachedInfo()
    fmt.Printf("Consumer: %s, 待处理: %d\n", info.Name, info.NumPending)
}
```

---

## 2.11 最佳实践

### 1. 选择合适的存储类型

| 场景 | 存储类型 | 说明 |
|------|----------|------|
| 实时日志 | 内存 | 低延迟，不需要持久 |
| 订单事件 | 文件 | 需要持久化 |
| 金融交易 | 文件 + 副本 | 高可靠性 |

### 2. 合理设置保留策略

```go
// 根据业务需求设置
Retention: jetstream.LimitsPolicy{
    MaxBytes: 1024 * 1024 * 1024 * 10,  // 10GB
    MaxAge:   time.Hour * 24 * 30,      // 30 天
    MaxMsgs:  1000000,                   // 100 万条
}
```

### 3. 消费者组设计

```go
// 根据处理能力设计消费者数量
ConsumerConfig: jetstream.ConsumerConfig{
    MaxAckPending:   100,   // 根据处理能力调整
    MaxDeliver:      3,     // 重试次数
    AckWait:         30 * time.Second,
}
```

---

## 2.12 相关资源

- [JetStream 官方文档](https://docs.nats.io/using-nats/jetstream)
- [JetStream API](https://pkg.go.dev/github.com/nats-io/nats.go/jetstream)
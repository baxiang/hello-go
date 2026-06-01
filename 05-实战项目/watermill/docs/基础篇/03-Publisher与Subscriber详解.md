# 03 - Publisher 与 Subscriber 详解

本章深入解析 Watermill 的 Publisher 和 Subscriber 机制，以 `GoChannel`（进程内实现）为切入点理解其设计。

## GoChannel 内部实现

GoChannel 是 Watermill 内置的进程内 Pub/Sub 实现，位于 `pubsub/gochannel` 包。其核心原理：维护一个 `map[string][]chan *Message` 结构，key 为 topic 名称，value 为该 topic 的所有订阅者 channel 列表。

```go
// basic/01-pubsub/main.go:17
pubSub := gochannel.NewGoChannel(gochannel.Config{}, logger)
```

`Config` 支持两个关键参数：
- **OutputChannelBuffer**（默认 0）：订阅 channel 的缓冲大小，非阻塞场景建议设置 >0
- **PublishBlocking**（默认 true）：设为 false 时 Publish 非阻塞，消息满 buffer 就丢弃

## Publish 生命周期

调用 `pubSub.Publish("hello", msg)` 时，GoChannel 的内部步骤如下：

1. 在 map 中查找 topic "hello" 的所有订阅者 channel
2. 将 `msg` 写入**每一个**订阅者 channel
3. 等待所有写入完成后返回（blocking 模式）

如果 topic 无订阅者，消息被丢弃。这符合 Pub/Sub 的"发布即忘"语义——发布者不关心谁在消费。

## Subscribe 生命周期

调用 `pubSub.Subscribe(ctx, "hello")` 时：

1. GoChannel 创建一个带缓冲的 `chan *Message`（大小 = OutputChannelBuffer）
2. 将 channel 注册到 topic "hello" 的订阅者列表中
3. 返回该 channel 给调用者

当 `ctx` 被取消时，GoChannel 自动关闭 channel 并从订阅者列表中移除，实现优雅退出。

在 `basic/01-pubsub/main.go:28` 中展示了典型的消费模式：

```go
go func() {
    for msg := range messages {
        fmt.Printf("收到消息: %s\n", string(msg.Payload))
        msg.Ack()
    }
}()
```

## Ack / Nack 机制

Watermill 的消息确认分为两个层面：

**订阅层面**（由 Pub/Sub 后端实现）。Kafka 等持久化后端需要显式提交 offset（`msg.Ack()`），GoChannel 中 Ack 简化为日志记录。

**消费层面**（由 Router 控制）。Handler 返回 `nil` 视为成功，Router 自动调 `msg.Ack()`；返回 `error` 视为失败，Router 自动调 `msg.Nack()`。对于手工消费（不使用 Router），需要自行调用 Ack/Nack。

## 多后端适配

GoChannel 仅适用于单进程场景（测试或本地开发）。生产环境需切换为 Kafka：

```go
// pkg/kafka/client.go:11-18
func NewPublisher(brokers []string, logger watermill.LoggerAdapter) (message.Publisher, error) {
    return kafka.NewPublisher(kafka.PublisherConfig{
        Brokers:   brokers,
        Marshaler: kafka.DefaultMarshaler{},
    }, logger)
}
```

注意函数签名返回 `message.Publisher` 接口——调用者代码完全不变，因为业务逻辑只依赖接口而非具体实现。这是 Watermill 设计的核心优势。

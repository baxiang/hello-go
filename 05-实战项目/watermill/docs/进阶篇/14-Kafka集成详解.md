# 14 - Kafka 集成详解

Watermill 通过 `watermill-kafka` 适配器实现 Kafka 集成。本章解析 `ecommerce/pkg/kafka/client.go` 的配置细节和内部机制。

## Publisher 配置

```go
func NewPublisher(brokers []string, logger watermill.LoggerAdapter) (message.Publisher, error) {
    return kafka.NewPublisher(kafka.PublisherConfig{
        Brokers:   brokers,
        Marshaler: kafka.DefaultMarshaler{},
    }, logger)
}
```

`PublisherConfig` 核心参数：

| 参数 | 默认值 | 说明 |
|------|--------|------|
| Brokers | 必填 | Kafka broker 地址列表 |
| Marshaler | DefaultMarshaler | 消息序列化器（见下方） |
| OverwriteSaramaConfig | nil | 底层 Sarama 配置覆盖 |
| Async | false | true 时异步发布，吞吐更高但可能丢失消息 |
| BatchMarshaler | nil | 批量序列化器 |

**DefaultMarshaler** 将 Watermill 的 `*message.Message` 转换为 Kafka 的 `*sarama.ProducerMessage`：
- `msg.UUID` → `Key`（用作 Kafka 分区键）
- `msg.Payload` → `Value`
- `msg.Metadata` → `Headers`

这是实现消息分区和顺序消费的关键——**消息 UUID 作为分区键**，同一 UUID 的消息落在同一个 partition。

## Subscriber 配置

```go
func NewSubscriber(brokers []string, consumerGroup string, logger watermill.LoggerAdapter) (message.Subscriber, error) {
    return kafka.NewSubscriber(kafka.SubscriberConfig{
        Brokers:       brokers,
        ConsumerGroup: consumerGroup,
        Unmarshaler:   kafka.DefaultMarshaler{},
    }, logger)
}
```

`SubscriberConfig` 核心参数：

| 参数 | 说明 |
|------|------|
| ConsumerGroup | 消费者组名。同一组内的实例共享 partition |
| Brokers | Kafka broker 地址 |
| Unmarshaler | 反序列化器 |
| InitializeTopicDetails | 自动创建 topic 配置（不推荐生产使用） |
| NackResendSleep | NACK 后重新消费的等待时间 |
| ReconnectRetrySleep | 断连重试间隔 |

## 分区策略

Kafka 分区决定消息的路由和消费顺序。Watermill 的 `DefaultMarshaler` 使用 `msg.UUID` 作为 Key，Sarama 默认使用 **Hash 分区器**：

```
partition = hash(key) % partition_count
```

要按业务键（如 order_id）分区，需要自定义 Marshaler：

```go
type OrderKeyMarshaler struct{}

func (m OrderKeyMarshaler) Marshal(topic string, msg *message.Message) (*sarama.ProducerMessage, error) {
    // 从 Metadata 中提取 order_id 作为分区键
    orderID := msg.Metadata.Get("order_id")
    return &sarama.ProducerMessage{
        Topic: topic,
        Key:   sarama.StringEncoder(orderID),
        Value: sarama.ByteEncoder(msg.Payload),
    }, nil
}
```

## Offset 管理

Watermill Kafka 使用消费者组自动提交 offset：

- **自动提交**（默认）。消费消息后自动提交 offset，可能丢消息（消息收到但未处理完就 crash）
- **手动提交**。通过 `SubscriberConfig.DisableAutoCommit = true` 禁用，配合 `message.Ack()` 手动提交

在电商项目中，由于使用 Router + Retry 中间件，消息未处理成功会触发 NACK + 重试，offset 仅在 `msg.Ack()` 时提交。这保证了 at-least-once 语义。

## 消费者组重平衡（Rebalance）

当消费者组中的实例增减时，Kafka 触发 rebalance——重新分配 partition 给各实例。期间所有消费者暂停消费。

优化建议：
- 增加 `session.timeout.ms` 和 `max.poll.interval.ms` 减少不必要的 rebalance
- 缩短单条消息处理时间（30s 以内）
- 使用 **Static Membership** 固定成员 ID，避免滚动重启时的 rebalance

## Kafka 日志查看

```bash
# 查看 topic 列表
docker exec watermill-kafka kafka-topics --list --bootstrap-server localhost:9092

# 查看消费者组状态
docker exec watermill-kafka kafka-consumer-groups \
  --bootstrap-server localhost:9092 \
  --group order-service --describe

# 查看 topic 最新消息
docker exec watermill-kafka kafka-console-consumer \
  --bootstrap-server localhost:9092 \
  --topic order.events --from-beginning --max-messages 5
```

# Kafka 概述

Apache Kafka 是分布式流处理平台，用于构建实时数据管道和流应用。

## 核心概念

### Producer（生产者）
发布消息到 Kafka 集群的应用。

### Consumer（消费者）
从 Kafka 订阅消息的应用。

### Topic（主题）
消息的分类，类似于数据库的表。

### Partition（分区）
Topic 分为多个分区，实现水平扩展。

### Broker
Kafka 集群中的服务器节点。

## 架构

```
Producer  →  Topic  →  Consumer
              │
         ┌────┴────┐
         ▼         ▼
     Partition  Partition
         │         │
     Broker 0   Broker 1
```

## 快速开始

### Docker Compose

```yaml
version: '3.8'
services:
  zookeeper:
    image: bitnami/zookeeper:3.9.1
    ports: ["2181:2181"]

  kafka:
    image: bitnami/kafka:3.7.0
    ports: ["9092:9092"]
    environment:
      KAFKA_ZOOKEEPER_CONNECT: zookeeper:2181
      KAFKA_ADVERTISED_LISTENERS: PLAINTEXT://localhost:9092
```

### Go Producer

```go
import "github.com/IBM/sarama"

producer, _ := sarama.NewSyncProducer([]string{"localhost:9092"}, nil)
msg := &sarama.ProducerMessage{
    Topic: "my-topic",
    Value: sarama.StringEncoder("Hello Kafka"),
}
producer.SendMessage(msg)
```

### Go Consumer

```go
consumer, _ := sarama.NewConsumer([]string{"localhost:9092"}, nil)
partitionConsumer, _ := consumer.ConsumePartition("my-topic", 0, sarama.OffsetNewest)

for msg := range partitionConsumer.Messages() {
    println(string(msg.Value))
}
```

## 版本信息

| 组件 | 版本 |
|------|------|
| Kafka | 3.7.0 |
| sarama | 1.43.0 |

在下一章中，我们将深入学习核心概念。

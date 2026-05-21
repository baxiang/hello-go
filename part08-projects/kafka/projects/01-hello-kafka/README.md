# 项目01：Hello Kafka

最基础的 Kafka 生产者+消费者示例。

## 功能
- 生产者向 `hello-topic` 发送 10 条消息
- 消费者从该主题消费消息并打印
- 支持多分区消费

## 环境准备

```bash
# 启动 Kafka（使用 Docker）
docker-compose up -d

# 创建主题（3个分区）
docker exec -it <kafka-container> kafka-topics.sh \
  --create --topic hello-topic --partitions 3 --replication-factor 1 \
  --bootstrap-server localhost:9092
```

## 运行

```bash
# 启动消费者（先启动，等待消息）
go run consumer/main.go

# 另开终端运行生产者
go run producer/main.go
```

## 代码结构
```
01-hello-kafka/
├── docker-compose.yml    # Kafka + Zookeeper
├── producer/
│   └── main.go           # 同步生产者
├── consumer/
│   └── main.go           # 分区消费者
└── go.mod
```

## 对应文档
- [01-Kafka概述](../docs/01-Kafka概述.md)
- [03-生产者与消费者](../docs/03-生产者与消费者.md)
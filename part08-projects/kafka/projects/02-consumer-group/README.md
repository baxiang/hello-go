# 项目02：Consumer Group

演示 Kafka 消费者组的工作机制和分区分配。

## 功能
- 生产者持续发送带 key 的消息（同一 key 进入同一分区）
- 多个消费者组成消费者组共同消费
- 支持手动 offset 提交和重平衡通知
- 启动/停止消费者观察分区重新分配

## 环境准备

```bash
# 启动 Kafka
docker-compose up -d

# 创建主题（6个分区，便于观察分配）
docker exec -it <kafka-container> kafka-topics.sh \
  --create --topic group-topic --partitions 6 --replication-factor 1 \
  --bootstrap-server localhost:9092
```

## 运行

```bash
# 1. 启动生产者（持续发送消息）
go run producer/main.go

# 2. 启动第一个消费者
go run consumer/main.go consumer-1

# 3. 另开终端启动第二个消费者
go run consumer/main.go consumer-2

# 4. 再开一个启动第三个消费者
go run consumer/main.go consumer-3
```

## 观察现象

- 初始只有 consumer-1：分配到所有 6 个分区
- 启动 consumer-2：分区重新分配，各消费 3 个
- 启动 consumer-3：各消费 2 个分区
- 停止某个消费者：分区重新分配给存活消费者

## 代码结构
```
02-consumer-group/
├── docker-compose.yml
├── producer/
│   └── main.go           # 带 key 的生产者
├── consumer/
│   └── main.go           # 消费者组（手动提交）
└── go.mod
```

## 对应文档
- [02-核心概念](../docs/02-核心概念.md)（分区、消费者组）
- [05-消费者组](../docs/05-消费者组.md)（重平衡、offset管理）
# 项目03：Advanced Producer

演示 Kafka 生产者的高级配置和性能优化。

## 功能
- 幂等性配置（Producer.Idempotent）
- 批量发送优化（Flush.Bytes/Messages/Frequency）
- Snappy 压缩
- 并发发送（10个 goroutine）
- 吞吐量统计

## 环境准备

```bash
# 启动 Kafka
docker-compose up -d

# 创建主题（3个分区）
docker exec -it <kafka-container> kafka-topics.sh \
  --create --topic advanced-topic --partitions 3 --replication-factor 1 \
  --bootstrap-server localhost:9092
```

## 运行

```bash
# 启动消费者
go run consumer/main.go

# 另开终端启动生产者（发送10000条消息）
go run producer/main.go
```

## 关键配置说明

### 可靠性
- `RequiredAcks: WaitForAll` - 等待所有 ISR 确认
- `Retry.Max: 5` - 失败重试 5 次
- `Idempotent: true` - 开启幂等性，避免重复

### 性能
- `Flush.Bytes: 1MB` - 批次达到 1MB 才发送
- `Flush.Messages: 1000` - 或累积 1000 条消息
- `Flush.Frequency: 100ms` - 或最长等待 100ms
- `Compression: Snappy` - Snappy 压缩算法

## 代码结构
```
03-advanced-producer/
├── docker-compose.yml
├── producer/
│   └── main.go           # 高并发批量生产者
├── consumer/
│   └── main.go           # 消费者组（自动提交）
└── go.mod
```

## 对应文档
- [04-分区与副本](../docs/04-分区与副本.md)
- [08-性能优化](../docs/08-性能优化.md)
- [09-可靠性保证](../docs/09-可靠性保证.md)
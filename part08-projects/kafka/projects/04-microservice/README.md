# 项目04：Microservice（微服务实战）

模拟真实微服务场景：订单系统 + 日志收集。

## 功能
- **订单服务**（生产者）：模拟生成订单事件，发送到 `orders` 和 `order-logs` topic
- **订单处理器**（消费者）：消费 `orders` topic，根据订单状态执行不同处理逻辑
- **日志收集器**（消费者）：消费 `order-logs` topic，分类打印日志

## 环境准备

```bash
# 启动 Kafka
docker-compose up -d

# 创建主题
docker exec -it <kafka-container> kafka-topics.sh \
  --create --topic orders --partitions 3 --replication-factor 1 \
  --bootstrap-server localhost:9092

docker exec -it <kafka-container> kafka-topics.sh \
  --create --topic order-logs --partitions 3 --replication-factor 1 \
  --bootstrap-server localhost:9092
```

## 运行

```bash
# 1. 启动订单处理器
go run order-processor/main.go

# 2. 另开终端启动日志收集器
go run log-collector/main.go

# 3. 另开终端启动订单服务（发送100条订单）
go run order-service/main.go
```

## 场景说明

1. 订单服务生成订单并发送到两个 topic
2. 订单处理器根据状态执行不同业务逻辑
3. 日志收集器实时收集并分类订单日志
4. 可启动多个订单处理器实例实现负载均衡

## 代码结构
```
04-microservice/
├── docker-compose.yml
├── order-service/
│   └── main.go           # 订单服务（生产者）
├── order-processor/
│   └── main.go           # 订单处理器（消费者组）
├── log-collector/
│   └── main.go           # 日志收集器（消费者组）
└── go.mod
```

## 对应文档
- [06-数据流处理](../docs/06-数据流处理.md)
- [07-集群管理](../docs/07-集群管理.md)
- [09-可靠性保证](../docs/09-可靠性保证.md)
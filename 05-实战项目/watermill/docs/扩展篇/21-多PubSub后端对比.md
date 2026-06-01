# 21 - 多 Pub/Sub 后端对比

Watermill 的最大价值是**统一抽象、多后端可互换**。选择正确的 Pub/Sub 后端直接影响系统的复杂度、性能和运维成本。

## 后端总览

| 后端 | 包路径 | 适用场景 |
|------|--------|---------|
| GoChannel | `pubsub/gochannel` | 单进程、测试、本地开发 |
| Kafka | `watermill-kafka` | 高吞吐、持久化、生产级事件总线 |
| NATS JetStream | `watermill-nats` | 轻量、低延迟、替代 Kafka 的简单方案 |
| Redis Stream | `watermill-redis` | 低延迟、利用现有 Redis 基础设施 |
| RabbitMQ | `watermill-amqp` | 复杂路由、成熟生态、传统企业 |
| SQL | `watermill-sql` | 无需新基础设施、利用已有数据库 |

## GoChannel

**原理**：进程内 `map[topic][]chan`，零 I/O 开销。

```go
pubSub := gochannel.NewGoChannel(gochannel.Config{
    OutputChannelBuffer: 1024,
}, logger)
```

- **优点**：零依赖、零配置、极快、测试友好
- **缺点**：无持久化、无跨进程通信、进程退出消息全部丢失
- **适用**：单元测试、开发环境、`go run *.go` 即用

在 `basic/` 目录的所有示例中，GoChannel 是默认后端。它让学习 Watermill 概念的门槛降到最低。

## Kafka

**原理**：分布式日志，partition 有序，磁盘持久化。

- **优点**：吞吐极高（百万级/秒）、天然水平扩展、数据持久化、成熟生态
- **缺点**：运维复杂（需 Zookeeper 或 KRaft）、延迟略高（1-10ms）、资源消耗大
- **适用**：大规模事件驱动系统、日志采集、流处理

电商项目使用 Kafka（`ecommerce/pkg/kafka/client.go`），展示了生产级配置方式。

## NATS JetStream

**原理**：NATS 的持久化层，支持 at-least-once 和流式消费。

- **优点**：轻量运维（单个二进制文件）、低延迟、多协议（TCP/WS/MQTT）
- **缺点**：生态小于 Kafka、社区规模较小
- **适用**：IoT、边缘计算、中小规模微服务

切换体验：只需修改 Publisher/Subscriber 创建代码，业务逻辑零改动。

## Redis Stream

**原理**：Redis 5.0+ 的流数据结构，支持消费者组。

- **优点**：超低延迟（微秒级）、利用现有 Redis 部署
- **缺点**：持久化依赖 Redis RDB/AOF 配置、内存成本高
- **适用**：实时性要求极高的场景、已有 Redis 基础设施的团队

## RabbitMQ

**原理**：AMQP 协议，Exchange 灵活路由。

- **优点**：成熟稳定、灵活路由策略（direct/topic/fanout/headers）、管理 UI 友好
- **缺点**：吞吐低于 Kafka、集群模式复杂
- **适用**：传统企业系统、需要复杂路由逻辑的场景

## SQL

**原理**：用数据库表模拟消息队列，轮询读取。

- **优点**：零新基础设施、事务一致性（Outbox 模式天然支持）
- **缺点**：吞吐极低、延迟高（秒级）、对业务数据库造成额外负载
- **适用**：低频事件通知、Outbox 模式的 Forwarder 源端

## 决策矩阵

| 维度 | GoChannel | Kafka | NATS | Redis | RabbitMQ | SQL |
|------|:---:|:---:|:---:|:---:|:---:|:---:|
| 吞吐量 | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐⭐ | ⭐ |
| 延迟 | ⭐⭐⭐⭐⭐ | ⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐ | ⭐ |
| 持久化 | ❌ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ |
| 运维复杂度 | ⭐⭐⭐⭐⭐ | ⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐⭐ | ⭐⭐⭐ | ⭐⭐⭐⭐ |
| 社区生态 | ⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐ |

**建议**：开发/测试用 GoChannel → 中小规模用 NATS → 大规模用 Kafka → 特殊需求用 Redis/RabbitMQ/SQL。

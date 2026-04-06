# NATS 技术指南系列

> 本系列是一份面向工程师的 NATS 系统性学习指南，从基础概念到生产实践，覆盖 Core NATS、JetStream、集群高可用、监控运维以及 Go 客户端编程。每篇文章独立成章，也可按学习路径顺序阅读。

---

## 目录

| 序号 | 文章 | 主题 | 难度 |
|------|------|------|------|
| 01 | [NATS 是什么](./01-NATS概述.md) | 起源、设计哲学、核心特性、架构全景 | ★☆☆ |
| 02 | [Subject 寻址系统](./02-Subject寻址系统.md) | 命名规范、通配符、层级设计 | ★☆☆ |
| 03 | [Pub/Sub 发布订阅](./03-发布订阅模式.md) | 模式详解、扇出、内部实现 | ★★☆ |
| 04 | [Request/Reply](./04-请求响应模式.md) | 同步 RPC 模式、Queue Groups 负载均衡 | ★★☆ |
| 05 | [JetStream Streams](./05-JetStream-Streams.md) | 持久化日志、存储策略、配置详解 | ★★☆ |
| 06 | [JetStream Consumers](./06-JetStream-Consumers.md) | Push vs Pull、Ack 策略、消费起点 | ★★★ |
| 07 | [KV Store 与 Object Store](./07-KV存储与对象存储.md) | 分布式状态存储 | ★★☆ |
| 08 | [集群架构](./08-集群架构.md) | Route Pool、num_routes 原理、Subject 路由 | ★★★ |
| 09 | [JetStream 高可用](./09-JetStream高可用.md) | Raft 共识、副本机制、故障恢复 | ★★★ |
| 10 | [监控与可观测性](./10-监控与可观测性.md) | 所有 HTTP 端点详解 | ★★☆ |
| 11 | [Go 客户端实战](./11-Go客户端实战.md) | 连接管理、各模式代码示例 | ★★★ |

---

## 推荐学习路径

### 路径 A：快速入门（2-3 小时）

```
01-overview  →  02-subjects  →  03-pub-sub  →  04-request-reply
```

适合：刚接触 NATS、希望快速理解其定位与基础使用方式的工程师。

### 路径 B：持久化消息（基于路径 A，再加 3-4 小时）

```
→  05-jetstream-streams  →  06-jetstream-consumers  →  07-jetstream-kv
```

适合：需要消息持久化、消费回溯、KV 状态管理等能力的工程师。

### 路径 C：生产运维（基于路径 B，再加 4-5 小时）

```
→  08-cluster  →  09-jetstream-ha  →  10-monitoring
```

适合：负责 NATS 集群部署、维护和监控告警的 SRE / 运维工程师。

### 路径 D：开发全栈（全系列 + 实践）

```
全部 11 篇  →  11-go-client（重点结合业务代码实践）
```

适合：希望在项目中深度集成 NATS 的 Go 后端工程师。

---

## 各章内容简介

### [01 - NATS 是什么](./01-NATS概述.md)

介绍 NATS 的起源（Derek Collison，2010 年，Cloud Foundry）、设计哲学（简单、高速、始终在线），以及四层架构全景（Core NATS → JetStream → Cluster → 网络拓扑）。对比 Kafka 和 RabbitMQ 的定位差异，明确 NATS 的适用与不适用场景。包含官方性能基准数据（单节点 18M msg/s）。

### [02 - Subject 寻址系统](./02-Subject寻址系统.md)

NATS 中一切通信基于 Subject（主题字符串）寻址，本章深入讲解 Subject 的命名规范、层级结构（以 `.` 分隔），以及两种通配符 `*`（单级）和 `>`（多级）的使用方式。给出生产环境中常见的 Subject 层级设计模式。

### [03 - Pub/Sub 发布订阅](./03-发布订阅模式.md)

Core NATS 最基础的通信模式。讲解发布者、订阅者模型，消息扇出（Fan-out）的行为，以及 NATS 服务器内部的订阅路由表（subscription routing）实现原理。说明 Core NATS 的 At-Most-Once 语义及其设计取舍。

### [04 - Request/Reply](./04-请求响应模式.md)

NATS 内置的同步 RPC 模式，基于临时 inbox subject 实现。讲解 `nats.Request()` 的底层机制，Queue Group 如何实现多实例负载均衡，以及与传统 HTTP 调用的对比。

### [05 - JetStream Streams](./05-JetStream-Streams.md)

JetStream 是 NATS 的持久化消息层。本章聚焦于 Stream（流）的概念：消息日志的存储结构、Retention Policy（限制/工作队列/关注策略）、Storage Backend（文件 vs 内存）、消息保留策略（MaxAge、MaxMsgs、MaxBytes）以及 Stream 配置的完整字段详解。

### [06 - JetStream Consumers](./06-JetStream-Consumers.md)

Consumer 是 JetStream 中读取 Stream 数据的游标。本章对比 Push Consumer（服务器推送）和 Pull Consumer（客户端拉取）的适用场景；详解 Ack Policy（Explicit / None / All）；讲解消费起点（DeliverPolicy）如何控制从头、从尾、从某时间点开始消费；以及 Durable Consumer 与 Ephemeral Consumer 的区别。

### [07 - KV Store 与 Object Store](./07-KV存储与对象存储.md)

基于 JetStream 之上封装的两种高级存储原语。KV Store 提供带版本历史的键值存储，支持 Watch（变更监听）和 CAS（Compare-and-Swap）操作。Object Store 提供大对象分块存储能力。本章讲解其底层实现（均为特殊 Stream），适用场景，以及与 Redis、etcd 的对比。

### [08 - 集群架构](./08-集群架构.md)

讲解 NATS 如何通过 Routes 组建集群：Route Pool 连接机制、`num_routes` 参数的配置原理与最佳实践、集群内 Subject 兴趣传播（Interest Propagation）、消息在节点间的转发路径。区分 Cluster、Leaf Node、Gateway 三种组网模式的用途。

### [09 - JetStream 高可用](./09-JetStream高可用.md)

深入讲解 JetStream 在集群模式下的高可用机制：Raft 共识算法在 NATS 中的实现，Stream 副本（R1/R3/R5）的写入流程，Leader 选举与 Quorum 要求，节点故障时的自动恢复流程，以及手动触发 Leader 切换的运维操作。

### [10 - 监控与可观测性](./10-监控与可观测性.md)

NATS Server 内置 HTTP 监控端点（默认 `:8222`），本章逐一讲解所有端点：`/varz`（服务器基础指标）、`/connz`（连接详情）、`/subsz`（订阅统计）、`/routez`（集群路由）、`/jsz`（JetStream 统计）、`/healthz` / `/readyz` / `/livez`（健康检查）等。给出与 Prometheus + Grafana 集成的方案。

### [11 - Go 客户端实战](./11-Go客户端实战.md)

使用官方 `nats.go` 客户端库进行完整的代码实战。涵盖：连接选项与重连配置、Core Pub/Sub 与 Request/Reply 代码示例、JetStream API（发布、Push/Pull Consumer）、KV Store 操作、错误处理模式，以及在生产环境中的常见陷阱与最佳实践。

---

## 快速概念速查

### 核心术语对照

| NATS 术语 | 含义 | 类比 |
|-----------|------|------|
| Subject | 消息地址字符串 | Kafka Topic / RabbitMQ Routing Key |
| Publisher | 消息发送方 | Producer |
| Subscriber | 消息接收方 | Consumer |
| Queue Group | 同名订阅者共享消息（LB） | Kafka Consumer Group |
| JetStream | NATS 持久化引擎 | Kafka 的持久化层 |
| Stream | 持久化消息日志 | Kafka Partition Log |
| Consumer | JetStream 读取游标 | Kafka Consumer Offset |
| KV Store | 基于 JetStream 的键值存储 | Redis / etcd |
| Leaf Node | 边缘接入节点 | Kafka MirrorMaker |

### 消息语义速查

| 模式 | 投递语义 | 持久化 | 重试 |
|------|----------|--------|------|
| Core Pub/Sub | At Most Once | 否 | 否 |
| Core Request/Reply | At Most Once | 否 | 否 |
| JetStream Pub | At Least Once | 是 | 是（Ack 机制） |
| JetStream KV | Exactly Once（CAS） | 是 | N/A |

### 通配符速查

```
*   匹配单个层级的任意 token
    例: sensor.*.temp  匹配 sensor.room1.temp，但不匹配 sensor.room1.floor2.temp

>   匹配剩余所有层级（只能出现在末尾）
    例: sensor.>  匹配 sensor.room1.temp、sensor.room1.floor2.temp 等所有子级
```

### 配置快速模板

```yaml
# nats-server.conf 最小集群配置模板
server_name: nats-1
port: 4222
http_port: 8222

jetstream: {
  store_dir: /data/nats
  max_memory_store: 1GB
  max_file_store: 10GB
}

cluster: {
  name: my-cluster
  port: 6222
  routes: [
    "nats://nats-2:6222"
    "nats://nats-3:6222"
  ]
}
```

---

## 版本兼容性说明

本指南基于以下版本编写：

| 组件 | 版本 | 说明 |
|------|------|------|
| NATS Server | 2.11.10 | 当前稳定版，JetStream 功能完善 |
| nats.go | 1.49.0 | 推荐使用新版 jetstream 包 |
| Go | 1.21+ | 客户端最低要求 |

**版本演进要点：**

- **NATS 2.0**：引入去中心化安全模型（NKEY + JWT）、Account 隔离
- **NATS 2.2**：JetStream 正式 GA，内置持久化引擎
- **NATS 2.8**：Consumer FilterSubjects（多 Subject 过滤）
- **NATS 2.9**：JetStream 性能大幅提升，新增压缩支持
- **NATS 2.10**：Stream Sources 增强，Consumer 优先级队列
- **NATS 2.11**：分布式消息追踪、Per-message TTL、Consumer 优先级组、Consumer 暂停、资产版本控制

---

## 常见问题快速导航

| 问题 | 参考章节 |
|------|----------|
| NATS 和 Kafka 怎么选？ | [01-NATS概述.md#6-与-kafka-rabbitmq-的对比](./01-NATS概述.md#6-与-kafka-rabbitmq-的对比) |
| Subject 怎么命名？ | [02-Subject寻址系统.md#4-subject-命名最佳实践](./02-Subject寻址系统.md#4-subject-命名最佳实践) |
| 如何实现负载均衡？ | [04-请求响应模式.md#9-queue-groups-概念](./04-请求响应模式.md#9-queue-groups-概念) |
| 消息如何持久化？ | [05-JetStream-Streams.md](./05-JetStream-Streams.md) |
| 如何保证消息不丢失？ | [06-JetStream-Consumers.md#6-ack-policy-全解](./06-JetStream-Consumers.md#6-ack-policy-全解) |
| 如何实现分布式锁？ | [07-KV存储与对象存储.md#8-分布式锁实现模式](./07-KV存储与对象存储.md#8-分布式锁实现模式) |
| 集群怎么部署？ | [08-集群架构.md#7-配置示例](./08-集群架构.md#7-配置示例) |
| 如何实现高可用？ | [09-JetStream高可用.md](./09-JetStream高可用.md) |
| 怎么监控 NATS？ | [10-监控与可观测性.md](./10-监控与可观测性.md) |
| Go 代码怎么写？ | [11-Go客户端实战.md](./11-Go客户端实战.md) |

---

## 参考资源

### 官方资源

- [NATS 官方文档](https://docs.nats.io) — 最权威的参考
- [NATS Server 源码](https://github.com/nats-io/nats-server) — Go 实现
- [nats.go 客户端库](https://github.com/nats-io/nats.go) — 官方 Go 客户端
- [nats-surveyor](https://github.com/nats-io/nats-surveyor) — Prometheus Exporter
- [NATS CLI](https://github.com/nats-io/natscli) — 命令行工具

### 学习资源

- [Synadia 官方博客](https://www.synadia.com/blog) — 技术文章和案例
- [NATS by Example](https://natsbyexample.com) — 代码示例集合
- [NATS YouTube 频道](https://www.youtube.com/c/NATSio) — 视频教程

### 社区资源

- [NATS Slack](https://natsio.slack.com) — 社区讨论
- [CNCF NATS 页面](https://www.cncf.io/projects/nats/) — CNCF 毕业项目

### 相关项目

- [NATS K8s Operator](https://github.com/nats-io/nats-operator) — Kubernetes 部署
- [NATS Helm Charts](https://github.com/nats-io/k8s) — Helm 部署模板
- [NATS Streaming (已废弃)](https://github.com/nats-io/nats-streaming-server) — 旧版持久化方案，请迁移到 JetStream


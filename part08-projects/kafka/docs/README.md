# Kafka 学习系列

本系列文档系统性地介绍 Apache Kafka 分布式消息系统，从基础概念到生产实践。

## 学习路径

### 第一阶段：基础概念

| 文档 | 描述 | 预计时间 |
|------|------|---------|
| 01-Kafka概述 | Kafka 简介、应用场景、架构概览 | 40 分钟 |
| 02-核心概念 | Topic、Partition、Consumer Group | 50 分钟 |
| 03-生产者与消费者 | Go 客户端开发、配置优化 | 60 分钟 |

### 第二阶段：进阶特性

| 文档 | 描述 | 预计时间 |
|------|------|---------|
| 04-分区与副本 | 分区策略、副本机制、ISR | 50 分钟 |
| 05-消费者组 | Rebalance、Offset 管理 | 45 分钟 |
| 06-数据流处理 | Kafka Streams、Connect | 60 分钟 |

### 第三阶段：生产实践

| 文档 | 描述 | 预计时间 |
|------|------|---------|
| 07-集群管理 | 部署、监控、运维 | 50 分钟 |
| 08-性能优化 | 吞吐量、延迟、资源优化 | 45 分钟 |
| 09-可靠性保证 | Exactly Once、事务 | 50 分钟 |

## 前置知识

- Go 语言基础
- 分布式系统基础
- 消息队列概念

## 学习建议

1. 理解 Kafka 核心概念
2. 按顺序学习理论文档
3. 完成对应 hands-on 练习
4. 参考 official 文档：https://kafka.apache.org/documentation/

## 相关资源

- [Kafka 官方文档](https://kafka.apache.org/documentation/)
- [Kafka GitHub](https://github.com/apache/kafka)
- [sarama Go 客户端](https://github.com/IBM/sarama)
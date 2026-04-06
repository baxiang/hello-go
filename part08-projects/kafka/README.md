# Kafka 学习笔记

本目录包含完整的 Apache Kafka 学习资料，包括理论文档、实战项目和代码示例。

## 目录结构

```
kafka/
├── docs/                    # 理论文档
│   ├── README.md           # 系列导航和学习路径
│   ├── 01-Kafka概述.md
│   ├── 02-核心概念.md
│   └── ...
├── projects/               # 实战项目
│   ├── 01-入门-HelloWorld/
│   ├── 02-进阶-消息系统/
│   └── ...
├── hands-on/               # 动手练习代码
└── example/                # 完整示例应用
```

## 学习路径

### 入门路径
docs/01-Kafka概述 → docs/02-核心概念 → docs/03-生产者与消费者

### 进阶路径
docs/04-分区与副本 → docs/05-消费者组 → docs/06-数据流处理

### 实战路径
完成理论学习后 → projects/01-入门-HelloWorld → projects/02-进阶-消息系统 → ...

## 本地开发环境

### 启动 Kafka

```bash
# 使用 Docker Compose（推荐）
docker-compose up -d

# 或单独启动 Zookeeper 和 Kafka
docker run -d --name zookeeper \
  -p 2181:2181 \
  bitnami/zookeeper:3.9.1

docker run -d --name kafka \
  -p 9092:9092 \
  -e KAFKA_ZOOKEEPER_CONNECT=zookeeper:2181 \
  -e KAFKA_ADVERTISED_LISTENERS=PLAINTEXT://localhost:9092 \
  bitnami/kafka:3.7.0
```

### 连接地址

- Kafka Broker: localhost:9092
- Zookeeper: localhost:2181

## 快速开始

1. 阅读 [docs/README.md](./docs/README.md) 了解完整学习路径
2. 从 [docs/01-Kafka概述.md](./docs/01-Kafka概述.md) 开始学习
3. 使用 hands-on/ 目录中的代码进行练习

## 版本信息

| 组件 | 版本 |
|------|------|
| Apache Kafka | 3.7.0 |
| sarama | 1.43.0 |
| Go | 1.21+ |
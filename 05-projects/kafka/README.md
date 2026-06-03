# Kafka 学习笔记

本目录包含完整的 Apache Kafka 学习资料，包括理论文档和配套实战项目。每个项目都是完整可运行的代码，可直接编译执行。

## 目录结构

```
kafka/
├── docs/                    # 理论文档（按学习顺序）
│   ├── README.md           # 系列导航
│   ├── 01-Kafka概述.md
│   ├── 02-核心概念.md
│   ├── 03-生产者与消费者.md
│   ├── 04-分区与副本.md
│   ├── 05-消费者组.md
│   ├── 06-数据流处理.md
│   ├── 07-集群管理.md
│   ├── 08-性能优化.md
│   └── 09-可靠性保证.md
└── projects/               # 实战项目（配套文档）
    ├── 01-hello-kafka/     # 基础生产者+消费者（对应 01、03）
    ├── 02-consumer-group/  # 消费者组+分区（对应 02、05）
    ├── 03-advanced-producer/ # 高级配置+性能优化（对应 04、08、09）
    └── 04-microservice/    # 微服务实战（对应 06、07）
```

## 学习路径

| 阶段 | 文档 | 实战项目 | 预计时间 |
|------|------|----------|----------|
| 基础 | 01-Kafka概述 → 02-核心概念 | - | 1.5 小时 |
| 入门 | 03-生产者与消费者 | [01-hello-kafka](./projects/01-hello-kafka/) | 1 小时 |
| 进阶 | 04-分区与副本 → 05-消费者组 | [02-consumer-group](./projects/02-consumer-group/) | 1.5 小时 |
| 进阶 | 08-性能优化 → 09-可靠性保证 | [03-advanced-producer](./projects/03-advanced-producer/) | 1 小时 |
| 生产 | 06-数据流处理 → 07-集群管理 | [04-microservice](./projects/04-microservice/) | 1.5 小时 |

## 环境准备

### 启动 Kafka（Docker）

每个项目目录下都有 `docker-compose.yml`：

```bash
cd projects/01-hello-kafka
docker-compose up -d

# 创建主题（3个分区）
docker exec -it <kafka-container> kafka-topics.sh \
  --create --topic hello-topic --partitions 3 --replication-factor 1 \
  --bootstrap-server localhost:9092
```

### 安装依赖

```bash
# 项目已初始化 go.mod，直接获取依赖
cd projects/01-hello-kafka
go mod tidy
```

## 快速开始

以 Hello Kafka 项目为例：

```bash
cd projects/01-hello-kafka

# 启动 Kafka
docker-compose up -d

# 启动消费者（先启动，等待消息）
go run consumer/main.go

# 另开终端运行生产者
go run producer/main.go
```

## 版本信息

| 组件 | 版本 |
|------|------|
| Apache Kafka | 3.7.0 |
| sarama | 1.49.0 |
| Go | 1.21+ |

## 参考

- [Kafka 官方文档](https://kafka.apache.org/documentation/)
- [sarama Go 客户端](https://github.com/IBM/sarama)
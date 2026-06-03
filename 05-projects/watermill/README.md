# Watermill 学习项目

基于 [Watermill](https://github.com/ThreeDotsLabs/watermill) 的 Go 事件驱动编程学习项目，包含基础概念示例和 Kafka 电商实战。

## 目录结构

```
watermill/
├── basic/              # 基础示例（GoChannel，零依赖）
│   ├── 01-pubsub/      # Publisher + Subscriber
│   ├── 02-router/      # Router 消息路由
│   ├── 03-middleware/  # 中间件：重试/超时/恢复
│   ├── 04-cqrs/        # CQRS 命令与事件分离
│   └── 05-metrics/     # Prometheus 指标采集
├── ecommerce/          # Kafka 事件驱动电商
│   ├── api/proto/      # Protobuf 事件定义
│   ├── cmd/            # 4 个微服务入口
│   ├── internal/       # Clean Architecture 分层
│   ├── pkg/            # 共享基础设施
│   ├── deploy/         # Docker Compose
│   └── configs/        # 各服务配置
├── docs/               # 24 篇中文学习文档
└── README.md
```

## 快速开始

### 基础示例
```bash
cd basic
go run 01-pubsub/main.go    # Pub/Sub 基础
go run 02-router/main.go    # Router 路由
go run 03-middleware/main.go  # 中间件
go run 04-cqrs/main.go       # CQRS
go run 05-metrics/main.go    # 指标监控
```

### 电商实战
```bash
# 1. 启动基础设施
cd ecommerce/deploy
docker-compose up -d

# 2. 启动服务（各开一个终端）
cd ecommerce
go run cmd/order-service/main.go --config=configs/order-service.yaml
go run cmd/inventory-service/main.go --config=configs/inventory-service.yaml
go run cmd/payment-service/main.go --config=configs/payment-service.yaml
go run cmd/notification-service/main.go --config=configs/notification-service.yaml

# 3. 创建订单（触发完整 Saga 流程）
curl -X POST http://localhost:8081/orders \
  -H 'Content-Type: application/json' \
  -d '{"order_id":"ORD-001","user_id":"U001","items":[{"product_id":"PROD-001","quantity":1,"price":99.99}],"total":99.99}'
```

## 学习路径

1. **新人入门**: 阅读 `docs/basics/01-06` → 运行 `basic/` 示例
2. **上手实战**: 阅读 `docs/practice/07-13` → 运行 `ecommerce/`
3. **深入原理**: 阅读 `docs/advanced/14-20`
4. **扩展视野**: 阅读 `docs/extensions/21-23`

## 技术栈

| 组件 | 用途 |
|------|------|
| Watermill v1.5 | 事件驱动框架 |
| Kafka | 消息队列（watermill-kafka） |
| GORM + MySQL | 数据持久化 |
| Protobuf | 事件序列化 |
| Viper | 配置管理 |
| Zap | 结构化日志 |
| Docker Compose | 基础设施编排 |

# Kratos 学习项目

包含两部分：**框架源码学习文档** + **微服务电商实战项目**。

## 📚 目录结构

```
kratos/
├── docs/              # Kratos 框架源码学习文档
│   ├── 01-为什么学习-Kratos.md
│   ├── 02-核心概念铺垫.md
│   ├── 03-整体架构.md
│   ├── 04-启动流程详解.md
│   ├── 05-HTTP-Server-详解.md
│   └── README.md
├── services/          # 微服务电商实战项目
│   ├── api/           # 公共 protobuf 定义
│   │   ├── user/v1/
│   │   ├── product/v1/
│   │   ├── order/v1/
│   │   └── payment/v1/
│   ├── pkg/           # 公共基础设施
│   │   ├── config/       # YAML 配置加载
│   │   ├── database/     # GORM + MySQL
│   │   ├── natsclient/   # NATS JetStream
│   │   ├── redisclient/  # Redis
│   │   ├── token/        # JWT
│   │   └── server/       # gRPC/HTTP 服务封装
│   ├── user-service/      # 用户服务 (gRPC :50051)
│   ├── product-service/   # 商品服务 (gRPC :50052)
│   ├── order-service/     # 订单服务 (gRPC :50053)，调用 product
│   ├── payment-service/   # 支付服务 (gRPC :50054)，调用 order
│   ├── api-gateway/       # HTTP 网关 (:8080)，调用所有后端
│   ├── deploy/            # Docker Compose / Prometheus / 初始化 SQL
│   └── go.mod
└── README.md
```

## 🎯 学习路径

1. 阅读 [docs/](./docs/) 理解 Kratos 框架原理
2. 阅读 [services/](./services/) 理解微服务实战代码
3. 启动基础设施 → 启动微服务 → 通过网关访问

## 🚀 快速开始

### 1. 启动基础设施
```bash
cd services/deploy/docker
docker-compose up -d
```

### 2. 启动各服务（每个开一个终端）
```bash
cd services/user-service && go run cmd/main.go
cd services/product-service && go run cmd/main.go
cd services/order-service && go run cmd/main.go
cd services/payment-service && go run cmd/main.go
cd services/api-gateway && go run cmd/main.go
```

### 3. 测试接口
```bash
# 创建用户
curl -X POST http://localhost:8080/api/v1/users \
  -H 'Content-Type: application/json' \
  -d '{"username":"alice","password":"123456","email":"alice@example.com"}'

# 登录获取 token
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"username":"alice","password":"123456"}'

# 查看商品
curl http://localhost:8080/api/v1/products \
  -H 'Authorization: Bearer <YOUR_TOKEN>'
```

## 🏗️ 微服务架构

```
                       ┌─────────────────┐
                       │ HTTP Client     │
                       └────────┬────────┘
                                ↓ HTTP
                       ┌─────────────────┐
                       │  api-gateway    │  :8080
                       │  (HTTP → gRPC)  │
                       └────────┬────────┘
                                ↓ gRPC
        ┌──────────────┬────────┴────────┬──────────────┐
        ↓              ↓                 ↓              ↓
   ┌─────────┐   ┌──────────┐    ┌──────────┐   ┌──────────┐
   │  user   │   │ product  │    │  order   │   │ payment  │
   │ :50051  │   │  :50052  │    │  :50053  │   │  :50054  │
   └────┬────┘   └─────┬────┘    └────┬─────┘   └─────┬────┘
        │              ↑              │ ↗            │
        │              └──────────────┘ │            │
        │              扣减/恢复库存    └────────────┘
        │                                  更新支付状态
        ↓
    ┌──────┐       ┌──────┐       ┌──────┐
    │MySQL │       │ NATS │       │Redis │
    └──────┘       └──────┘       └──────┘
```

**跨服务调用**：
- `order-service` → `product-service`（gRPC 客户端调用扣减/恢复库存）
- `payment-service` → `order-service`（gRPC 客户端更新订单状态）

## 🛠️ 技术栈

| 组件 | 用途 |
|------|------|
| Go 1.21+ | 主语言 |
| gRPC + Protobuf | 服务间通信 |
| GORM + MySQL | 数据持久化 |
| Redis | 缓存（预留） |
| NATS JetStream | 消息队列（预留） |
| JWT | 身份认证 |
| gorilla/mux | HTTP 路由（网关） |
| zap | 结构化日志 |
| viper | 配置加载 |
| Docker Compose | 基础设施编排 |

## 🔧 重新生成 Protobuf

```bash
cd services
protoc --go_out=. --go_opt=paths=source_relative \
       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
       api/user/v1/user.proto \
       api/product/v1/product.proto \
       api/order/v1/order.proto \
       api/payment/v1/payment.proto
```

## ✅ 验证

```bash
cd services
go build ./...   # 全部编译通过
go vet ./...     # 静态检查通过
```

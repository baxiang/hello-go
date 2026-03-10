# Kratos 微服务电商平台

基于 Kratos 框架构建的工业级微服务电商平台，使用 NATS 作为消息队列。

## 系统架构

```
┌─────────────────────────────────────────────────────────────────────────┐
│                              API Gateway (8080)                          │
│                         Kratos + CORS + Tracing                         │
└─────────────────────────────────┬───────────────────────────────────────┘
                                  │
        ┌─────────────────────────┼─────────────────────────┐
        │                         │                         │
        ▼                         ▼                         ▼
┌───────────────┐         ┌───────────────┐         ┌───────────────┐
│ User Service  │         │Product Service│         │ Order Service │
│   (8081)      │         │   (8082)      │         │   (8083)      │
│   gRPC+HTTP   │         │  gRPC+HTTP    │         │  gRPC+HTTP    │
└───────┬───────┘         └───────┬───────┘         └───────┬───────┘
        │                         │                         │
        └─────────────────────────┼─────────────────────────┘
                                  │
                                  ▼
                    ┌───────────────────────────┐
                    │   Payment Service (8084)  │
                    │      gRPC + HTTP          │
                    └───────────────────────────┘
                                  │
                                  ▼
                    ┌───────────────────────────┐
                    │    NATS JetStream         │
                    │   (消息队列 + 持久化)      │
                    └───────────────────────────┘
```

## 服务列表

| 服务 | 端口 | 协议 | 说明 |
|------|------|------|------|
| api-gateway | 8080 | HTTP | API 网关 |
| user-service | 8081 | gRPC + HTTP | 用户服务 |
| product-service | 8082 | gRPC + HTTP | 商品服务 |
| order-service | 8083 | gRPC + HTTP | 订单服务 |
| payment-service | 8084 | gRPC + HTTP | 支付服务 |

## 技术栈

- **框架**: Kratos
- **消息队列**: NATS JetStream
- **数据库**: MySQL + GORM
- **缓存**: Redis
- **认证**: JWT
- **监控**: Prometheus + Grafana
- **链路追踪**: Jaeger
- **容器化**: Docker + Docker Compose
- **编排**: Kubernetes

## 快速开始

### 前置要求

- Go 1.20+
- Docker & Docker Compose
- protobuf 编译器
- NATS Server

### 启动服务

```bash
# 1. 启动 NATS 和 MySQL
docker-compose up -d nats mysql redis

# 2. 启动各个服务
cd user-service && go run .
cd product-service && go run .
cd order-service && go run .
cd payment-service && go run .
cd api-gateway && go run .
```

### API 文档

- User Service: http://localhost:8081/helloworld.Greeter/SayHello
- Product Service: http://localhost:8082/helloworld.Greeter/SayHello
- Order Service: http://localhost:8083/helloworld.Greeter/SayHello
- Payment Service: http://localhost:8084/helloworld.Greeter/SayHello
- API Gateway: http://localhost:8080/

## 项目结构

```
kratos/
├── api/                          # Protobuf API 定义
│   └── user/
│       └── v1/
│           ├── user.proto
│           └── user_grpc.pb.go
├── cmd/                          # 入口文件
│   ├── api-gateway/
│   ├── user-service/
│   ├── product-service/
│   ├── order-service/
│   └── payment-service/
├── internal/                     # 内部包
│   ├── biz/                      # 业务逻辑层
│   ├── data/                    # 数据访问层
│   ├── server/                  # 服务器配置
│   ├── service/                 # 服务实现
│   └── client/                  # 客户端
├── third_party/                  # 第三方 proto
├── configs/                     # 配置文件
├── deployments/                 # 部署配置
│   ├── docker/
│   └── k8s/
└── go.mod
```

## 消息流

1. **创建订单**: Client -> API Gateway -> Order Service -> NATS (order.created)
2. **扣减库存**: Order Service -> Product Service (via NATS)
3. **处理支付**: Product Service -> Payment Service (via NATS)
4. **发送通知**: Payment Service -> Notification (via NATS)

## 开发指南

### 生成 Protobuf 代码

```bash
# 安装 protoc 插件
go install github.com/go-kratos/kratos/cmd/protoc-gen-kratos/v2@latest
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# 生成代码
protoc --proto_path=. \
       --proto_path=./third_party \
       --go_out=paths=source_relative:. \
       --go-grpc_out=paths=source_relative:. \
       --kratos_out=paths=source_relative:. \
       api/user/v1/user.proto
```

### 添加新服务

1. 在 `cmd/` 下创建服务目录
2. 在 `api/` 下定义 Protobuf 接口
3. 实现 `internal/biz`、`internal/data`、`internal/service`
4. 在 `cmd/main.go` 中注册服务
5. 在 `deployments/docker/docker-compose.yml` 中添加服务

## 监控

- Prometheus: http://localhost:9090
- Jaeger: http://localhost:16686
- NATS Monitor: http://localhost:8222
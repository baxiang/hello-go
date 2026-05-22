# NATS 边学边练

> 结合 `nats://localhost:4222` 实际服务器，理论与实践结合学习 NATS

## 学习环境

```bash
# NATS 服务器地址
NATS_URL="nats://localhost:4222"

# 安装依赖
cd docs/nats-guide/hands-on
go mod tidy
```

## 学习路径

| 章节 | 主题 | 理论文档 | 实践内容 |
|------|------|----------|----------|
| 01 | Subject 寻址 | [02-subjects.md](../02-subjects.md) | 通配符匹配、命名规范 |
| 02 | Pub/Sub 发布订阅 | [03-pub-sub.md](../03-pub-sub.md) | 扇出、异步订阅 |
| 03 | Request/Reply | [04-request-reply.md](../04-request-reply.md) | RPC 调用、超时处理 |
| 04 | JetStream | [05-jetstream-streams.md](../05-jetstream-streams.md) | 消息持久化、Consumer |
| 05 | KV Store | [07-jetstream-kv.md](../07-jetstream-kv.md) | 设备状态管理、Watch |

## 快速开始

```bash
# 第一章：Subject 寻址
make subject
# 或
go run ./01-subjects/main.go

# 第二章：Pub/Sub
make pubsub
# 或
go run ./02-pub-sub/main.go

# 第三章：Request/Reply
make rpc
# 或
go run ./03-request-reply/main.go

# 第四章：JetStream
make jetstream
# 或
go run ./04-jetstream/main.go

# 第五章：KV Store
make kv
# 或
go run ./05-kv/main.go
```

## 目录结构

```
hands-on/
├── README.md           # 本文件
├── go.mod              # Go 模块
├── Makefile            # 构建脚本
├── 01-subjects/        # Subject 寻址实践
│   └── main.go
├── 02-pub-sub/         # Pub/Sub 实践
│   └── main.go
├── 03-request-reply/   # Request/Reply 实践
│   └── main.go
├── 04-jetstream/       # JetStream 实践
│   └── main.go
└── 05-kv/              # KV Store 实践
    └── main.go
```

## 学习建议

1. **先读理论**：阅读对应的理论文档，理解核心概念
2. **运行示例**：运行实践代码，观察输出
3. **修改实验**：修改代码参数，观察行为变化
4. **举一反三**：基于示例实现自己的业务场景
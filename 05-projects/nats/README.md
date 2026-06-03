# NATS 学习笔记

本目录包含完整的 NATS 学习资料，包括理论文档、实战项目和代码示例。

## 目录结构

```
nats/
├── docs/                    # 理论文档（11篇系统性学习指南）
│   ├── README.md           # 系列导航和学习路径
│   ├── 01-NATS概述.md
│   ├── 02-Subject寻址系统.md
│   ├── ...
│   └── 11-Go客户端实战.md
├── projects/               # 实战项目（5个渐进式项目）
│   ├── 01-nats-core-concepts/
│   ├── 02-jetstream-persistence/
│   ├── ...
│   └── 05-microservice-ecommerce-order/
├── hands-on/               # 动手练习代码
└── example/                # 代码示例
```

## 学习路径

### 入门路径
docs/01-NATS概述 → docs/02-Subject寻址系统 → docs/03-发布订阅模式 → docs/04-请求响应模式

### 进阶路径
docs/05-JetStream-Streams → docs/06-JetStream-Consumers → docs/07-KV存储与对象存储

### 生产运维路径
docs/08-集群架构 → docs/09-JetStream高可用 → docs/10-监控与可观测性

### 开发实战路径
完成理论学习后 → projects/01-nats-core-concepts → projects/02-jetstream-persistence → ...

## 本地开发环境

### 启动 NATS Server

```bash
# 使用 Docker
docker run -d --name nats -p 4222:4222 -p 8222:8222 nats:2.11.10

# 或使用 Homebrew (macOS)
brew install nats-server
nats-server

# 或使用 Go 安装
go install github.com/nats-io/nats-server/v2@latest
nats-server
```

### 连接地址

- NATS 端口: nats://localhost:4222
- 监控端口: http://localhost:8222

## 快速开始

1. 阅读 [docs/README.md](./docs/README.md) 了解完整学习路径
2. 从 [docs/01-NATS概述.md](./docs/01-NATS概述.md) 开始学习
3. 使用 hands-on/ 目录中的代码进行练习

## 版本信息

| 组件 | 版本 |
|------|------|
| NATS Server | 2.11.10 |
| nats.go | 1.49.0 |
| Go | 1.21+ |
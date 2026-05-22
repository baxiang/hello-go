# 入门 - 配置中心

本项目演示如何使用 etcd 构建分布式配置中心。

## 项目概述

- 实现配置的集中管理
- 支持配置的实时更新和监听
- 多应用共享配置

## 目录结构

```
02-入门-配置中心/
├── README.md           # 项目说明
├── config-server/     # 配置管理服务端
│   └── main.go
├── config-client/     # 配置使用客户端
│   └── main.go
└── go.mod             # Go 模块定义
```

## 运行步骤

### 1. 启动 etcd

```bash
docker run -d --name etcd -p 2379:2379 quay.io/coreos/etcd:v3.5.17
```

### 2. 运行配置服务端

```bash
cd config-server
go run main.go
```

### 3. 运行配置客户端

```bash
cd config-client
go run main.go
```

## 功能演示

1. 配置服务端写入配置
2. 配置客户端监听配置变更
3. 配置更新时客户端自动接收通知

## 预期输出

**服务端输出：**
```
Configuration server started
Config updated: database.host = localhost
Config updated: database.port = 3306
```

**客户端输出：**
```
Watching configuration changes...
Received PUT event: database.host = localhost
Received PUT event: database.port = 3306
```

## 学习要点

1. **前缀监听**：使用 `WithPrefix()` 监听配置变更
2. **事件处理**：处理 PUT 和 DELETE 事件
3. **实时更新**：配置变更实时推送到客户端

## 相关文档

- [docs/07-Go-客户端-Watch和Lease.md](../../docs/07-Go-客户端-Watch和Lease.md)
- [docs/06-Go-客户端-KV操作.md](../../docs/06-Go-客户端-KV操作.md)
# 进阶 - 服务发现

本项目演示如何使用 etcd 实现服务注册与发现。

## 项目概述

- 实现服务注册（带租约）
- 实现服务发现（查询可用服务）
- 实现健康检查（基于租约 KeepAlive）

## 目录结构

```
03-advanced-service-discovery/
├── README.md            # 项目说明
├── service-provider/   # 服务提供者（注册）
│   └── main.go
├── service-consumer/   # 服务消费者（发现）
│   └── main.go
└── go.mod              # Go 模块定义
```

## 运行步骤

### 1. 启动 etcd

```bash
docker run -d --name etcd -p 2379:2379 quay.io/coreos/etcd:v3.5.17
```

### 2. 运行服务提供者

```bash
cd service-provider
go run main.go
```

### 3. 运行服务消费者

```bash
cd service-consumer
go run main.go
```

## 功能演示

1. 服务提供者注册到 etcd
2. 租约自动续期保持服务在线
3. 服务消费者查询可用服务实例
4. 服务下线后自动从列表移除

## 预期输出

**服务提供者输出：**
```
Service registered: userservice/instance-1 at localhost:8080
KeepAlive running...
Service is alive
```

**服务消费者输出：**
```
Discovering services...
Found 1 instance(s) of userservice
Instance 1: localhost:8080
```

## 学习要点

1. **租约机制**：使用 Lease 创建临时键
2. **KeepAlive**：保持租约存活，防止过期
3. **前缀查询**：查询指定服务的所有实例
4. **服务注册格式**：`/services/<service>/<instance>`

## 相关文档

- [docs/07-Go-客户端-Watch和Lease.md](../../docs/07-Go-客户端-Watch和Lease.md)
- [docs/06-Go-客户端-KV操作.md](../../docs/06-Go-客户端-KV操作.md)
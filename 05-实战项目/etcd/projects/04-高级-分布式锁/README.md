# 高级 - 分布式锁

本项目演示如何使用 etcd 实现分布式锁。

## 项目概述

- 实现分布式互斥锁
- 支持锁的超时和自动释放
- 模拟多进程竞争锁的场景

## 目录结构

```
04-高级-分布式锁/
├── README.md           # 项目说明
├── lock-demo/         # 锁使用示例
│   └── main.go
├── lock-client1/      # 客户端 1
│   └── main.go
├── lock-client2/      # 客户端 2
│   └── main.go
└── go.mod             # Go 模块定义
```

## 运行步骤

### 1. 启动 etcd

```bash
docker run -d --name etcd -p 2379:2379 quay.io/coreos/etcd:v3.5.17
```

### 2. 运行单个锁示例

```bash
cd lock-demo
go run main.go
```

### 3. 运行多个客户端竞争锁

```bash
# 终端 1
cd lock-client1
go run main.go

# 终端 2
cd lock-client2
go run main.go
```

## 功能演示

1. 客户端尝试获取分布式锁
2. 只有获得锁的客户端能执行关键操作
3. 锁释放后其他客户端可获取锁
4. 锁超时自动释放（基于租约）

## 预期输出

**lock-demo 输出：**
```
Trying to acquire lock...
Lock acquired!
Doing critical work...
Lock released
```

**lock-client1 输出：**
```
Client 1: Trying to acquire lock...
Client 1: Lock acquired!
Client 1: Doing work...
Client 1: Lock released
```

**lock-client2 输出：**
```
Client 2: Trying to acquire lock...
Client 2: Waiting for lock...
Client 2: Lock acquired (after client 1 released)
Client 2: Doing work...
Client 2: Lock released
```

## 学习要点

1. **事务 CAS**：使用事务实现原子性锁获取
2. **租约绑定**：锁键绑定租约，超时自动释放
3. **KeepAlive**：保持锁存活
4. **锁竞争**：多个客户端竞争同一锁
5. **SDK 使用**：使用 `concurrency.Mutex`

## 相关文档

- [docs/08-Go-客户端-事务和锁.md](../../docs/08-Go-客户端-事务和锁.md)
- [docs/07-Go-客户端-Watch和Lease.md](../../docs/07-Go-客户端-Watch和Lease.md)
# etcd 概述

## 什么是 etcd

etcd 是一个高可用的分布式键值存储系统，由 CoreOS 团队开发，用于存储需要被分布式系统或机器集群访问的关键数据。它在分布式系统中扮演着至关重要的角色，是 Kubernetes 等云原生系统的核心组件。

### 核心特性

- **简单易用**：提供标准的 gRPC API 和 HTTP API，支持多种编程语言
- **强一致性**：基于 Raft 共识算法，保证数据一致性
- **高可用**：支持多节点集群部署，自动故障转移
- **可扩展**：支持动态增删节点
- **安全**：支持 TLS 加密通信和基于角色的访问控制（RBAC）
- **可靠**：数据持久化存储，支持快照和增量备份

## 核心架构

### 整体架构

```
┌─────────────────────────────────────────────────────────┐
│                      Client Applications                │
└────────────────────┬────────────────────────────────────┘
                     │ gRPC/HTTP
┌────────────────────┴────────────────────────────────────┐
│                      etcd Server                         │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  │
│  │   API Layer  │  │  Raft Layer  │  │ Storage Layer│  │
│  │  (gRPC/HTTP) │  │ (Consensus)  │  │   (WAL+BBolt)│  │
│  └──────────────┘  └──────────────┘  └──────────────┘  │
└─────────────────────────────────────────────────────────┘
```

### 核心组件

1. **API 层**：处理客户端请求，提供 gRPC 和 HTTP 接口
2. **Raft 层**：实现共识算法，保证分布式一致性
3. **存储层**：数据持久化，使用 WAL（预写日志）和 BBolt 数据库

## 应用场景

### 1. 配置管理

存储分布式系统的配置信息，支持实时更新和监听。

```go
// 存储配置
put /config/database/host "localhost"
put /config/database/port "3306"

// 应用监听配置变更
watch /config/database/
```

### 2. 服务发现

注册和发现服务实例，实现动态服务发现。

```
/services/userservice/instance1 -> {"host": "192.168.1.10", "port": 8080}
/services/userservice/instance2 -> {"host": "192.168.1.11", "port": 8080}
```

### 3. 分布式锁

实现跨进程、跨机器的互斥锁，保证资源访问的安全性。

```go
// 获取分布式锁
lock /locks/my-resource
// 执行关键操作
// ...
// 释放锁
unlock /locks/my-resource
```

### 4. Leader 选举

在多个实例中选举出主节点，实现主备切换。

```
/election/leader -> "node-1"
```

### 5. Kubernetes 集群状态

Kubernetes 使用 etcd 存储所有集群状态数据，是 Kubernetes 的核心数据存储。

```
/registry/pods/default/my-pod
/registry/services/default/my-service
/registry/configmaps/default/my-config
```

## etcd vs 其他存储系统

| 特性 | etcd | Redis | ZooKeeper | Consul |
|------|------|-------|-----------|--------|
| 一致性模型 | 强一致 | 最终一致 | 强一致 | 强一致 |
| 共识算法 | Raft | 无 | ZAB | Raft |
| 数据模型 | KV | 多种数据结构 | 层级命名空间 | KV |
| API | gRPC/HTTP | Redis 协议 | 自定义 | HTTP/gRPC |
| 性能 | 高 | 很高 | 中等 | 高 |
| 适用场景 | 配置、元数据 | 缓存、会话 | 配置、协调 | 服务发现 |

## 基本概念

### 键值对

etcd 的基本数据单元，键和值都是字节数组。

```bash
# 设置键值
etcdctl put /message "Hello, etcd"

# 获取值
etcdctl get /message
```

### Revision（修订版本）

全局递增的版本号，每次修改操作都会增加 revision。

```
Revision 1: put /foo bar
Revision 2: put /foo bar2
Revision 3: put /baz qux
```

### Lease（租约）

键的生存时间机制，租约到期后关联的键自动删除。

```bash
# 创建 60 秒租约
etcdctl lease grant 60
lease 7581594885984314185 granted with TTL(60s)

# 将键绑定到租约
etcdctl put --lease=7581594885984314185 /temp "temporary"
```

### Watch（监听）

监听键或前缀的变更事件，实现实时通知。

```bash
# 监听 /foo 的变更
etcdctl watch /foo
```

### Transaction（事务）

原子性执行多个操作，支持条件判断。

```go
// 伪代码
txn.if(key.value == "expected")
   .then(put(key, "new-value"))
   .else(get(key))
```

## 快速开始

### 启动单节点

```bash
# 使用 Docker
docker run -d --name etcd \
  -p 2379:2379 \
  -p 2380:2380 \
  quay.io/coreos/etcd:v3.5.17 \
  /usr/local/bin/etcd \
  --name s1 \
  --data-dir /etcd-data \
  --listen-client-urls http://0.0.0.0:2379 \
  --advertise-client-urls http://0.0.0.0:2379 \
  --listen-peer-urls http://0.0.0.0:2380 \
  --initial-advertise-peer-urls http://0.0.0.0:2380 \
  --initial-cluster s1=http://0.0.0.0:2380 \
  --initial-cluster-token tkn \
  --initial-cluster-state new
```

### 使用 etcdctl

```bash
# 安装 etcdctl
# macOS: brew install etcd
# Linux: 从 GitHub releases 下载

# 设置键值
etcdctl put /message "Hello, etcd"

# 获取值
etcdctl get /message

# 获取所有键
etcdctl get "" --prefix

# 删除键
etcdctl del /message
```

### Go 客户端示例

```go
package main

import (
    "context"
    "fmt"
    "time"

    clientv3 "go.etcd.io/etcd/client/v3"
)

func main() {
    // 创建客户端
    cli, err := clientv3.New(clientv3.Config{
        Endpoints:   []string{"localhost:2379"},
        DialTimeout: 5 * time.Second,
    })
    if err != nil {
        panic(err)
    }
    defer cli.Close()

    // 设置键值
    ctx, cancel := context.WithTimeout(context.Background(), time.Second)
    _, err = cli.Put(ctx, "/message", "Hello, etcd")
    cancel()
    if err != nil {
        panic(err)
    }

    // 获取值
    ctx, cancel = context.WithTimeout(context.Background(), time.Second)
    resp, err := cli.Get(ctx, "/message")
    cancel()
    if err != nil {
        panic(err)
    }

    for _, ev := range resp.Kvs {
        fmt.Printf("%s : %s\n", ev.Key, ev.Value)
    }
}
```

## 性能特点

- **读性能**：单节点可达 10,000+ QPS
- **写性能**：单节点可达 10,000+ QPS
- **延迟**：P99 延迟通常在 10ms 以内
- **数据量**：建议单集群数据量不超过 8GB
- **集群规模**：建议 3、5、7 个节点（奇数）

## 最佳实践

1. **集群规模**：生产环境至少 3 个节点，推荐 5 个
2. **硬件配置**：使用 SSD 存储，保证磁盘 I/O 性能
3. **网络**：节点间网络延迟应小于 10ms
4. **备份策略**：定期创建快照备份
5. **监控**：监控关键指标（延迟、吞吐量、存储大小）
6. **安全**：生产环境启用 TLS 和 RBAC

## 总结

etcd 是云原生生态系统的核心组件，以其强一致性、高可用性和简洁的 API 设计成为分布式系统的首选配置存储方案。理解 etcd 的核心概念和架构对于构建可靠的分布式系统至关重要。

在接下来的章节中，我们将深入学习：
- 数据模型和存储机制
- Raft 共识算法原理
- 集群管理和运维
- Go 客户端开发实战
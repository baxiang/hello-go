# Redis 概述

## 什么是 Redis

Redis（Remote Dictionary Server）是一个开源的内存数据结构存储系统，由 Salvatore Sanfilippo 于 2009 年开发。它可以用作数据库、缓存、消息中间件和流处理引擎。

### 核心特性

- **内存存储**：数据存储在内存中，读写性能极高
- **数据结构丰富**：支持 String、Hash、List、Set、ZSet、Stream 等
- **持久化**：支持 RDB 和 AOF 两种持久化方式
- **高可用**：支持主从复制、哨兵、集群模式
- **功能强大**：支持事务、Lua 脚本、Pub/Sub、Stream 等

## 应用场景

### 1. 缓存

最常见的使用场景，提升应用性能。

```go
// 查询缓存
value, err := rdb.Get(ctx, "user:1001").Result()
if err == redis.Nil {
    // 缓存未命中，从数据库查询
    value = getUserFromDB(1001)
    // 写入缓存，设置过期时间
    rdb.Set(ctx, "user:1001", value, 30*time.Minute)
}
```

### 2. 会话存储

存储用户会话信息，支持分布式环境。

```
session:token123 → {"user_id": 1001, "login_time": 1234567890}
```

### 3. 排行榜

使用 ZSet 实现实时排行榜。

```go
// 添加分数
rdb.ZAdd(ctx, "leaderboard", redis.Z{Score: 95.5, Member: "user1"})

// 获取排名
rdb.ZRevRange(ctx, "leaderboard", 0, 9, "WITHSCORES")
```

### 4. 消息队列

使用 List 或 Stream 实现简单消息队列。

```go
// 生产者
rdb.LPush(ctx, "task_queue", taskData)

// 消费者
result, _ := rdb.BRPop(ctx, 0, "task_queue").Result()
```

### 5. 分布式锁

使用 SET NX EX 实现分布式锁。

```go
// 获取锁
ok, _ := rdb.SetNX(ctx, "lock:resource", "owner", 10*time.Second).Result()
if ok {
    // 获取锁成功
    defer rdb.Del(ctx, "lock:resource")
}
```

### 6. 计数器

使用 INCR 实现原子计数。

```go
// 页面访问计数
rdb.Incr(ctx, "page:home:views")

// 用户点赞
rdb.Incr(ctx, "article:123:likes")
```

## 架构设计

### 单机架构

```
┌─────────────────┐
│   Application   │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  Redis Server   │
│    (Memory)     │
└─────────────────┘
```

适用于：
- 开发测试环境
- 小规模应用
- 单点缓存

### 主从复制架构

```
┌─────────────────┐
│   Application   │
└────────┬────────┘
         │
    ┌────┴────┐
    ▼         ▼
┌────────┐ ┌────────┐
│ Master │ │ Slave  │
│ (RW)   │ │ (R)    │
└────────┘ └────────┘
```

适用于：
- 读写分离
- 数据备份
- 高可用（需配合哨兵）

### 哨兵架构

```
┌──────────┐  ┌──────────┐  ┌──────────┐
│ Sentinel │  │ Sentinel │  │ Sentinel │
└────┬─────┘  └────┬─────┘  └────┬─────┘
     │             │             │
     └─────────────┼─────────────┘
                   │ monitor
          ┌────────┴────────┐
          ▼                 ▼
    ┌──────────┐      ┌──────────┐
    │  Master  │──────▶│  Slave   │
    └──────────┘      └──────────┘
```

功能：
- 监控：检查主从节点是否正常
- 提醒：发现问题时发送通知
- 自动故障转移：主节点故障时选举新主

### 集群架构

```
┌──────────┐  ┌──────────┐  ┌──────────┐
│ Master 1 │  │ Master 2 │  │ Master 3 │
│ 0-5460   │  │ 5461-10922│ │ 10923-16383│
└────┬─────┘  └────┬─────┘  └────┬─────┘
     │             │             │
┌────┴────┐  ┌────┴────┐  ┌────┴────┐
│ Slave 1 │  │ Slave 2 │  │ Slave 3 │
└─────────┘  └─────────┘  └─────────┘
```

特点：
- 数据分片：数据分布在多个节点
- 水平扩展：支持 16384 个槽位
- 高可用：每个主节点有从节点

## Redis vs Memcached

| 特性 | Redis | Memcached |
|------|-------|-----------|
| 数据类型 | 丰富（String、Hash、List等） | 仅 String |
| 持久化 | 支持（RDB、AOF） | 不支持 |
| 集群 | 原生支持 | 需要客户端分片 |
| 线程模型 | 单线程 | 多线程 |
| 事务 | 支持 | 不支持 |
| 性能 | 高（单线程避免锁） | 高（多线程） |
| 内存管理 | 更灵活 | 更简单 |

## Redis vs etcd

| 特性 | Redis | etcd |
|------|-------|------|
| 定位 | 内存数据结构存储 | 分布式配置存储 |
| 一致性 | 最终一致 | 强一致（Raft） |
| 数据模型 | KV + 数据结构 | KV + Revision |
| 性能 | 更高（内存） | 高（磁盘+缓存） |
| 应用场景 | 缓存、会话、队列 | 配置中心、服务发现 |

## 性能特点

- **QPS**：单节点可达 10万+ QPS
- **延迟**：P99 < 1ms（内存访问）
- **内存**：建议单个实例内存 < 10GB
- **网络**：支持管道（Pipeline）批量操作

## 安装部署

### Docker 部署

```bash
# 基础部署
docker run -d --name redis \
  -p 6379:6379 \
  redis:7.2.4

# 带配置文件部署
docker run -d --name redis \
  -p 6379:6379 \
  -v /path/to/redis.conf:/usr/local/etc/redis/redis.conf \
  redis:7.2.4 \
  redis-server /usr/local/etc/redis/redis.conf
```

### 编译安装

```bash
# 下载源码
wget https://github.com/redis/redis/archive/refs/tags/7.2.4.tar.gz

# 编译
tar xzf 7.2.4.tar.gz
cd redis-7.2.4
make

# 启动
./src/redis-server
```

### 配置文件

```conf
# redis.conf

# 绑定地址
bind 0.0.0.0

# 端口
port 6379

# 密码
requirepass your_password

# 持久化
save 900 1
save 300 10
save 60 10000

# AOF
appendonly yes
appendfsync everysec

# 最大内存
maxmemory 2gb
maxmemory-policy allkeys-lru
```

## 基本命令

### 连接

```bash
# 连接本地
redis-cli

# 连接指定主机和端口
redis-cli -h 127.0.0.1 -p 6379

# 带密码连接
redis-cli -h 127.0.0.1 -p 6379 -a your_password
```

### 数据操作

```bash
# String
SET key value
GET key
DEL key

# Hash
HSET user:1001 name "Alice"
HGET user:1001 name

# List
LPUSH mylist value
RPOP mylist

# Set
SADD myset member
SMEMBERS myset

# ZSet
ZADD myzset 1 member
ZRANGE myzset 0 -1
```

### 服务器管理

```bash
# 查看信息
INFO

# 查看统计
STATS

# 监控命令
MONITOR

# 清空数据库
FLUSHDB
FLUSHALL
```

## 客户端库

### go-redis（推荐）

```go
import "github.com/redis/go-redis/v9"

rdb := redis.NewClient(&redis.Options{
    Addr:     "localhost:6379",
    Password: "",
    DB:       0,
})
```

### redigo

```go
import "github.com/gomodule/redigo/redis"

c, _ := redis.Dial("tcp", "localhost:6379")
defer c.Close()
```

## 最佳实践

1. **键命名规范**
   - 使用冒号分隔：`user:1001:profile`
   - 避免过长：键名会影响内存占用

2. **内存优化**
   - 使用 Hash 替代多个 String
   - 合理设置过期时间
   - 使用合适的数据结构

3. **性能优化**
   - 使用管道批量操作
   - 避免大 Key（> 10KB）
   - 使用连接池

4. **安全配置**
   - 设置密码
   - 禁用危险命令（FLUSHALL、CONFIG）
   - 使用防火墙限制访问

## 监控指标

```
# 内存使用
used_memory
used_memory_peak

# 连接数
connected_clients

# 命令统计
total_commands_processed
instantaneous_ops_per_sec

# 持久化
rdb_last_save_time
aof_last_rewrite_time_sec

# 复制
connected_slaves
```

## 学习路径

1. 掌握基本数据类型和操作
2. 理解持久化机制
3. 学习主从复制和哨兵
4. 掌握集群部署和运维
5. 实战应用开发

在下一章中，我们将深入学习 Redis 的数据类型和操作。
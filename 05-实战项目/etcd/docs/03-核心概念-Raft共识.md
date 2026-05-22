# 核心概念 - Raft 共识

etcd 的分布式一致性依赖于 Raft 共识算法。本章将深入探讨 Raft 算法的原理及其在 etcd 中的实现。

## 分布式一致性问题

### 问题背景

在分布式系统中，多个节点需要就数据达成一致，面临以下挑战：

1. **网络分区**：节点间通信可能中断
2. **节点故障**：节点可能随时宕机
3. **消息延迟/丢失**：网络不可靠

### CAP 定理

分布式系统最多只能同时满足以下三项中的两项：

- **Consistency（一致性）**：所有节点看到相同的数据
- **Availability（可用性）**：每个请求都能得到响应
- **Partition Tolerance（分区容错）**：网络分区时系统仍能运行

etcd 选择 CP：强一致性 + 分区容错。

## Raft 算法概述

Raft 是一种易于理解的分布式共识算法，由 Diego Ongaro 和 John Ousterhout 在 2014 年提出。

### 核心目标

1. **Leader 选举**：选出一个 Leader 负责管理日志复制
2. **日志复制**：Leader 接收客户端请求并复制到其他节点
3. **安全性**：保证所有已提交的日志条目不会被覆盖

### 节点角色

```
┌─────────────────────────────────────────────────────┐
│                                                     │
│   ┌──────────┐                                      │
│   │  Leader  │ ◄─── 接收客户端请求                  │
│   └────┬─────┘                                      │
│        │ 心跳/日志复制                              │
│   ┌────┴────┐                                       │
│   ▼         ▼                                       │
│ ┌────────┐ ┌────────┐                              │
│ │Follower│ │Follower│                              │
│ └────────┘ └────────┘                              │
│                                                     │
│  可能存在 Candidate（候选者）参与选举               │
└─────────────────────────────────────────────────────┘
```

- **Follower（跟随者）**：被动响应 Leader 和 Candidate 的请求
- **Candidate（候选者）**：参与选举，尝试成为 Leader
- **Leader（领导者）**：处理所有客户端请求，复制日志到 Follower

## Leader 选举

### 选举过程

```
时间线：
┌─────────────────────────────────────────────────────>
     │                   │                   │
  选举超时            选举超时            选举超时
     │                   │                   │
     ▼                   ▼                   ▼
 Follower           Follower           Follower
     │                   │                   │
     │ 超时无心跳        │                   │
     ▼                   │                   │
 Candidate ────────────>│ 发起投票          │
     │ RequestVote      │                   │
     │                  ▼                   │
     │              投票给 Candidate       │
     │                  │                   │
     │◄─────────────────┘                   │
     │ 获得多数票                           │
     ▼                                      │
  Leader ────────────────>心跳─────────────>│
```

### 选举超时

- 每个节点有随机选举超时时间（如 150-300ms）
- 超时期间未收到 Leader 心跳，转变为 Candidate
- 随机化避免同时发起选举（减少选票瓜分）

### 选举规则

1. 每个 term 最多投一票，先到先得
2. Candidate 需要获得多数节点投票才能成为 Leader
3. 新 Leader 立即发送心跳确立领导地位

### etcd 中的实现

```go
// 选举超时配置
type RaftConfig struct {
    // 选举超时 = base + rand(0, electionTick)
    ElectionTick    int    // 默认 10
    HeartbeatTick   int    // 默认 1
}
```

```bash
# 查看当前 Leader
etcdctl endpoint status --cluster -w table

# 输出示例
+------------------------+------------------+---------+---------+-----------+------------+
|        ENDPOINT        |        ID        | VERSION | DB SIZE | IS LEADER | RAFT INDEX |
+------------------------+------------------+---------+---------+-----------+------------+
| http://localhost:2379 | 8e9e05c52164694d |  3.5.17 |   20 kB |     false |          10|
| http://localhost:22379| fd422379fda50e48 |  3.5.17 |   20 kB |     false |          10|
| http://localhost:32379| 7a61a32f8669483a |  3.5.17 |   20 kB |      true |          10|
+------------------------+------------------+---------+---------+-----------+------------+
```

## 日志复制

### 日志结构

```
日志条目（Log Entry）：
┌──────────────────────────────────────────┐
│ Index | Term | Command                   │
├───────┼──────┼───────────────────────────┤
│   1   |  1   | Set /foo = "bar1"         │
│   2   |  1   | Set /baz = "qux"          │
│   3   |  2   | Set /foo = "bar2"         │
│   4   |  2   | Delete /baz              │
└───────┴──────┴───────────────────────────┘
```

- **Index**：日志索引，递增
- **Term**：任期号，用于检测过时 Leader
- **Command**：具体操作（Put、Delete 等）

### 复制流程

```
Client                Leader              Follower
  │                     │                    │
  │  Put /foo = "bar"   │                    │
  │────────────────────>│                    │
  │                     │                    │
  │                     │  AppendEntries RPC │
  │                     │───────────────────>│
  │                     │                    │ 写入本地日志
  │                     │  AppendEntries Resp │
  │                     │<───────────────────│ (success)
  │                     │                    │
  │                     │ 应用到状态机        │
  │                     │ (Commit Index)     │
  │                     │                    │
  │      Response       │                    │
  │<────────────────────│                    │
```

### 提交（Commit）

- **已提交（Committed）**：日志被复制到多数节点
- **已应用（Applied）**：已提交的日志应用到状态机
- Leader 更新 commitIndex，通知 Follower 应用日志

### 一致性保证

1. **日志匹配特性**：如果两个日志在相同索引有相同 term，则之前的日志都相同
2. **Leader 完整性**：Leader 包含所有已提交的日志
3. **状态机安全性**：所有节点按相同顺序应用相同日志

## 安全性规则

### Term 机制

- Term 是逻辑时钟，每次选举递增
- 节点发现更高 term 时，立即转为 Follower
- 保证同一时间最多一个 Leader

### 投票限制

Candidate 请求投票时，携带自己的日志信息：

```
RequestVote RPC:
- term: Candidate 的 term
- candidateId: Candidate ID
- lastLogIndex: 最后一条日志的索引
- lastLogTerm: 最后一条日志的 term
```

投票者只投票给日志至少和自己一样新的 Candidate：

```go
// 日志比较逻辑
func isUpToDate(lastLogTerm, lastLogIndex int, myLastLogTerm, myLastLogIndex int) bool {
    if lastLogTerm != myLastLogTerm {
        return lastLogTerm > myLastLogTerm
    }
    return lastLogIndex >= myLastLogIndex
}
```

### 防止脑裂

```
网络分区场景：
┌─────────────┐         ┌─────────────┐
│ Partition A │         │ Partition B │
│             │         │             │
│   Leader    │ ─ ─ ─ ─ │  旧 Leader  │
│  (Term 3)   │  网络    │  (Term 2)   │
│   Follower  │  分区    │             │
└─────────────┘         └─────────────┘
    3节点                  2节点
    多数派                 少数派

Partition A 可以继续工作（3/5 = 多数）
Partition B 旧 Leader 的写操作会失败（无法获得多数确认）
```

## etcd 中的 Raft 实现

### 关键组件

```
┌────────────────────────────────────────────────────┐
│                    etcd Server                      │
│  ┌──────────────┐  ┌──────────────┐               │
│  │   Raft Node  │  │   Storage    │               │
│  │              │  │  (WAL+BBolt) │               │
│  └──────┬───────┘  └──────────────┘               │
│         │                                          │
│  ┌──────┴───────┐  ┌──────────────┐               │
│  │   Transport  │  │   Apply      │               │
│  │  (gRPC)      │  │   Channel    │               │
│  └──────────────┘  └──────────────┘               │
└────────────────────────────────────────────────────┘
```

### 状态持久化

```
持久化状态（写入 WAL）：
- currentTerm：当前任期
- votedFor：投票给谁
- log[]：日志条目

易失状态（内存）：
- commitIndex：已知的最高已提交日志索引
- lastApplied：最后应用到状态机的日志索引
- nextIndex[]：对每个节点，下一个要发送的日志索引
- matchIndex[]：对每个节点，已复制的最高日志索引
```

### ReadIndex 读优化

为保证线性一致性读，Leader 需要：

```
1. 记录当前 commitIndex 为 readIndex
2. 向多数节点发送心跳确认自己是 Leader
3. 等待状态机应用到 readIndex
4. 执行读操作
```

```bash
# etcd 线性一致性读（默认）
etcdctl get /foo

# 串行化读（性能更高，可能读到旧数据）
etcdctl get /foo --endpoints=http://localhost:2379 --consistency=s
```

### Lease 机制

etcd 的租约与 Raft 紧密结合：

```
1. 客户端创建租约，指定 TTL
2. Leader 维护租约到期时间
3. Leader 崩溃时，新 Leader 从日志恢复租约状态
4. 租约到期后，Leader 删除关联的键
```

## 故障处理

### Leader 故障

```
Leader 故障恢复：
1. Follower 选举超时
2. 发起选举，获得多数票
3. 新 Leader 上任
4. 新 Leader 发送心跳，同步日志
```

### Follower 故障

```
Follower 故障恢复：
1. Leader 持续重试 AppendEntries
2. Follower 重启后，从 Leader 同步缺失日志
3. 日志匹配后继续正常工作
```

### 网络分区

```
网络分区处理：
1. 少数派分区的 Leader 无法提交日志（无多数确认）
2. 多数派选举新 Leader
3. 网络恢复后，旧 Leader 发现更高 term
4. 旧 Leader 转为 Follower，回滚未提交日志
```

## 性能优化

### 批量处理

- Leader 批量发送日志给 Follower
- 减少网络往返次数

### 管道化

- Leader 不等待 Follower 确认就发送下一批日志
- 提高吞吐量

### 写入优化

```go
// etcd 配置
type Config struct {
    // 每批最大条目数
    MaxBatchSize int
    // 批量等待时间
    MaxBatchWait time.Duration
}
```

## 监控指标

```bash
# 查看集群健康状态
etcdctl endpoint health --cluster

# 查看 Raft 状态指标
curl http://localhost:2379/metrics | grep etcd_server_has_leader
curl http://localhost:2379/metrics | grep etcd_server_leader_changes_seen_total

# 关键指标
# - etcd_server_has_leader: 是否有 Leader
# - etcd_server_leader_changes_seen_total: Leader 变更次数
# - etcd_raft_commit_index: 已提交日志索引
# - etcd_raft_applied_index: 已应用日志索引
```

## 总结

Raft 算法通过以下机制保证分布式一致性：

1. **Leader 选举**：选出一个 Leader 协调日志复制
2. **日志复制**：Leader 将客户端操作复制到多数节点
3. **安全性规则**：保证已提交日志不被覆盖
4. **Term 机制**：检测过时 Leader，防止脑裂

etcd 基于 Raft 实现了：
- 强一致性读写
- 高可用性（容忍少数节点故障）
- 线性一致性保证

理解 Raft 对于正确使用和运维 etcd 至关重要。在下一章中，我们将学习如何搭建和管理 etcd 集群。
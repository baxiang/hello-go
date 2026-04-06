# JetStream 高可用与 Raft 共识

## 目录

1. [JetStream HA 整体架构](#1-jetstream-ha-整体架构)
2. [Raft 共识算法基础](#2-raft-共识算法基础)
3. [NATS 中的 Raft Groups](#3-nats-中的-raft-groups)
4. [Leader Election 流程](#4-leader-election-流程)
5. [写入流程（Quorum 确认）](#5-写入流程quorum-确认)
6. [Quorum 规则](#6-quorum-规则)
7. [Stream Replicas 配置](#7-stream-replicas-配置)
8. [节点故障场景](#8-节点故障场景)
9. [节点恢复流程](#9-节点恢复流程)
10. [JetStream 集群监控](#10-jetstream-集群监控)
11. [meta_cluster.leader 与 Stream Leader 的区别](#11-meta_clusterleader-与-stream-leader-的区别)
12. [滚动升级期间 Raft 的行为](#12-滚动升级期间-raft-的行为)
13. [代码示例](#13-代码示例)

---

## 1. JetStream HA 整体架构

JetStream 的高可用依赖 NATS 集群（Cluster Mode）。在集群模式下，每个 NATS Server 既是消息路由节点，也是 JetStream 数据节点。

```
                   ┌─────────────────────────────────────────┐
                   │          NATS Cluster (3 Nodes)          │
                   │                                          │
                   │  ┌──────────┐  ┌──────────┐  ┌────────┐ │
                   │  │  Node-1  │  │  Node-2  │  │ Node-3 │ │
                   │  │          │  │          │  │        │ │
                   │  │ JS Store │  │ JS Store │  │JS Store│ │
                   │  │  Leader* │  │ Follower │  │Follower│ │
                   │  └──────────┘  └──────────┘  └────────┘ │
                   │       ↕  Raft 内部通信  ↕                 │
                   └─────────────────────────────────────────┘
                           ↑                   ↑
                     Client 1             Client 2
                  (发布消息)            (消费消息)
```

**核心设计原则：**

1. **元数据高可用（Meta Group）**：Stream/Consumer 的配置元数据通过 Raft 在所有节点间复制，任一节点宕机不影响集群管理功能。

2. **数据高可用（Stream Raft Group）**：每个 replicas>1 的 Stream 有独立的 Raft Group 管理数据副本，Leader 负责接收写入，Followers 负责复制。

3. **消费状态高可用（Consumer Raft Group）**：Durable Consumer 的 ACK 状态通过 Raft 持久化，节点重启后消费进度不丢失。

---

## 2. Raft 共识算法基础

Raft 是一种相对容易理解的共识算法，用于在分布式系统中保证多个节点的数据一致性。

### 2.1 三种节点角色

```
┌─────────────────────────────────────────────────────────┐
│                       Raft 角色                          │
│                                                         │
│  Leader    ──────  唯一处理写入请求的节点                   │
│               ├──  心跳保持 Follower 的续期                │
│               └──  发起日志复制（AppendEntries RPC）        │
│                                                         │
│  Follower  ──────  接收并应用 Leader 的日志               │
│               ├──  响应 Leader 心跳                       │
│               └──  当心跳超时后转为 Candidate              │
│                                                         │
│  Candidate ──────  向其他节点发起投票请求                   │
│               ├──  获得多数票 → 成为 Leader                │
│               └──  收到 Leader 心跳 → 回退为 Follower      │
└─────────────────────────────────────────────────────────┘
```

### 2.2 三个核心概念

**Term（任期）**
- 每次选举产生新的 Term 编号，单调递增
- Term 号更大的节点拥有更新的状态
- 解决"脑裂"：旧 Leader 重连后发现 Term 落后，自动降为 Follower

**Log（日志）**
- 所有写操作先写入 Leader 的日志（Uncommitted）
- Leader 将日志复制到多数 Follower 后，标记为 Committed
- Committed 的日志才会应用到状态机（JetStream 数据存储）

**Safety（安全性）**
- **选举安全**：一个 Term 内最多一个 Leader
- **日志匹配**：相同 index + term 的日志，内容相同
- **Leader 完整性**：Leader 一定拥有所有已提交的日志条目

### 2.3 Raft 关键参数（NATS 配置）

```yaml
# nats-server.conf 集群配置
cluster {
  name: "my-cluster"
  listen: "0.0.0.0:6222"

  routes: [
    "nats://nats-1:6222",
    "nats://nats-2:6222",
    "nats://nats-3:6222"
  ]
}

# JetStream 配置
jetstream {
  store_dir: "/data/nats"
  max_memory: 4GB
  max_file: 100GB
}
```

---

## 3. NATS 中的 Raft Groups

NATS JetStream 使用三类 Raft Group，各自管理不同范围的数据：

### 3.1 Meta Group（元数据 Raft）

```
Meta Group 管理范围：
  - 所有 Stream 的配置（名称、Subject、副本数等）
  - 所有 Consumer 的配置
  - 账号（Account）的 JetStream 资源配额

Meta Group 大小：
  - 等于集群节点数（所有节点都参与）
  - 3节点集群 → Meta Group Size = 3

Meta Leader（meta_leader）：
  - 处理所有 JetStream 资源创建/删除/更新请求
  - 例如：CreateStream、DeleteConsumer 必须经过 meta_leader
  - 用 /jsz 接口查看：meta_cluster.leader 字段
```

### 3.2 Stream Raft Group（数据 Raft）

```
每个 replicas > 1 的 Stream 有独立的 Raft Group：

Stream A（Replicas=3）  →  Raft Group: {Node-1, Node-2, Node-3}
Stream B（Replicas=3）  →  Raft Group: {Node-1, Node-2, Node-3}（Leader 可能不同）
Stream C（Replicas=1）  →  无 Raft Group（单节点，无副本）

Stream Raft Group 的 Leader（stream_leader）：
  - 接收该 Stream 的所有 Publish 消息
  - 将消息复制到 Follower
  - 处理该 Stream 的 Consumer Fetch 请求
```

### 3.3 Consumer Raft Group（消费状态 Raft）

```
Durable Consumer 的 ACK 状态也通过 Raft 持久化：

- Consumer 所在的 Raft Group 通常是 Stream Raft Group 的子集
- 记录：已 ACK 的消息序列号、Pending 消息数、RedeliveryCount
- 节点重启后，Consumer 从 Raft 日志中恢复消费状态
- Ephemeral Consumer（临时）不使用 Raft，状态仅在内存中
```

### 3.4 三类 Raft Group 的关系图

```
     集群节点：Node-1, Node-2, Node-3
     ────────────────────────────────
     Meta Group:
       Leader = Node-1
       Followers = Node-2, Node-3

     Stream "ORDERS" (R=3):
       Stream Leader = Node-2
       Followers = Node-1, Node-3

     Stream "EVENTS" (R=3):
       Stream Leader = Node-3
       Followers = Node-1, Node-2

     Consumer "ORDERS.processor" (Durable):
       Consumer Leader = Node-2（通常与 Stream Leader 相同）
       Followers = Node-1, Node-3
```

Leader 均匀分布在不同节点，避免单节点成为热点。

---

## 4. Leader Election 流程

### 4.1 正常情况：集群启动

```
时间轴 ──────────────────────────────────────────────────────▶

Node-1: [Follower] ──等待心跳超时──▶ [Candidate] ──获得多数票──▶ [Leader]
         election timer: 150ms         Term=1, vote=self          发送心跳
                                       req vote → Node-2 ✓
                                       req vote → Node-3 ✓

Node-2: [Follower] ──收到投票请求──▶ [Follower]  ──收到心跳──▶ [Follower]
         election timer: 200ms         投票给 Node-1              重置 timer

Node-3: [Follower] ──收到投票请求──▶ [Follower]  ──收到心跳──▶ [Follower]
         election timer: 180ms         投票给 Node-1              重置 timer
```

### 4.2 Leader 故障重新选举

```
时间轴 ──────────────────────────────────────────────────────▶

Node-1: [Leader] ──X 崩溃 X──
         Term=1

Node-2: [Follower] ──心跳超时──▶ [Candidate] ──获得 Node-3 票──▶ [Leader]
         timer 190ms              Term=2                           发送心跳
                                  req vote → Node-3 ✓             Term=2

Node-3: [Follower] ──心跳超时──▶ [Follower]  ──投票 Node-2──▶   [Follower]
         timer 210ms（更长）       收到 Node-2 的投票请求           Term=2
         比 Node-2 晚超时          Term=2 > Term=1，投票
```

**选举超时时间是随机的**（NATS 默认 election timeout 在 2s 附近随机化），避免多个节点同时发起选举造成分票（Split Vote）。

### 4.3 投票规则

节点 B 收到节点 A 的投票请求时，B 同意投票的条件：

```
1. A 的 Term > B 当前 Term（B 更新自己的 Term 并投票）
   OR
   A 的 Term == B 当前 Term 且 B 还未在本 Term 投票

AND

2. A 的日志不比 B 旧（保证 Leader 拥有所有已提交的日志）
   即：A.lastLogTerm > B.lastLogTerm
   OR  A.lastLogTerm == B.lastLogTerm AND A.lastLogIndex >= B.lastLogIndex
```

---

## 5. 写入流程（Quorum 确认）

### 5.1 完整写入序列图

```
Client                  Leader (Node-2)          Follower (Node-1)       Follower (Node-3)
  │                          │                          │                       │
  │── Publish("ORDERS") ────▶│                          │                       │
  │                          │                          │                       │
  │                          │── AppendEntries(log) ───▶│                       │
  │                          │── AppendEntries(log) ─────────────────────────▶ │
  │                          │                          │                       │
  │                          │                          │◀─ ACK (Seq=42) ───────│
  │                          │◀─ ACK (Seq=42) ──────────│                       │
  │                          │                          │                       │
  │                          │  [收到 2/3 节点 ACK]       │                       │
  │                          │  [Quorum 达成，提交日志]    │                       │
  │                          │                          │                       │
  │◀─ PubAck(Seq=42) ────────│                          │                       │
  │                          │── Commit Notify ────────▶│                       │
  │                          │── Commit Notify ───────────────────────────────▶│
  │                          │                          │                       │
```

**关键步骤说明：**

1. Client 将消息发送到 Stream Leader（如果连接到 Follower，Follower 会内部转发给 Leader）
2. Leader 将消息写入本地 WAL（Write-Ahead Log）
3. Leader 同时向所有 Follower 发送 AppendEntries RPC（包含日志条目）
4. 收到多数节点（包括自己）的 ACK 后，Leader 认为 Quorum 达成
5. Leader 提交日志，应用到 JetStream 存储，返回 PubAck 给 Client
6. 异步通知 Follower 也进行提交（Commit Index 推进）

### 5.2 发布确认级别

NATS JetStream 的 PubAck 保证的是 Quorum 持久化：

```go
// 等待 PubAck（Quorum 确认后返回）
ack, err := js.Publish(ctx, "ORDERS.placed", data)
if err != nil {
    // 网络错误或 Leader 不可用
    log.Printf("发布失败: %v", err)
    return
}
fmt.Printf("消息已持久化: stream=%s seq=%d\n", ack.Stream, ack.Sequence)
// 此时消息已在多数节点持久化，Leader 宕机也不会丢失
```

### 5.3 写入延迟分析

```
总延迟 = 网络往返（Client→Leader） + 本地 WAL 写入 + Raft 复制 + 网络往返（Leader→Client）

典型局域网：
  - 网络往返：~1ms
  - WAL 写入：~1-5ms（SSD）
  - Raft 复制到 Follower：~1-2ms（局域网）
  总计：约 5-10ms

跨机房（同城）：
  - 网络往返：~3-10ms
  - 总计：约 15-30ms
```

---

## 6. Quorum 规则

### 6.1 Quorum 计算公式

```
Quorum（最少需要确认的节点数） = ⌊N/2⌋ + 1

其中 N 为 Raft Group 的节点总数
```

### 6.2 节点数对比表

| 集群节点数 N | Quorum（写入需要） | 可容忍故障节点数 | 说明 |
|:-----------:|:-----------------:|:---------------:|------|
| 1 | 1 | 0 | 单节点，无 HA |
| 2 | 2 | 0 | 任一节点故障即不可写，不推荐 |
| 3 | 2 | 1 | **生产最小推荐，可容忍 1 个节点故障** |
| 4 | 3 | 1 | 容错能力与 3 节点相同，浪费资源 |
| 5 | 3 | 2 | 可容忍 2 个节点故障，适合高可用要求高的场景 |
| 6 | 4 | 2 | 与 5 节点相同容错 |
| 7 | 4 | 3 | 可容忍 3 个节点故障，大规模部署 |

### 6.3 为什么推荐奇数节点？

```
4节点集群：
  故障容忍 = 4 - 3 = 1（与 3 节点相同）
  但多了 1 个节点的资源成本
  且网络分区时：2-2 分裂，两侧都无法达到 Quorum（都不可用）

5节点集群：
  故障容忍 = 5 - 3 = 2（比 4 节点多容忍 1 个节点故障）
  网络分区时：3-2 分裂，拥有 3 个节点的一侧可以继续工作

结论：偶数节点集群既多花资源，又没有提升容错能力，
      还在特定网络分区场景下表现更差。奇数节点是最优选择。
```

### 6.4 Stream Replicas 与集群关系

```
集群节点数=3，Stream Replicas 的选择：

Replicas=1:
  Stream 只在 1 个节点上存储数据
  无 Raft 开销，写入最快
  节点故障 → Stream 该节点数据不可访问

Replicas=2:
  有 Raft，但 Quorum=2
  任一节点故障 → 不可写（不推荐）

Replicas=3:
  Quorum=2，可容忍 1 个节点故障
  写入需要等待 2/3 节点确认
  （推荐生产配置）

Replicas > 集群节点数：
  错误！NATS 创建 Stream 时会返回错误
```

---

## 7. Stream Replicas 配置

### 7.1 Replicas=1（单副本，开发环境）

```go
// 单副本，适合开发/测试，无 Raft 开销
stream, err := js.CreateOrUpdateStream(ctx, jetstream.StreamConfig{
    Name:     "DEV_EVENTS",
    Subjects: []string{"dev.events.>"},
    Replicas: 1,
    Storage:  jetstream.FileStorage,
})
```

**特性：**
- 写入无需 Raft 确认，延迟最低
- 节点重启后数据完整（文件存储）
- 节点故障期间 Stream 完全不可用

### 7.2 Replicas=3（生产推荐）

```go
// 3副本，生产推荐，可容忍 1 个节点故障
stream, err := js.CreateOrUpdateStream(ctx, jetstream.StreamConfig{
    Name:        "PROD_ORDERS",
    Subjects:    []string{"orders.>"},
    Replicas:    3,
    Storage:     jetstream.FileStorage,
    MaxAge:      7 * 24 * time.Hour,
    MaxBytes:    10 * 1024 * 1024 * 1024, // 10GB
    Retention:   jetstream.LimitsPolicy,
    Discard:     jetstream.DiscardOld,
})
```

### 7.3 R3 Stream 在 3 节点集群的数据分布

```
3节点集群：Node-1, Node-2, Node-3

Stream "ORDERS"（Replicas=3）数据分布：

  Node-1 (Follower)          Node-2 (Leader)           Node-3 (Follower)
  ┌─────────────────┐        ┌─────────────────┐        ┌─────────────────┐
  │ /data/nats/     │        │ /data/nats/      │        │ /data/nats/     │
  │  jetstream/     │        │  jetstream/      │        │  jetstream/     │
  │   ORDERS/       │◀─ 复制 ─│   ORDERS/        │─ 复制 ─▶│   ORDERS/       │
  │   msg.db        │        │   msg.db (主)    │        │   msg.db        │
  │   seq.db        │        │   seq.db         │        │   seq.db        │
  └─────────────────┘        └─────────────────┘        └─────────────────┘
  写入：转发给 Leader          写入：直接处理              写入：转发给 Leader
  读取：可以本地读取            读取：本地读取              读取：可以本地读取

  注：Consumer Fetch 默认路由到 Stream Leader，以保证一致性读
```

### 7.4 多 Stream 的 Leader 分布

NATS 会尽量将不同 Stream 的 Leader 均匀分布在不同节点上：

```
Node-1:  Meta Leader
         Stream "ORDERS" Leader

Node-2:  Stream "EVENTS" Leader
         Stream "NOTIFICATIONS" Leader

Node-3:  Stream "AUDIT" Leader

（实际分布由 NATS 自动负载均衡决定，可通过 LeaderRebalance 触发）
```

---

## 8. 节点故障场景

### 8.1 场景一：Follower 故障（不影响读写）

```
正常状态：Node-1(Leader), Node-2(Follower), Node-3(Follower)

Node-3 突然宕机：
  ┌─────────────────────────────────────────────────┐
  │  Node-1 (Leader)                                │
  │    - 向 Node-2 发送 AppendEntries：成功           │
  │    - 向 Node-3 发送 AppendEntries：失败（超时）    │
  │    - 已有 2/3 节点 ACK（含自己） → Quorum 达成     │
  │    - 继续正常处理写入请求                          │
  └─────────────────────────────────────────────────┘

影响：
  - 写入：不受影响（2/3 Quorum 仍可满足）
  - 读取：不受影响
  - 消费：不受影响

Node-3 恢复后：
  - 从 Node-1(Leader) 拉取缺失的日志条目
  - 追上进度后重新参与 Raft
```

### 8.2 场景二：Leader 故障（短暂不可用）

```
初始状态：Node-1(Leader), Node-2(Follower), Node-3(Follower)

Node-1 突然宕机：
                          T=0ms     Node-1 宕机
                          T=100ms   Node-2 收不到心跳
                          T=200ms   Node-3 收不到心跳
                          T=选举超时  Node-2 或 Node-3 发起选举（约 150-300ms）
                          T=选举完成  新 Leader 产生（通常 < 500ms）

不可用窗口：从 Node-1 宕机 到 新 Leader 完成选举
            通常 200ms ~ 1s（取决于选举超时配置）

在此期间：
  - 发布：返回 ErrNoResponders 或超时
  - 消费：返回超时（Consumer Fetch）

客户端处理：
  - 使用重试机制，等待新 Leader 上任
  - nats.go 默认会自动重连
```

### 8.3 场景三：多数节点故障（集群不可写）

```
初始状态：Node-1(Leader), Node-2(Follower), Node-3(Follower)

Node-2 和 Node-3 同时宕机（2个节点故障）：

  Node-1 发现只有自己（1/3 节点），无法达到 Quorum
  Node-1 自动 Step Down（放弃 Leader 身份）

  此时：
  - 集群没有 Leader
  - 写入：所有 Publish 失败（ErrNoResponders）
  - 读取：Consumer Fetch 可能失败（取决于实现）
  - JetStream 不可用

恢复条件：至少 2 个节点重新上线
```

### 8.4 场景四：网络分区

```
3节点集群，发生脑裂：

分区 A：Node-1, Node-2（能互相通信）
分区 B：Node-3（孤立节点）

行为：
  分区 A：
    - 有 2/3 节点，Quorum 可以达成
    - 如果原 Leader 在 A：继续正常工作
    - 如果原 Leader 在 B：Node-1/Node-2 重新选举

  分区 B（Node-3 孤立）：
    - 只有 1/3 节点，无法达到 Quorum
    - Node-3 无法选出 Leader（无法获得多数票）
    - Node-3 的旧 Leader 状态自动 Step Down
    - 拒绝处理写入请求

结果：系统 CAP 中选择了 CP（一致性 + 分区容忍），放弃了 AP
      只有多数分区可以继续服务，少数分区停止服务
```

---

## 9. 节点恢复流程

### 9.1 全新节点加入集群（Snapshot + Log Replay）

```
步骤 1：新节点启动，发现已有 Leader（Node-2）

步骤 2：新节点向 Leader 请求加入 Raft Group

步骤 3：Leader 检查新节点的日志落后程度
  - 如果落后太多（超过 Leader 保留的 Log 范围）：
      → 触发 Snapshot Transfer（快照传输）

步骤 4：Snapshot Transfer（快照传输）
  Leader:                          New Node:
  ├── 生成当前状态快照              │
  └── 分批发送快照数据 ────────────▶├── 接收并写入本地存储
                                   └── 应用快照（恢复 Stream 当前状态）

步骤 5：Log Replay（日志回放）
  快照之后的增量日志，通过 AppendEntries RPC 逐条同步

步骤 6：追上 Leader 的最新 CommitIndex 后，节点成为正式 Follower

步骤 7：Meta Group 将新节点加入 Raft 投票组（开始参与 Quorum 决策）
```

### 9.2 已有节点重新加入（仅 Log Replay）

```
节点宕机后重启，本地有完整的数据文件：

步骤 1：节点启动，读取本地 WAL，恢复到上次 CommitIndex

步骤 2：向 Leader 发送最新的 LogIndex

步骤 3：Leader 计算缺失的 Log 条目范围

步骤 4：Log Replay（仅增量）
  如果宕机时间短（<1min）：
    差异很小，几十条 Log，追上极快（< 1s）
  如果宕机时间长：
    可能需要 Snapshot Transfer（同全新节点流程）

步骤 5：追上后重新参与 Quorum
```

### 9.3 节点替换操作步骤

在 Kubernetes 等容器环境中替换故障节点：

```bash
# 1. 确认集群状态（查看 meta_cluster 和 stream leaders）
nats server report jetstream

# 2. 如果要永久移除旧节点：
nats server raft peer-remove <node-name>

# 3. 启动新节点，配置相同的 cluster routes

# 4. 新节点自动加入集群，开始 Snapshot/Log 同步

# 5. 验证新节点已加入 Raft
curl -s http://new-node:8222/jsz | jq '.meta_cluster.peers'
```

---

## 10. JetStream 集群监控

### 10.1 /jsz 端点详解

```bash
curl -s http://nats-server:8222/jsz | jq .
```

响应结构（关键字段）：

```json
{
  "server_id": "NBQXXX...",
  "now": "2024-01-15T10:30:00Z",
  "config": {
    "max_memory": 4294967296,
    "max_storage": 107374182400,
    "store_dir": "/data/nats/jetstream",
    "sync_interval": 2000000000
  },
  "memory": 1048576,
  "storage": 536870912,
  "reserved_memory": 0,
  "reserved_storage": 0,
  "accounts": 1,
  "ha_assets": 6,
  "api": {
    "total": 1234,
    "errors": 0
  },
  "meta_cluster": {
    "name": "my-cluster",
    "leader": "Node-1",       ← 当前 meta_leader
    "peer": "NBQXXX...",
    "replicas": [
      {
        "name": "Node-1",
        "current": true,       ← 是否是当前节点
        "active": "100ms",     ← 最后一次 Raft 活跃时间
        "lag": 0               ← 日志落后 Leader 的条目数
      },
      {
        "name": "Node-2",
        "current": false,
        "active": "120ms",
        "lag": 0
      },
      {
        "name": "Node-3",
        "current": false,
        "active": "115ms",
        "lag": 0
      }
    ]
  },
  "streams": 5,               ← Stream 总数
  "consumers": 12,            ← Consumer 总数
  "messages": 1000000,        ← 消息总数
  "bytes": 536870912          ← 存储字节数
}
```

### 10.2 查询 Stream 详情（含 Raft 状态）

```bash
# 获取所有 Stream 的详细信息（含副本状态）
curl -s "http://nats-server:8222/jsz?streams=1&account=\$G" | jq '.account_details[0].stream_detail'

# 关键字段：
# cluster.leader: 该 Stream 的当前 Leader 节点
# cluster.replicas[].lag: 各副本的日志落后量（应接近 0）
# cluster.replicas[].active: 副本最后活跃时间
# state.messages: 当前消息数
# state.bytes: 当前存储大小
```

响应示例：

```json
{
  "config": {
    "name": "ORDERS",
    "subjects": ["orders.>"],
    "replicas": 3
  },
  "state": {
    "messages": 50000,
    "bytes": 10485760,
    "first_seq": 1,
    "last_seq": 50000
  },
  "cluster": {
    "name": "my-cluster",
    "leader": "Node-2",
    "replicas": [
      { "name": "Node-1", "current": true,  "active": "10ms", "lag": 0 },
      { "name": "Node-3", "current": true,  "active": "12ms", "lag": 0 }
    ]
  }
}
```

### 10.3 关键监控指标

| 指标 | 含义 | 告警阈值 |
|------|------|---------|
| `meta_cluster.replicas[].lag` | Meta Raft 日志落后量 | > 1000 告警 |
| `cluster.replicas[].lag` | Stream Raft 落后量 | > 1000 告警 |
| `cluster.replicas[].active` | 副本最后活跃时间 | > 5s 告警（可能节点故障） |
| `api.errors` | JetStream API 错误数 | 突增时告警 |
| `ha_assets` | HA 资源数（Stream+Consumer）| 监控趋势 |

### 10.4 常用监控命令

```bash
# 查看集群整体健康（元数据视角）
curl -s http://nats-1:8222/jsz | jq '{
  meta_leader: .meta_cluster.leader,
  streams: .streams,
  consumers: .consumers,
  ha_assets: .ha_assets,
  api_errors: .api.errors
}'

# 查看每个 Stream 的 Leader 分布
curl -s "http://nats-1:8222/jsz?streams=1" | \
  jq '.account_details[0].stream_detail[] | {name: .config.name, leader: .cluster.leader}'

# 检查是否有副本落后（lag > 0 表示复制延迟）
curl -s "http://nats-1:8222/jsz?streams=1" | \
  jq '.account_details[0].stream_detail[] |
      {stream: .config.name, replicas: [.cluster.replicas[]? | select(.lag > 0)]}'

# 查看 Consumer 的消费进度
curl -s "http://nats-1:8222/jsz?consumers=1" | \
  jq '.account_details[0].stream_detail[] |
      {stream: .config.name,
       consumers: [.consumer_detail[]? |
         {name: .name, pending: .num_pending, ack_pending: .num_ack_pending}]}'
```

---

## 11. meta_cluster.leader 与 Stream Leader 的区别

这是一个常见的混淆点：

```
meta_cluster.leader（元数据 Leader）
───────────────────────────────────
职责：
  - 处理所有 JetStream 管理 API（创建/删除/更新 Stream、Consumer 等）
  - 维护全局 JetStream 配置的一致性
  - NATS CLI 和 SDK 的 CreateStream/DeleteConsumer 等调用必须经过此节点

如何查看：
  GET /jsz → .meta_cluster.leader

特性：
  - 整个集群只有一个 meta_leader
  - 通常不是性能瓶颈（管理 API 调用频率低）

Stream Leader（数据 Leader）
────────────────────────────
职责：
  - 接收该 Stream 的所有 Publish 消息
  - 协调 Raft 日志复制
  - 处理 Consumer Fetch 请求

如何查看：
  GET /jsz?streams=1 → .stream_detail[].cluster.leader

特性：
  - 每个 Stream 有独立的 Leader（可能在不同节点）
  - 是写入性能的关键节点
  - Leader 宕机后选出新 Leader，中断 <1s

实际影响：
  - meta_leader 宕机：
      无法创建/删除 Stream 和 Consumer
      已有 Stream 的 Publish/Subscribe 不受影响 ✓
      约 200ms-1s 内完成重新选举

  - Stream Leader 宕机：
      该 Stream 的 Publish 短暂失败（重新选举期间）
      其他 Stream 不受影响 ✓
```

---

## 12. 滚动升级期间 Raft 的行为

### 12.1 推荐滚动升级步骤（3节点集群）

```bash
# 步骤 1：升级非 Leader 节点（Node-3，Follower）
kubectl rollout restart deployment/nats-3

# 等待 Node-3 重新加入集群并追上 Raft 日志
# 检查：Node-3 的 lag 应回到 0
watch 'curl -s http://nats-3:8222/jsz | jq ".meta_cluster.replicas[] | select(.name == \"Node-3\") | .lag"'

# 步骤 2：升级另一个 Follower（Node-2）
kubectl rollout restart deployment/nats-2
# 等待恢复...

# 步骤 3：升级 Leader（Node-1）
# 升级前主动触发 Leader 迁移，减少中断时间
nats server raft step-down   # 让 Node-1 主动放弃 Leader
# Node-2 或 Node-3 会成为新 Leader，等待选举完成（< 500ms）
kubectl rollout restart deployment/nats-1
```

### 12.2 滚动升级期间的 Raft 状态变化

```
初始状态：Node-1(Leader), Node-2(Follower), Node-3(Follower)
         当前版本：v2.10.0 → 目标：v2.11.0

阶段 1：升级 Node-3
  Node-3 重启（~5s 停机）
  集群：Node-1(Leader), Node-2(Follower)   ← 仍有 2/3 Quorum，正常工作
  Node-3 重启完成：追赶日志，重新加入

阶段 2：升级 Node-2
  Node-2 重启（~5s 停机）
  集群：Node-1(Leader), Node-3(Follower)   ← 仍有 2/3 Quorum
  Node-2 重启完成：追赶日志，重新加入

阶段 3：升级 Node-1（Leader）
  主动 Step Down → Node-2 选举为新 Leader（~200ms 中断）
  Node-1 重启
  Node-1 重启完成：作为 Follower 追赶日志

整个过程中断时间：约 200ms（仅 Leader 切换时）
```

### 12.3 混合版本兼容性

NATS 保证相邻主版本兼容，滚动升级期间集群可以运行混合版本（如 v2.10 和 v2.11 节点共存），无需停机。

---

## 13. 代码示例

### 13.1 创建 R3 Stream

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/nats-io/nats.go"
    "github.com/nats-io/nats.go/jetstream"
)

func createR3Stream() {
    // 连接到集群（任意节点均可，内部自动路由）
    nc, err := nats.Connect(
        "nats://nats-1:4222,nats://nats-2:4222,nats://nats-3:4222",
        nats.Name("stream-manager"),
        nats.MaxReconnects(-1),
    )
    if err != nil {
        log.Fatal("连接集群失败:", err)
    }
    defer nc.Drain()

    js, err := jetstream.New(nc)
    if err != nil {
        log.Fatal("创建 JetStream context 失败:", err)
    }

    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    // 创建 R3 Stream（3副本，生产推荐）
    stream, err := js.CreateOrUpdateStream(ctx, jetstream.StreamConfig{
        Name:        "ORDERS",
        Description: "订单事件流（生产环境，3副本）",
        Subjects:    []string{"orders.>"},
        Replicas:    3,                      // 核心：3副本
        Storage:     jetstream.FileStorage,  // 磁盘持久化
        Retention:   jetstream.LimitsPolicy, // 按限制保留
        Discard:     jetstream.DiscardOld,   // 超限时丢弃旧消息

        // 数据保留策略
        MaxAge:      7 * 24 * time.Hour, // 最长保留 7 天
        MaxBytes:    10 << 30,           // 最多 10GB
        MaxMsgs:     10_000_000,         // 最多 1000 万条

        // 性能调优
        MaxMsgSize:  1 << 20,            // 单条消息最大 1MB
        NoAck:       false,              // 需要 ACK（保证持久化）

        // 去重窗口（防止重复发布）
        Duplicates:  5 * time.Minute,
    })
    if err != nil {
        log.Fatal("创建 Stream 失败:", err)
    }

    info := stream.CachedInfo()
    fmt.Printf("Stream 创建成功:\n")
    fmt.Printf("  名称:      %s\n", info.Config.Name)
    fmt.Printf("  副本数:    %d\n", info.Config.Replicas)
    fmt.Printf("  集群 Leader: %s\n", info.Cluster.Leader)
    fmt.Printf("  副本列表:\n")
    for _, r := range info.Cluster.Replicas {
        fmt.Printf("    - %s (lag=%d)\n", r.Name, r.Lag)
    }
}
```

### 13.2 检查 Stream 状态（含 HA 健康检查）

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/nats-io/nats.go"
    "github.com/nats-io/nats.go/jetstream"
)

type StreamHealthReport struct {
    Name        string
    Leader      string
    IsHealthy   bool
    Issues      []string
    Replicas    []ReplicaStatus
}

type ReplicaStatus struct {
    Name    string
    Current bool
    Lag     uint64
    Active  time.Duration
}

func checkStreamHealth(js jetstream.JetStream, streamName string) (*StreamHealthReport, error) {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    stream, err := js.Stream(ctx, streamName)
    if err != nil {
        return nil, fmt.Errorf("获取 Stream 失败: %w", err)
    }

    // 获取最新状态（绕过缓存，直接请求服务端）
    info, err := stream.Info(ctx)
    if err != nil {
        return nil, fmt.Errorf("获取 Stream 状态失败: %w", err)
    }

    report := &StreamHealthReport{
        Name:      info.Config.Name,
        IsHealthy: true,
    }

    // 检查 Cluster 信息（仅 Replicas > 1 的 Stream 有此字段）
    if info.Cluster != nil {
        report.Leader = info.Cluster.Leader

        if info.Cluster.Leader == "" {
            report.IsHealthy = false
            report.Issues = append(report.Issues, "Stream 当前没有 Leader（选举中或节点故障）")
        }

        for _, r := range info.Cluster.Replicas {
            rs := ReplicaStatus{
                Name:    r.Name,
                Current: r.Current,
                Lag:     r.Lag,
                Active:  r.Active,
            }
            report.Replicas = append(report.Replicas, rs)

            // 检查副本健康
            if !r.Current {
                report.Issues = append(report.Issues,
                    fmt.Sprintf("副本 %s 不在线（not current）", r.Name))
                report.IsHealthy = false
            }
            if r.Lag > 1000 {
                report.Issues = append(report.Issues,
                    fmt.Sprintf("副本 %s 日志落后 %d 条（可能复制延迟）", r.Name, r.Lag))
            }
            if r.Active > 5*time.Second {
                report.Issues = append(report.Issues,
                    fmt.Sprintf("副本 %s 最后活跃时间 %v 前（可能节点故障）", r.Name, r.Active))
                report.IsHealthy = false
            }
        }
    } else {
        // Replicas=1，无 Raft，无集群信息
        report.Leader = "本节点（单副本）"
    }

    return report, nil
}

func monitorStreams() {
    nc, _ := nats.Connect("nats://nats-1:4222,nats://nats-2:4222,nats://nats-3:4222",
        nats.MaxReconnects(-1))
    defer nc.Drain()

    js, _ := jetstream.New(nc)

    // 定期检查
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()

    streams := []string{"ORDERS", "EVENTS", "NOTIFICATIONS"}

    for range ticker.C {
        for _, name := range streams {
            report, err := checkStreamHealth(js, name)
            if err != nil {
                log.Printf("[ERROR] 检查 Stream %s 失败: %v", name, err)
                continue
            }

            if report.IsHealthy {
                fmt.Printf("[OK]    Stream %-20s Leader=%-10s 副本均健康\n",
                    report.Name, report.Leader)
            } else {
                fmt.Printf("[WARN]  Stream %-20s Leader=%-10s 存在问题:\n",
                    report.Name, report.Leader)
                for _, issue := range report.Issues {
                    fmt.Printf("        - %s\n", issue)
                }
            }

            // 打印副本详情
            for _, r := range report.Replicas {
                status := "✓"
                if !r.Current {
                    status = "✗"
                }
                fmt.Printf("        副本 %-10s %s lag=%-5d active=%v\n",
                    r.Name, status, r.Lag, r.Active.Round(time.Millisecond))
            }
        }
    }
}

func main() {
    createR3Stream()
    // monitorStreams() // 取消注释以运行监控循环
}
```

---

## 总结

| 关键概念 | 要点 |
|---------|------|
| Meta Group | 管理 Stream/Consumer 配置，全集群一个 meta_leader |
| Stream Raft Group | 每个 R>1 Stream 独立的 Raft，管理数据副本 |
| Quorum | N/2+1，3节点集群需要 2 节点确认 |
| Leader Election | 选举超时随机，通常 <1s 完成重新选举 |
| 故障容忍 | 3节点容忍 1 节点故障，5节点容忍 2 节点故障 |
| 奇数节点 | 比偶数节点更高效，推荐 3/5/7 节点 |
| Lag | 副本日志落后量，应接近 0，>1000 时告警 |
| 滚动升级 | 先升级 Follower，最后升级 Leader，中断 <1s |

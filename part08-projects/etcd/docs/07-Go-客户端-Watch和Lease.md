# Go 客户端 - Watch 和 Lease

本章将学习 etcd 的监听机制和租约机制，这两个特性在实时通知和临时数据场景中非常重要。

## Watch 监听

### 基本概念

Watch 用于监听键或前缀的变更事件，实现实时通知。

```
客户端              etcd
  │                  │
  │ Watch /foo       │
  │─────────────────>│
  │                  │
  │ WatchResponse    │
  │<─────────────────│  Put /foo = "bar"
  │                  │
  │ WatchResponse    │
  │<─────────────────│  Put /foo = "bar2"
  │                  │
  │ WatchResponse    │
  │<─────────────────│  Delete /foo
  │                  │
```

### 监听单个键

```go
package main

import (
    "context"
    "fmt"
    "time"

    clientv3 "go.etcd.io/etcd/client/v3"
)

func main() {
    cli, _ := clientv3.New(clientv3.Config{
        Endpoints:   []string{"localhost:2379"},
        DialTimeout: 5 * time.Second,
    })
    defer cli.Close()

    // 监听单个键
    watchCh := cli.Watch(context.Background(), "/foo")

    for watchResp := range watchCh {
        for _, event := range watchResp.Events {
            switch event.Type {
            case clientv3.EventTypePut:
                fmt.Printf("Put: %s = %s\n", event.Kv.Key, event.Kv.Value)
            case clientv3.EventTypeDelete:
                fmt.Printf("Delete: %s\n", event.Kv.Key)
            }
        }
    }
}
```

### 监听前缀

```go
// 监听所有配置变更
watchCh := cli.Watch(context.Background(), "/config/", clientv3.WithPrefix())

for watchResp := range watchCh {
    for _, event := range watchResp.Events {
        fmt.Printf("Event: %s, Key: %s, Value: %s\n",
            event.Type, event.Kv.Key, event.Kv.Value)
    }
}
```

### 监听范围

```go
// 监听 [start, end) 范围
watchCh := cli.Watch(context.Background(), "/a", clientv3.WithRange("/b"))
```

### 监听历史变更

```go
// 从指定 revision 开始监听（包括历史事件）
// 获取当前 revision
getResp, _ := cli.Get(context.Background(), "/foo")
startRev := getResp.Header.Revision

// 监听从 startRev 开始的所有变更
watchCh := cli.Watch(context.Background(), "/foo", clientv3.WithRev(startRev))

for watchResp := range watchCh {
    for _, event := range watchResp.Events {
        fmt.Printf("Rev: %d, Type: %s, Key: %s\n",
            event.Kv.ModRevision, event.Type, event.Kv.Key)
    }
}
```

### Watch 事件类型

```go
type EventType int

const (
    EventTypePut    EventType = 0  // Put 或 Update
    EventTypeDelete EventType = 1  // Delete
)

// 事件详情
type Event struct {
    Type EventType
    Kv   *KeyValue       // 当前值（Put 时有）
    PrevKv *KeyValue     // 先前值（WithPrevKV() 时有）
}
```

### 获取先前值

```go
// 监听时获取变更前的值
watchCh := cli.Watch(context.Background(), "/foo", clientv3.WithPrevKV())

for watchResp := range watchCh {
    for _, event := range watchResp.Events {
        if event.PrevKv != nil {
            fmt.Printf("Old: %s, New: %s\n", event.PrevKv.Value, event.Kv.Value)
        }
    }
}
```

### Watch 响应处理

```go
func handleWatchResponse(watchResp clientv3.WatchResponse) {
    // 检查错误
    if watchResp.Err() != nil {
        fmt.Printf("Watch error: %v\n", watchResp.Err())
        return
    }

    // 检查取消
    if watchResp.Canceled {
        fmt.Println("Watch canceled")
        return
    }

    // 处理事件
    for _, event := range watchResp.Events {
        switch event.Type {
        case clientv3.EventTypePut:
            fmt.Printf("Put: %s = %s (Rev: %d)\n",
                event.Kv.Key, event.Kv.Value, event.Kv.ModRevision)
        case clientv3.EventTypeDelete:
            fmt.Printf("Delete: %s (Rev: %d)\n",
                event.Kv.Key, event.Kv.ModRevision)
        }
    }
}
```

### Watch 进度通知

```go
// 创建 Watch 时请求进度通知
watchCh := cli.Watch(context.Background(), "/foo", 
    clientv3.WithProgressNotify())

for watchResp := range watchCh {
    if watchResp.IsProgressNotify() {
        fmt.Printf("Progress notify, Revision: %d\n", watchResp.Header.Revision)
        continue
    }

    for _, event := range watchResp.Events {
        fmt.Printf("Event: %s\n", event.Type)
    }
}
```

### 取消监听

```go
ctx, cancel := context.WithCancel(context.Background())

watchCh := cli.Watch(ctx, "/foo")

// 在需要时取消
cancel()

// Watch channel 会关闭
for watchResp := range watchCh {
    if watchResp.Canceled {
        break
    }
}
```

### Watch 重连

```go
func watchWithRetry(cli *clientv3.Client, key string) {
    var watchRev int64 = 0

    for {
        ctx, cancel := context.WithCancel(context.Background())

        opts := []clientv3.OpOption{}
        if watchRev > 0 {
            opts = append(opts, clientv3.WithRev(watchRev))
        }

        watchCh := cli.Watch(ctx, key, opts...)

        for watchResp := range watchCh {
            if watchResp.Err() != nil {
                fmt.Printf("Watch error: %v, reconnecting...\n", watchResp.Err())
                cancel()
                time.Sleep(1 * time.Second)
                break
            }

            for _, event := range watchResp.Events {
                watchRev = event.Kv.ModRevision + 1
                fmt.Printf("Event: %s\n", event.Type)
            }
        }

        cancel()
    }
}
```

## Lease 租约

### 基本概念

租约（Lease）为键设置生存时间（TTL），租约到期后关联的键自动删除。

```
┌──────────────────────────────────────────────┐
│                Lease (TTL=60s)               │
│                                              │
│  ┌─────────┐ ┌─────────┐ ┌─────────┐        │
│  │Key1     │ │Key2     │ │Key3     │        │
│  └─────────┘ └─────────┘ └─────────┘        │
│                                              │
│  租约到期 → 所有键自动删除                    │
└──────────────────────────────────────────────┘
```

### 创建租约

```go
func main() {
    cli, _ := clientv3.New(clientv3.Config{
        Endpoints:   []string{"localhost:2379"},
        DialTimeout: 5 * time.Second,
    })
    defer cli.Close()

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    // 创建 60 秒租约
    leaseResp, err := cli.Lease.Grant(ctx, 60)
    if err != nil {
        panic(err)
    }

    fmt.Printf("Lease ID: %d, TTL: %d\n", leaseResp.ID, leaseResp.TTL)

    // 将键绑定到租约
    _, err = cli.Put(ctx, "/services/app/instance1", "192.168.1.10:8080",
        clientv3.WithLease(leaseResp.ID))
    if err != nil {
        panic(err)
    }
}
```

### 查询租约

```go
// 查询租约剩余时间
leaseID := leaseResp.ID

ttlResp, err := cli.Lease.TimeToLive(ctx, leaseID)
if err != nil {
    panic(err)
}

fmt.Printf("Lease ID: %d, TTL: %d, GrantedTTL: %d\n",
    ttlResp.ID, ttlResp.TTL, ttlResp.GrantedTTL)

// 查询租约关联的键
if ttlResp.Keys != nil {
    for _, key := range ttlResp.Keys {
        fmt.Printf("Key: %s\n", key)
    }
}
```

### KeepAlive 续租

```go
// 单次续租
keepAliveResp, err := cli.Lease.KeepAliveOnce(ctx, leaseID)
if err != nil {
    panic(err)
}

fmt.Printf("KeepAlive succeeded, TTL: %d\n", keepAliveResp.TTL)
```

### 自动续租

```go
// 持续自动续租
func keepAlive(cli *clientv3.Client, leaseID clientv3.LeaseID) {
    keepAliveCh, err := cli.Lease.KeepAlive(context.Background(), leaseID)
    if err != nil {
        panic(err)
    }

    for resp := range keepAliveCh {
        if resp == nil {
            fmt.Println("KeepAlive channel closed")
            return
        }
        fmt.Printf("KeepAlive: LeaseID=%d, TTL=%d\n", resp.ID, resp.TTL)
    }
}

// 在后台运行
go keepAlive(cli, leaseResp.ID)
```

### 撤销租约

```go
// 手动撤销租约（立即删除关联的键）
_, err := cli.Lease.Revoke(ctx, leaseID)
if err != nil {
    panic(err)
}

fmt.Println("Lease revoked")
```

### 租约列表

```go
// 查询所有租约
leasesResp, err := cli.Lease.Leases(ctx)
if err != nil {
    panic(err)
}

for _, lease := range leasesResp.Leases {
    fmt.Printf("Lease ID: %d\n", lease.ID)
}
```

## 实用示例

### 动态配置监听

```go
type ConfigWatcher struct {
    cli    *clientv3.Client
    prefix string
    config map[string]string
}

func NewConfigWatcher(cli *clientv3.Client, prefix string) *ConfigWatcher {
    cw := &ConfigWatcher{
        cli:    cli,
        prefix: prefix,
        config: make(map[string]string),
    }

    // 初始化：读取当前配置
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    resp, _ := cli.Get(ctx, prefix, clientv3.WithPrefix())
    cancel()

    for _, kv := range resp.Kvs {
        key := string(kv.Key)[len(prefix):]
        cw.config[key] = string(kv.Value)
    }

    // 启动监听
    go cw.watch()

    return cw
}

func (cw *ConfigWatcher) watch() {
    watchCh := cw.cli.Watch(context.Background(), cw.prefix, clientv3.WithPrefix())

    for watchResp := range watchCh {
        for _, event := range watchResp.Events {
            key := string(event.Kv.Key)[len(cw.prefix):]

            switch event.Type {
            case clientv3.EventTypePut:
                cw.config[key] = string(event.Kv.Value)
                fmt.Printf("Config updated: %s = %s\n", key, cw.config[key])
            case clientv3.EventTypeDelete:
                delete(cw.config, key)
                fmt.Printf("Config deleted: %s\n", key)
            }
        }
    }
}

func (cw *ConfigWatcher) Get(key string) string {
    return cw.config[key]
}
```

### 服务注册（带租约）

```go
type ServiceRegistry struct {
    cli       *clientv3.Client
    leaseID   clientv3.LeaseID
    keepAlive context.CancelFunc
}

func RegisterService(cli *clientv3.Client, serviceName, instanceID, addr string, ttl int64) (*ServiceRegistry, error) {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

    // 创建租约
    leaseResp, err := cli.Lease.Grant(ctx, ttl)
    if err != nil {
        cancel()
        return nil, err
    }

    // 注册服务
    key := fmt.Sprintf("/services/%s/%s", serviceName, instanceID)
    _, err = cli.Put(ctx, key, addr, clientv3.WithLease(leaseResp.ID))
    if err != nil {
        cancel()
        return nil, err
    }
    cancel()

    // 启动 KeepAlive
    keepAliveCtx, keepAliveCancel := context.WithCancel(context.Background())
    keepAliveCh, err := cli.Lease.KeepAlive(keepAliveCtx, leaseResp.ID)
    if err != nil {
        keepAliveCancel()
        return nil, err
    }

    // 处理 KeepAlive 响应
    go func() {
        for {
            select {
            case resp := <-keepAliveCh:
                if resp == nil {
                    fmt.Println("KeepAlive stopped")
                    return
                }
            case <-keepAliveCtx.Done():
                return
            }
        }
    }()

    return &ServiceRegistry{
        cli:       cli,
        leaseID:   leaseResp.ID,
        keepAlive: keepAliveCancel,
    }, nil
}

func (sr *ServiceRegistry) Unregister() {
    // 取消 KeepAlive
    sr.keepAlive()

    // 撤销租约
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    sr.cli.Lease.Revoke(ctx, sr.leaseID)
    cancel()
}
```

### 分布式锁（基于 Lease）

```go
type DistributedLock struct {
    cli     *clientv3.Client
    key     string
    leaseID clientv3.LeaseID
}

func AcquireLock(cli *clientv3.Client, key string, ttl int64) (*DistributedLock, error) {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    // 创建租约
    leaseResp, err := cli.Lease.Grant(ctx, ttl)
    if err != nil {
        return nil, err
    }

    // 尝试获取锁（事务）
    txnResp, err := cli.Txn(ctx).
        If(clientv3.Compare(clientv3.CreateRevision(key), "=", 0)).
        Then(clientv3.OpPut(key, "locked", clientv3.WithLease(leaseResp.ID))).
        Else(clientv3.OpGet(key)).
        Commit()

    if err != nil {
        cli.Lease.Revoke(ctx, leaseResp.ID)
        return nil, err
    }

    if !txnResp.Succeeded {
        cli.Lease.Revoke(ctx, leaseResp.ID)
        return nil, fmt.Errorf("lock already held")
    }

    // 启动 KeepAlive
    go func() {
        keepAliveCh, _ := cli.Lease.KeepAlive(context.Background(), leaseResp.ID)
        for resp := range keepAliveCh {
            if resp == nil {
                return
            }
        }
    }()

    return &DistributedLock{
        cli:     cli,
        key:     key,
        leaseID: leaseResp.ID,
    }, nil
}

func (lock *DistributedLock) Release() error {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    // 撤销租约（自动删除键）
    _, err := lock.cli.Lease.Revoke(ctx, lock.leaseID)
    return err
}
```

### 会话管理（租约）

```go
type Session struct {
    cli     *clientv3.Client
    leaseID clientv3.LeaseID
    prefix  string
}

func NewSession(cli *clientv3.Client, ttl int64) (*Session, error) {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    leaseResp, err := cli.Lease.Grant(ctx, ttl)
    if err != nil {
        return nil, err
    }

    keepAliveCh, err := cli.Lease.KeepAlive(context.Background(), leaseResp.ID)
    if err != nil {
        cli.Lease.Revoke(ctx, leaseResp.ID)
        return nil, err
    }

    go func() {
        for resp := range keepAliveCh {
            if resp == nil {
                return
            }
        }
    }()

    return &Session{
        cli:     cli,
        leaseID: leaseResp.ID,
        prefix:  fmt.Sprintf("/session/%d/", leaseResp.ID),
    }, nil
}

func (s *Session) Put(key, value string) error {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    fullKey := s.prefix + key
    _, err := s.cli.Put(ctx, fullKey, value, clientv3.WithLease(s.leaseID))
    return err
}

func (s *Session) Close() error {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    _, err := s.cli.Lease.Revoke(ctx, s.leaseID)
    return err
}
```

## 性能优化

### Watch 批量处理

```go
// Watch 响应是批量返回的，减少处理频率
for watchResp := range watchCh {
    // 批量处理事件
    events := watchResp.Events
    for _, event := range events {
        // ...
    }
}
```

### 租约复用

```go
// 多个键共享一个租约
leaseResp, _ := cli.Lease.Grant(ctx, 60)

cli.Put(ctx, "/session/user1/data1", "value1", clientv3.WithLease(leaseResp.ID))
cli.Put(ctx, "/session/user1/data2", "value2", clientv3.WithLease(leaseResp.ID))
cli.Put(ctx, "/session/user1/data3", "value3", clientv3.WithLease(leaseResp.ID))

// 租约到期时，所有键一起删除
```

### 减少网络往返

```go
// 使用 ProgressNotify 减少不必要的网络请求
watchCh := cli.Watch(context.Background(), "/foo", clientv3.WithProgressNotify())

// KeepAlive 批量发送
// etcd 会批量处理 KeepAlive 请求
```

## 最佳实践

1. **Watch 使用**
   - 使用前缀监听而不是多个单键监听
   - 处理错误和取消情况
   - 合理设置 revision 起始点

2. **KeepAlive 管理**
   - 确保 KeepAlive goroutine 能正确退出
   - 使用 context 控制取消
   - 处理 KeepAlive channel 关闭

3. **租约设计**
   - TTL 设置合理的值（如 10-60 秒）
   - 相关数据共享租约
   - 及时撤销不需要的租约

4. **资源清理**
   - 取消监听时关闭 context
   - 退出时撤销租约
   - 避免 goroutine 泄漏

## 总结

本章学习了：
- Watch 监听：监听变更、获取事件、重连处理
- Lease 租约：创建、查询、续租、撤销
- 实用示例：动态配置、服务注册、分布式锁、会话管理

在下一章中，我们将学习事务和分布式锁的高级用法。
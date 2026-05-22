# Go 客户端 - KV 操作

本章将深入学习 etcd 的键值操作，包括 Put、Get、Delete、范围查询等核心功能。

## Put 操作

### 基本写入

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

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    // 基本写入
    resp, err := cli.Put(ctx, "/config/database/host", "localhost")
    if err != nil {
        panic(err)
    }

    fmt.Printf("Put succeeded, revision: %d\n", resp.Header.Revision)
}
```

### 带租约写入

```go
func putWithLease(cli *clientv3.Client, key, value string, ttl int64) error {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    // 创建租约
    leaseResp, err := cli.Lease.Grant(ctx, ttl)
    if err != nil {
        return err
    }

    // 带租约写入
    _, err = cli.Put(ctx, key, value, clientv3.WithLease(leaseResp.ID))
    return err
}

// 使用
putWithLease(cli, "/services/app/instance1", "192.168.1.10:8080", 60)
```

### 带先前值检查（CAS）

```go
// 只有当键不存在时才写入（Create 操作）
resp, err := cli.Put(ctx, "/lock/my-resource", "locked",
    clientv3.WithPrevKV())

if resp.PrevKv != nil {
    fmt.Println("键已存在，Put 操作成功但返回先前值")
    fmt.Printf("先前值: %s\n", resp.PrevKv.Value)
}
```

## Get 操作

### 精确查询

```go
func get(cli *clientv3.Client, key string) (string, error) {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    resp, err := cli.Get(ctx, key)
    if err != nil {
        return "", err
    }

    if len(resp.Kvs) == 0 {
        return "", fmt.Errorf("key not found: %s", key)
    }

    return string(resp.Kvs[0].Value), nil
}
```

### 查询详细信息

```go
func getWithDetails(cli *clientv3.Client, key string) {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    resp, _ := cli.Get(ctx, key)
    if len(resp.Kvs) > 0 {
        kv := resp.Kvs[0]
        fmt.Printf("Key: %s\n", kv.Key)
        fmt.Printf("Value: %s\n", kv.Value)
        fmt.Printf("CreateRevision: %d\n", kv.CreateRevision)
        fmt.Printf("ModRevision: %d\n", kv.ModRevision)
        fmt.Printf("Version: %d\n", kv.Version)
        fmt.Printf("Lease: %d\n", kv.Lease)
    }
}
```

### 获取 Revision

```go
// 获取当前全局 revision
resp, _ := cli.Get(ctx, "any-key")
currentRevision := resp.Header.Revision
fmt.Printf("Current cluster revision: %d\n", currentRevision)
```

## 范围查询

### 前缀查询

```go
func getByPrefix(cli *clientv3.Client, prefix string) ([]string, error) {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    resp, err := cli.Get(ctx, prefix, clientv3.WithPrefix())
    if err != nil {
        return nil, err
    }

    values := make([]string, 0, len(resp.Kvs))
    for _, kv := range resp.Kvs {
        values = append(values, string(kv.Value))
    }
    return values, nil
}

// 查询所有配置
configs, _ := getByPrefix(cli, "/config/")
for i, config := range configs {
    fmt.Printf("Config %d: %s\n", i+1, config)
}
```

### 范围查询

```go
// 查询 [start, end) 范围内的键
func getByRange(cli *clientv3.Client, start, end string) {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    resp, _ := cli.Get(ctx, start, clientv3.WithRange(end))

    for _, kv := range resp.Kvs {
        fmt.Printf("Key: %s, Value: %s\n", kv.Key, kv.Value)
    }
}

// 查询 /a 到 /b 之间的键（不包括 /b）
getByRange(cli, "/a", "/b")
```

### 限制数量

```go
// 只返回前 10 个键
resp, _ := cli.Get(ctx, "/services/", 
    clientv3.WithPrefix(),
    clientv3.WithLimit(10))

fmt.Printf("Total keys: %d, returned: %d\n", resp.Count, len(resp.Kvs))

// 如果 resp.More == true，表示还有更多数据
if resp.More {
    fmt.Println("还有更多数据")
}
```

### 分页查询

```go
func paginate(cli *clientv3.Client, prefix string, pageSize int64) {
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    var lastKey string
    pageNum := 1

    for {
        opts := []clientv3.OpOption{
            clientv3.WithPrefix(),
            clientv3.WithLimit(pageSize),
            clientv3.WithSort(clientv3.SortByKey, clientv3.SortAscend),
        }

        if lastKey != "" {
            opts = append(opts, clientv3.WithFromKey())
        }

        resp, err := cli.Get(ctx, lastKey, opts...)
        if err != nil {
            panic(err)
        }

        fmt.Printf("Page %d: %d keys\n", pageNum, len(resp.Kvs))
        for _, kv := range resp.Kvs {
            fmt.Printf("  %s\n", kv.Key)
        }

        if !resp.More {
            break
        }

        lastKey = string(resp.Kvs[len(resp.Kvs)-1].Key) + "\x00"
        pageNum++
    }
}
```

### 排序查询

```go
// 按键升序排序（默认）
resp, _ := cli.Get(ctx, "/services/",
    clientv3.WithPrefix(),
    clientv3.WithSort(clientv3.SortByKey, clientv3.SortAscend))

// 按版本降序排序
resp, _ := cli.Get(ctx, "/services/",
    clientv3.WithPrefix(),
    clientv3.WithSort(clientv3.SortByVersion, clientv3.SortDescend))

// 按 ModRevision 排序
resp, _ := cli.Get(ctx, "/services/",
    clientv3.WithPrefix(),
    clientv3.WithSort(clientv3.SortByModRevision, clientv3.SortAscend))
```

### 只获取键数量

```go
// 只统计数量，不返回数据
resp, _ := cli.Get(ctx, "/services/", 
    clientv3.WithPrefix(),
    clientv3.WithCountOnly())

fmt.Printf("Total keys: %d\n", resp.Count)
```

### 获取所有键

```go
// 获取所有键（从空字符串开始）
resp, _ := cli.Get(ctx, "", clientv3.WithPrefix())

fmt.Printf("Total keys in cluster: %d\n", resp.Count)
for _, kv := range resp.Kvs {
    fmt.Printf("Key: %s\n", kv.Key)
}
```

## 历史版本查询

### 查询指定 Revision

```go
// 查询 revision 100 时的值
resp, _ := cli.Get(ctx, "/foo", clientv3.WithRev(100))
if len(resp.Kvs) > 0 {
    fmt.Printf("Value at revision 100: %s\n", resp.Kvs[0].Value)
}
```

### 查询修改历史

```go
// 注意：需要先设置 --auto-compaction-retention 才能查询历史版本
// 默认情况下，历史版本很快会被压缩删除

// 查询键的所有历史版本（从创建到当前）
func getAllVersions(cli *clientv3.Client, key string) {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    // 获取当前 revision
    currentResp, _ := cli.Get(ctx, key)
    currentRev := currentResp.Header.Revision

    // 获取创建 revision
    if len(currentResp.Kvs) > 0 {
        createRev := currentResp.Kvs[0].CreateRevision

        // 查询历史
        for rev := createRev; rev <= currentRev; rev++ {
            resp, _ := cli.Get(ctx, key, clientv3.WithRev(rev))
            if len(resp.Kvs) > 0 {
                fmt.Printf("Revision %d: %s\n", rev, resp.Kvs[0].Value)
            }
        }
    }
}
```

## Delete 操作

### 删除单个键

```go
func delete(cli *clientv3.Client, key string) error {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    resp, err := cli.Delete(ctx, key)
    if err != nil {
        return err
    }

    fmt.Printf("Deleted %d keys\n", resp.Deleted)
    return nil
}
```

### 删除前缀

```go
func deleteByPrefix(cli *clientv3.Client, prefix string) {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    resp, _ := cli.Delete(ctx, prefix, clientv3.WithPrefix())
    fmt.Printf("Deleted %d keys\n", resp.Deleted)
}

// 删除所有服务注册信息
deleteByPrefix(cli, "/services/app/")
```

### 范围删除

```go
// 删除 [start, end) 范围内的键
resp, _ := cli.Delete(ctx, "/a", clientv3.WithRange("/b"))
fmt.Printf("Deleted %d keys\n", resp.Deleted)
```

### 返回被删除的键值

```go
resp, _ := cli.Delete(ctx, "/foo", clientv3.WithPrevKV())

fmt.Printf("Deleted %d keys\n", resp.Deleted)
if len(resp.PrevKvs) > 0 {
    for _, kv := range resp.PrevKvs {
        fmt.Printf("Deleted: %s = %s\n", kv.Key, kv.Value)
    }
}
```

## 批量操作

### 批量写入

使用事务实现批量操作：

```go
func batchPut(cli *clientv3.Client, kvs map[string]string) error {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    // 构建 Put 操作列表
    ops := make([]clientv3.Op, 0, len(kvs))
    for key, value := range kvs {
        ops = append(ops, clientv3.OpPut(key, value))
    }

    // 执行事务（无条件执行）
    _, err := cli.Txn(ctx).Then(ops...).Commit()
    return err
}

// 使用
batchPut(cli, map[string]string{
    "/config/database/host":     "localhost",
    "/config/database/port":     "3306",
    "/config/database/username": "root",
    "/config/database/password": "password",
})
```

### 批量读取

```go
func batchGet(cli *clientv3.Client, keys []string) (map[string]string, error) {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    // 构建 Get 操作列表
    ops := make([]clientv3.Op, 0, len(keys))
    for _, key := range keys {
        ops = append(ops, clientv3.OpGet(key))
    }

    // 执行事务
    txnResp, err := cli.Txn(ctx).Then(ops...).Commit()
    if err != nil {
        return nil, err
    }

    // 解析结果
    results := make(map[string]string)
    for i, resp := range txnResp.Responses {
        getResp := (*clientv3.GetResponse)(resp.GetResponseRange())
        if len(getResp.Kvs) > 0 {
            results[keys[i]] = string(getResp.Kvs[0].Value)
        }
    }
    return results, nil
}
```

## Op 操作

### 使用 Op 对象

```go
// 创建 Op 对象
putOp := clientv3.OpPut("/foo", "bar")
getOp := clientv3.OpGet("/foo")
deleteOp := clientv3.OpDelete("/foo")

// 执行 Op
resp, err := cli.Do(ctx, putOp)
resp, err = cli.Do(ctx, getOp)
resp, err = cli.Do(ctx, deleteOp)
```

### Op 配置

```go
// 带 Lease 的 Put Op
leaseResp, _ := cli.Lease.Grant(ctx, 60)
putWithLease := clientv3.OpPut("/foo", "bar", clientv3.WithLease(leaseResp.ID))

// 前缀查询 Op
getByPrefix := clientv3.OpGet("/services/", clientv3.WithPrefix())

// 带 Revision 的 Get Op
getWithRev := clientv3.OpGet("/foo", clientv3.WithRev(100))
```

## 实用示例

### 配置管理

```go
type ConfigManager struct {
    cli    *clientv3.Client
    prefix string
}

func NewConfigManager(cli *clientv3.Client, prefix string) *ConfigManager {
    return &ConfigManager{cli: cli, prefix: prefix}
}

func (cm *ConfigManager) Set(key, value string) error {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    _, err := cm.cli.Put(ctx, cm.prefix+key, value)
    return err
}

func (cm *ConfigManager) Get(key string) (string, error) {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    resp, err := cm.cli.Get(ctx, cm.prefix+key)
    if err != nil {
        return "", err
    }
    if len(resp.Kvs) == 0 {
        return "", fmt.Errorf("config not found")
    }
    return string(resp.Kvs[0].Value), nil
}

func (cm *ConfigManager) GetAll() (map[string]string, error) {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    resp, err := cm.cli.Get(ctx, cm.prefix, clientv3.WithPrefix())
    if err != nil {
        return nil, err
    }

    configs := make(map[string]string)
    for _, kv := range resp.Kvs {
        key := string(kv.Key)[len(cm.prefix):]
        configs[key] = string(kv.Value)
    }
    return configs, nil
}

func (cm *ConfigManager) Delete(key string) error {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    _, err := cm.cli.Delete(ctx, cm.prefix+key)
    return err
}
```

### 服务注册

```go
func registerService(cli *clientv3.Client, serviceName, instanceID, addr string, ttl int64) error {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    key := fmt.Sprintf("/services/%s/%s", serviceName, instanceID)

    // 创建租约
    leaseResp, err := cli.Lease.Grant(ctx, ttl)
    if err != nil {
        return err
    }

    // 注册服务
    _, err = cli.Put(ctx, key, addr, clientv3.WithLease(leaseResp.ID))
    if err != nil {
        return err
    }

    // 保持租约（在后台运行）
    go func() {
        keepAliveCtx, keepAliveCancel := context.WithCancel(context.Background())
        defer keepAliveCancel()

        keepAliveCh, err := cli.Lease.KeepAlive(keepAliveCtx, leaseResp.ID)
        if err != nil {
            return
        }

        for {
            select {
            case resp := <-keepAliveCh:
                if resp == nil {
                    return
                }
            case <-keepAliveCtx.Done():
                return
            }
        }
    }()

    return nil
}
```

### 服务发现

```go
func discoverServices(cli *clientv3.Client, serviceName string) ([]string, error) {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    prefix := fmt.Sprintf("/services/%s/", serviceName)
    resp, err := cli.Get(ctx, prefix, clientv3.WithPrefix())
    if err != nil {
        return nil, err
    }

    instances := make([]string, 0, len(resp.Kvs))
    for _, kv := range resp.Kvs {
        instances = append(instances, string(kv.Value))
    }
    return instances, nil
}
```

## 性能优化

### 使用串行化读

```go
// 读取频繁但一致性要求不高时，使用串行化读
resp, _ := cli.Get(ctx, "/config/", 
    clientv3.WithPrefix(),
    clientv3.WithSerializable())
```

### 批量操作

```go
// 一次性获取多个键，减少网络往返
txnResp, _ := cli.Txn(ctx).Then(
    clientv3.OpGet("/key1"),
    clientv3.OpGet("/key2"),
    clientv3.OpGet("/key3"),
).Commit()
```

### 避免大值

```go
// 大值建议存储在对象存储，etcd 只存引用
largeData := "very large data..." // 1MB+
// 不推荐：cli.Put(ctx, "/large", largeData)

// 推荐：存储引用
cli.Put(ctx, "/large", "s3://bucket/path/to/data")
```

## 最佳实践

1. **键命名**
   - 使用有意义的前缀组织数据
   - 避免过长的键名

2. **查询优化**
   - 使用前缀查询减少网络往返
   - 只需要数量时用 WithCountOnly
   - 合理使用分页避免大量数据传输

3. **写入优化**
   - 批量写入使用事务
   - 临时数据使用租约
   - 避免频繁写入大值

4. **错误处理**
   - 区分键不存在和异常错误
   - 网络错误时重试

5. **资源管理**
   - 及时释放租约
   - 避免创建过多 Watch

## 总结

本章学习了：
- Put 操作：基本写入、带租约写入、CAS
- Get 操作：精确查询、前缀查询、范围查询、排序
- Delete 操作：删除键、删除前缀、返回被删除值
- 批量操作：使用事务批量读写
- 实用示例：配置管理、服务注册发现

在下一章中，我们将学习 Watch 和 Lease 的高级用法。
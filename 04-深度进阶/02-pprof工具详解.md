# 10.2 pprof工具详解

## pprof基础概念

### 什么是pprof？

pprof是Go官方提供的性能分析工具，用于：
- CPU性能分析
- 内存分配分析
- Goroutine分析
- 阻塞分析
- 互斥锁分析

### 数据采集方式

```go
// 方式1: 通过HTTP endpoint（推荐）
import _ "net/http/pprof"

func main() {
    go func() {
        http.ListenAndServe("localhost:6060", nil)
    }()
    // 应用代码...
}

// 访问方式
// http://localhost:6060/debug/pprof/

// 方式2: 手动采集
import "runtime/pprof"

func main() {
    // CPU分析
    f, _ := os.Create("cpu.prof")
    pprof.StartCPUProfile(f)
    defer pprof.StopCPUProfile()
    
    // 应用代码...
    
    // 内存分析
    f2, _ := os.Create("mem.prof")
    pprof.WriteHeapProfile(f2)
    f2.Close()
}
```

---

## CPU性能分析实战

### 采集CPU数据

```bash
# 方式1: 通过HTTP endpoint采集（30秒）
curl -o cpu.prof http://localhost:6060/debug/pprof/profile?seconds=30

# 方式2: 使用go tool pprof
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30

# 方式3: 在benchmark中采集
go test -cpuprofile=cpu.prof -bench=.
```

### 分析CPU数据

```bash
# 交互式分析
go tool pprof cpu.prof

# 常用命令
(pprof) top10        # Top 10 CPU消耗函数
(pprof) top10 -cum   # 按累计时间排序
(pprof) list main    # 查看函数详情
(pprof) web          # 生成可视化图（需要graphviz）

# 示例输出
      flat  flat%   sum%        cum   cum%
     250ms 25.00% 25.00%      250ms 25.00%  runtime.memmove
     200ms 20.00% 45.00%      450ms 45.00%  main.processData
     150ms 15.00% 60.00%      150ms 15.00%  runtime.mallocgc
```

### 火焰图分析

```bash
# 生成火焰图（需要go 1.11+）
go tool pprof -http=:8080 cpu.prof

# 浏览器访问 http://localhost:8080
# 选择 VIEW -> Flame Graph
```

### CPU优化案例

**问题定位**:
```bash
(pprof) top5
     400ms 40.00%  runtime.memmove
     300ms 30.00%  main.processData
     200ms 20.00%  runtime.mallocgc
```

**优化前代码**:
```go
// ❌ 频繁切片append导致内存分配和拷贝
func processData(data []byte) []byte {
    var result []byte
    for _, chunk := range splitData(data) {
        result = append(result, chunk...)  // 频繁扩容
    }
    return result
}
```

**优化后代码**:
```go
// ✅ 预分配容量
func processData(data []byte) []byte {
    chunks := splitData(data)
    totalSize := 0
    for _, chunk := range chunks {
        totalSize += len(chunk)
    }
    
    result := make([]byte, 0, totalSize)  // 预分配
    for _, chunk := range chunks {
        result = append(result, chunk...)
    }
    return result
}

// 性能提升: 40% → 15% CPU消耗
```

---

## 内存性能分析实战

### 采集内存数据

```bash
# 方式1: HTTP endpoint
curl -o mem.prof http://localhost:6060/debug/pprof/heap

# 方式2: benchmark
go test -memprofile=mem.prof -bench=.

# 方式3: 程序退出时保存
func main() {
    defer func() {
        f, _ := os.Create("mem.prof")
        pprof.WriteHeapProfile(f)
        f.Close()
    }()
    // ...
}
```

### 分析内存数据

```bash
# 交互式分析
go tool pprof -sample_index=alloc_space mem.prof  # 总分配
go tool pprof -sample_index=inuse_space mem.prof  # 当前使用

# Top 10内存分配
(pprof) top10
     100MB 30.00% 30.00%     100MB 30.00%  main.buildRequest
      80MB 24.00% 54.00%      80MB 24.00%  main.processData
      50MB 15.00% 69.00%      50MB 15.00%  runtime.malg

# 查看函数详情
(pprof) list main.buildRequest
```

### 内存泄露排查

**步骤1: 建立基线**
```bash
# 记录初始内存使用
curl http://localhost:6060/debug/pprof/heap > base.prof
```

**步骤2: 压测后对比**
```bash
# 压测一段时间
ab -n 10000 http://localhost:8080/api

# 记录压测后内存
curl http://localhost:6060/debug/pprof/heap > after.prof

# 对比差异
go tool pprof -base base.prof after.prof
```

**步骤3: 定位泄露**
```bash
(pprof) top
     50MB 100.00%  main.GlobalCache.Set
     30MB  60.00%  main.NewRequest
```

**内存泄露案例**:
```go
// ❌ 全局map无限增长
var GlobalCache = make(map[string]*Data)

func Set(key string, data *Data) {
    GlobalCache[key] = data  // 永远不删除
}

// ✅ 使用LRU缓存或定期清理
type LRUCache struct {
    mu    sync.Mutex
    data  map[string]*Data
    queue []string
    max   int
}

func (c *LRUCache) Set(key string, data *Data) {
    c.mu.Lock()
    defer c.mu.Unlock()
    
    if len(c.data) >= c.max {
        // 删除最老的数据
        oldest := c.queue[0]
        delete(c.data, oldest)
        c.queue = c.queue[1:]
    }
    
    c.data[key] = data
    c.queue = append(c.queue, key)
}
```

---

## Goroutine分析

### 采集Goroutine数据

```bash
# 方式1: HTTP endpoint
curl -o goroutine.prof http://localhost:6060/debug/pprof/goroutine

# 方式2: runtime
pprof.Lookup("goroutine").WriteTo(os.Stdout, 1)

# 方式3: pprof工具
go tool pprof http://localhost:6060/debug/pprof/goroutine
```

### 分析Goroutine泄露

```bash
# 查看goroutine数量
(pprof) top
     1000  50.00%  main.worker
      500  25.00%  main.handler

# 查看堆栈
(pprof) traces
-----------+-------------------------------------------------------
      1000   main.worker
             runtime.main
```

**Goroutine泄露案例**:
```go
// ❌ Goroutine泄露
func handler() {
    for {
        conn, err := listener.Accept()
        if err != nil {
            log.Println(err)
            continue
        }
        go handleConn(conn)  // 无限创建，永不退出
    }
}

func handleConn(conn net.Conn) {
    // 没有超时和退出机制
    for {
        data := make([]byte, 1024)
        conn.Read(data)  // 如果客户端不发送数据，goroutine永久阻塞
    }
}

// ✅ 正确处理
func handleConn(conn net.Conn) {
    defer conn.Close()
    
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    
    done := make(chan struct{})
    go func() {
        defer close(done)
        // 处理连接...
    }()
    
    select {
    case <-done:
        return
    case <-ctx.Done():
        return  // 超时退出
    }
}
```

---

## Block和Mutex分析

### Block分析（阻塞）

```go
// 开启block profiling
runtime.SetBlockProfileRate(1)  // 采样率：1=全部采样
```

```bash
# 采集数据
curl -o block.prof http://localhost:6060/debug/pprof/block

# 分析
go tool pprof block.prof
(pprof) top
     500ms  main.processData
     200ms  main.writeToDB
```

### Mutex分析（互斥锁）

```go
// 开启mutex profiling
runtime.SetMutexProfileFraction(1)  // 采样率
```

```bash
# 采集数据
curl -o mutex.prof http://localhost:6060/debug/pprof/mutex

# 分析
go tool pprof mutex.prof
(pprof) top
     300ms  main.processRequest
     150ms  main.updateCache
```

**锁优化案例**:
```go
// ❌ 细粒度锁导致高竞争
var mu sync.Mutex
var cache = make(map[string]string)

func Get(key string) string {
    mu.Lock()
    defer mu.Unlock()
    return cache[key]
}

// ✅ 使用读写锁
var mu sync.RWMutex
var cache = make(map[string]string)

func Get(key string) string {
    mu.RLock()         // 读锁，允许并发读
    defer mu.RUnlock()
    return cache[key]
}

// ✅ 使用sync.Map（高并发场景）
var cache sync.Map

func Get(key string) (string, bool) {
    value, ok := cache.Load(key)
    if !ok {
        return "", false
    }
    return value.(string), true
}
```

---

## Trace工具

### 使用trace

```bash
# 采集trace数据
curl -o trace.out http://localhost:6060/debug/pprof/trace?seconds=5

# 可视化分析
go tool trace trace.out
```

### trace分析维度

```
# 浏览器打开后可以看到：
1. View trace: 时间线视图
   - Goroutines: Goroutine调度情况
   - Network: 网络阻塞
   - Syscalls: 系统调用
   - GC: 垃圾回收

2. Goroutine analysis: Goroutine分析
   - 每个goroutine的运行时间
   - 阻塞时间
   - 同步等待时间

3. Network/Syscall/Heap profile: 各维度分析
```

---

## 性能分析SOP

### 标准流程

```
1. 建立基线
   - 采集初始性能数据
   - 记录关键指标（QPS、延迟、CPU、内存）

2. 问题定位
   - 使用pprof分析瓶颈
   - 查看top N函数
   - 生成火焰图

3. 优化实施
   - 针对性优化代码
   - 避免过早优化
   - 一次只改一个点

4. 验证效果
   - 重新采集数据
   - 对比优化前后
   - 确保性能提升

5. 监控告警
   - 建立性能监控
   - 设置告警阈值
   - 持续跟踪
```

---

## 常用命令速查

```bash
# CPU分析
go tool pprof cpu.prof
(pprof) top10
(pprof) list functionName
(pprof) web

# 内存分析
go tool pprof -sample_index=inuse_space mem.prof
go tool pprof -sample_index=alloc_space mem.prof
(pprof) top10

# 对比分析
go tool pprof -base base.prof after.prof

# Goroutine分析
go tool pprof http://localhost:6060/debug/pprof/goroutine

# Trace分析
go tool trace trace.out

# Benchmark
go test -bench=. -benchmem -cpuprofile=cpu.prof
```

---

## 最佳实践

```
[ ] 生产环境开启pprof endpoint
[ ] 定期采集性能数据
[ ] 建立性能基线
[ ] 使用火焰图快速定位
[ ] 一次只优化一个瓶颈
[ ] 数据驱动决策
[ ] Benchmark验证效果
[ ] 建立监控告警
[ ] 文档化优化过程
[ ] 团队分享经验
```

---

## 学习检查点

完成本章节后，验证：

- [ ] 掌握pprof各种采集方式
- [ ] 能分析CPU和内存瓶颈
- [ ] 排查过内存泄露问题
- [ ] 检测并修复Goroutine泄露
- [ ] 使用trace分析并发问题
- [ ] 会生成和分析火焰图
- [ ] 建立完整的性能分析流程

---

## 延伸阅读

- [pprof官方文档](https://github.com/google/pprof/blob/master/doc/README.md)
- [Go性能分析](https://golang.org/doc/diagnostics.html)
- [火焰图介绍](http://www.brendangregg.com/flamegraphs.html)
- [Uber Go性能指南](https://github.com/uber-go/guide/blob/master/style.md)
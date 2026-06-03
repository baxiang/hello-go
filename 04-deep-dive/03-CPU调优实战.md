# 10.3 CPU调优实战

## CPU性能瓶颈定位

### 常见CPU问题

1. **算法复杂度过高**
2. **频繁内存分配**
3. **锁竞争严重**
4. **系统调用频繁**
5. **GC压力过大**

### CPU性能指标

```bash
# 查看CPU使用率
top -p $(pidof myapp)

# Go runtime指标
curl http://localhost:6060/debug/vars | jq '.cmdline'

# Prometheus metrics
process_cpu_seconds_total
go_goroutines
go_gc_duration_seconds
```

---

## CPU热点优化案例

### 案例1: 字符串拼接优化

**问题代码**:
```go
// ❌ 频繁内存分配
func buildMessage(items []string) string {
    var result string
    for _, item := range items {
        result += item + ","  // 每次分配新字符串
    }
    return result
}

// Benchmark
// BenchmarkBuildMessage-8    100000    12000 ns/op    3200 B/op    99 allocs/op
```

**优化方案**:
```go
// ✅ 使用strings.Builder
func buildMessage(items []string) string {
    var builder strings.Builder
    builder.Grow(len(items) * 20)  // 预估容量
    
    for i, item := range items {
        if i > 0 {
            builder.WriteByte(',')
        }
        builder.WriteString(item)
    }
    return builder.String()
}

// Benchmark
// BenchmarkBuildMessage-8    500000    2500 ns/op    1024 B/op    1 allocs/op
```

**性能提升**: **4.8倍** (12000ns → 2500ns)

---

### 案例2: 切片预分配

**问题代码**:
```go
// ❌ 频繁扩容
func processItems(n int) []int {
    var result []int
    for i := 0; i < n; i++ {
        result = append(result, i*2)  // 多次扩容
    }
    return result
}
```

**优化方案**:
```go
// ✅ 预分配容量
func processItems(n int) []int {
    result := make([]int, 0, n)  // 预分配
    for i := 0; i < n; i++ {
        result = append(result, i*2)
    }
    return result
}

// 或者直接使用长度
func processItems(n int) []int {
    result := make([]int, n)
    for i := 0; i < n; i++ {
        result[i] = i * 2
    }
    return result
}
```

---

### 案例3: 避免反射

**问题代码**:
```go
// ❌ 使用反射
func getField(obj interface{}, field string) interface{} {
    v := reflect.ValueOf(obj)
    f := v.FieldByName(field)
    return f.Interface()
}
```

**优化方案**:
```go
// ✅ 类型断言或接口
type Getter interface {
    GetField(name string) interface{}
}

func getField(obj Getter, field string) interface{} {
    return obj.GetField(field)
}

// 或直接访问
type User struct {
    Name string
    Age  int
}

func (u *User) GetField(name string) interface{} {
    switch name {
    case "Name":
        return u.Name
    case "Age":
        return u.Age
    default:
        return nil
    }
}
```

**性能对比**:
```
反射方式:      500 ns/op
接口方式:       10 ns/op
直接访问:         2 ns/op
提升:         250倍
```

---

## 并发优化策略

### 案例4: 并行计算优化

**问题代码**:
```go
// ❌ 串行处理
func sum(numbers []int) int {
    total := 0
    for _, n := range numbers {
        total += n
    }
    return total
}
```

**优化方案**:
```go
// ✅ 并行计算
func sumParallel(numbers []int) int {
    n := len(numbers)
    if n < 1000 {
        return sum(numbers)  // 小数据量不值得并行
    }
    
    workers := runtime.NumCPU()
    chunkSize := (n + workers - 1) / workers
    
    results := make([]int, workers)
    var wg sync.WaitGroup
    
    for i := 0; i < workers; i++ {
        wg.Add(1)
        go func(idx int) {
            defer wg.Done()
            start := idx * chunkSize
            end := start + chunkSize
            if end > n {
                end = n
            }
            results[idx] = sum(numbers[start:end])
        }(i)
    }
    
    wg.Wait()
    
    total := 0
    for _, r := range results {
        total += r
    }
    return total
}
```

**性能对比** (100万个数字):
```
串行:   2.5 ms
并行:   0.8 ms
提升:   3.1倍
```

---

## 锁优化技巧

### 案例5: 读写锁优化

**问题代码**:
```go
// ❌ 使用互斥锁
type Cache struct {
    mu   sync.Mutex
    data map[string]string
}

func (c *Cache) Get(key string) string {
    c.mu.Lock()
    defer c.mu.Unlock()
    return c.data[key]
}

func (c *Cache) Set(key, value string) {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.data[key] = value
}
```

**优化方案1: 读写锁**:
```go
// ✅ 使用RWMutex
type Cache struct {
    mu   sync.RWMutex
    data map[string]string
}

func (c *Cache) Get(key string) string {
    c.mu.RLock()         // 读锁允许并发
    defer c.mu.RUnlock()
    return c.data[key]
}

func (c *Cache) Set(key, value string) {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.data[key] = value
}
```

**优化方案2: sync.Map**:
```go
// ✅ 使用sync.Map（读多写少场景）
type Cache struct {
    data sync.Map
}

func (c *Cache) Get(key string) (string, bool) {
    value, ok := c.data.Load(key)
    if !ok {
        return "", false
    }
    return value.(string), true
}

func (c *Cache) Set(key, value string) {
    c.data.Store(key, value)
}
```

**性能对比** (读多写少场景，90%读10%写):
```
Mutex:     850 ns/op
RWMutex:   120 ns/op
sync.Map:   50 ns/op
提升:      17倍
```

---

## GC优化策略

### 案例6: 减少GC压力

**问题代码**:
```go
// ❌ 频繁分配
func processRequest() {
    for i := 0; i < 1000; i++ {
        data := make([]byte, 1024)  // 每次分配
        processData(data)
    }
}
```

**优化方案**:
```go
// ✅ 使用sync.Pool
var bufferPool = sync.Pool{
    New: func() interface{} {
        return make([]byte, 1024)
    },
}

func processRequest() {
    for i := 0; i < 1000; i++ {
        data := bufferPool.Get().([]byte)
        processData(data)
        bufferPool.Put(data)  // 复用
    }
}
```

**GC压力对比**:
```
优化前:  GC 15次/秒
优化后:  GC  2次/秒
降低:    87%
```

---

## CPU调优检查清单

```
[ ] 使用pprof定位CPU热点
[ ] 优化算法复杂度（O(n²) → O(n log n)）
[ ] 减少内存分配频率
[ ] 使用strings.Builder拼接字符串
[ ] 切片预分配容量
[ ] 避免频繁反射
[ ] 合理使用并发（数据量判断）
[ ] 使用读写锁替代互斥锁
[ ] sync.Pool复用对象
[ ] 减少GC压力
[ ] 使用benchmark验证优化效果
```

---

## 学习检查点

完成本章节后，验证：

- [ ] 使用pprof定位过CPU热点
- [ ] 优化过字符串拼接性能
- [ ] 实现过切片预分配优化
- [ ] 使用过读写锁优化性能
- [ ] 使用过sync.Pool减少GC
- [ ] 并行优化过计算密集型任务
- [ ] 使用benchmark验证过优化效果

---

## 延伸阅读

- [Go性能优化技巧](https://github.com/dgryski/go-perfbook)
- [高性能Go代码](https://dave.cheney.net/high-performance-go-workshop/dotgo-paris.html)
- [Uber Go风格指南-性能](https://github.com/uber-go/guide/blob/master/style.md#performance)
# 第七部分：高级主题

## 7.1 内存管理

### 内存分配机制

```go
package main

import (
    "fmt"
    "runtime"
)

// ========== Go 内存分配器 ==========
/*
Go 使用 tcmalloc 的改进版本

内存层次:
1. sync.Pool - 对象池 (最快)
2. mspan - 小对象分配 (2KB - 32KB)
3. mheap - 大对象分配 (>32KB)
4. mmap - 直接向系统申请

分配策略:
- 小对象 (< 32KB): 从 mspan 分配
- 大对象 (> 32KB): 直接从 mheap 分配
- 超大对象 (> 32MB): 直接向系统 mmap

逃逸分析:
- 编译器决定变量分配在栈上还是堆上
- 栈分配更快，自动释放
- 堆分配需要 GC
*/

// ========== 查看内存统计 ==========
func printMemStats() {
    var m runtime.MemStats
    runtime.ReadMemStats(&m)

    fmt.Printf("Alloc = %v KB", m.Alloc/1024)
    fmt.Printf(", TotalAlloc = %v KB", m.TotalAlloc/1024)
    fmt.Printf(", Sys = %v KB", m.Sys/1024)
    fmt.Printf(", NumGC = %v\n", m.NumGC)
}

// ========== 减少内存分配 ==========

// 不好：循环内分配
func badAlloc() []byte {
    var result []byte
    for i := 0; i < 1000; i++ {
        result = append(result, make([]byte, 100)...)  // 每次都分配
    }
    return result
}

// 好：预设容量
func goodAlloc() []byte {
    result := make([]byte, 0, 1000*100)  // 一次分配
    for i := 0; i < 1000; i++ {
        // 使用预分配的容量
    }
    return result
}

// ========== 使用 sync.Pool ==========
var bufferPool = sync.Pool{
    New: func() interface{} {
        return make([]byte, 1024)
    },
}

func usePool() {
    buf := bufferPool.Get().([]byte)
    defer bufferPool.Put(buf)

    // 使用 buf
    // 使用完后归还到池中
}

func main() {
    printMemStats()
}
```

### 逃逸分析

```go
package main

// ========== 什么是逃逸分析 ==========
/*
编译器分析变量的作用域，决定分配位置:
- 栈分配：函数返回后自动释放，快速
- 堆分配：GC 管理，有开销

逃逸规则:
1. 返回局部变量指针 -> 逃逸
2. 闭包捕获变量 -> 可能逃逸
3. 接口类型转换 -> 可能逃逸
4. 切片/Map 底层数组 -> 可能逃逸
5. 超大栈对象 -> 逃逸
*/

// ========== 逃逸示例 ==========

// 1. 返回局部变量指针 (逃逸)
func escapeToHeap() *int {
    x := 42
    return &x  // x 逃逸到堆上
}

// 2. 不逃逸 (栈分配)
func noEscape() int {
    x := 42
    return x  // x 在栈上
}

// 3. 闭包捕获 (逃逸)
func closureEscape() func() int {
    x := 42
    return func() int {
        return x  // x 逃逸到堆上
    }
}

// 4. 接口转换 (可能逃逸)
func interfaceEscape() interface{} {
    x := 42
    return x  // x 逃逸到堆上
}

// 5. 大数组 (逃逸)
func largeArray() {
    var buf [1024 * 1024]byte  // 1MB, 可能逃逸
    _ = buf
}

// ========== 查看逃逸分析 ==========
// go build -gcflags="-m" main.go

/*
示例输出:
./main.go:10:6: can inline escapeToHeap
./main.go:11:2: &x escapes to heap

./main.go:16:6: can inline noEscape
./main.go:17:2: x does not escape
*/

// ========== 优化建议 ==========

// 1. 避免不必要的指针返回
func newValue() int {    // 返回值拷贝，栈分配
    return 42
}

// 2. 使用值类型代替接口
func processInt(x int)  // 无逃逸
func processAny(x interface{})  // 可能逃逸

// 3. 预分配切片容量
func appendLoop(n int) []int {
    s := make([]int, 0, n)  // 预设容量
    for i := 0; i < n; i++ {
        s = append(s, i)
    }
    return s
}

// 4. 使用 sync.Pool 复用对象
var pool = sync.Pool{
    New: func() interface{} {
        return &Buffer{}
    },
}
```

### 垃圾回收 (GC) 原理

```go
// ========== GC 算法 ==========
/*
Go 1.5+ 使用并发标记 - 清扫 (Concurrent Mark-Sweep)
Go 1.8+ 使用混合写屏障 (Hybrid Write Barrier)

GC 阶段:
1. 标记准备 (STW, 短暂)
2. 并发标记 (应用程序继续运行)
3. 标记结束 (STW, 短暂)
4. 并发清扫 (应用程序继续运行)

写屏障:
- 保证并发标记的正确性
- 混合写屏障减少 STW 时间
*/

// ========== GC 调优 ==========

// 1. 调整 GOGC (默认 100)
// GOGC=50  -> 更频繁 GC，更低内存
// GOGC=200 -> 更少 GC，更高内存
// export GOGC=50

// 2. 减少分配
// - 预设容量
// - 对象复用
// - 避免逃逸

// 3. 使用 sync.Pool
var pool = sync.Pool{
    New: func() interface{} {
        return make([]byte, 1024)
    },
}

// ========== 查看 GC 统计 ==========
func printGCStats() {
    var m runtime.MemStats
    runtime.ReadMemStats(&m)

    fmt.Printf("GC 次数：%d\n", m.NumGC)
    fmt.Printf("上次 GC: %d ms 前\n", (m.LastGC-time.Now().UnixNano())/1000000)
    fmt.Printf("GC 暂停时间：%v\n", time.Duration(m.PauseTotalNs))
}

// ========== GC 调优工具 ==========
// GODEBUG=gctrace=1 go run main.go
/*
输出示例:
gc 1 @0.012s 0%: 0.12+0.5+0.015 ms clock, 0.5+0.1/5/10+0.060 MB, 56 -> 12 -> 0 MB
格式：gc N @t s cpum%: mark+clean+... MB, heap -> heap after GC -> retained MB
*/
```

### 内存优化技巧

```go
package main

import (
    "bytes"
    "strings"
    "sync"
)

// ========== 1. 字符串拼接 ==========

// 不好：O(n²) 时间复杂度
func badConcat(parts []string) string {
    result := ""
    for _, p := range parts {
        result += p  // 每次都创建新字符串
    }
    return result
}

// 好：使用 strings.Builder
func goodConcat(parts []string) string {
    var sb strings.Builder
    for _, p := range parts {
        sb.WriteString(p)
    }
    return sb.String()
}

// ========== 2. 字节切片操作 ==========

// 使用 bytes.Buffer
func useBuffer(data []byte) []byte {
    var buf bytes.Buffer
    buf.Write(data)
    return buf.Bytes()
}

// ========== 3. Map 预分配 ==========

// 不好
func badMap() map[int]int {
    m := make(map[int]int)
    for i := 0; i < 1000; i++ {
        m[i] = i  // 可能多次扩容
    }
    return m
}

// 好：预设容量提示
func goodMap() map[int]int {
    m := make(map[int]int, 1000)
    for i := 0; i < 1000; i++ {
        m[i] = i
    }
    return m
}

// ========== 4. 切片复用 ==========
var slicePool = sync.Pool{
    New: func() interface{} {
        return make([]byte, 4096)
    },
}

func useSlice() {
    buf := slicePool.Get().([]byte)
    defer slicePool.Put(buf)
    // 使用 buf
}

// ========== 5. 零拷贝 ==========

// 使用切片操作避免复制
func zeroCopy(data []byte, start, end int) []byte {
    return data[start:end]  // 共享底层数组
}

// ========== 6. 避免接口装箱 ==========

// 不好
func processInt(i interface{}) {
    // 装箱操作
}

// 好
func processInt(i int) {
    // 直接操作
}

// ========== 7. 结构体字段对齐 ==========

// 不好：有内存空洞
type BadStruct struct {
    A bool    // 1 byte
    B int64   // 8 bytes (对齐后有 7 bytes 空洞)
    C bool    // 1 byte
}  // 总共 24 bytes

// 好：按大小排序
type GoodStruct struct {
    B int64   // 8 bytes
    A bool    // 1 byte
    C bool    // 1 byte (紧凑排列)
}  // 总共 16 bytes
```

---

## 7.2 反射与 Unsafe

### reflect 包基础

```go
package main

import (
    "fmt"
    "reflect"
)

type Person struct {
    Name string `json:"name"`
    Age  int    `json:"age"`
}

func main() {
    p := Person{Name: "Alice", Age: 25}

    // ========== 获取类型信息 ==========
    t := reflect.TypeOf(p)

    fmt.Println("类型:", t)           // main.Person
    fmt.Println("种类:", t.Kind())    // struct
    fmt.Println("字段数:", t.NumField())  // 2

    // 遍历字段
    for i := 0; i < t.NumField(); i++ {
        field := t.Field(i)
        fmt.Printf("字段 %d: %s (%s) tag=%s\n",
            i, field.Name, field.Type, field.Tag)
    }

    // ========== 获取值信息 ==========
    v := reflect.ValueOf(p)

    fmt.Println("值:", v)
    fmt.Println("字段值:")
    for i := 0; i < v.NumField(); i++ {
        fmt.Printf("  %s = %v\n", v.Type().Field(i).Name, v.Field(i).Interface())
    }

    // ========== 修改值 (需要指针) ==========
    vp := reflect.ValueOf(&p).Elem()  // 解引用

    vp.FieldByName("Name").SetString("Bob")
    vp.FieldByName("Age").SetInt(30)

    fmt.Println("修改后:", p)

    // ========== 调用方法 ==========
    m := reflect.ValueOf(&p).MethodByName("String")
    if m.IsValid() {
        result := m.Call(nil)
        fmt.Println("方法返回:", result[0])
    }

    // ========== 创建实例 ==========
    t2 := reflect.TypeOf(Person{})
    v2 := reflect.New(t2)  // 返回指针
    newInstance := v2.Interface().(*Person)
    fmt.Println("新实例:", newInstance)
}
```

### 反射的应用场景

```go
package main

import (
    "encoding/json"
    "fmt"
    "reflect"
)

// ========== 1. 通用 JSON 解析 ==========
func parseJSON(data []byte, v interface{}) error {
    rv := reflect.ValueOf(v)
    if rv.Kind() != reflect.Ptr || rv.IsNil() {
        return fmt.Errorf("需要非空指针")
    }

    return json.Unmarshal(data, v)
}

// ========== 2. 结构体验证 ==========
func validate(v interface{}) error {
    val := reflect.ValueOf(v)
    typ := val.Type()

    for i := 0; i < val.NumField(); i++ {
        field := val.Field(i)
        fieldType := typ.Field(i)

        // 读取 validate tag
        validateTag := fieldType.Tag.Get("validate")
        if validateTag == "" {
            continue
        }

        // 执行验证
        if validateTag == "required" && field.IsZero() {
            return fmt.Errorf("%s 是必填项", fieldType.Name)
        }
    }
    return nil
}

// ========== 3. 深度拷贝 ==========
func deepCopy(dst, src interface{}) error {
    dstVal := reflect.ValueOf(dst)
    srcVal := reflect.ValueOf(src)

    if dstVal.Kind() != reflect.Ptr || dstVal.IsNil() {
        return fmt.Errorf("dst 必须是非空指针")
    }

    dstVal.Elem().Set(srcVal)
    return nil
}

// ========== 4. 动态调用 ==========
type Handler func(string) string

func invokeHandler(handler Handler, input string) string {
    hv := reflect.ValueOf(handler)
    args := []reflect.Value{reflect.ValueOf(input)}
    results := hv.Call(args)
    return results[0].String()
}

// ========== 5. ORM 映射 ==========
func mapToStruct(data map[string]interface{}, v interface{}) error {
    val := reflect.ValueOf(v).Elem()
    typ := val.Type()

    for i := 0; i < typ.NumField(); i++ {
        field := typ.Field(i)
        jsonTag := field.Tag.Get("json")
        if jsonTag == "" || jsonTag == "-" {
            continue
        }

        if val, ok := data[jsonTag]; ok {
            fieldVal := val.(reflect.Value)
            if fieldVal.Type().ConvertibleTo(field.Type) {
                val.Field(i).Set(fieldVal.Convert(field.Type))
            }
        }
    }
    return nil
}
```

### unsafe 包

```go
package main

import (
    "fmt"
    "unsafe"
)

func main() {
    // ========== unsafe.Pointer ==========
    // unsafe.Pointer 可以转换为任意类型的指针

    x := 42
    p := &x
    ptr := unsafe.Pointer(p)

    // 转换为其他类型指针
    intPtr := (*int)(ptr)
    fmt.Println(*intPtr)  // 42

    // ========== uintptr ==========
    // uintptr 存储指针的整数值

    addr := uintptr(ptr)
    fmt.Printf("地址：0x%x\n", addr)

    // ========== 访问结构体字段 ==========
    type Person struct {
        Name string
        Age  int
    }

    p2 := Person{Name: "Alice", Age: 25}
    p2Ptr := unsafe.Pointer(&p2)

    // Name 字段偏移量 = 0
    namePtr := (*string)(p2Ptr)
    fmt.Println(*namePtr)  // Alice

    // Age 字段偏移量 = string 的大小
    ageOffset := unsafe.Sizeof(string(""))
    agePtr := (*int)(unsafe.Pointer(uintptr(p2Ptr) + ageOffset))
    fmt.Println(*agePtr)  // 25

    // ========== 切片操作 ==========
    arr := [5]int{1, 2, 3, 4, 5}

    // 创建指向数组中间的切片
    ptr3 := unsafe.Pointer(&arr[2])
    slice := unsafe.Slice((*int)(ptr3), 3)  // Go 1.17+
    fmt.Println(slice)  // [3 4 5]

    // ========== 字节转换 ==========
    // 零拷贝字节转换

    s := "hello"
    // 字符串转字节切片 (零拷贝)
    b := unsafe.Slice(
        (*byte)(unsafe.Pointer(unsafe.StringData(s))),
        len(s),
    )
    fmt.Println(string(b))  // hello

    // ========== 注意 ==========
    // unsafe 破坏了类型安全，应谨慎使用
    // 主要用于:
    // 1. 与 C 代码交互
    // 2. 高性能场景 (零拷贝)
    // 3. 底层系统编程
}
```

---

## 7.3 汇编基础

### Go 汇编简介

```go
// ========== Go 汇编文件 ==========
// 文件名: add_amd64.s

/*
TEXT ·Add(SB), NOSPLIT, $0-16
    MOVQ a+0(FP), AX      // 加载参数 a 到 AX 寄存器
    MOVQ b+8(FP), BX      // 加载参数 b 到 BX 寄存器
    ADDQ BX, AX           // AX = AX + BX
    MOVQ AX, ret+16(FP)   // 存储结果到返回值
    RET

// 对应的 Go 声明
// func Add(a, b int64) int64
*/

// ========== 汇编调用 Go ==========
// main.go
/*
package main

import "C"

//export GoFunction
func GoFunction(x int) int {
    return x * 2
}
*/

// ========== 查看汇编代码 ==========
// go tool compile -S main.go
// go build -gcflags="-S" main.go
```

### 性能优化场景

```go
// ========== 使用内联优化 ==========

//go:noinline
func noInlineFunction(x int) int {
    return x * 2
}

//go:inline
func inlineFunction(x int) int {
    return x * 2
}

// ========== 使用 intrinsic 函数 ==========
// Go 编译器自动使用 CPU 指令优化

// bytes.Count 使用 SIMD 指令
n := bytes.Count(data, []byte{0})

// ========== 手动优化热点代码 ==========
// 对于性能关键代码，可以:
// 1. 使用汇编
// 2. 使用 SSA 优化
// 3. 使用 CPU 特定指令
```

---

## 7.4 设计模式

### 单例模式

```go
package main

import "sync"

// ========== sync.Once 实现 ==========
type Singleton struct {
    value string
}

var (
    instance *Singleton
    once     sync.Once
)

func GetInstance() *Singleton {
    once.Do(func() {
        instance = &Singleton{value: "singleton"}
    })
    return instance
}

// ========== 带参数的单例 ==========
type Config struct {
    data map[string]string
}

var (
    configInstance *Config
    configOnce     sync.Once
)

func GetConfig(path string) *Config {
    configOnce.Do(func() {
        configInstance = loadConfig(path)
    })
    return configInstance
}

func loadConfig(path string) *Config {
    // 加载配置
    return &Config{data: make(map[string]string)}
}
```

### 工厂模式

```go
package main

// ========== 简单工厂 ==========
type Product interface {
    Use()
}

type ConcreteProductA struct{}
func (p *ConcreteProductA) Use() { fmt.Println("Using A") }

type ConcreteProductB struct{}
func (p *ConcreteProductB) Use() { fmt.Println("Using B") }

func NewProduct(kind string) Product {
    switch kind {
    case "A":
        return &ConcreteProductA{}
    case "B":
        return &ConcreteProductB{}
    default:
        return nil
    }
}

// ========== 工厂方法 ==========
type Factory interface {
    Create() Product
}

type FactoryA struct{}
func (f *FactoryA) Create() Product { return &ConcreteProductA{} }

type FactoryB struct{}
func (f *FactoryB) Create() Product { return &ConcreteProductB{} }
```

### 选项模式 (Functional Options)

```go
package main

import "time"

// ========== 选项模式 ==========
type Server struct {
    host    string
    port    int
    timeout time.Duration
    retries int
}

// Option 函数类型
type Option func(*Server)

// 选项函数
func WithHost(host string) Option {
    return func(s *Server) {
        s.host = host
    }
}

func WithPort(port int) Option {
    return func(s *Server) {
        s.port = port
    }
}

func WithTimeout(timeout time.Duration) Option {
    return func(s *Server) {
        s.timeout = timeout
    }
}

func WithRetries(retries int) Option {
    return func(s *Server) {
        s.retries = retries
    }
}

// 构造函数
func NewServer(opts ...Option) *Server {
    // 默认值
    server := &Server{
        host:    "localhost",
        port:    8080,
        timeout: 30 * time.Second,
        retries: 3,
    }

    // 应用选项
    for _, opt := range opts {
        opt(server)
    }

    return server
}

// 使用
func main() {
    s := NewServer(
        WithHost("example.com"),
        WithPort(443),
        WithTimeout(time.Minute),
    )
}
```

### 依赖注入

```go
package main

// ========== 手动依赖注入 ==========
type Database interface {
    Query(string) ([]interface{}, error)
}

type UserRepository struct {
    db Database
}

func NewUserRepository(db Database) *UserRepository {
    return &UserRepository{db: db}
}

type UserService struct {
    repo *UserRepository
}

func NewUserService(repo *UserRepository) *UserService {
    return &UserService{repo: repo}
}

// 组合所有依赖
func NewApp() *UserService {
    db := NewMySQLDatabase()
    repo := NewUserRepository(db)
    return NewUserService(repo)
}

// ========== 使用 dig 容器 ==========
/*
import "go.uber.org/dig"

func main() {
    container := dig.New()

    // 提供依赖
    container.Provide(NewMySQLDatabase)
    container.Provide(NewUserRepository)
    container.Provide(NewUserService)

    // 调用
    container.Invoke(func(service *UserService) {
        // 使用 service
    })
}
*/
```

### 观察者模式

```go
package main

import "sync"

// ========== 观察者模式 ==========
type Observer interface {
    Update(event string, data interface{})
}

type Subject struct {
    mu        sync.RWMutex
    observers []Observer
}

func (s *Subject) Attach(o Observer) {
    s.mu.Lock()
    defer s.mu.Unlock()
    s.observers = append(s.observers, o)
}

func (s *Subject) Detach(o Observer) {
    s.mu.Lock()
    defer s.mu.Unlock()
    for i, obs := range s.observers {
        if obs == o {
            s.observers = append(s.observers[:i], s.observers[i+1:]...)
            break
        }
    }
}

func (s *Subject) Notify(event string, data interface{}) {
    s.mu.RLock()
    defer s.mu.RUnlock()
    for _, o := range s.observers {
        o.Update(event, data)
    }
}

// 使用
type LogObserver struct{}
func (o *LogObserver) Update(event string, data interface{}) {
    fmt.Printf("Event: %s, Data: %v\n", event, data)
}
```

### 策略模式

```go
package main

// ========== 策略模式 ==========
type Strategy interface {
    Execute(data []int) []int
}

type BubbleSort struct{}
func (s *BubbleSort) Execute(data []int) []int {
    // 冒泡排序实现
    return data
}

type QuickSort struct{}
func (s *QuickSort) Execute(data []int) []int {
    // 快速排序实现
    return data
}

type Sorter struct {
    strategy Strategy
}

func (s *Sorter) SetStrategy(strategy Strategy) {
    s.strategy = strategy
}

func (s *Sorter) Sort(data []int) []int {
    return s.strategy.Execute(data)
}

// 使用
func main() {
    sorter := &Sorter{}
    sorter.SetStrategy(&QuickSort{})
    result := sorter.Sort([]int{3, 1, 4, 1, 5})
}
```

---

## 7.5 性能优化

### 性能分析方法

```bash
# ========== pprof 使用 ==========

# 1. HTTP 端点 (自动)
import _ "net/http/pprof"
// 访问 http://localhost:6060/debug/pprof/

# 2. CPU 分析
go tool pprof http://localhost:6060/debug/pprof/profile

# 3. 内存分析
go tool pprof http://localhost:6060/debug/pprof/heap

# 4. 阻塞分析
go tool pprof http://localhost:6060/debug/pprof/block

# 5. 互斥锁分析
go tool pprof http://localhost:6060/debug/pprof/mutex

# ========== 生成火焰图 ==========
go tool pprof -http=:8080 cpu.prof
go tool pprof -http=:8080 mem.prof


# ========== 使用 trace ==========
go test -trace=trace.out
go tool trace trace.out
```

### CPU 分析

```go
package main

import (
    "os"
    "runtime/pprof"
)

func main() {
    // ========== CPU 分析 ==========
    f, _ := os.Create("cpu.prof")
    defer f.Close()

    pprof.StartCPUProfile(f)
    defer pprof.StopCPUProfile()

    // 运行代码
    for i := 0; i < 1000000; i++ {
        process(i)
    }
}

func process(n int) int {
    result := 0
    for j := 0; j < n; j++ {
        result += j
    }
    return result
}

// 分析:
// go tool pprof cpu.prof
// (pprof) top10
// (pprof) list process
```

### 内存分析

```go
package main

import (
    "os"
    "runtime"
    "runtime/pprof"
)

func main() {
    // ========== 内存分析 ==========
    // 运行代码产生内存分配

    // 写入内存快照
    f, _ := os.Create("mem.prof")
    defer f.Close()

    runtime.GC()  // 先执行 GC
    pprof.WriteHeapProfile(f)

    // 分析:
    // go tool pprof mem.prof
    // (pprof) top
    // (pprof) list functionName
}

// ========== 常见内存问题 ==========

// 1. 内存泄漏 (goroutine 泄漏)
func leak() {
    ch := make(chan int)
    go func() {
        <-ch  // 永远阻塞
    }()
}

// 2. 不必要的分配
func badConcat() string {
    s := ""
    for i := 0; i < 1000; i++ {
        s += "x"  // 每次都分配
    }
    return s
}

// 3. 大对象分配
func largeAlloc() {
    _ = make([]byte, 100*1024*1024)  // 100MB
}
```

### 性能优化案例

```go
// ========== 案例 1: 字符串拼接优化 ==========

// 优化前: 100ms
func badConcat(parts []string) string {
    result := ""
    for _, p := range parts {
        result += p
    }
    return result
}

// 优化后：1ms
func goodConcat(parts []string) string {
    var sb strings.Builder
    sb.Grow(len(parts) * 10)  // 预设容量
    for _, p := range parts {
        sb.WriteString(p)
    }
    return sb.String()
}


// ========== 案例 2: 切片预分配 ==========

// 优化前：频繁扩容
func badAppend(n int) []int {
    s := make([]int, 0)
    for i := 0; i < n; i++ {
        s = append(s, i)
    }
    return s
}

// 优化后：一次分配
func goodAppend(n int) []int {
    s := make([]int, 0, n)
    for i := 0; i < n; i++ {
        s = append(s, i)
    }
    return s
}


// ========== 案例 3: Map 预分配 ==========

// 优化前
func badMap() map[int]int {
    m := make(map[int]int)
    for i := 0; i < 10000; i++ {
        m[i] = i
    }
    return m
}

// 优化后
func goodMap() map[int]int {
    m := make(map[int]int, 10000)
    for i := 0; i < 10000; i++ {
        m[i] = i
    }
    return m
}


// ========== 案例 4: 使用 sync.Pool ==========

// 优化前
func process() {
    buf := make([]byte, 1024)
    // 使用...
    // 每次调用都分配
}

// 优化后
var pool = sync.Pool{
    New: func() interface{} {
        return make([]byte, 1024)
    },
}

func process() {
    buf := pool.Get().([]byte)
    defer pool.Put(buf)
    // 使用...
    // 复用缓冲区
}


// ========== 案例 5: 避免接口装箱 ==========

// 优化前
func ProcessAny(v interface{}) {
    // 装箱操作
    switch x := v.(type) {
    case int:
        // ...
    }
}

// 优化后
func ProcessInt(v int) {
    // 直接处理
}
```

---

## 第七部分完

接下来可以继续学习第八部分：实战项目

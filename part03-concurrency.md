# 第三部分：并发编程

## 3.1 Goroutine

### Goroutine 概念与原理

```go
package main

import (
    "fmt"
    "time"
)

func main() {
    // ========== 启动 Goroutine ==========
    // 使用 go 关键字启动 goroutine

    // 主 goroutine
    go sayHello("World")  // 新 goroutine

    // 匿名函数
    go func() {
        fmt.Println("匿名 goroutine")
    }()

    // 带参数的匿名函数
    go func(msg string) {
        fmt.Println(msg)
    }("带参数")

    // 等待 goroutine 完成
    time.Sleep(100 * time.Millisecond)

    fmt.Println("主 goroutine 结束")
}

func sayHello(name string) {
    fmt.Println("Hello,", name)
}
```

### Goroutine 调度模型 (GMP 模型)

```go
// ========== GMP 模型详解 ==========

/*
GMP 模型是 Go 运行时调度器的核心设计

G (Goroutine):
- 每个 G 代表一个 goroutine
- 包含：栈、指令指针、状态等
- 轻量级，初始栈大小 2KB

M (Machine):
- 代表操作系统线程
- 真正执行计算的实体
- 默认数量：CPU 核心数

P (Processor):
- 逻辑处理器/调度上下文
- 管理 G 和 M 的调度
- 默认数量：GOMAXPROCS (CPU 核心数)

调度流程:
1. G 创建后放入 P 的本地队列
2. M 从 P 获取 G 执行
3. P 的本地队列为空时，从其他 P 偷取 G (work stealing)
4. G 阻塞时，M 和 P 分离，执行其他 G

优点:
- 用户态调度，开销小
- Work Stealing 平衡负载
- 网络/系统调用不阻塞线程
*/

// ========== 查看 GOMAXPROCS ==========
/*
import "runtime"

// 获取
n := runtime.GOMAXPROCS(0)

// 设置
runtime.GOMAXPROCS(4)
*/
```

### Goroutine 与线程的区别

```go
package main

import (
    "fmt"
    "runtime"
    "time"
)

func main() {
    // ========== 资源消耗对比 ==========
    /*
    | 特性          | Goroutine      | 操作系统线程    |
    |---------------|----------------|----------------|
    | 栈大小        | 2KB (动态增长)  | 1-8MB (固定)    |
    | 创建开销      | 微秒级          | 毫秒级          |
    | 切换开销      | 纳秒级          | 微秒级          |
    | 数量限制      | 百万级          | 数千            |
    | 调度方式      | 用户态调度      | 内核调度        |
    */

    // ========== 大量 Goroutine 示例 ==========
    var m runtime.MemStats

    runtime.ReadMemStats(&m)
    fmt.Printf("创建前 Alloc = %v KB\n", m.Alloc/1024)

    // 创建 10 万个 goroutine
    for i := 0; i < 100000; i++ {
        go func() {
            time.Sleep(time.Second)
        }()
    }

    runtime.ReadMemStats(&m)
    fmt.Printf("创建后 Alloc = %v KB\n", m.Alloc/1024)

    // 等待完成
    time.Sleep(2 * time.Second)

    fmt.Println("完成")
}
```

### Goroutine 泄露问题

```go
package main

import (
    "context"
    "fmt"
    "time"
)

// ========== Goroutine 泄露场景 ==========

// 场景 1: 阻塞的 channel 操作
func leakyChannel() {
    ch := make(chan int)
    go func() {
        v := <-ch  // 永远阻塞，如果没有发送者
        fmt.Println(v)
    }()
    // goroutine 永远无法退出
}

// 场景 2: 无限循环没有退出条件
func leakyLoop() {
    go func() {
        for {
            // 没有退出条件的循环
            time.Sleep(time.Second)
        }
    }()
}

// 场景 3: 子 goroutine 阻塞，父 goroutine 无法继续
func leakyParent() {
    done := make(chan bool)
    go func() {
        // 子 goroutine 做某事
        time.Sleep(10 * time.Second)
        done <- true
    }()

    // 如果这里超时返回，子 goroutine 就泄露了
    <-done
}

// ========== 防止 Goroutine 泄露 ==========

// 方法 1: 使用 context 控制生命周期
func safeWithContext(ctx context.Context) {
    go func() {
        for {
            select {
            case <-ctx.Done():
                return  // 优雅退出
            default:
                // 做工作
                time.Sleep(100 * time.Millisecond)
            }
        }
    }()
}

// 方法 2: 使用带缓冲的 channel 和超时
func safeWithTimeout() {
    done := make(chan bool, 1)  // 带缓冲

    go func() {
        // 工作
        time.Sleep(100 * time.Millisecond)
        done <- true
    }()

    // 超时控制
    select {
    case <-done:
        fmt.Println("完成")
    case <-time.After(time.Second):
        fmt.Println("超时")
    }
}

// 方法 3: 使用 WaitGroup 确保完成
func safeWithWaitGroup() {
    // 见 3.3 节 sync.WaitGroup
}

func main() {
    ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
    defer cancel()

    safeWithContext(ctx)
    time.Sleep(time.Second)
}
```

---

## 3.2 Channel

### Channel 的创建与使用

```go
package main

import "fmt"

func main() {
    // ========== Channel 创建 ==========

    // 无缓冲 channel
    ch1 := make(chan int)

    // 带缓冲 channel (容量 10)
    ch2 := make(chan int, 10)

    // ========== 发送和接收 ==========
    ch := make(chan string)

    // 发送端
    go func() {
        ch <- "Hello"  // 发送
    }()

    // 接收端
    msg := <-ch  // 接收
    fmt.Println(msg)

    // ========== Channel 方向 ==========

    // 只发送 channel
    func sendOnly(ch chan<- int) {
        ch <- 42
        // v := <-ch  // 错误！不能从只发送 channel 接收
    }

    // 只接收 channel
    func recvOnly(ch <-chan int) {
        v := <-ch
        // ch <- 1  // 错误！不能向只接收 channel 发送
        fmt.Println(v)
    }

    sendOnly(ch1)
    recvOnly(ch1)
}
```

### 无缓冲 vs 有缓冲 Channel

```go
package main

import (
    "fmt"
    "time"
)

func main() {
    // ========== 无缓冲 Channel ==========
    // 发送和接收必须同时准备好 (同步)

    unbuffered := make(chan int)

    go func() {
        fmt.Println("准备发送")
        unbuffered <- 42  // 阻塞，直到有接收者
        fmt.Println("发送完成")
    }()

    time.Sleep(100 * time.Millisecond)
    val := <-unbuffered  // 接收
    fmt.Println("接收:", val)

    // ========== 有缓冲 Channel ==========
    // 可以在缓冲区内异步发送

    buffered := make(chan int, 3)

    buffered <- 1  // 不阻塞
    buffered <- 2  // 不阻塞
    buffered <- 3  // 不阻塞

    // buffered <- 4  // 阻塞！缓冲区已满

    fmt.Println(<-buffered)  // 1
    fmt.Println(<-buffered)  // 2
    fmt.Println(<-buffered)  // 3

    // ========== 选择建议 ==========
    /*
    使用无缓冲:
    - 需要同步/握手
    - 简单的请求 - 响应
    - 确保接收者准备好

    使用有缓冲:
    - 生产者消费者模式
    - 需要吞吐性能
    - 缓冲突发流量
    */
}
```

### Channel 的关闭

```go
package main

import "fmt"

func main() {
    ch := make(chan int, 5)

    // ========== 关闭 Channel ==========
    ch <- 1
    ch <- 2
    close(ch)  // 关闭发送端

    // ========== 接收关闭的 Channel ==========
    // 可以继续接收，直到缓冲区为空
    fmt.Println(<-ch)  // 1
    fmt.Println(<-ch)  // 2

    // 缓冲区空后，接收会返回零值
    fmt.Println(<-ch)  // 0 (零值)

    // ========== 检查 Channel 是否关闭 ==========
    ch2 := make(chan int)
    close(ch2)

    val, ok := <-ch2
    fmt.Println(val, ok)  // 0 false (ok=false 表示已关闭)

    // ========== 遍历 Channel ==========
    ch3 := make(chan int, 3)
    ch3 <- 1
    ch3 <- 2
    ch3 <- 3
    close(ch3)

    // range 自动检测关闭
    for v := range ch3 {
        fmt.Println(v)
    }

    // ========== 关闭注意事项 ==========
    // 1. 只能关闭一次
    // close(ch3)  // panic!

    // 2. 不能向关闭的 channel 发送
    // ch3 <- 4  // panic!

    // 3. 接收端不应该关闭 channel
    // 应该由发送端关闭
}
```

### range 遍历 Channel

```go
package main

import "fmt"

func producer(ch chan<- int, n int) {
    for i := 0; i < n; i++ {
        ch <- i
    }
    close(ch)  // 生产完成，关闭 channel
}

func main() {
    ch := make(chan int, 10)

    go producer(ch, 5)

    // ========== range 遍历 ==========
    // 自动接收直到 channel 关闭
    for val := range ch {
        fmt.Println("收到:", val)
    }
    fmt.Println("所有数据接收完成")

    // ========== 多 Channel 遍历 ==========
    // 使用 select 而不是 range
    ch1 := make(chan string)
    ch2 := make(chan string)

    go func() {
        ch1 <- "from ch1"
        ch2 <- "from ch2"
        close(ch1)
        close(ch2)
    }()

    // 不能直接 range 多个 channel
    // 需要使用 select
}
```

### select 多路复用

```go
package main

import (
    "fmt"
    "time"
)

func main() {
    ch1 := make(chan string)
    ch2 := make(chan string)

    // ========== 基本 select ==========
    go func() {
        time.Sleep(100 * time.Millisecond)
        ch1 <- "消息 1"
    }()

    go func() {
        time.Sleep(50 * time.Millisecond)
        ch2 <- "消息 2"
    }()

    // select 等待多个 channel 操作
    select {
    case msg1 := <-ch1:
        fmt.Println("收到:", msg1)
    case msg2 := <-ch2:
        fmt.Println("收到:", msg2)
    }

    // ========== 带超时的 select ==========
    select {
    case msg := <-ch1:
        fmt.Println("收到:", msg)
    case <-time.After(time.Second):
        fmt.Println("超时")
    }

    // ========== 带 default 的 select (非阻塞) ==========
    ch3 := make(chan int, 1)

    select {
    case v := <-ch3:
        fmt.Println("收到:", v)
    default:
        fmt.Println("没有数据，不阻塞")
    }

    // ========== 无限 select ==========
    done := make(chan bool)

    go func() {
        for {
            select {
            case <-done:
                fmt.Println("收到退出信号")
                return
            default:
                // 做工作
                time.Sleep(100 * time.Millisecond)
            }
        }
    }()

    time.Sleep(500 * time.Millisecond)
    done <- true
    time.Sleep(100 * time.Millisecond)
}
```

---

## 3.3 同步原语

### sync.WaitGroup

```go
package main

import (
    "fmt"
    "sync"
    "time"
)

func main() {
    var wg sync.WaitGroup

    // ========== 基本使用 ==========
    for i := 0; i < 5; i++ {
        wg.Add(1)  // 计数器 +1

        go func(id int) {
            defer wg.Done()  // 计数器 -1
            fmt.Printf("Goroutine %d 开始\n", id)
            time.Sleep(100 * time.Millisecond)
            fmt.Printf("Goroutine %d 完成\n", id)
        }(i)
    }

    wg.Wait()  // 等待计数器归零
    fmt.Println("所有 goroutine 完成")

    // ========== Add 可以在 goroutine 外部调用 ==========
    wg2 := sync.WaitGroup{}
    wg2.Add(2)

    go worker(1, &wg2)
    go worker(2, &wg2)

    wg2.Wait()
}

func worker(id int, wg *sync.WaitGroup) {
    defer wg.Done()
    fmt.Printf("Worker %d 工作\n", id)
    time.Sleep(100 * time.Millisecond)
}
```

### sync.Mutex 与 sync.RWMutex

```go
package main

import (
    "fmt"
    "sync"
    "time"
)

// ========== 使用 Mutex 保护共享数据 ==========
type Counter struct {
    mu    sync.Mutex
    value int
}

func (c *Counter) Incr() {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.value++
}

func (c *Counter) Value() int {
    c.mu.Lock()
    defer c.mu.Unlock()
    return c.value
}

// ========== RWMutex (读写锁) ==========
// 适用于读多写少的场景
type Cache struct {
    mu   sync.RWMutex
    data map[string]string
}

func (c *Cache) Get(key string) string {
    c.mu.RLock()      // 读锁 (可多个同时持有)
    defer c.mu.RUnlock()
    return c.data[key]
}

func (c *Cache) Set(key, value string) {
    c.mu.Lock()       // 写锁 (排他)
    defer c.mu.Unlock()
    c.data[key] = value
}

func main() {
    // ========== Mutex 示例 ==========
    counter := &Counter{}
    var wg sync.WaitGroup

    for i := 0; i < 1000; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            counter.Incr()
        }()
    }

    wg.Wait()
    fmt.Println("最终值:", counter.Value())

    // ========== RWMutex 示例 ==========
    cache := &Cache{data: make(map[string]string)}

    // 多个 goroutine 同时读
    var rwg sync.WaitGroup
    for i := 0; i < 10; i++ {
        rwg.Add(1)
        go func(id int) {
            defer rwg.Done()
            cache.Get("key")
        }(i)
    }
    rwg.Wait()
}
```

### sync.Cond

```go
package main

import (
    "fmt"
    "sync"
    "time"
)

// ========== Cond 条件变量 ==========
// 用于 goroutine 等待某个条件成立

type Queue struct {
    mu    sync.Mutex
    cond  *sync.Cond
    items []int
}

func NewQueue() *Queue {
    q := &Queue{items: make([]int, 0)}
    q.cond = sync.NewCond(&q.mu)
    return q
}

func (q *Queue) Push(v int) {
    q.mu.Lock()
    defer q.mu.Unlock()

    q.items = append(q.items, v)
    q.cond.Signal()  // 唤醒一个等待者
}

func (q *Queue) Pop() int {
    q.mu.Lock()
    defer q.mu.Unlock()

    // 等待队列非空
    for len(q.items) == 0 {
        q.cond.Wait()  // 释放锁并等待
    }

    v := q.items[0]
    q.items = q.items[1:]
    return v
}

func main() {
    queue := NewQueue()

    // 消费者
    go func() {
        for i := 0; i < 5; i++ {
            v := queue.Pop()
            fmt.Println("消费:", v)
        }
    }()

    // 生产者
    go func() {
        for i := 1; i <= 5; i++ {
            time.Sleep(500 * time.Millisecond)
            queue.Push(i)
            fmt.Println("生产:", i)
        }
    }()

    time.Sleep(3 * time.Second)
}
```

### sync.Once

```go
package main

import (
    "fmt"
    "sync"
)

// ========== Once 确保只执行一次 ==========

var once sync.Once
var config map[string]string

func initConfig() {
    fmt.Println("初始化配置...")
    config = map[string]string{
        "host": "localhost",
        "port": "8080",
    }
}

func getConfig() map[string]string {
    once.Do(initConfig)  // 只执行一次
    return config
}

func main() {
    // 多个 goroutine 同时调用
    var wg sync.WaitGroup

    for i := 0; i < 5; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            cfg := getConfig()
            fmt.Printf("Goroutine %d: %v\n", id, cfg)
        }(i)
    }

    wg.Wait()
    // 输出：只有一次"初始化配置..."
}

// ========== 单例模式 ==========
type Singleton struct {
    value string
}

var (
    instance *Singleton
    once2    sync.Once
)

func GetInstance() *Singleton {
    once2.Do(func() {
        instance = &Singleton{value: "single"}
    })
    return instance
}
```

### atomic 原子操作

```go
package main

import (
    "fmt"
    "sync"
    "sync/atomic"
)

func main() {
    // ========== 原子整数操作 ==========
    var counter int64
    var wg sync.WaitGroup

    for i := 0; i < 1000; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            atomic.AddInt64(&counter, 1)
        }()
    }

    wg.Wait()
    fmt.Println("计数:", atomic.LoadInt64(&counter))

    // ========== 原子操作函数 ==========

    // 加法
    atomic.AddInt64(&counter, 10)

    // 加载 (读取)
    val := atomic.LoadInt64(&counter)

    // 存储 (写入)
    atomic.StoreInt64(&counter, 0)

    // 比较并交换 (CAS)
    swapped := atomic.CompareAndSwapInt64(&counter, 0, 1)
    fmt.Println("CAS 成功:", swapped)

    // 交换
    old := atomic.SwapInt64(&counter, 100)

    // ========== 原子指针 ==========
    var ptr atomic.Value
    ptr.Store("initial")

    val2 := ptr.Load()
    fmt.Println("原子值:", val2)

    // ========== 性能对比 ==========
    // atomic 比 Mutex 更快，但功能有限
    // 适用于简单计数器场景
}
```

---

## 3.4 Context

### Context 接口详解

```go
package main

import "context"

// ========== Context 接口定义 ==========
/*
type Context interface {
    Deadline() (deadline time.Time, ok bool)  // 截止时间
    Done() <-chan struct{}                    // 完成信号
    Err() error                               // 错误原因
    Value(key any) any                        // 获取值
}
*/

// ========== 创建 Context ==========
func createContexts() {
    // 1. 根 context (不取消)
    ctx := context.Background()

    // 2. 可取消的 context
    ctx2, cancel := context.WithCancel(context.Background())
    defer cancel()

    // 3. 带超时的 context
    ctx3, cancel3 := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel3()

    // 4. 带截止时间的 context
    ctx4, cancel4 := context.WithDeadline(context.Background(), time.Now().Add(time.Hour))
    defer cancel4()

    // 5. 带值的 context
    ctx5 := context.WithValue(context.Background(), "key", "value")
}
```

### Context 的创建与传递

```go
package main

import (
    "context"
    "fmt"
    "time"
)

// ========== Context 传递 ==========
func process(ctx context.Context, id int) {
    select {
    case <-ctx.Done():
        fmt.Printf("任务 %d 被取消\n", id)
        return
    case <-time.After(time.Second):
        fmt.Printf("任务 %d 完成\n", id)
    }
}

func main() {
    // ========== 传递 context ==========
    ctx := context.Background()

    // 创建可取消的 context
    ctx, cancel := context.WithCancel(ctx)
    defer cancel()

    // 传递给子 goroutine
    for i := 0; i < 5; i++ {
        go process(ctx, i)
    }

    // 取消所有
    time.Sleep(500 * time.Millisecond)
    cancel()

    time.Sleep(time.Second)
}

// ========== 值传递 ==========
type contextKey string

const userIDKey contextKey = "userID"

func withUserID(ctx context.Context, userID string) context.Context {
    return context.WithValue(ctx, userIDKey, userID)
}

func getUserID(ctx context.Context) string {
    if v := ctx.Value(userIDKey); v != nil {
        return v.(string)
    }
    return ""
}

func handleRequest(ctx context.Context) {
    userID := getUserID(ctx)
    fmt.Println("用户 ID:", userID)
}
```

### Context 取消机制

```go
package main

import (
    "context"
    "fmt"
    "time"
)

// ========== 优雅取消 ==========
func worker(ctx context.Context, id int) {
    for {
        select {
        case <-ctx.Done():
            fmt.Printf("Worker %d 退出\n", id)
            return
        default:
            // 做工作
            time.Sleep(100 * time.Millisecond)
        }
    }
}

// ========== 级联取消 ==========
func cascade() {
    parent, cancel := context.WithCancel(context.Background())
    defer cancel()

    // 子 context 会继承父的取消
    child1, cancel1 := context.WithCancel(parent)
    defer cancel1()

    child2, cancel2 := context.WithCancel(parent)
    defer cancel2()

    // 取消 parent，所有子都会取消
    cancel()

    fmt.Println("child1 被取消:", child1.Err())
    fmt.Println("child2 被取消:", child2.Err())
}

// ========== 部分取消 ==========
func partialCancel() {
    parent := context.Background()

    // 只取消其中一个分支
    child1, cancel1 := context.WithCancel(parent)
    child2, cancel2 := context.WithCancel(parent)

    cancel1()  // 只取消 child1 及其子树

    fmt.Println("child1 被取消:", child1.Err())     // context canceled
    fmt.Println("child2 被取消:", child2.Err())     // <nil>
}

func main() {
    ctx, cancel := context.WithCancel(context.Background())

    for i := 0; i < 5; i++ {
        go worker(ctx, i)
    }

    time.Sleep(500 * time.Millisecond)
    cancel()

    time.Sleep(time.Second)
}
```

### Context 与超时控制

```go
package main

import (
    "context"
    "fmt"
    "time"
)

// ========== WithTimeout ==========
func withTimeoutExample() {
    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
    defer cancel()

    result := make(chan string, 1)

    go func() {
        time.Sleep(3 * time.Second)  // 模拟慢操作
        result <- "完成"
    }()

    select {
    case r := <-result:
        fmt.Println(r)
    case <-ctx.Done():
        fmt.Println("超时:", ctx.Err())
    }
}

// ========== WithDeadline ==========
func withDeadlineExample() {
    deadline := time.Now().Add(2 * time.Second)
    ctx, cancel := context.WithDeadline(context.Background(), deadline)
    defer cancel()

    select {
    case <-ctx.Done():
        fmt.Println("超时或取消:", ctx.Err())
    case <-time.After(3 * time.Second):
        fmt.Println("完成")
    }
}

// ========== HTTP 请求超时 ==========
func httpTimeoutExample() {
    /*
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    req, _ := http.NewRequestWithContext(ctx, "GET", "http://example.com", nil)
    _, err := http.DefaultClient.Do(req)
    if err != nil {
        if ctx.Err() == context.DeadlineExceeded {
            fmt.Println("请求超时")
        }
    }
    */
}

// ========== 组合使用 ==========
func combinedExample() {
    // 基础超时
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    // 可以手动提前取消
    go func() {
        // 提前完成时取消
        // cancel()
    }()

    select {
    case <-ctx.Done():
        fmt.Println("完成或超时")
    }
}

func main() {
    withTimeoutExample()
}
```

### Context 使用最佳实践

```go
package main

import (
    "context"
    "fmt"
)

// ========== 最佳实践 ==========

// 1. Context 作为第一个参数
func DoWork(ctx context.Context, arg string) error {
    // ...
    return nil
}

// 2. 不要将 Context 存在结构体中
// 错误
type BadWorker struct {
    ctx context.Context  // 不要这样
}

// 正确
type GoodWorker struct {
    // context 在方法参数中传递
}

func (w *GoodWorker) Run(ctx context.Context) {
    // ...
}

// 3. 不要传递 nil context
// 错误：DoWork(nil, "arg")
// 正确：DoWork(context.Background(), "arg")

// 4. 取消函数必须由调用者调用
func caller() {
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()  // 确保调用

    DoWork(ctx, "arg")
}

// 5. 使用 context 传递请求作用域的值
type requestIDKey struct{}

func WithRequestID(ctx context.Context, id string) context.Context {
    return context.WithValue(ctx, requestIDKey{}, id)
}

func GetRequestID(ctx context.Context) string {
    if id, ok := ctx.Value(requestIDKey{}).(string); ok {
        return id
    }
    return ""
}

// 6. 每个请求使用独立的 context
func handleRequest(ctx context.Context) {
    // 为子操作创建派生 context
    childCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()

    // 使用 childCtx
    _ = childCtx
}

func main() {
    ctx := context.Background()
    ctx = WithRequestID(ctx, "12345")

    id := GetRequestID(ctx)
    fmt.Println("Request ID:", id)
}
```

---

## 3.5 并发模式

### Worker Pool 模式

```go
package main

import (
    "fmt"
    "sync"
    "time"
)

// ========== Worker Pool ==========
type Job struct {
    ID   int
    Data string
}

type Result struct {
    JobID  int
    Output string
}

func worker(id int, jobs <-chan Job, results chan<- Result, wg *sync.WaitGroup) {
    defer wg.Done()

    for job := range jobs {
        fmt.Printf("Worker %d 处理任务 %d\n", id, job.ID)
        time.Sleep(100 * time.Millisecond)
        results <- Result{JobID: job.ID, Output: "完成"}
    }
}

func workerPool() {
    const numWorkers = 3
    const numJobs = 10

    jobs := make(chan Job, numJobs)
    results := make(chan Result, numJobs)

    var wg sync.WaitGroup

    // 启动 worker
    for i := 1; i <= numWorkers; i++ {
        wg.Add(1)
        go worker(i, jobs, results, &wg)
    }

    // 发送任务
    for j := 1; j <= numJobs; j++ {
        jobs <- Job{ID: j, Data: fmt.Sprintf("task-%d", j)}
    }
    close(jobs)

    // 等待所有 worker 完成
    go func() {
        wg.Wait()
        close(results)
    }()

    // 收集结果
    for r := range results {
        fmt.Printf("结果：%d - %s\n", r.JobID, r.Output)
    }
}

func main() {
    workerPool()
}
```

### Producer-Consumer 模式

```go
package main

import (
    "fmt"
    "sync"
)

// ========== 基础版本 ==========
func producerConsumer() {
    ch := make(chan int, 10)
    var wg sync.WaitGroup

    // 生产者
    wg.Add(1)
    go func() {
        defer wg.Done()
        for i := 0; i < 10; i++ {
            ch <- i
        }
        close(ch)
    }()

    // 消费者
    wg.Add(1)
    go func() {
        defer wg.Done()
        for v := range ch {
            fmt.Println("消费:", v)
        }
    }()

    wg.Wait()
}

// ========== 多生产者多消费者 ==========
func multiProducersConsumers() {
    ch := make(chan int, 100)
    var wg sync.WaitGroup

    // 多个生产者
    for p := 0; p < 3; p++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            for i := 0; i < 10; i++ {
                ch <- id*100 + i
            }
        }(p)
    }

    // 等待生产者完成后关闭 channel
    go func() {
        wg.Wait()
        close(ch)
    }()

    // 多个消费者
    for c := 0; c < 5; c++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            for v := range ch {
                fmt.Printf("消费者 %d: %d\n", id, v)
            }
        }(c)
    }
}
```

### Fan-In/Fan-Out 模式

```go
package main

import (
    "fmt"
    "sync"
)

// ========== Fan-Out (分发) ==========
// 一个输入，多个处理者

func fanOut(input <-chan int, numWorkers int) []<-chan int {
    outputs := make([]<-chan int, numWorkers)

    for i := 0; i < numWorkers; i++ {
        out := make(chan int)
        outputs[i] = out

        go func(ch chan<- int) {
            defer close(ch)
            for v := range input {
                if v%numWorkers == i {
                    ch <- v
                }
            }
        }(out)
    }

    return outputs
}

// ========== Fan-In (合并) ==========
// 多个输入，一个输出

func fanIn(channels ...<-chan int) <-chan int {
    out := make(chan int)

    var wg sync.WaitGroup
    wg.Add(len(channels))

    for _, ch := range channels {
        go func(c <-chan int) {
            defer wg.Done()
            for v := range c {
                out <- v
            }
        }(ch)
    }

    go func() {
        wg.Wait()
        close(out)
    }()

    return out
}

func main() {
    // 创建输入
    input := make(chan int, 100)
    for i := 0; i < 100; i++ {
        input <- i
    }
    close(input)

    // Fan-out
    workers := fanOut(input, 3)

    // 处理
    processed := make([]<-chan int, 3)
    for i, ch := range workers {
        out := make(chan int)
        processed[i] = out
        go func(c <-chan int, out chan<- int) {
            defer close(out)
            for v := range c {
                out <- v * 2  // 处理
            }
        }(ch, out)
    }

    // Fan-in
    result := fanIn(processed...)

    // 收集结果
    for v := range result {
        fmt.Println(v)
    }
}
```

### Pipeline 模式

```go
package main

import (
    "fmt"
)

// ========== 管道阶段 ==========
func stage1(in <-chan int) <-chan int {
    out := make(chan int)
    go func() {
        defer close(out)
        for n := range in {
            out <- n * 2  // 第一阶段处理
        }
    }()
    return out
}

func stage2(in <-chan int) <-chan int {
    out := make(chan int)
    go func() {
        defer close(out)
        for n := range in {
            out <- n + 10  // 第二阶段处理
        }
    }()
    return out
}

func stage3(in <-chan int) <-chan int {
    out := make(chan int)
    go func() {
        defer close(out)
        for n := range in {
            out <- n * n  // 第三阶段处理
        }
    }()
    return out
}

// ========== 构建管道 ==========
func pipeline() {
    // 创建输入
    input := make(chan int, 10)
    for i := 1; i <= 10; i++ {
        input <- i
    }
    close(input)

    // 连接管道
    p1 := stage1(input)
    p2 := stage2(p1)
    p3 := stage3(p2)

    // 输出结果
    for result := range p3 {
        fmt.Println(result)
    }
    // 1 -> 2 -> 12 -> 144
}

// ========== 带错误的管道 ==========
type Result struct {
    Value int
    Error error
}

func pipelineWithError() {
    // 可以传递 Result 结构体来处理错误
}
```

### ErrGroup 使用

```go
package main

import (
    "context"
    "fmt"
    "golang.org/x/sync/errgroup"
    "time"
)

// ========== ErrGroup 基础 ==========
func errgroupBasic() error {
    var g errgroup.Group

    for i := 0; i < 3; i++ {
        i := i
        g.Go(func() error {
            fmt.Printf("任务 %d\n", i)
            if i == 2 {
                return fmt.Errorf("任务 %d 失败", i)
            }
            return nil
        })
    }

    return g.Wait()
}

// ========== 带 Context 的 ErrGroup ==========
func errgroupWithContext() error {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    g, ctx := errgroup.WithContext(ctx)

    for i := 0; i < 5; i++ {
        i := i
        g.Go(func() error {
            select {
            case <-ctx.Done():
                return ctx.Err()
            default:
                fmt.Printf("任务 %d 执行\n", i)
                time.Sleep(time.Second)
                return nil
            }
        })
    }

    return g.Wait()
}

// ========== 限制并发数 ==========
func errgroupWithLimit() error {
    g, ctx := errgroup.WithContext(context.Background())
    g.SetLimit(3)  // 最多 3 个并发

    for i := 0; i < 10; i++ {
        i := i
        g.Go(func() error {
            select {
            case <-ctx.Done():
                return ctx.Err()
            default:
                fmt.Printf("任务 %d\n", i)
                time.Sleep(500 * time.Millisecond)
                return nil
            }
        })
    }

    return g.Wait()
}

func main() {
    if err := errgroupBasic(); err != nil {
        fmt.Println("错误:", err)
    }
}
```

---

## 第三部分完

接下来可以继续学习第四部分：标准库精讲

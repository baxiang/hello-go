# 2. Channel 高级用法与模式

> "不要通过共享内存来通信；而是通过通信来共享内存。"

## 2.1 Channel 核心概念

### 2.1.1 Channel 内部结构

```
┌─────────────────────────────────────────────────────────────┐
│                   Channel 结构体                             │
├─────────────────────────────────────────────────────────────┤
│  ┌─────────────────────────────────────────────────────┐   │
│  │              环形缓冲区 (hbuf)                       │   │
│  │  [ ][ ][ ][ ][ ][ ][ ][ ][ ][ ]                    │   │
│  │   ↑                                      ↑          │   │
│  │  sendx (发送位置)                    recvx (接收位置) │   │
│  └─────────────────────────────────────────────────────┘   │
│                                                             │
│  sendq: *waitq  等待发送的 goroutine 队列                  │
│  recvq: *waitq  等待接收的 goroutine 队列                  │
│  closed: bool   是否已关闭                                  │
│  elemtype: *type  元素类型                                 │
└─────────────────────────────────────────────────────────────┘
```

### 2.1.2 Channel 类型

```go
package main

import "fmt"

func main() {
    // 1. 无缓冲 channel (同步)
    ch1 := make(chan int)
    
    // 2. 有缓冲 channel (异步)
    ch2 := make(chan int, 10)
    
    // 3. 只读 channel
    var readOnly <-chan int = ch1
    
    // 4. 只写 channel
    var writeOnly chan<- int = ch1
    
    // 5. 双向 channel (默认)
    ch3 := make(chan int)
    
    fmt.Printf("ch1: %T\n", ch1)
    fmt.Printf("ch2: %T\n", ch2)
    fmt.Printf("readOnly: %T\n", readOnly)
    fmt.Printf("writeOnly: %T\n", writeOnly)
}
```

## 2.2 Select 高级用法

### 2.2.1 多路复用

```go
package main

import (
    "fmt"
    "time"
)

func main() {
    ch1 := make(chan string)
    ch2 := make(chan string)
    
    // Goroutine 1
    go func() {
        time.Sleep(100 * time.Millisecond)
        ch1 <- "one"
    }()
    
    // Goroutine 2
    go func() {
        time.Sleep(200 * time.Millisecond)
        ch2 <- "two"
    }()
    
    // 使用 select 等待多个 channel
    for i := 0; i < 2; i++ {
        select {
        case msg1 := <-ch1:
            fmt.Println("received:", msg1)
        case msg2 := <-ch2:
            fmt.Println("received:", msg2)
        }
    }
}
```

### 2.2.2 超时处理

```go
package main

import (
    "fmt"
    "time"
)

func main() {
    ch := make(chan string)
    
    go func() {
        time.Sleep(2 * time.Second)
        ch <- "result"
    }()
    
    // 超时处理
    select {
    case result := <-ch:
        fmt.Println("收到结果:", result)
    case <-time.After(500 * time.Millisecond):
        fmt.Println("超时!")
    }
    
    // 非阻塞尝试
    select {
    case msg := <-ch:
        fmt.Println("收到:", msg)
    default:
        fmt.Println("没有消息可用")
    }
}
```

### 2.2.3 退出信号

```go
package main

import (
    "fmt"
    "time"
)

func worker(done chan struct{}) {
    ticker := time.NewTicker(time.Second)
    defer ticker.Stop()
    
    for {
        select {
        case <-ticker.C:
            fmt.Println("工作中...")
        case <-done:
            fmt.Println("收到退出信号")
            return
        }
    }
}

func main() {
    done := make(chan struct{})
    
    go worker(done)
    
    time.Sleep(3 * time.Second)
    close(done)  // 发送退出信号
    
    time.Sleep(time.Second)
}
```

## 2.3 高级 Channel 模式

### 2.3.1 管道 (Pipeline)

```go
package main

import "fmt"

// 生成器：产生数据
func generate(nums ...int) <-chan int {
    out := make(chan int)
    go func() {
        for _, n := range nums {
            out <- n
        }
        close(out)
    }()
    return out
}

// 平方计算
func square(in <-chan int) <-chan int {
    out := make(chan int)
    go func() {
        for n := range in {
            out <- n * n
        }
        close(out)
    }()
    return out
}

// 过滤
func filter(in <-chan int, predicate func(int) bool) <-chan int {
    out := make(chan int)
    go func() {
        for n := range in {
            if predicate(n) {
                out <- n
            }
        }
        close(out)
    }()
    return out
}

func main() {
    // 构建管道：generate -> square -> filter
    pipeline := square(generate(1, 2, 3, 4, 5))
    pipeline = filter(pipeline, func(n int) bool { return n > 10 })
    
    // 消费结果
    sum := 0
    for n := range pipeline {
        sum += n
        fmt.Println("got:", n)
    }
    fmt.Println("sum:", sum)
}
```

### 2.3.2 扇出扇入 (Fan-out, Fan-in)

```go
package main

import (
    "fmt"
    "sync"
)

// 扇出：多个 worker 处理同一输入
func fanOut(in <-chan int, workers int) []<-chan int {
    var wg sync.WaitGroup
    outputs := make([]<-chan int, workers)
    
    for i := 0; i < workers; i++ {
        wg.Add(1)
        ch := make(chan int, 10)
        outputs[i] = ch
        
        go func() {
            defer wg.Done()
            for n := range in {
                ch <- n * 2  // 模拟处理
            }
            close(ch)
        }()
    }
    return outputs
}

// 扇入：合并多个 channel
func fanIn(inputs ...<-chan int) <-chan int {
    var wg sync.WaitGroup
    out := make(chan int)
    
    output := func(ch <-chan int) {
        defer wg.Done()
        for n := range ch {
            out <- n
        }
    }
    
    wg.Add(len(inputs))
    for _, ch := range inputs {
        go output(ch)
    }
    
    go func() {
        wg.Wait()
        close(out)
    }()
    
    return out
}

func main() {
    // 创建输入
    in := make(chan int)
    go func() {
        for i := 0; i < 10; i++ {
            in <- i
        }
        close(in)
    }()
    
    // 扇出到 3 个 worker
    outputs := fanOut(in, 3)
    
    // 扇入结果
    result := fanIn(outputs...)
    
    // 打印结果
    for n := range result {
        fmt.Println(n)
    }
}
```

### 2.3.3 限流器

```go
package main

import (
    "fmt"
    "time"
)

// RateLimiter 速率限制器
type RateLimiter struct {
    requests chan struct{}
}

func NewRateLimiter(rate int) *RateLimiter {
    limiter := &RateLimiter{
        requests: make(chan struct{}, rate),
    }
    
    // 定期补充令牌
    go func() {
        ticker := time.NewTicker(time.Second)
        for range ticker.C {
            select {
            case limiter.requests <- struct{}{}:
            default:
            }
        }
    }()
    
    return limiter
}

func (r *RateLimiter) Allow() bool {
    select {
    case <-r.requests:
        return true
    default:
        return false
    }
}

func main() {
    limiter := NewRateLimiter(5)  // 每秒 5 个请求
    
    for i := 0; i < 20; i++ {
        if limiter.Allow() {
            fmt.Printf("请求 %d: 允许\n", i)
        } else {
            fmt.Printf("请求 %d: 拒绝\n", i)
        }
        time.Sleep(100 * time.Millisecond)
    }
}
```

## 2.4 实际案例：Docker 中的 Channel 使用

### 2.4.1 Container 状态管理

```go
// docker/container/state.go (简化)

// State 容器状态
type State struct {
    stateC  chan StateType
    current StateType
}

// NewState 创建状态机
func NewState() *State {
    s := &State{
        stateC: make(chan StateType, 1),
    }
    go s.watch()
    return s
}

// watch 监听状态变化
func (s *State) watch() {
    for state := range s.stateC {
        s.current = state
        // 处理状态变化
    }
}

// Transition 状态转换
func (s *State) Transition(to StateType) {
    select {
    case s.stateC <- to:
    default:
        // 通道已满，状态转换失败
    }
}
```

### 2.4.2 Exec 容器命令

```go
// docker/container/exec.go (简化)

// ExecProcess 执行进程
type ExecProcess struct {
    stdin  io.ReadCloser
    stdout io.Writer
    stderr io.Writer
    exitCh chan int
}

// Wait 等待执行完成
func (p *ExecProcess) Wait() (int, error) {
    // 从 channel 获取退出码
    code := <-p.exitCh
    return code, nil
}
```

## 2.5 Channel 最佳实践

### 2.5.1 关闭 Channel 的原则

```go
package main

import "sync"

// 原则 1: 由发送方关闭 channel
func sender(ch chan<- int) {
    for i := 0; i < 10; i++ {
        ch <- i
    }
    close(ch)  // 发送方负责关闭
}

// 原则 2: 使用 defer 确保关闭
func safeSender(ch chan int) {
    defer close(ch)
    for i := 0; i < 10; i++ {
        ch <- i
    }
}

// 原则 3: 使用 sync.Once 防止重复关闭
type safeChannel struct {
    ch   chan int
    once sync.Once
}

func (s *safeChannel) Close() {
    s.once.Do(func() {
        close(s.ch)
    })
}
```

### 2.5.2 避免 Channel 泄漏

```go
package main

import (
    "fmt"
    "time"
)

// 错误示例：发送后无人接收
func badExample() {
    ch := make(chan int)
    go func() {
        ch <- 1  // 泄漏！
    }()
}

// 正确示例 1: 使用带缓冲的 channel
func correct1() {
    ch := make(chan int, 1)  // 有缓冲
    go func() {
        ch <- 1  // 不会阻塞
    }()
    <-ch  // 接收
}

// 正确示例 2: 使用 select + default
func correct2() {
    ch := make(chan int)
    select {
    case ch <- 1:
        fmt.Println("sent")
    default:
        fmt.Println("channel full, skip")
    }
}

// 正确示例 3: 使用 context 取消
func correct3() {
    ch := make(chan int)
    done := make(chan struct{})
    
    go func() {
        select {
        case ch <- 1:
            fmt.Println("sent")
        case <-done:
            fmt.Println("cancelled")
        }
    }()
    
    time.Sleep(time.Millisecond)
    close(done)
}
```

## 2.6 本章小结

```
┌─────────────────────────────────────────────────────────────┐
│                      本章总结                                │
├─────────────────────────────────────────────────────────────┤
│ ✓ Channel 是 Go 并发的核心通信方式                          │
│   └── 环形缓冲区 + 等待队列                                │
│                                                             │
│ ✓ 常见模式：管道、扇入扇出、限流                         │
│                                                             │
│ ✓ Select 是处理多 channel 的利器                          │
│   └── 超时、默认分支、退出信号                            │
│                                                             │
│ ✓ 正确关闭 channel: 发送方负责、防止重复关闭              │
│                                                             │
│ ✓ 避免泄漏：带缓冲、select、context                     │
│                                                             │
│ ✓ Docker 等项目的实际应用                                │
└─────────────────────────────────────────────────────────────┘
```

---

[上一章：← Goroutine 深度解析](./01-Goroutine 深度解析与 GMP 模型.md) | [下一章：同步原语详解 →](./03-同步原语详解.md)
# 3.1 Goroutine

## 基础使用

```go
package main

import (
    "fmt"
    "time"
)

func main() {
    // 启动 goroutine
    go sayHello("World")

    // 匿名函数
    go func() {
        fmt.Println("匿名 goroutine")
    }()

    // 带参数
    go func(msg string) {
        fmt.Println(msg)
    }("带参数")

    time.Sleep(100 * time.Millisecond)
}

func sayHello(name string) {
    fmt.Println("Hello,", name)
}
```

---

## GMP 模型

```
G (Goroutine):
- 每个 G 代表一个 goroutine
- 包含：栈、指令指针、状态等
- 轻量级，初始栈大小 2KB

M (Machine):
- 代表操作系统线程
- 真正执行计算的实体

P (Processor):
- 逻辑处理器/调度上下文
- 默认数量：GOMAXPROCS (CPU 核心数)

调度流程:
1. G 创建后放入 P 的本地队列
2. M 从 P 获取 G 执行
3. P 的本地队列为空时，从其他 P 偷取 G
```

---

## Goroutine 泄露

```go
package main

import (
    "context"
    "time"
)

// 泄露场景
func leakyChannel() {
    ch := make(chan int)
    go func() {
        v := <-ch  // 永远阻塞
        _ = v
    }()
}

// 防止泄露 - 使用 context
func safeWithContext(ctx context.Context) {
    go func() {
        for {
            select {
            case <-ctx.Done():
                return
            default:
                time.Sleep(100 * time.Millisecond)
            }
        }
    }()
}

// 防止泄露 - 使用带缓冲的 channel
func safeWithTimeout() {
    done := make(chan bool, 1)
    go func() {
        time.Sleep(100 * time.Millisecond)
        done <- true
    }()

    select {
    case <-done:
        fmt.Println("完成")
    case <-time.After(time.Second):
        fmt.Println("超时")
    }
}
```

# 3.2 Channel

## 基础使用

```go
package main

import "fmt"

func main() {
    // 无缓冲 channel
    ch1 := make(chan int)

    // 带缓冲 channel
    ch2 := make(chan int, 10)

    // 发送和接收
    ch := make(chan string)
    go func() {
        ch <- "Hello"  // 发送
    }()
    msg := <-ch  // 接收
    fmt.Println(msg)

    // 只发送 channel
    func sendOnly(ch chan<- int) {
        ch <- 42
    }

    // 只接收 channel
    func recvOnly(ch <-chan int) {
        v := <-ch
        fmt.Println(v)
    }
}
```

---

## 无缓冲 vs 有缓冲

```go
package main

import "fmt"

func main() {
    // 无缓冲 - 同步
    unbuffered := make(chan int)
    go func() {
        unbuffered <- 42  // 阻塞直到有接收者
    }()
    val := <-unbuffered

    // 有缓冲 - 异步
    buffered := make(chan int, 3)
    buffered <- 1  // 不阻塞
    buffered <- 2
    buffered <- 3
    // buffered <- 4  // 阻塞！缓冲区已满

    fmt.Println(<-buffered)
}
```

---

## Channel 关闭

```go
package main

import "fmt"

func main() {
    ch := make(chan int, 5)
    ch <- 1
    ch <- 2
    close(ch)  // 关闭

    // 接收
    fmt.Println(<-ch)  // 1
    fmt.Println(<-ch)  // 2

    // 检查是否关闭
    val, ok := <-ch
    fmt.Println(val, ok)  // 0 false

    // range 遍历
    ch2 := make(chan int, 3)
    ch2 <- 1
    ch2 <- 2
    ch2 <- 3
    close(ch2)

    for v := range ch2 {
        fmt.Println(v)
    }

    // 注意:
    // 1. 只能关闭一次
    // 2. 不能向关闭的 channel 发送
    // 3. 应该由发送端关闭
}
```

---

## select 多路复用

```go
package main

import (
    "fmt"
    "time"
)

func main() {
    ch1 := make(chan string)
    ch2 := make(chan string)

    go func() {
        time.Sleep(100 * time.Millisecond)
        ch1 <- "消息 1"
    }()

    go func() {
        time.Sleep(50 * time.Millisecond)
        ch2 <- "消息 2"
    }()

    // 基本 select
    select {
    case msg1 := <-ch1:
        fmt.Println("收到:", msg1)
    case msg2 := <-ch2:
        fmt.Println("收到:", msg2)
    }

    // 带超时
    select {
    case msg := <-ch1:
        fmt.Println("收到:", msg)
    case <-time.After(time.Second):
        fmt.Println("超时")
    }

    // 非阻塞
    ch3 := make(chan int, 1)
    select {
    case v := <-ch3:
        fmt.Println("收到:", v)
    default:
        fmt.Println("没有数据")
    }

    // 无限 select
    done := make(chan bool)
    go func() {
        for {
            select {
            case <-done:
                return
            default:
                time.Sleep(100 * time.Millisecond)
            }
        }
    }()
}
```

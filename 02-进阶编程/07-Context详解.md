# 4. Context 深度解析

> Context 是 Go 并发控制的核心，用于取消信号、截止时间和值传递。

## 4.1 Context 接口

```go
type Context interface {
    Deadline() (deadline time.Time, ok bool)
    Done() <-chan struct{}
    Err() error
    Value(key interface{}) interface{}
}
```

## 4.2 取消信号

### 4.2.1 基本用法

```go
package main

import (
    "context"
    "fmt"
    "time"
)

func worker(ctx context.Context) {
    for {
        select {
        case <-ctx.Done():
            fmt.Println("收到取消信号")
            return
        default:
            fmt.Println("工作中...")
            time.Sleep(time.Second)
        }
    }
}

func main() {
    ctx, cancel := context.WithCancel(context.Background())
    
    go worker(ctx)
    
    time.Sleep(3 * time.Second)
    cancel()  // 发送取消信号
    
    time.Sleep(time.Second)
}
```

### 4.2.2 级联取消

```go
package main

import (
    "context"
    "fmt"
    "time"
)

func main() {
    ctx, cancel := context.WithCancel(context.Background())
    
    // 创建子 context
    ctx1, cancel1 := context.WithCancel(ctx)
    ctx2, cancel2 := context.WithCancel(ctx)
    
    go func() {
        <-ctx1.Done()
        fmt.Println("ctx1 取消")
    }()
    
    go func() {
        <-ctx2.Done()
        fmt.Println("ctx2 取消")
    }()
    
    time.Sleep(time.Second)
    cancel()  // 父 context 取消，子 context 也会取消
    
    time.Sleep(time.Second)
    cancel1()
    cancel2()
}
```

## 4.3 截止时间

### 4.3.1 WithDeadline

```go
package main

import (
    "context"
    "fmt"
    "time"
)

func main() {
    d := time.Now().Add(50 * time.Millisecond)
    ctx, cancel := context.WithDeadline(context.Background(), d)
    defer cancel()
    
    select {
    case <-time.After(1 * time.Second):
        fmt.Println("超时")
    case <-ctx.Done():
        fmt.Println("context 取消:", ctx.Err())
    }
}
```

### 4.3.2 WithTimeout

```go
package main

import (
    "context"
    "fmt"
    "time"
)

func slowOperation(ctx context.Context) error {
    select {
    case <-time.After(2 * time.Second):
        return nil
    case <-ctx.Done():
        return ctx.Err()
    }
}

func main() {
    ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
    defer cancel()
    
    if err := slowOperation(ctx); err != nil {
        fmt.Println("错误:", err)
    }
}
```

## 4.4 值传递

```go
package main

import (
    "context"
    "fmt"
)

type contextKey string

const userIDKey contextKey = "userID"

func main() {
    ctx := context.WithValue(context.Background(), userIDKey, "user-123")
    
    handler(ctx)
}

func handler(ctx context.Context) {
    userID := ctx.Value(userIDKey).(string)
    fmt.Println("User ID:", userID)
}
```

## 4.5 实际案例

### 4.5.1 HTTP 请求超时

```go
package main

import (
    "context"
    "fmt"
    "net/http"
    "time"
)

func main() {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    req, _ := http.NewRequest("GET", "http://example.com", nil)
    req = req.WithContext(ctx)
    
    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        fmt.Println("错误:", err)
        return
    }
    defer resp.Body.Close()
    
    fmt.Println("请求成功")
}
```

### 4.5.2 gRPC 超时

```go
package main

import (
    "context"
    "fmt"
    "time"
    
    "google.golang.org/grpc"
)

func main() {
    conn, _ := grpc.Dial("localhost:50051", grpc.WithInsecure())
    client := NewServiceClient(conn)
    
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    resp, err := client.GetData(ctx, &Request{})
    if err != nil {
        fmt.Println("错误:", err)
        return
    }
    
    fmt.Println("响应:", resp)
}
```

## 4.6 最佳实践

### 4.6.1 不要存储 Context

```go
// 错误示例
type Server struct {
    ctx context.Context  // 不要这样做
}

// 正确示例
type Server struct{}

func (s *Server) Serve(ctx context.Context) {
    // 使用传入的 ctx
}
```

### 4.6.2 第一个参数传递 Context

```go
// 正确：Context 作为第一个参数
func DoSomething(ctx context.Context, arg1, arg2 string) error

// 错误：Context 不是第一个参数
func DoSomething(arg1, arg2 string, ctx context.Context) error
```

## 4.7 本章小结

```
┌─────────────────────────────────────────────────────────────┐
│                      本章总结                                │
├─────────────────────────────────────────────────────────────┤
│ ✓ Context 用于取消信号、截止时间、值传递                  │
│                                                             │
│ ✓ WithCancel: 手动取消                                    │
│ ✓ WithDeadline: 指定截止时间                              │
│ ✓ WithTimeout: 指定超时时间                               │
│ ✓ WithValue: 传递请求作用域的值                          │
│                                                             │
│ ✓ 最佳实践：不要存储 Context，作为第一个参数传递        │
└─────────────────────────────────────────────────────────────┘
```

---

[上一章：← 同步原语详解](./03-同步原语详解.md) | [下一章：并发模式与最佳实践 →](./05-并发模式与最佳实践.md)
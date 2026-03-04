# 3.4 Context

## 基础使用

```go
package main

import (
    "context"
    "fmt"
    "time"
)

func main() {
    // 根 context
    ctx := context.Background()

    // 可取消的 context
    ctx, cancel := context.WithCancel(ctx)
    defer cancel()

    // 带超时的 context
    ctx2, cancel2 := context.WithTimeout(ctx, 5*time.Second)
    defer cancel2()

    // 带截止时间的 context
    ctx3, cancel3 := context.WithDeadline(ctx, time.Now().Add(time.Hour))
    defer cancel3()

    // 带值的 context
    ctx4 := context.WithValue(ctx, "key", "value")
}
```

---

## Context 传递

```go
package main

import (
    "context"
    "fmt"
    "time"
)

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
    ctx := context.Background()
    ctx, cancel := context.WithCancel(ctx)
    defer cancel()

    for i := 0; i < 5; i++ {
        go process(ctx, i)
    }

    time.Sleep(500 * time.Millisecond)
    cancel()  // 取消所有
    time.Sleep(time.Second)
}
```

---

## 值传递

```go
package main

import (
    "context"
    "fmt"
)

type contextKey string

const userIDKey contextKey = "userID"

func WithUserID(ctx context.Context, userID string) context.Context {
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

func main() {
    ctx := context.Background()
    ctx = WithUserID(ctx, "12345")
    handleRequest(ctx)
}
```

---

## 超时控制

```go
package main

import (
    "context"
    "fmt"
    "time"
)

func withTimeoutExample() {
    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
    defer cancel()

    result := make(chan string, 1)
    go func() {
        time.Sleep(3 * time.Second)
        result <- "完成"
    }()

    select {
    case r := <-result:
        fmt.Println(r)
    case <-ctx.Done():
        fmt.Println("超时:", ctx.Err())
    }
}

func main() {
    withTimeoutExample()
}
```

---

## 最佳实践

```go
// 1. Context 作为第一个参数
func DoWork(ctx context.Context, arg string) error {
    return nil
}

// 2. 不要将 Context 存在结构体中
// 错误
type BadWorker struct {
    ctx context.Context  // 不要这样
}

// 正确
type GoodWorker struct{}
func (w *GoodWorker) Run(ctx context.Context) {}

// 3. 不要传递 nil context
// 使用 context.Background()

// 4. 取消函数必须由调用者调用
func caller() {
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()  // 确保调用
    DoWork(ctx, "arg")
}

// 5. 使用 context 传递请求作用域的值
// 6. 每个请求使用独立的 context
```

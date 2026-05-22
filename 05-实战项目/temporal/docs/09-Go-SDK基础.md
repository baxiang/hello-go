# Go SDK 基础

本文档介绍 Temporal Go SDK 的安装、客户端连接和基本配置。

## 9.1 安装与依赖

### 安装 SDK

```bash
go get go.temporal.io/sdk@latest
```

### 项目初始化

```bash
mkdir my-temporal-app && cd my-temporal-app
go mod init my-temporal-app
go get go.temporal.io/sdk
```

### 核心包

| 包 | 用途 |
|---|-----|
| `go.temporal.io/sdk/client` | 客户端连接 |
| `go.temporal.io/sdk/worker` | Worker 管理 |
| `go.temporal.io/sdk/workflow` | Workflow 定义 |
| `go.temporal.io/sdk/activity` | Activity 定义 |
| `go.temporal.io/sdk/temporal` | 错误和选项 |

---

## 9.2 客户端连接

### 基本连接

```go
package main

import (
    "log"
    
    "go.temporal.io/sdk/client"
)

func main() {
    // 创建客户端
    c, err := client.Dial(client.Options{
        HostPort: "localhost:7233",
    })
    if err != nil {
        log.Fatalln("无法创建客户端", err)
    }
    defer c.Close()
    
    log.Println("已连接 Temporal 服务器")
}
```

### 连接选项

```go
c, err := client.Dial(client.Options{
    // 服务器地址
    HostPort: "localhost:7233",
    
    // 命名空间
    Namespace: "default",
    
    // 日志配置
    Logger: log.NewStdoutLogger(),
    
    // 连接超时
    ConnectionOptions: client.ConnectionOptions{
        DialTimeout: 10 * time.Second,
    },
})
```

### TLS 配置

```go
import "crypto/tls"

c, err := client.Dial(client.Options{
    HostPort: "temporal.example.com:7233",
    Namespace: "production",
    ConnectionOptions: client.ConnectionOptions{
        TLS: &tls.Config{
            InsecureSkipVerify: false,
        },
    },
})
```

---

## 9.3 基本工作流结构

一个完整的 Temporal 应用包含以下组件：

```
my-temporal-app/
├── workflow.go        # Workflow 定义
├── activities.go      # Activity 定义
├── worker/
│   └── main.go       # Worker 启动
└── starter/
    └── main.go       # Workflow 启动
```

### 最简示例

**workflow.go**
```go
package app

import (
    "time"
    "go.temporal.io/sdk/workflow"
)

// GreetingWorkflow 简单的问候工作流
func GreetingWorkflow(ctx workflow.Context, name string) (string, error) {
    ao := workflow.ActivityOptions{
        StartToCloseTimeout: 10 * time.Second,
    }
    ctx = workflow.WithActivityOptions(ctx, ao)
    
    var result string
    err := workflow.ExecuteActivity(ctx, SayHello, name).Get(ctx, &result)
    return result, err
}
```

**activities.go**
```go
package app

import (
    "context"
    "fmt"
)

// SayHello 活动：生成问候语
func SayHello(ctx context.Context, name string) (string, error) {
    return fmt.Sprintf("Hello, %s!", name), nil
}
```

**worker/main.go**
```go
package main

import (
    "log"
    
    "go.temporal.io/sdk/client"
    "go.temporal.io/sdk/worker"
    "my-temporal-app/app"
)

func main() {
    c, err := client.Dial(client.Options{})
    if err != nil {
        log.Fatalln("无法创建客户端", err)
    }
    defer c.Close()
    
    w := worker.New(c, "greeting-task-queue", worker.Options{})
    w.RegisterWorkflow(app.GreetingWorkflow)
    w.RegisterActivity(app.SayHello)
    
    err = w.Run(worker.InterruptCh())
    if err != nil {
        log.Fatalln("Worker 启动失败", err)
    }
}
```

**starter/main.go**
```go
package main

import (
    "context"
    "log"
    
    "go.temporal.io/sdk/client"
    "my-temporal-app/app"
)

func main() {
    c, err := client.Dial(client.Options{})
    if err != nil {
        log.Fatalln("无法创建客户端", err)
    }
    defer c.Close()
    
    we, err := c.ExecuteWorkflow(context.Background(),
        client.StartWorkflowOptions{
            TaskQueue: "greeting-task-queue",
        },
        app.GreetingWorkflow,
        "World",
    )
    if err != nil {
        log.Fatalln("启动 Workflow 失败", err)
    }
    
    var result string
    err = we.Get(context.Background(), &result)
    if err != nil {
        log.Fatalln("获取结果失败", err)
    }
    
    log.Printf("结果: %s", result)
}
```

---

## 9.4 运行第一个应用

### 启动 Temporal 服务

```bash
# 使用 Temporal CLI
temporal server start-dev

# 或使用 Docker
docker run -d --name temporal -p 7233:7233 -p 8233:8233 temporalio/server:latest
```

### 运行 Worker

```bash
go run ./worker/main.go
```

### 启动 Workflow

```bash
go run ./starter/main.go
```

### 查看结果

访问 Web UI: http://localhost:8233

---

## 9.5 常用客户端方法

### 执行 Workflow

```go
// 异步执行
we, err := c.ExecuteWorkflow(ctx, options, MyWorkflow, input)

// 同步等待结果
var result MyResult
err = we.Get(ctx, &result)

// 获取 Workflow ID
workflowID := we.GetID()
```

### 发送 Signal

```go
err := c.SignalWorkflow(ctx, workflowID, runID, "my-signal", signalValue)
```

### 查询 Workflow

```go
resp, err := c.QueryWorkflow(ctx, workflowID, runID, "my-query")
var result MyResult
resp.Get(&result)
```

### 取消 Workflow

```go
err := c.CancelWorkflow(ctx, workflowID, runID)
```

### 获取 Workflow 状态

```go
resp, err := c.DescribeWorkflowExecution(ctx, workflowID, runID)
```

---

## 下一步

- [10-Go-SDK-Workflow开发](./10-Go-SDK-Workflow开发.md) - 深入学习 Workflow 定义
- [11-Go-SDK-Activity开发](./11-Go-SDK-Activity开发.md) - 学习 Activity 开发
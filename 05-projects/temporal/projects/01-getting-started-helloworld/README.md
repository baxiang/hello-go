# 01-getting-started-helloworld

一个最简单的 Temporal 工作流项目，帮助您快速上手 Temporal 开发。

## 项目简介

本项目演示 Temporal 工作流的基本概念：
- 定义一个简单的 Workflow
- 实现一个 Activity
- 启动 Worker 注册工作流
- 通过客户端启动工作流执行

## 项目结构

```
01-getting-started-helloworld/
├── README.md           # 项目说明
├── workflow.go         # Workflow 定义
├── activities.go       # Activity 实现
├── worker/
│   └── main.go         # Worker 启动程序
└── starter/
    └── main.go         # 工作流启动程序
```

## 代码实现

### workflow.go - 工作流定义

```go
package helloworld

import (
    "time"
    
    "go.temporal.io/sdk/workflow"
)

// HelloWorldWorkflow 是一个简单的问候工作流
func HelloWorldWorkflow(ctx workflow.Context, name string) (string, error) {
    // 设置 Activity 选项
    ao := workflow.ActivityOptions{
        StartToCloseTimeout: 10 * time.Second,
    }
    ctx = workflow.WithActivityOptions(ctx, ao)
    
    // 调用 Activity
    var result string
    err := workflow.ExecuteActivity(ctx, SayHello, name).Get(ctx, &result)
    if err != nil {
        return "", err
    }
    
    return result, nil
}
```

### activities.go - Activity 实现

```go
package helloworld

import (
    "fmt"
)

// SayHello 是一个简单的问候 Activity
func SayHello(name string) (string, error) {
    greeting := fmt.Sprintf("Hello, %s!", name)
    return greeting, nil
}
```

### worker/main.go - Worker 启动

```go
package main

import (
    "log"
    
    "go.temporal.io/sdk/client"
    "go.temporal.io/sdk/worker"
    
    "github.com/baxiang/hello-go/08-projects/temporal/projects/01-getting-started-helloworld"
)

func main() {
    // 创建 Temporal 客户端
    c, err := client.Dial(client.Options{
        HostPort: "localhost:7233",
    })
    if err != nil {
        log.Fatalln("Unable to create Temporal client", err)
    }
    defer c.Close()
    
    // 创建 Worker
    w := worker.New(c, "hello-world-task-queue", worker.Options{})
    
    // 注册 Workflow 和 Activity
    w.RegisterWorkflow(helloworld.HelloWorldWorkflow)
    w.RegisterActivity(helloworld.SayHello)
    
    // 启动 Worker
    err = w.Run(worker.InterruptCh())
    if err != nil {
        log.Fatalln("Unable to start worker", err)
    }
}
```

### starter/main.go - 工作流启动器

```go
package main

import (
    "context"
    "log"
    
    "go.temporal.io/sdk/client"
    
    "github.com/baxiang/hello-go/08-projects/temporal/projects/01-getting-started-helloworld"
)

func main() {
    // 创建 Temporal 客户端
    c, err := client.Dial(client.Options{
        HostPort: "localhost:7233",
    })
    if err != nil {
        log.Fatalln("Unable to create Temporal client", err)
    }
    defer c.Close()
    
    // 启动工作流
    workflowOptions := client.StartWorkflowOptions{
        ID:        "hello-world-workflow-id",
        TaskQueue: "hello-world-task-queue",
    }
    
    we, err := c.ExecuteWorkflow(context.Background(), workflowOptions, helloworld.HelloWorldWorkflow, "Temporal")
    if err != nil {
        log.Fatalln("Unable to execute workflow", err)
    }
    
    log.Printf("Started workflow: WorkflowID=%s, RunID=%s", we.GetID(), we.GetRunID())
    
    // 获取工作流结果
    var result string
    err = we.Get(context.Background(), &result)
    if err != nil {
        log.Fatalln("Unable to get workflow result", err)
    }
    
    log.Printf("Workflow result: %s", result)
}
```

## 运行说明

### 前置条件

1. 安装 Temporal 服务器（推荐使用 Docker）:
   ```bash
   docker run -d --name temporal -p 7233:7233 temporalio/auto-setup:latest
   ```

2. 访问 Temporal Web UI:
   ```bash
   docker run -d --name temporal-ui -p 8080:8080 --link temporal temporalio/ui:latest
   ```
   打开浏览器访问 http://localhost:8080

### 运行步骤

1. 初始化模块（在项目根目录）:
   ```bash
   go mod init github.com/baxiang/hello-go/08-projects/temporal/projects/01-getting-started-helloworld
   go mod tidy
   ```

2. 启动 Worker（在一个终端）:
   ```bash
   go run worker/main.go
   ```

3. 启动工作流（在另一个终端）:
   ```bash
   go run starter/main.go
   ```

### 预期输出

starter 终端输出:
```
2024/01/01 10:00:00 Started workflow: WorkflowID=hello-world-workflow-id, RunID=...
2024/01/01 10:00:01 Workflow result: Hello, Temporal!
```

## 核心概念

### Workflow（工作流）

Workflow 是业务流程的编排逻辑，具有以下特点：
- **确定性**: 工作流必须是确定性的，不能使用随机数、当前时间等
- **持久化**: 工作流状态自动持久化，支持长时间运行
- **可恢复**: 工作流失败后可以从断点恢复执行

### Activity（活动）

Activity 是具体的业务操作，具有以下特点：
- **非确定性**: 可以调用外部服务、数据库操作等
- **可重试**: 支持配置重试策略
- **超时控制**: 支持多种超时配置

### Worker（工作器）

Worker 负责执行工作流和活动：
- 监听 Task Queue 获取任务
- 执行注册的 Workflow 和 Activity
- 向 Temporal 服务器报告执行状态

### Task Queue（任务队列）

Task Queue 是工作流任务和活动任务的分发机制：
- 解耦工作流启动和执行
- 支持多个 Worker 并行处理
- 支持任务路由和负载均衡

## 下一步

完成本项目后，建议继续学习：
- [02-getting-started-order-processing](../02-getting-started-order-processing/) - 实际业务场景的工作流
- [hands-on/01-workflow-basics](../../hands-on/01-workflow-basics/) - 工作流基础练习
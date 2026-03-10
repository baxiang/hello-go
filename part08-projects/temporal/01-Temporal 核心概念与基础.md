# Temporal 核心概念与基础

Temporal 是一个强大的分布式工作流编排平台，用于构建可靠的应用程序。它通过将业务逻辑与基础设施分离，使开发者能够专注于编写代码，而无需担心分布式系统的复杂性。

## 1.1 Temporal 简介

### 什么是 Temporal？

Temporal 是一个开源的工作流编排引擎，用于构建持久化、可扩展的应用程序。它源自 Uber 的 Cadence 项目，是一个用于构建可靠分布式系统的编程模型。

### 核心特性

| 特性 | 说明 |
|------|------|
| **持久化状态** | 工作流状态自动持久化，即使系统故障也不会丢失 |
| **可重试执行** | 活动失败时自动重试，无需手动编写重试逻辑 |
| **时间旅行** | 支持回溯和重新执行工作流 |
| **活动取消** | 支持优雅取消正在执行的工作流 |
| **可见性** | 提供完整的执行历史和调试工具 |
| **水平扩展** | 支持高并发和大规模工作流执行 |

### Temporal 与其他技术对比

| 特性 | Temporal | AWS Step Functions | Camunda | Airflow |
|------|----------|---------------------|---------|----------|
| 编程模型 | 代码优先 | 声明式 | 声明式 | DAG |
| 持久化执行 | 原生支持 | AWS 托管 | 需要配置 | 需要配置 |
| 活动重试 | 自动 | 有限 | 有限 | 有限 |
| 长时间运行 | 支持 | 有超时限制 | 支持 | 有超时限制 |
| 分布式事务 | Saga 支持 | 有限 | Saga 支持 | 不支持 |
| 多语言支持 | 丰富 | AWS SDK | 丰富 | 丰富 |

### 适用场景

- **微服务编排**：协调多个微服务之间的交互
- **业务流程自动化**：订单处理、审批流程等
- **数据处理管道**：ETL、批处理等
- **事件驱动架构**：响应事件并执行复杂逻辑
- **长时间运行流程**：人工审批、等待外部回调

---

## 1.2 Temporal 架构

### 系统架构

```
┌─────────────────────────────────────────────────────────────────────────┐
│                              Temporal Cluster                           │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐   │
│  │  Frontend   │  │  History    │  │  Matching   │  │  Worker     │   │
│  │   Service   │  │   Service   │  │   Service   │  │  Service    │   │
│  └──────┬──────┘  └──────┬──────┘  └──────┬──────┘  └──────┬──────┘   │
│         │                │                 │                 │          │
│         └────────────────┴────────┬────────┴────────────────┘          │
│                                  │                                     │
│                         ┌────────▼────────┐                           │
│                         │   Persistence    │                           │
│                         │ (MySQL/Postgres)│                           │
│                         └─────────────────┘                           │
└─────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                           Worker Processes                              │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │  Worker Node                                                    │   │
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐           │   │
│  │  │  Workflow   │  │   Activity  │  │  Activity   │           │   │
│  │  │  Worker 1   │  │   Worker 1  │  │   Worker 2  │           │   │
│  │  └─────────────┘  └─────────────┘  └─────────────┘           │   │
│  └─────────────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                          Client Applications                             │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐                    │
│  │   Starter   │  │   Starter   │  │   Starter   │                    │
│  │  (Go/Java)  │  │  (Go/Java)  │  │  (Go/Java)  │                    │
│  └─────────────┘  └─────────────┘  └─────────────┘                    │
└─────────────────────────────────────────────────────────────────────────┘
```

### 核心组件

#### Frontend Service（前端服务）
- 接收客户端请求
- 限流和认证
- 任务路由

#### History Service（历史服务）
- 管理工作流执行状态
- 维护执行历史
- 事件存储

#### Matching Service（匹配服务）
- 任务队列管理
- 将任务分发给 Worker
- 负责活动执行调度

#### Worker Service（工作服务）
- 执行工作流和活动代码
- 从任务队列拉取任务
- 与 Temporal Server 通信

### 数据存储

Temporal 使用持久化存储来保存：
- 工作流执行状态
- 事件历史
- 任务队列
- 命名空间配置

支持的存储后端：
- **MySQL**：生产环境推荐
- **PostgreSQL**：生产环境推荐
- **SQLite**：开发环境
- **Cassandra**：大规模部署

---

## 1.3 安装与配置

### 安装 Temporal Server

#### 使用 Docker Compose（开发环境）

```bash
# 创建 docker-compose.yml
version: '3.8'

services:
  temporal:
    image: temporalio/auto-setup:1.22.0
    ports:
      - "7233:7233"
    environment:
      - DB=postgresql
      - DB_PORT=5432
      - POSTGRES_USER=temporal
      - POSTGRES_PWD=temporal
      - POSTGRES_SEEDS=temporal-postgresql
    volumes:
      - temporal-data:/var/lib/temporal

  temporal-postgresql:
    image: postgres:13
    environment:
      POSTGRES_USER: temporal
      POSTGRES_PASSWORD: temporal
    volumes:
      - postgres-data:/var/lib/postgresql/data

volumes:
  temporal-data:
  postgres-data:
```

```bash
# 启动服务
docker-compose up -d

# 验证服务
docker-compose ps
```

#### 使用 Helm（Kubernetes）

```bash
# 添加 Temporal Helm 仓库
helm repo add temporalio https://helm.temporal.io

# 安装 Temporal
helm install temporal temporalio/temporal \
  --namespace temporal \
  --create-namespace \
  --set server.replicaCount=1
```

### 安装 Temporal CLI

```bash
# macOS
brew install temporal

# Linux
curl -sSL https://github.com/temporalio/cli/releases/latest/download/temporal_linux_amd64.tar.gz | tar -xz
sudo mv temporal /usr/local/bin/

# Windows
choco install temporal
```

### CLI 基本命令

```bash
# 启动 Temporal 服务
temporal server start-dev

# 查看命名空间
temporal namespace list

# 创建命名空间
temporal namespace create my-namespace

# 查看工作流执行
temporal workflow list

# 显示工作流历史
temporal workflow show --workflow-id <id>
```

---

## 1.4 核心概念

### 1.4.1 工作流（Workflow）

工作流是 Temporal 中的核心概念，代表一个持久的业务逻辑流程。

```go
// 定义工作流
func OrderProcessingWorkflow(ctx workflow.Context, orderID string) error {
    // 工作流逻辑
    return nil
}
```

**工作流特点：**
- 确定性执行：相同输入产生相同输出
- 持久化状态：自动保存执行状态
- 可重试：失败后自动重试
- 长时间运行：可持续运行数天甚至数月

### 1.4.2 活动（Activity）

活动是工作流中执行的单个原子操作。

```go
// 定义活动
func ProcessPayment(ctx context.Context, payment Payment) error {
    // 支付处理逻辑
    return nil
}
```

**活动特点：**
- 可重试：失败时自动重试
- 有超时：可设置执行超时
- 有心跳：支持长时间运行活动的心跳检测
- 幂等性：支持幂等执行

### 1.4.3 工作流执行

```
┌─────────────────────────────────────────────────────────────────────────┐
│                        Workflow Execution                                │
│                                                                          │
│  ┌─────────┐    ┌─────────┐    ┌─────────┐    ┌─────────┐              │
│  │ Start   │───▶│Activity1│───▶│Activity2│───▶│ Complete│              │
│  └─────────┘    └─────────┘    └─────────┘    └─────────┘              │
│       │              │              │              │                      │
│       ▼              ▼              ▼              ▼                      │
│  ┌─────────────────────────────────────────────────────────┐            │
│  │                    Event History                         │            │
│  │  WorkflowStarted                                         │            │
│  │  Activity1Scheduled  Activity1Completed                  │            │
│  │  Activity2Scheduled  Activity2Completed                  │            │
│  │  WorkflowCompleted                                       │            │
│  └─────────────────────────────────────────────────────────┘            │
└─────────────────────────────────────────────────────────────────────────┘
```

### 1.4.4 任务队列（Task Queue）

任务队列是 Worker 从 Temporal Server 获取任务的机制。

```go
// 启动 Worker 并指定任务队列
worker := worker.New(client, "my-task-queue", worker.Options{})
worker.RegisterWorkflow(MyWorkflow)
worker.RegisterActivity(MyActivity)
worker.Start()
```

### 1.4.5 命名空间（Namespace）

命名空间用于隔离工作流执行。

```bash
# 创建命名空间
temporal namespace create payments

# 查看命名空间
temporal namespace describe payments
```

---

## 1.5 工作流类型

### 1.5.1 同步工作流

```go
func SyncWorkflow(ctx workflow.Context, input string) (string, error) {
    // 同步调用活动
    var result string
    err := workflow.ExecuteActivity(ctx, MyActivity, input).Get(ctx, &result)
    if err != nil {
        return "", err
    }
    return result, nil
}
```

### 1.5.2 异步工作流

```go
func AsyncWorkflow(ctx workflow.Context, input string) error {
    // 异步启动活动
    future := workflow.ExecuteActivity(ctx, LongRunningActivity, input)
    
    // 等待活动完成
    // 或者继续执行其他逻辑
    
    return future.Get(ctx, nil)
}
```

### 1.5.3 定时工作流

```go
func CronWorkflow(ctx workflow.Context) error {
    // 每天凌晨执行的任务
    logger := workflow.GetLogger(ctx)
    logger.Info("Cron job started")
    
    // 执行任务
    return workflow.ExecuteActivity(ctx, DailyTask).Get(ctx, nil)
}

// 启动时配置 Cron
workflowOptions := client.StartWorkflowOptions{
    ID:        "cron-workflow",
    CronSchedule: "0 0 * * *",  // 每天午夜
}
```

### 1.5.4 持续工作流

```go
func ContinuousWorkflow(ctx workflow.Context) error {
    logger := workflow.GetLogger(ctx)
    
    for {
        // 等待下一个事件或定时触发
        selector := workflow.NewSelector(ctx)
        
        // 定时检查
        timer := workflow.NewTimer(ctx, time.Hour)
        selector.AddFuture(timer, func(f workflow.Future) {
            f.Get(ctx, nil)
            // 执行定期任务
            workflow.ExecuteActivity(ctx, CheckStatus).Get(ctx, nil)
        })
        
        selector.Select(ctx)
    }
}
```

---

## 1.6 活动类型

### 1.6.1 本地活动

本地活动在同一个 Worker 进程中执行，减少网络开销。

```go
func LocalActivityWorkflow(ctx workflow.Context) error {
    // 使用本地活动
    ao := workflow.LocalActivityOptions{
        StartToCloseTimeout: 10 * time.Second,
    }
    ctx = workflow.WithLocalActivityOptions(ctx, ao)
    
    return workflow.ExecuteActivity(ctx, QuickTask).Get(ctx, nil)
}
```

### 1.6.2 长时间运行活动

支持心跳检测的活动，用于需要长时间运行的任务。

```go
func LongRunningActivity(ctx context.Context, taskID string) error {
    logger := activity.GetLogger(ctx)
    
    progress := 0
    for progress < 100 {
        // 报告进度
        activity.RecordHeartbeat(ctx, progress)
        
        // 模拟耗时操作
        time.Sleep(5 * time.Second)
        progress += 10
    }
    
    logger.Info("Task completed", "taskID", taskID)
    return nil
}
```

### 1.6.3 幂等活动

```go
func IdempotentActivity(ctx context.Context, request Request) error {
    logger := activity.GetLogger(ctx)
    
    // 检查是否已处理
    if activity.HasHeartbeatDetails(ctx) {
        var completed bool
        activity.GetHeartbeatDetails(ctx, &completed)
        if completed {
            logger.Info("Already processed", "requestID", request.ID)
            return nil  // 幂等返回
        }
    }
    
    // 处理业务逻辑
    processRequest(request)
    
    // 记录完成状态
    activity.RecordHeartbeat(ctx, true)
    
    return nil
}
```

---

## 1.7 工作流选项

### 1.7.1 启动选项

```go
// 基本启动选项
options := client.StartWorkflowOptions{
    ID:        "workflow-id",
    TaskQueue: "my-task-queue",
}

// 带超时
options := client.StartWorkflowOptions{
    ID:              "workflow-id",
    TaskQueue:       "my-task-queue",
    StartToCloseTimeout: 10 * time.Minute,
}

// 带重试策略
options := client.StartWorkflowOptions{
    ID:              "workflow-id",
    TaskQueue:       "my-task-queue",
    RetryPolicy: &temporal.RetryPolicy{
        InitialInterval:    time.Second,
        BackoffCoefficient: 2.0,
        MaximumInterval:    time.Minute,
        MaximumAttempts:    5,
    },
}

// 带 Cron
options := client.StartWorkflowOptions{
    ID:            "cron-workflow",
    TaskQueue:     "my-task-queue",
    CronSchedule: "0 0 * * *",
}
```

### 1.7.2 执行选项

```go
// 活动执行选项
activityOptions := workflow.ActivityOptions{
    StartToCloseTimeout: 5 * time.Minute,
    ScheduleToStartTimeout: 1 * time.Minute,
    ScheduleToCloseTimeout: 6 * time.Minute,
    HeartbeatTimeout: 30 * time.Second,
    RetryPolicy: &temporal.RetryPolicy{
        InitialInterval:    time.Second,
        BackoffCoefficient: 2.0,
        MaximumAttempts:    3,
    },
}
ctx = workflow.WithActivityOptions(ctx, activityOptions)
```

### 1.7.3 子工作流选项

```go
// 子工作流选项
childWorkflowOptions := workflow.ChildWorkflowOptions{
    WorkflowID: "child-workflow-id",
    ParentClosePolicy: temporal.ParentClosePolicyRequestCancel,
}
ctx = workflow.WithChildOptions(ctx, childWorkflowOptions)
```

---

## 1.8 错误处理

### 1.8.1 工作流中的错误处理

```go
func OrderWorkflow(ctx workflow.Context, orderID string) error {
    logger := workflow.GetLogger(ctx)
    
    // 使用 Select 处理多个活动
    selector := workflow.NewSelector(ctx)
    
    // 支付活动
    paymentFuture := workflow.ExecuteActivity(ctx, ProcessPayment, orderID)
    selector.AddFuture(paymentFuture, func(f workflow.Future) {
        var err error
        f.Get(ctx, &err)
        if err != nil {
            logger.Error("Payment failed", "error", err)
        }
    })
    
    // 库存活动
    inventoryFuture := workflow.ExecuteActivity(ctx, ReserveInventory, orderID)
    selector.AddFuture(inventoryFuture, func(f workflow.Future) {
        var err error
        f.Get(ctx, &err)
        if err != nil {
            logger.Error("Inventory failed", "error", err)
        }
    })
    
    selector.Select(ctx)
    
    return nil
}
```

### 1.8.2 活动中的错误处理

```go
func ProcessPaymentActivity(ctx context.Context, payment Payment) error {
    logger := activity.GetLogger(ctx)
    
    // 业务逻辑
    err := processPayment(payment)
    if err != nil {
        // 区分可重试和不可重试错误
        if isRetryable(err) {
            return err  // 可重试，Temporal 会自动重试
        }
        // 不可重试，返回非重试错误
        return temporal.NewNonRetryableApplicationError(
            err.Error(),
            "payment.failed",
            nil,
        )
    }
    
    return nil
}
```

### 1.8.3 重试策略

```go
// 全局重试策略
retryPolicy := &temporal.RetryPolicy{
    InitialInterval:    time.Second,        // 初始间隔
    BackoffCoefficient: 2.0,                // 退避系数
    MaximumInterval:    time.Minute,        // 最大间隔
    MaximumAttempts:    5,                  // 最大重试次数
    NonRetryableErrorTypes: []string{       // 不重试的错误类型
        "InvalidInput",
        "AuthenticationFailed",
    },
}

// 活动级别重试
activityOptions := workflow.ActivityOptions{
    RetryPolicy: retryPolicy,
}
```

---

## 1.9 最佳实践

### 1.9.1 工作流设计原则

1. **保持工作流确定性**
   - 不使用随机数
   - 不使用当前时间（使用 Now()）
   - 不依赖外部状态

2. **避免长时间阻塞**
   - 使用定时器而非 sleep
   - 使用信号而非轮询

3. **幂等性设计**
   - 活动应支持幂等执行
   - 使用唯一标识符防止重复处理

### 1.9.2 活动设计原则

1. **保持活动简单**
   - 每个活动只做一件事
   - 避免复杂业务逻辑

2. **正确设置超时**
   - StartToCloseTimeout：活动执行时间
   - ScheduleToStartTimeout：活动等待时间
   - HeartbeatTimeout：心跳间隔

3. **实现心跳**
   - 长时间运行的活动必须实现心跳
   - 定期报告进度

### 1.9.3 错误处理原则

1. **区分可重试错误**
   - 网络错误：可重试
   - 业务错误：通常不可重试

2. **使用父工作流策略**
   - RequestCancel：请求取消子工作流
   - Terminate：终止子工作流
   - Abandon：忽略子工作流

---

## 1.10 相关资源

- [Temporal 官方文档](https://docs.temporal.io/)
- [Temporal Go SDK](https://github.com/temporalio/sdk-go)
- [Temporal 示例](https://github.com/temporalio/samples-go)
- [Temporal 社区](https://community.temporal.io/)
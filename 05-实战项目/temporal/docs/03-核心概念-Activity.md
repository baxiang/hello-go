# 核心概念 - Activity（活动）

活动（Activity）是工作流中执行的单个原子操作。与工作流不同，活动可以包含任何非确定性代码，如网络调用、数据库操作、文件 I/O 等。本文档详细介绍 Temporal 活动的核心概念和最佳实践。

---

## Activity Definition（活动定义）

活动定义是一个普通函数，执行具体的业务逻辑。活动可以使用标准 Go 代码，不受确定性约束的限制。

### 基本定义

```go
package activity

import (
    "context"
    "fmt"
)

// 简单的活动定义
func SayHello(ctx context.Context, name string) (string, error) {
    return fmt.Sprintf("Hello, %s!", name), nil
}

// 无返回值的活动
func SendEmail(ctx context.Context, to, subject, body string) error {
    // 发送邮件逻辑
    return nil
}

// 多参数的活动
func ProcessPayment(ctx context.Context, orderID string, amount float64, currency string) error {
    // 处理支付逻辑
    return nil
}
```

### 活动参数和返回值

```go
// 基本类型参数
func BasicActivity(ctx context.Context, id string, count int, enabled bool) error {
    return nil
}

// 结构体参数
type PaymentRequest struct {
    OrderID  string
    Amount   float64
    Currency string
    Metadata map[string]string
}

type PaymentResult struct {
    TransactionID string
    Status        string
    ProcessedAt   time.Time
}

func ProcessPaymentActivity(ctx context.Context, req PaymentRequest) (PaymentResult, error) {
    // 处理支付
    return PaymentResult{
        TransactionID: "txn-123",
        Status:        "completed",
        ProcessedAt:   time.Now(),
    }, nil
}

// 多返回值
func DivideActivity(ctx context.Context, a, b int) (quotient, remainder int, err error) {
    if b == 0 {
        return 0, 0, fmt.Errorf("division by zero")
    }
    return a / b, a % b, nil
}
```

### 活动注册

```go
package main

import (
    "go.temporal.io/sdk/client"
    "go.temporal.io/sdk/worker"
    "myapp/activity"
)

func main() {
    c, _ := client.Dial(client.Options{})
    defer c.Close()

    w := worker.New(c, "my-task-queue", worker.Options{})
    
    // 方式一：注册活动函数
    w.RegisterActivity(activity.SayHello)
    w.RegisterActivity(activity.ProcessPayment)
    
    // 方式二：注册活动结构体方法
    w.RegisterActivity(&PaymentActivity{})
    
    w.Run(worker.InterruptCh())
}
```

### 结构体活动

```go
// 使用结构体组织相关活动
type PaymentActivity struct {
    PaymentGateway PaymentGateway
    Logger         *zap.Logger
}

// 结构体方法作为活动
func (a *PaymentActivity) ProcessPayment(ctx context.Context, req PaymentRequest) (PaymentResult, error) {
    a.Logger.Info("Processing payment", "orderID", req.OrderID)
    return a.PaymentGateway.Process(req)
}

func (a *PaymentActivity) RefundPayment(ctx context.Context, transactionID string) error {
    a.Logger.Info("Refunding payment", "transactionID", transactionID)
    return a.PaymentGateway.Refund(transactionID)
}

// 注册结构体活动
func main() {
    w := worker.New(c, "task-queue", worker.Options{})
    
    paymentActivity := &PaymentActivity{
        PaymentGateway: NewPaymentGateway(),
        Logger:         zap.NewExample(),
    }
    
    w.RegisterActivity(paymentActivity)
}
```

---

## Activity Execution（活动执行）

活动由工作流调用执行，执行结果会被持久化到事件历史中。

### 从工作流调用活动

```go
func OrderWorkflow(ctx workflow.Context, orderID string) error {
    // 配置活动选项
    ao := workflow.ActivityOptions{
        StartToCloseTimeout: 5 * time.Minute,
    }
    ctx = workflow.WithActivityOptions(ctx, ao)
    
    // 调用活动
    var result string
    err := workflow.ExecuteActivity(ctx, ProcessOrderActivity, orderID).Get(ctx, &result)
    if err != nil {
        return err
    }
    
    return nil
}
```

### 活动执行流程

```
┌─────────────────────────────────────────────────────────────────────────┐
│                        Activity Execution Flow                           │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│  Workflow Worker                    Temporal Server                     │
│  ┌─────────────┐                   ┌─────────────┐                      │
│  │ Workflow    │                   │  Matching   │                      │
│  │ Execution   │                   │  Service    │                      │
│  └──────┬──────┘                   └──────┬──────┘                      │
│         │                                 │                             │
│         │ 1. Schedule Activity             │                             │
│         │ ──────────────────────────────▶ │                             │
│         │                                 │                             │
│         │                                 │ 2. Add to Task Queue        │
│         │                                 │                             │
│         │                                 ▼                             │
│         │                          ┌─────────────┐                      │
│         │                          │ Task Queue  │                      │
│         │                          └─────────────┘                      │
│         │                                 │                             │
│         │                                 │ 3. Dispatch to Worker       │
│         │                                 ▼                             │
│         │                          ┌─────────────┐                      │
│         │                          │  Activity   │                      │
│         │                          │   Worker    │                      │
│         │                          └──────┬──────┘                      │
│         │                                 │                             │
│         │                                 │ 4. Execute Activity        │
│         │                                 │                             │
│         │ 5. Report Result               │                             │
│         │ ◀────────────────────────────── │                             │
│         │                                 │                             │
│         ▼                                 ▼                             │
│  ┌─────────────────────────────────────────────────────────┐            │
│  │                    Event History                         │            │
│  │  - ActivityTaskScheduled                                │            │
│  │  - ActivityTaskStarted                                  │            │
│  │  - ActivityTaskCompleted                                │            │
│  └─────────────────────────────────────────────────────────┘            │
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
```

### 执行选项详解

```go
ao := workflow.ActivityOptions{
    // 超时设置
    ScheduleToCloseTimeout: 10 * time.Minute,  // 从调度到完成的总超时
    ScheduleToStartTimeout:  1 * time.Minute,   // 从调度到开始执行的超时
    StartToCloseTimeout:     5 * time.Minute,  // 从开始执行到完成的超时
    HeartbeatTimeout:        30 * time.Second, // 心跳超时
    
    // 重试策略
    RetryPolicy: &temporal.RetryPolicy{
        InitialInterval:    time.Second,        // 初始重试间隔
        BackoffCoefficient: 2.0,                // 退避系数
        MaximumInterval:    time.Minute,       // 最大重试间隔
        MaximumAttempts:    5,                 // 最大重试次数
        NonRetryableErrorTypes: []string{      // 不可重试的错误类型
            "InvalidInput",
            "AuthenticationFailed",
        },
    },
    
    // 其他选项
    TaskQueue:                "custom-task-queue",  // 指定任务队列
    ActivityID:               "unique-activity-id", // 活动唯一ID
    WaitForCancellation:      true,                  // 取消时等待活动完成
    OriginalTaskQueueName:   "original-queue",     // 原始任务队列
    
    // 心跳超时配置
    HeartbeatTimeout: 30 * time.Second,
}

ctx = workflow.WithActivityOptions(ctx, ao)
```

### 超时类型对比

```
┌─────────────────────────────────────────────────────────────────────────┐
│                        Activity Timeout Types                            │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│  ScheduleToStartTimeout    ScheduleToCloseTimeout                       │
│  ┌────────────────┐        ┌──────────────────────────────────────┐     │
│  │                │        │                                      │     │
│  │   Schedule     │        │   Schedule        Complete           │     │
│  │      │         │        │      │                               │     │
│  │      ▼         │        │      ▼                               │     │
│  │   [Queue]──────│───────▶│   [Queue]──▶[Execute]──▶[Done]     │     │
│  │                │        │                                      │     │
│  └────────────────┘        └──────────────────────────────────────┘     │
│                                                                          │
│  StartToCloseTimeout                                                     │
│  ┌────────────────────────────────┐                                     │
│  │                                │                                     │
│  │   Start              Close     │                                     │
│  │     │                  │       │                                     │
│  │     ▼                  ▼       │                                     │
│  │   [Execute]──────▶[Done]      │                                     │
│  │                                │                                     │
│  └────────────────────────────────┘                                     │
│                                                                          │
│  HeartbeatTimeout                                                        │
│  ┌────────────────────────────────────────────────────────┐             │
│  │  [Start]──▶[Heartbeat]──▶[Heartbeat]──▶...──▶[Done]   │             │
│  │              │              │                          │             │
│  │              └──必须在此时间内发送下一个心跳──┘           │             │
│  └────────────────────────────────────────────────────────┘             │
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
```

### 异步执行

```go
func AsyncWorkflow(ctx workflow.Context) error {
    ao := workflow.ActivityOptions{
        StartToCloseTimeout: time.Minute,
    }
    ctx = workflow.WithActivityOptions(ctx, ao)
    
    // 异步启动活动，不等待完成
    future := workflow.ExecuteActivity(ctx, LongRunningActivity, "input")
    
    // 可以继续执行其他逻辑
    workflow.GetLogger(ctx).Info("Activity started")
    
    // 稍后等待结果
    var result string
    err := future.Get(ctx, &result)
    if err != nil {
        return err
    }
    
    return nil
}
```

---

## 重试策略配置

Temporal 提供灵活的重试策略，活动失败后会自动重试。

### 重试策略配置

```go
retryPolicy := &temporal.RetryPolicy{
    // 初始重试间隔
    InitialInterval: time.Second,
    
    // 退避系数：每次重试间隔 = 前一次间隔 * 退避系数
    BackoffCoefficient: 2.0,
    
    // 最大重试间隔（避免指数增长过大）
    MaximumInterval: time.Minute,
    
    // 最大重试次数
    MaximumAttempts: 5,
    
    // 不可重试的错误类型
    NonRetryableErrorTypes: []string{
        "InvalidInput",
        "AuthenticationFailed",
        "ResourceNotFound",
    },
}
```

### 重试间隔计算

```
重试次数    间隔计算                  实际间隔
────────────────────────────────────────────────
第 1 次     InitialInterval          1s
第 2 次     1s × 2.0                  2s
第 3 次     2s × 2.0                  4s
第 4 次     4s × 2.0                  8s
第 5 次     8s × 2.0 = 16s            16s
第 6 次     16s × 2.0 = 32s           32s
第 7 次     32s × 2.0 = 64s           60s (MaximumInterval)
第 8 次     60s                       60s (MaximumInterval)
```

### 在活动中指定错误类型

```go
func ProcessPaymentActivity(ctx context.Context, req PaymentRequest) error {
    logger := activity.GetLogger(ctx)
    
    // 业务逻辑
    err := validateRequest(req)
    if err != nil {
        // 不可重试的错误
        return temporal.NewNonRetryableApplicationError(
            err.Error(),
            "InvalidInput",
            nil,
        )
    }
    
    // 可重试的错误（如网络问题）
    err = callPaymentGateway(req)
    if err != nil {
        if isNetworkError(err) {
            // 可重试，直接返回错误
            return err
        }
        // 不可重试
        return temporal.NewNonRetryableApplicationError(
            err.Error(),
            "PaymentFailed",
            nil,
        )
    }
    
    return nil
}
```

### 自定义重试策略示例

```go
// 不同场景使用不同重试策略
func WorkflowWithDifferentRetryPolicies(ctx workflow.Context) error {
    // 支付活动：快速失败，不重试
    paymentAO := workflow.ActivityOptions{
        StartToCloseTimeout: time.Minute,
        RetryPolicy: &temporal.RetryPolicy{
            MaximumAttempts: 1,  // 不重试
        },
    }
    ctx1 := workflow.WithActivityOptions(ctx, paymentAO)
    workflow.ExecuteActivity(ctx1, ProcessPayment, payment).Get(ctx, nil)
    
    // API 调用：适度重试
    apiAO := workflow.ActivityOptions{
        StartToCloseTimeout: 5 * time.Minute,
        RetryPolicy: &temporal.RetryPolicy{
            InitialInterval:    5 * time.Second,
            BackoffCoefficient: 2.0,
            MaximumAttempts:    3,
        },
    }
    ctx2 := workflow.WithActivityOptions(ctx, apiAO)
    workflow.ExecuteActivity(ctx2, CallExternalAPI, data).Get(ctx, nil)
    
    // 数据同步：激进重试
    syncAO := workflow.ActivityOptions{
        StartToCloseTimeout: 30 * time.Minute,
        RetryPolicy: &temporal.RetryPolicy{
            InitialInterval:    time.Second,
            BackoffCoefficient: 1.5,
            MaximumInterval:    time.Minute,
            MaximumAttempts:    20,
        },
    }
    ctx3 := workflow.WithActivityOptions(ctx, syncAO)
    workflow.ExecuteActivity(ctx3, SyncData, source).Get(ctx, nil)
    
    return nil
}
```

---

## 心跳（Heartbeat）

对于长时间运行的活动，心跳用于报告活动仍在执行，防止被误判为超时。

### 心跳机制

```
┌─────────────────────────────────────────────────────────────────────────┐
│                        Activity Heartbeat Flow                          │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│  Activity Worker                                                        │
│  ┌─────────────────────────────────────────────────────────────┐       │
│  │                                                               │       │
│  │  func LongRunningActivity(ctx context.Context) error {       │       │
│  │      for i := 0; i < 100; i++ {                              │       │
│  │          // 处理任务                                         │       │
│  │          processItem(i)                                       │       │
│  │                                                              │       │
│  │          // 发送心跳                                         │       │
│  │          activity.RecordHeartbeat(ctx, i)                   │       │
│  │      }                                                       │       │
│  │      return nil                                              │       │
│  │  }                                                           │       │
│  │                                                               │       │
│  └──────────────────────────┬──────────────────────────────────┬─┘       │
│                             │                                  │         │
│                             │ Heartbeat                        │ Result  │
│                             ▼                                  ▼         │
│                    ┌──────────────────────────────────────────────┐      │
│                    │          Temporal Server                    │      │
│                    │  ┌────────────────────────────────────────┐ │      │
│                    │  │ HeartbeatTimeout: 30s                 │ │      │
│                    │  │ LastHeartbeat: 2024-01-01 10:00:25   │ │      │
│                    │  │ Status: Running                        │ │      │
│                    │  └────────────────────────────────────────┘ │      │
│                    └──────────────────────────────────────────────┘      │
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
```

### 实现心跳

```go
func LongRunningActivity(ctx context.Context, taskID string) error {
    logger := activity.GetLogger(ctx)
    
    items := getItems(taskID)
    total := len(items)
    
    for i, item := range items {
        // 检查是否被取消
        if activity.IsCancelRequested(ctx) {
            logger.Info("Activity cancelled")
            return activity.NewCanceledError()
        }
        
        // 处理项目
        err := processItem(item)
        if err != nil {
            return err
        }
        
        // 报告心跳和进度
        progress := (i + 1) * 100 / total
        activity.RecordHeartbeat(ctx, progress)
        logger.Info("Progress", "percent", progress)
    }
    
    logger.Info("Task completed", "taskID", taskID)
    return nil
}
```

### 心跳详情（恢复进度）

```go
func ResumableActivity(ctx context.Context, taskID string) error {
    logger := activity.GetLogger(ctx)
    
    // 获取上次心跳详情（用于恢复进度）
    var lastProcessedIndex int
    if activity.HasHeartbeatDetails(ctx) {
        activity.GetHeartbeatDetails(ctx, &lastProcessedIndex)
        logger.Info("Resuming from", "index", lastProcessedIndex)
    }
    
    items := getItems(taskID)
    
    for i := lastProcessedIndex; i < len(items); i++ {
        item := items[i]
        
        // 处理项目
        if err := processItem(item); err != nil {
            return err
        }
        
        // 记录进度（下次恢复时从这里开始）
        activity.RecordHeartbeat(ctx, i+1)
    }
    
    return nil
}
```

### 心跳超时配置

```go
func WorkflowWithHeartbeat(ctx workflow.Context) error {
    ao := workflow.ActivityOptions{
        StartToCloseTimeout: 10 * time.Minute,  // 活动总超时
        HeartbeatTimeout:     30 * time.Second,  // 心跳超时
    }
    ctx = workflow.WithActivityOptions(ctx, ao)
    
    return workflow.ExecuteActivity(ctx, LongRunningActivity, "task-123").Get(ctx, nil)
}
```

### 心跳最佳实践

```go
func BestPracticeHeartbeatActivity(ctx context.Context, taskID string) error {
    logger := activity.GetLogger(ctx)
    
    // 1. 定期发送心跳
    ticker := time.NewTicker(10 * time.Second)
    defer ticker.Stop()
    
    // 2. 获取恢复点
    var lastProcessedIndex int
    if activity.HasHeartbeatDetails(ctx) {
        activity.GetHeartbeatDetails(ctx, &lastProcessedIndex)
    }
    
    items := getItems(taskID)
    
    for i := lastProcessedIndex; i < len(items); i++ {
        select {
        case <-ticker.C:
            // 定期发送心跳
            activity.RecordHeartbeat(ctx, i)
            
        default:
            // 继续处理
        }
        
        // 3. 检查取消请求
        if activity.IsCancelRequested(ctx) {
            // 保存进度
            activity.RecordHeartbeat(ctx, i)
            return activity.NewCanceledError()
        }
        
        // 处理项目
        if err := processItem(items[i]); err != nil {
            // 4. 错误时也保存进度
            activity.RecordHeartbeat(ctx, i)
            return err
        }
    }
    
    return nil
}
```

---

## 幂等性

活动可能被多次执行，因此必须实现幂等性，确保重复执行不会产生副作用。

### 为什么需要幂等性

```
┌─────────────────────────────────────────────────────────────────────────┐
│                     Activity Execution Scenarios                         │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│  场景 1：网络超时重试                                                    │
│  ┌─────────────┐         ┌─────────────┐         ┌─────────────┐        │
│  │  Activity   │ ──────▶ │  External   │         │  Activity   │        │
│  │  Execution  │         │   Service   │         │   Result    │        │
│  └─────────────┘         └─────────────┘         └─────────────┘        │
│         │                      │                       ▲                 │
│         │                      │                       │                 │
│         │   ◀──── 超时 ────▶   │                       │                 │
│         │                      │                       │                 │
│         │                      │     重试执行          │                 │
│         └──────────────────────┼───────────────────────┘                 │
│                                │                                          │
│  场景 2：Worker 崩溃                                                     │
│  ┌─────────────┐         ┌─────────────┐         ┌─────────────┐        │
│  │  Activity   │         │   Worker    │         │  New Worker │        │
│  │  Started    │ ──────▶ │   Crashes   │         │   Retries   │        │
│  └─────────────┘         └─────────────┘         └─────────────┘        │
│                                                                          │
│  场景 3：心跳超时                                                        │
│  ┌─────────────┐         ┌─────────────┐         ┌─────────────┐        │
│  │  Activity   │         │  Heartbeat  │         │   Retries   │        │
│  │  Running    │ ──────▶ │   Timeout   │ ──────▶ │   Again     │        │
│  └─────────────┘         └─────────────┘         └─────────────┘        │
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
```

### 幂等性实现策略

#### 1. 使用唯一标识符

```go
func ProcessPaymentActivity(ctx context.Context, req PaymentRequest) error {
    logger := activity.GetLogger(ctx)
    
    // 使用唯一事务ID
    transactionID := req.TransactionID
    
    // 检查是否已处理
    processed, err := checkTransactionProcessed(transactionID)
    if err != nil {
        return err
    }
    if processed {
        logger.Info("Transaction already processed", "transactionID", transactionID)
        return nil
    }
    
    // 处理支付
    err = processPayment(req)
    if err != nil {
        return err
    }
    
    // 标记为已处理
    return markTransactionProcessed(transactionID)
}
```

#### 2. 使用心跳详情

```go
func IdempotentActivity(ctx context.Context, req Request) error {
    logger := activity.GetLogger(ctx)
    
    // 获取上次执行进度
    var progress struct {
        Step1Completed bool
        Step2Completed bool
        Step3Completed bool
    }
    
    if activity.HasHeartbeatDetails(ctx) {
        activity.GetHeartbeatDetails(ctx, &progress)
    }
    
    // 第一步
    if !progress.Step1Completed {
        if err := step1(req); err != nil {
            return err
        }
        progress.Step1Completed = true
        activity.RecordHeartbeat(ctx, progress)
    }
    
    // 第二步
    if !progress.Step2Completed {
        if err := step2(req); err != nil {
            return err
        }
        progress.Step2Completed = true
        activity.RecordHeartbeat(ctx, progress)
    }
    
    // 第三步
    if !progress.Step3Completed {
        if err := step3(req); err != nil {
            return err
        }
        progress.Step3Completed = true
        activity.RecordHeartbeat(ctx, progress)
    }
    
    return nil
}
```

#### 3. 数据库事务

```go
func DatabaseActivity(ctx context.Context, req Request) error {
    logger := activity.GetLogger(ctx)
    
    // 使用数据库事务确保幂等性
    tx, err := db.BeginTx(ctx, nil)
    if err != nil {
        return err
    }
    defer tx.Rollback()
    
    // 检查是否已处理
    var exists bool
    err = tx.QueryRow(
        "SELECT EXISTS(SELECT 1 FROM processed_requests WHERE request_id = $1)",
        req.ID,
    ).Scan(&exists)
    if err != nil {
        return err
    }
    
    if exists {
        logger.Info("Request already processed", "requestID", req.ID)
        return nil
    }
    
    // 处理请求
    err = processRequest(tx, req)
    if err != nil {
        return err
    }
    
    // 记录已处理
    _, err = tx.Exec(
        "INSERT INTO processed_requests (request_id, processed_at) VALUES ($1, NOW())",
        req.ID,
    )
    if err != nil {
        return err
    }
    
    return tx.Commit()
}
```

### 幂等性最佳实践

```go
// 最佳实践：幂等活动模板
func IdempotentActivityTemplate(ctx context.Context, req Request) error {
    logger := activity.GetLogger(ctx)
    
    // 1. 生成或使用唯一ID
    requestID := req.ID
    if requestID == "" {
        requestID = uuid.New().String()
    }
    
    // 2. 检查幂等性键
    processed, err := isProcessed(requestID)
    if err != nil {
        return fmt.Errorf("failed to check processed status: %w", err)
    }
    if processed {
        logger.Info("Request already processed", "requestID", requestID)
        return nil
    }
    
    // 3. 获取恢复点
    var checkpoint Checkpoint
    if activity.HasHeartbeatDetails(ctx) {
        activity.GetHeartbeatDetails(ctx, &checkpoint)
    }
    
    // 4. 执行业务逻辑（从恢复点开始）
    result, err := processWithCheckpoint(req, checkpoint)
    if err != nil {
        // 记录失败（可选）
        return err
    }
    
    // 5. 标记为已处理
    if err := markAsProcessed(requestID, result); err != nil {
        return fmt.Errorf("failed to mark as processed: %w", err)
    }
    
    return nil
}
```

---

## Local Activity vs Remote Activity

Temporal 提供两种活动类型：本地活动和远程活动。

### 对比表

| 特性 | Local Activity | Remote Activity |
|------|----------------|-----------------|
| 执行位置 | 同一 Worker 进程 | 任意 Worker |
| 延迟 | 低（无网络开销） | 高（有网络开销） |
| 超时限制 | 短（< 1 分钟） | 长（可数小时） |
| 重试 | 有限 | 完整支持 |
| 心跳 | 不支持 | 支持 |
| 网络故障 | 无影响 | 可能影响 |
| 适用场景 | 快速本地操作 | 外部服务调用 |

### Local Activity

```go
func WorkflowWithLocalActivity(ctx workflow.Context) error {
    // 配置本地活动选项
    lao := workflow.LocalActivityOptions{
        StartToCloseTimeout: time.Minute,  // 必须较短
    }
    ctx = workflow.WithLocalActivityOptions(ctx, lao)
    
    // 执行本地活动
    var result string
    err := workflow.ExecuteLocalActivity(ctx, LocalActivity, "input").Get(ctx, &result)
    if err != nil {
        return err
    }
    
    return nil
}

// 本地活动定义（与普通活动相同）
func LocalActivity(ctx context.Context, input string) (string, error) {
    // 快速本地操作
    return "processed: " + input, nil
}
```

### Local Activity 适用场景

```go
// 适合：本地计算
func ComputeHashActivity(ctx context.Context, data []byte) (string, error) {
    hash := sha256.Sum256(data)
    return hex.EncodeToString(hash[:]), nil
}

// 适合：本地缓存操作
func CacheGetActivity(ctx context.Context, key string) (string, error) {
    return cache.Get(key), nil
}

// 不适合：外部API调用（应该用远程活动）
func CallExternalAPIActivity(ctx context.Context, url string) error {
    resp, err := http.Get(url)  // 网络调用，应该用远程活动
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    return nil
}
```

### Remote Activity

```go
func WorkflowWithRemoteActivity(ctx workflow.Context) error {
    // 配置远程活动选项
    ao := workflow.ActivityOptions{
        StartToCloseTimeout: 5 * time.Minute,
        HeartbeatTimeout:     30 * time.Second,  // 支持心跳
        RetryPolicy: &temporal.RetryPolicy{      // 完整重试支持
            MaximumAttempts: 5,
        },
    }
    ctx = workflow.WithActivityOptions(ctx, ao)
    
    // 执行远程活动
    var result string
    err := workflow.ExecuteActivity(ctx, RemoteActivity, "input").Get(ctx, &result)
    if err != nil {
        return err
    }
    
    return nil
}
```

### 混合使用

```go
func MixedWorkflow(ctx workflow.Context, order Order) error {
    // 本地活动：验证订单（快速、本地计算）
    lao := workflow.LocalActivityOptions{
        StartToCloseTimeout: 10 * time.Second,
    }
    localCtx := workflow.WithLocalActivityOptions(ctx, lao)
    
    if err := workflow.ExecuteLocalActivity(localCtx, ValidateOrder, order).Get(ctx, nil); err != nil {
        return err
    }
    
    // 远程活动：处理支付（需要调用外部服务）
    ao := workflow.ActivityOptions{
        StartToCloseTimeout: 5 * time.Minute,
        RetryPolicy: &temporal.RetryPolicy{
            MaximumAttempts: 3,
        },
    }
    remoteCtx := workflow.WithActivityOptions(ctx, ao)
    
    var paymentID string
    if err := workflow.ExecuteActivity(remoteCtx, ProcessPayment, order).Get(ctx, &paymentID); err != nil {
        return err
    }
    
    // 本地活动：更新缓存
    if err := workflow.ExecuteLocalActivity(localCtx, UpdateCache, order.ID, paymentID).Get(ctx, nil); err != nil {
        // 缓存更新失败不影响主流程
        workflow.GetLogger(ctx).Warn("Failed to update cache", "error", err)
    }
    
    return nil
}
```

---

## Go 代码示例

### 完整示例：电商订单处理

```go
package activity

import (
    "context"
    "fmt"
    "time"
    
    "go.temporal.io/sdk/activity"
    "go.temporal.io/sdk/temporal"
)

// 订单结构
type Order struct {
    ID         string
    CustomerID string
    Items      []OrderItem
    Total      float64
}

type OrderItem struct {
    ProductID string
    Quantity  int
    Price     float64
}

// 库存服务接口
type InventoryService interface {
    CheckAvailability(productID string, quantity int) (bool, error)
    Reserve(productID string, quantity int) error
    Release(productID string, quantity int) error
}

// 支付服务接口
type PaymentService interface {
    Process(orderID string, amount float64) (string, error)
    Refund(transactionID string) error
}

// 物流服务接口
type ShippingService interface {
    Ship(orderID string, address string) (string, error)
}

// 活动结构体
type OrderActivity struct {
    Inventory InventoryService
    Payment   PaymentService
    Shipping  ShippingService
}

// 验证订单活动
func (a *OrderActivity) ValidateOrder(ctx context.Context, order Order) error {
    logger := activity.GetLogger(ctx)
    logger.Info("Validating order", "orderID", order.ID)
    
    // 验证订单项
    if len(order.Items) == 0 {
        return temporal.NewNonRetryableApplicationError(
            "order has no items",
            "InvalidOrder",
            nil,
        )
    }
    
    // 验证总额
    var total float64
    for _, item := range order.Items {
        total += item.Price * float64(item.Quantity)
    }
    if total != order.Total {
        return fmt.Errorf("order total mismatch: expected %.2f, got %.2f", total, order.Total)
    }
    
    return nil
}

// 检查库存活动
func (a *OrderActivity) CheckInventory(ctx context.Context, order Order) error {
    logger := activity.GetLogger(ctx)
    
    // 获取恢复进度
    var checked map[string]bool
    if activity.HasHeartbeatDetails(ctx) {
        activity.GetHeartbeatDetails(ctx, &checked)
    } else {
        checked = make(map[string]bool)
    }
    
    // 检查每个商品
    for _, item := range order.Items {
        if checked[item.ProductID] {
            continue
        }
        
        available, err := a.Inventory.CheckAvailability(item.ProductID, item.Quantity)
        if err != nil {
            return fmt.Errorf("failed to check inventory for %s: %w", item.ProductID, err)
        }
        if !available {
            return temporal.NewNonRetryableApplicationError(
                fmt.Sprintf("product %s not available", item.ProductID),
                "OutOfStock",
                nil,
            )
        }
        
        checked[item.ProductID] = true
        activity.RecordHeartbeat(ctx, checked)
    }
    
    logger.Info("Inventory checked", "orderID", order.ID)
    return nil
}

// 预留库存活动（幂等）
func (a *OrderActivity) ReserveInventory(ctx context.Context, order Order) error {
    logger := activity.GetLogger(ctx)
    
    // 检查是否已预留
    reservationID := fmt.Sprintf("reserve-%s", order.ID)
    reserved, err := isReserved(reservationID)
    if err != nil {
        return err
    }
    if reserved {
        logger.Info("Inventory already reserved", "orderID", order.ID)
        return nil
    }
    
    // 预留库存
    for _, item := range order.Items {
        if err := a.Inventory.Reserve(item.ProductID, item.Quantity); err != nil {
            // 预留失败，释放已预留的
            for _, i := range order.Items {
                if i.ProductID == item.ProductID {
                    break
                }
                a.Inventory.Release(i.ProductID, i.Quantity)
            }
            return fmt.Errorf("failed to reserve %s: %w", item.ProductID, err)
        }
    }
    
    // 标记为已预留
    if err := markAsReserved(reservationID); err != nil {
        return err
    }
    
    logger.Info("Inventory reserved", "orderID", order.ID)
    return nil
}

// 处理支付活动（幂等）
func (a *OrderActivity) ProcessPayment(ctx context.Context, order Order) (string, error) {
    logger := activity.GetLogger(ctx)
    
    // 检查是否已支付
    paymentID, paid, err := getPaymentStatus(order.ID)
    if err != nil {
        return "", err
    }
    if paid {
        logger.Info("Payment already processed", "orderID", order.ID, "paymentID", paymentID)
        return paymentID, nil
    }
    
    // 处理支付
    transactionID, err := a.Payment.Process(order.ID, order.Total)
    if err != nil {
        return "", fmt.Errorf("payment failed: %w", err)
    }
    
    // 保存支付状态
    if err := savePaymentStatus(order.ID, transactionID); err != nil {
        return "", err
    }
    
    logger.Info("Payment processed", "orderID", order.ID, "transactionID", transactionID)
    return transactionID, nil
}

// 发货活动
func (a *OrderActivity) ShipOrder(ctx context.Context, order Order, address string) (string, error) {
    logger := activity.GetLogger(ctx)
    
    // 发送心跳
    ticker := time.NewTicker(10 * time.Second)
    defer ticker.Stop()
    
    done := make(chan string, 1)
    errCh := make(chan error, 1)
    
    go func() {
        trackingNumber, err := a.Shipping.Ship(order.ID, address)
        if err != nil {
            errCh <- err
            return
        }
        done <- trackingNumber
    }()
    
    for {
        select {
        case trackingNumber := <-done:
            logger.Info("Order shipped", "orderID", order.ID, "trackingNumber", trackingNumber)
            return trackingNumber, nil
            
        case err := <-errCh:
            return "", fmt.Errorf("shipping failed: %w", err)
            
        case <-ticker.C:
            // 发送心跳
            activity.RecordHeartbeat(ctx, "shipping in progress")
            
        case <-ctx.Done():
            // 检查取消请求
            if activity.IsCancelRequested(ctx) {
                return "", activity.NewCanceledError()
            }
        }
    }
}

// 取消订单活动
func (a *OrderActivity) CancelOrder(ctx context.Context, order Order, reason string) error {
    logger := activity.GetLogger(ctx)
    logger.Info("Cancelling order", "orderID", order.ID, "reason", reason)
    
    // 释放库存
    for _, item := range order.Items {
        a.Inventory.Release(item.ProductID, item.Quantity)
    }
    
    // 退款（如果已支付）
    paymentID, paid, _ := getPaymentStatus(order.ID)
    if paid {
        a.Payment.Refund(paymentID)
    }
    
    logger.Info("Order cancelled", "orderID", order.ID)
    return nil
}

// 辅助函数（示例实现）
func isReserved(reservationID string) (bool, error) {
    // 查询数据库
    return false, nil
}

func markAsReserved(reservationID string) error {
    // 更新数据库
    return nil
}

func getPaymentStatus(orderID string) (string, bool, error) {
    // 查询数据库
    return "", false, nil
}

func savePaymentStatus(orderID, transactionID string) error {
    // 更新数据库
    return nil
}
```

---

## 最佳实践

### 1. 活动设计原则

- **单一职责**：每个活动只做一件事
- **幂等性**：确保重复执行不产生副作用
- **合理超时**：根据实际执行时间设置超时
- **心跳报告**：长时间运行的活动必须发送心跳

### 2. 错误处理

```go
func RobustActivity(ctx context.Context, req Request) error {
    logger := activity.GetLogger(ctx)
    
    // 区分错误类型
    err := doSomething(req)
    if err != nil {
        if isBusinessError(err) {
            // 业务错误：不可重试
            return temporal.NewNonRetryableApplicationError(
                err.Error(),
                "BusinessError",
                nil,
            )
        }
        if isNetworkError(err) {
            // 网络错误：可重试
            return err
        }
        // 其他错误
        return fmt.Errorf("unexpected error: %w", err)
    }
    
    return nil
}
```

### 3. 资源管理

```go
func ResourceManagementActivity(ctx context.Context, req Request) error {
    // 获取资源
    resource, err := acquireResource()
    if err != nil {
        return err
    }
    
    // 确保释放资源
    defer releaseResource(resource)
    
    // 处理业务逻辑
    return processWithResource(resource, req)
}
```

### 4. 日志和监控

```go
func ObservabilityActivity(ctx context.Context, req Request) error {
    logger := activity.GetLogger(ctx)
    
    // 记录开始
    startTime := time.Now()
    logger.Info("Activity started", "requestID", req.ID)
    
    // 记录指标
    activity.GetMetricsHandler(ctx).Counter("activity_starts").Inc(1)
    
    // 处理业务逻辑
    err := processRequest(req)
    
    // 记录结果
    duration := time.Since(startTime)
    if err != nil {
        logger.Error("Activity failed", "error", err, "duration", duration)
        activity.GetMetricsHandler(ctx).Counter("activity_failures").Inc(1)
    } else {
        logger.Info("Activity completed", "duration", duration)
        activity.GetMetricsHandler(ctx).Counter("activity_completions").Inc(1)
    }
    
    return err
}
```

---

## 相关资源

- [Temporal 官方文档 - Activity](https://docs.temporal.io/activities)
- [Go SDK Activity 指南](https://docs.temporal.io/dev-guide/go/activities)
- [Activity Heartbeats](https://docs.temporal.io/activities#heartbeats)
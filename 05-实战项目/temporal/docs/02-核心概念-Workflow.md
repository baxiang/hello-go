# 核心概念 - Workflow（工作流）

工作流（Workflow）是 Temporal 中的核心概念，代表一个持久的业务逻辑流程。本文档详细介绍 Temporal 工作流的核心概念和最佳实践。

---

## Workflow Definition（工作流定义）

工作流定义是一个函数，描述了业务流程的执行逻辑。在 Go SDK 中，工作流函数使用 `workflow.Context` 作为第一个参数。

### 基本定义

```go
package workflow

import (
    "time"
    "go.temporal.io/sdk/workflow"
)

// 简单的工作流定义
func OrderProcessingWorkflow(ctx workflow.Context, orderID string) error {
    // 工作流逻辑
    return nil
}

// 带返回值的工作流
func GreetingWorkflow(ctx workflow.Context, name string) (string, error) {
    return "Hello, " + name, nil
}

// 多参数的工作流
func PaymentWorkflow(ctx workflow.Context, userID string, amount float64, currency string) error {
    // 处理支付逻辑
    return nil
}
```

### 工作流参数和返回值

工作流支持多种类型的参数和返回值：

```go
// 基本类型参数
func Workflow(ctx workflow.Context, id string, count int, enabled bool) error {
    return nil
}

// 结构体参数
type Order struct {
    ID        string
    Items     []Item
    Total     float64
    CreatedAt time.Time
}

func OrderWorkflow(ctx workflow.Context, order Order) error {
    return nil
}

// 多个返回值
func CalculationWorkflow(ctx workflow.Context, a, b int) (int, int, error) {
    return a + b, a * b, nil
}

// 结构体返回值
type Result struct {
    Success bool
    Message string
    Data    interface{}
}

func ProcessWorkflow(ctx workflow.Context) (Result, error) {
    return Result{
        Success: true,
        Message: "Completed",
    }, nil
}
```

### 工作流注册

```go
package main

import (
    "go.temporal.io/sdk/client"
    "go.temporal.io/sdk/worker"
    "myapp/workflow"
)

func main() {
    c, _ := client.Dial(client.Options{})
    defer c.Close()

    w := worker.New(c, "my-task-queue", worker.Options{})
    
    // 注册工作流
    w.RegisterWorkflow(workflow.OrderProcessingWorkflow)
    w.RegisterWorkflow(workflow.PaymentWorkflow)
    
    w.Run(worker.InterruptCh())
}
```

---

## Workflow Type（工作流类型）

工作流类型是工作流的唯一标识符，用于在启动工作流时指定要执行的工作流。

### 类型命名规则

```go
// 默认类型名：函数名
func OrderProcessingWorkflow(ctx workflow.Context, orderID string) error {
    return nil
}
// 类型名为: OrderProcessingWorkflow

// 自定义类型名
func main() {
    w := worker.New(c, "task-queue", worker.Options{})
    
    // 使用默认类型名
    w.RegisterWorkflow(OrderProcessingWorkflow)
    
    // 自定义类型名
    w.RegisterWorkflowWithOptions(
        OrderProcessingWorkflow,
        workflow.RegisterOptions{
            Name: "custom-order-workflow",
        },
    )
}
```

### 启动工作流时指定类型

```go
func main() {
    c, _ := client.Dial(client.Options{})
    defer c.Close()
    
    // 方式一：使用函数名作为类型
    we, _ := c.ExecuteWorkflow(
        context.Background(),
        client.StartWorkflowOptions{
            ID:        "order-123",
            TaskQueue: "my-task-queue",
        },
        OrderProcessingWorkflow,  // 传递函数
        "order-123",
    )
    
    // 方式二：使用字符串类型名
    we, _ := c.ExecuteWorkflow(
        context.Background(),
        client.StartWorkflowOptions{
            ID:        "order-456",
            TaskQueue: "my-task-queue",
        },
        "custom-order-workflow",  // 传递字符串类型名
        "order-456",
    )
}
```

### 工作流类型最佳实践

```go
// 命名约定：
// 1. 使用有意义的名称描述工作流功能
// 2. 采用 PascalCase 命名风格
// 3. 包含业务领域前缀（可选）

// 好的命名示例
func OrderProcessingWorkflow(ctx workflow.Context, orderID string) error { }
func PaymentRefundWorkflow(ctx workflow.Context, paymentID string) error { }
func UserRegistrationWorkflow(ctx workflow.Context, userID string) error { }
func DataSyncWorkflow(ctx workflow.Context, source, target string) error { }

// 不推荐：过于简单的命名
func ProcessWorkflow(ctx workflow.Context) error { }
func HandleWorkflow(ctx workflow.Context) error { }
```

---

## Workflow Execution（工作流执行）

工作流执行是工作流类型的一次具体运行实例。每个执行都有唯一的标识符和执行历史。

### 启动工作流执行

```go
func main() {
    c, _ := client.Dial(client.Options{})
    defer c.Close()
    
    // 基本启动
    options := client.StartWorkflowOptions{
        ID:        "order-123",
        TaskQueue: "order-task-queue",
    }
    
    we, err := c.ExecuteWorkflow(
        context.Background(),
        options,
        OrderProcessingWorkflow,
        "order-123",
    )
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("WorkflowID: %s, RunID: %s\n", we.GetID(), we.GetRunID())
    
    // 等待工作流完成
    var result string
    err = we.Get(context.Background(), &result)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Println("Result:", result)
}
```

### 执行选项详解

```go
options := client.StartWorkflowOptions{
    // 必填参数
    ID:        "workflow-unique-id",    // 工作流唯一标识
    TaskQueue: "my-task-queue",          // 任务队列名
    
    // 超时设置
    WorkflowExecutionTimeout:    10 * time.Minute,  // 执行总超时
    WorkflowRunTimeout:          5 * time.Minute,   // 单次运行超时
    WorkflowTaskTimeout:         10 * time.Second,  // 任务处理超时
    
    // 重试策略
    RetryPolicy: &temporal.RetryPolicy{
        InitialInterval:    time.Second,      // 初始重试间隔
        BackoffCoefficient: 2.0,              // 退避系数
        MaximumInterval:    time.Minute,      // 最大重试间隔
        MaximumAttempts:    5,                // 最大重试次数
        NonRetryableErrorTypes: []string{     // 不可重试的错误类型
            "InvalidInput",
            "AuthenticationFailed",
        },
    },
    
    // Cron 调度
    CronSchedule: "0 0 * * *",  // 每天午夜执行
    
    // 其他选项
    Memo:              map[string]interface{}{"key": "value"},  // 备注
    SearchAttributes:  map[string]interface{}{"type": "order"}, // 搜索属性
    WorkflowIDReusePolicy: enums.WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE,
}
```

### 工作流执行状态

```
┌─────────────────────────────────────────────────────────────┐
│                   Workflow Execution States                  │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  ┌──────────┐     ┌──────────┐     ┌──────────┐            │
│  │  Running │────▶│ Completed│     │  Failed  │            │
│  └──────────┘     └──────────┘     └──────────┘            │
│       │                                  ▲                  │
│       │                                  │                  │
│       ▼                                  │                  │
│  ┌──────────┐                            │                  │
│  │ Canceled │────────────────────────────┘                  │
│  └──────────┘                                               │
│       │                                                      │
│       ▼                                                      │
│  ┌──────────┐                                               │
│  │Terminated│                                               │
│  └──────────┘                                               │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

### 获取执行结果

```go
func main() {
    c, _ := client.Dial(client.Options{})
    defer c.Close()
    
    // 启动工作流
    we, _ := c.ExecuteWorkflow(context.Background(), options, MyWorkflow, "input")
    
    // 方式一：阻塞等待结果
    var result MyResult
    err := we.Get(context.Background(), &result)
    
    // 方式二：使用带超时的上下文
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
    defer cancel()
    err := we.Get(ctx, &result)
    
    // 方式三：查询已存在的工作流
    we := c.GetWorkflow(context.Background(), "workflow-id", "")
    we.Get(context.Background(), &result)
}
```

### 查询工作流状态

```go
func main() {
    c, _ := client.Dial(client.Options{})
    defer c.Close()
    
    // 获取工作流描述
    resp, err := c.DescribeWorkflowExecution(
        context.Background(),
        "workflow-id",
        "",  // run-id 可选
    )
    
    fmt.Printf("Status: %s\n", resp.WorkflowExecutionInfo.Status)
    fmt.Printf("StartTime: %v\n", resp.WorkflowExecutionInfo.StartTime)
    fmt.Printf("CloseTime: %v\n", resp.WorkflowExecutionInfo.CloseTime)
}
```

---

## Event History（事件历史）

事件历史是工作流执行的完整记录，是实现持久化执行的关键。

### 事件历史的作用

1. **状态恢复**：Worker 重启后可以从事件历史恢复执行状态
2. **调试审计**：提供完整的执行轨迹，便于问题排查
3. **重放执行**：支持重新执行历史事件

### 事件类型

```
┌─────────────────────────────────────────────────────────────┐
│                    Event History Example                     │
├─────────────────────────────────────────────────────────────┤
│ EventID: 1   WorkflowExecutionStarted                        │
│ EventID: 2   ActivityTaskScheduled      (ActivityID: 1)     │
│ EventID: 3   ActivityTaskStarted        (ActivityID: 1)     │
│ EventID: 4   ActivityTaskCompleted      (ActivityID: 1)     │
│ EventID: 5   TimerStarted               (TimerID: 1)        │
│ EventID: 6   TimerFired                 (TimerID: 1)        │
│ EventID: 7   ActivityTaskScheduled      (ActivityID: 2)     │
│ EventID: 8   ActivityTaskStarted        (ActivityID: 2)     │
│ EventID: 9   ActivityTaskFailed         (ActivityID: 2)     │
│ EventID: 10  ActivityTaskRetry          (ActivityID: 2)     │
│ EventID: 11  ActivityTaskCompleted      (ActivityID: 2)     │
│ EventID: 12  WorkflowExecutionCompleted                        │
└─────────────────────────────────────────────────────────────┘
```

### 查看事件历史

```bash
# 使用 CLI 查看事件历史
temporal workflow show --workflow-id order-123

# 输出示例
Progress:
  ID           Time                     Type                     Details
  1            2024-01-01T10:00:00Z     WorkflowExecutionStarted
  2            2024-01-01T10:00:01Z     ActivityTaskScheduled    ActivityID: 1, Type: ProcessPayment
  3            2024-01-01T10:00:05Z     ActivityTaskCompleted    ActivityID: 1
  4            2024-01-01T10:00:06Z     WorkflowExecutionCompleted
```

### 事件历史与确定性

工作流的确定性是通过事件历史重放实现的：

```go
func MyWorkflow(ctx workflow.Context, input string) error {
    // 第一次执行：执行活动 A，记录事件
    // 第二次执行（恢复）：重放事件，跳过实际执行，直接返回结果
    
    var result string
    err := workflow.ExecuteActivity(ctx, ActivityA, input).Get(ctx, &result)
    // Event: ActivityTaskScheduled -> ActivityTaskCompleted
    // 重放时：看到 ActivityTaskCompleted，直接使用其结果
    
    return nil
}
```

---

## 确定性约束

工作流必须具有确定性，即相同的输入和事件历史必须产生相同的执行路径。

### 确定性约束详解

#### 1. 禁止使用随机数

```go
// 错误：使用随机数
func BadWorkflow(ctx workflow.Context) error {
    rand.Seed(time.Now().UnixNano())  // 错误！
    n := rand.Intn(100)               // 错误！
    return nil
}

// 正确：使用确定性值或从参数传入
func GoodWorkflow(ctx workflow.Context, randomValue int) error {
    return nil
}
```

#### 2. 禁止使用当前时间

```go
// 错误：使用 time.Now()
func BadWorkflow(ctx workflow.Context) error {
    now := time.Now()  // 错误！
    return nil
}

// 正确：使用 workflow.Now()
func GoodWorkflow(ctx workflow.Context) error {
    now := workflow.Now(ctx)  // 正确！
    return nil
}
```

#### 3. 禁止使用原生并发

```go
// 错误：使用 goroutine
func BadWorkflow(ctx workflow.Context) error {
    go func() {  // 错误！
        // ...
    }()
    return nil
}

// 正确：使用 workflow.Go
func GoodWorkflow(ctx workflow.Context) error {
    workflow.Go(ctx, func(ctx workflow.Context) {
        // 正确！
    })
    return nil
}
```

#### 4. 禁止使用原生 Channel

```go
// 错误：使用原生 channel
func BadWorkflow(ctx workflow.Context) error {
    ch := make(chan int)  // 错误！
    return nil
}

// 正确：使用 workflow.Channel
func GoodWorkflow(ctx workflow.Context) error {
    ch := workflow.NewBufferedChannel(ctx, 10)  // 正确！
    return nil
}
```

#### 5. 禁止阻塞操作

```go
// 错误：使用 time.Sleep
func BadWorkflow(ctx workflow.Context) error {
    time.Sleep(time.Hour)  // 错误！
    return nil
}

// 正确：使用 workflow.NewTimer
func GoodWorkflow(ctx workflow.Context) error {
    workflow.NewTimer(ctx, time.Hour).Get(ctx, nil)  // 正确！
    return nil
}
```

### 确定性示例

```go
// 非确定性工作流（错误示例）
func NonDeterministicWorkflow(ctx workflow.Context) error {
    // 错误：每次执行可能产生不同结果
    if rand.Intn(2) == 0 {
        return workflow.ExecuteActivity(ctx, ActivityA).Get(ctx, nil)
    } else {
        return workflow.ExecuteActivity(ctx, ActivityB).Get(ctx, nil)
    }
}

// 确定性工作流（正确示例）
func DeterministicWorkflow(ctx workflow.Context, branch int) error {
    // 正确：基于输入决定执行路径
    if branch == 0 {
        return workflow.ExecuteActivity(ctx, ActivityA).Get(ctx, nil)
    } else {
        return workflow.ExecuteActivity(ctx, ActivityB).Get(ctx, nil)
    }
}
```

### 确定性检测工具

```go
import "go.temporal.io/sdk/worker"

func main() {
    c, _ := client.Dial(client.Options{})
    defer c.Close()
    
    w := worker.New(c, "task-queue", worker.Options{
        // 启用确定性检测
        DisableWorkflowWorker: false,
    })
    
    // 工作流代码会在沙箱中执行，检测非确定性操作
}
```

---

## 长时间运行的工作流

Temporal 支持运行数天、数月甚至数年的工作流。

### 长时间运行的设计模式

#### 1. 等待外部信号

```go
func ApprovalWorkflow(ctx workflow.Context, requestID string) error {
    var approved bool
    
    // 等待审批信号（可能需要数天）
    approvedChan := workflow.GetSignalChannel(ctx, "approval-signal")
    
    // 设置超时（30 天）
    timeout := workflow.NewTimer(ctx, 30*24*time.Hour)
    
    selector := workflow.NewSelector(ctx)
    selector.AddReceive(approvedChan, func(c workflow.ReceiveChannel, more bool) {
        c.Receive(ctx, &approved)
    })
    selector.AddFuture(timeout, func(f workflow.Future) {
        // 超时处理
    })
    
    selector.Select(ctx)
    
    if approved {
        return workflow.ExecuteActivity(ctx, ProcessApproval, requestID).Get(ctx, nil)
    }
    
    return errors.New("approval timeout")
}
```

#### 2. 定期检查

```go
func MonitoringWorkflow(ctx workflow.Context, serviceURL string) error {
    for {
        // 执行检查
        var status string
        err := workflow.ExecuteActivity(ctx, CheckService, serviceURL).Get(ctx, &status)
        if err != nil {
            workflow.GetLogger(ctx).Error("Service check failed", "error", err)
        }
        
        // 等待下一次检查
        workflow.NewTimer(ctx, time.Hour).Get(ctx, nil)
    }
}
```

#### 3. 继续-作为-新执行

```go
func LongRunningWorkflow(ctx workflow.Context, iteration int) error {
    // 执行一些工作
    workflow.ExecuteActivity(ctx, ProcessBatch, iteration).Get(ctx, nil)
    
    // 检查是否需要继续
    if iteration < 100 {
        // 继续作为新的工作流执行，避免事件历史过长
        return workflow.NewContinueAsNewError(ctx, LongRunningWorkflow, iteration+1)
    }
    
    return nil
}
```

### 事件历史大小管理

```go
func BatchProcessingWorkflow(ctx workflow.Context, items []string) error {
    batchSize := 100
    
    for i := 0; i < len(items); i += batchSize {
        end := i + batchSize
        if end > len(items) {
            end = len(items)
        }
        
        batch := items[i:end]
        workflow.ExecuteActivity(ctx, ProcessBatch, batch).Get(ctx, nil)
        
        // 检查事件历史大小
        info := workflow.GetInfo(ctx)
        if info.GetCurrentHistoryLength() > 40000 {
            // 事件历史接近上限（50000），启动新执行
            return workflow.NewContinueAsNewError(
                ctx, 
                BatchProcessingWorkflow, 
                items[end:],
            )
        }
    }
    
    return nil
}
```

---

## Schedule 和 Cron Job

Temporal 支持定时调度工作流。

### Cron 调度

```go
func main() {
    c, _ := client.Dial(client.Options{})
    defer c.Close()
    
    // 使用 Cron 表达式启动工作流
    options := client.StartWorkflowOptions{
        ID:           "daily-report",
        TaskQueue:    "report-task-queue",
        CronSchedule: "0 0 * * *",  // 每天午夜执行
    }
    
    we, _ := c.ExecuteWorkflow(context.Background(), options, ReportWorkflow)
    fmt.Println("Started cron workflow:", we.GetID())
}

func ReportWorkflow(ctx workflow.Context) error {
    logger := workflow.GetLogger(ctx)
    logger.Info("Running daily report")
    
    return workflow.ExecuteActivity(ctx, GenerateReport).Get(ctx, nil)
}
```

### Cron 表达式示例

```
┌───────────── 分钟 (0 - 59)
│ ┌───────────── 小时 (0 - 23)
│ │ ┌───────────── 日期 (1 - 31)
│ │ │ ┌───────────── 月份 (1 - 12)
│ │ │ │ ┌───────────── 星期 (0 - 6) (0 是周日)
│ │ │ │ │
* * * * *

示例：
0 0 * * *       - 每天午夜
0 */2 * * *     - 每 2 小时
0 9 * * 1-5     - 周一到周五早上 9 点
0 0 1 * *       - 每月 1 号午夜
0 0 1 1 *       - 每年 1 月 1 号午夜
```

### 使用 Schedule API（推荐）

```go
func main() {
    c, _ := client.Dial(client.Options{})
    defer c.Close()
    
    // 创建 Schedule 客户端
    scheduleClient := client.NewScheduleClient(c)
    
    // 创建 Schedule
    schedule := &schedule.Schedule{
        Spec: schedule.Spec{
            Calendars: []schedule.CalendarSpec{
                {
                    Second:      schedule.NewRange(0, 0, 1),
                    Minute:      schedule.NewRange(0, 0, 1),
                    Hour:        schedule.NewRange(9, 17, 1),  // 9am to 5pm
                    DayOfWeek:   schedule.NewRange(1, 5, 1),   // Mon-Fri
                },
            },
        },
        Action: &schedule.StartWorkflowAction{
            ID:        "scheduled-workflow",
            Workflow:  MyWorkflow,
            TaskQueue: "my-task-queue",
        },
    }
    
    _, err := scheduleClient.Create(context.Background(), schedule.ClientOptions{
        ID: "my-schedule",
    })
    if err != nil {
        log.Fatal(err)
    }
    
    // 暂停 Schedule
    handle, _ := scheduleClient.GetHandle(context.Background(), "my-schedule")
    handle.Pause(context.Background(), "Paused for maintenance")
    
    // 恢复 Schedule
    handle.Unpause(context.Background(), "Resumed")
    
    // 删除 Schedule
    handle.Delete(context.Background())
}
```

---

## Go 代码示例

### 完整示例：订单处理工作流

```go
package workflow

import (
    "errors"
    "time"
    "go.temporal.io/sdk/temporal"
    "go.temporal.io/sdk/workflow"
)

// 订单结构
type Order struct {
    ID         string
    CustomerID string
    Items      []Item
    Total      float64
}

type Item struct {
    ProductID string
    Quantity  int
    Price     float64
}

// 订单处理工作流
func OrderProcessingWorkflow(ctx workflow.Context, order Order) error {
    logger := workflow.GetLogger(ctx)
    logger.Info("Processing order", "OrderID", order.ID)
    
    // 配置活动选项
    ao := workflow.ActivityOptions{
        StartToCloseTimeout: 5 * time.Minute,
        RetryPolicy: &temporal.RetryPolicy{
            InitialInterval:    time.Second,
            BackoffCoefficient: 2.0,
            MaximumAttempts:    3,
        },
    }
    ctx = workflow.WithActivityOptions(ctx, ao)
    
    // 第一步：验证订单
    err := workflow.ExecuteActivity(ctx, ValidateOrder, order).Get(ctx, nil)
    if err != nil {
        logger.Error("Order validation failed", "error", err)
        return err
    }
    
    // 第二步：处理支付（可取消）
    var paymentID string
    paymentFuture := workflow.ExecuteActivity(ctx, ProcessPayment, order)
    
    // 设置取消信号监听
    cancelChan := workflow.GetSignalChannel(ctx, "cancel-order")
    selector := workflow.NewSelector(ctx)
    
    var cancelled bool
    selector.AddReceive(cancelChan, func(c workflow.ReceiveChannel, more bool) {
        c.Receive(ctx, &cancelled)
    })
    selector.AddFuture(paymentFuture, func(f workflow.Future) {
        f.Get(ctx, &paymentID)
    })
    
    selector.Select(ctx)
    
    if cancelled {
        logger.Info("Order cancelled", "OrderID", order.ID)
        return errors.New("order cancelled")
    }
    
    if paymentID == "" {
        return errors.New("payment failed")
    }
    
    // 第三步：更新库存
    err = workflow.ExecuteActivity(ctx, UpdateInventory, order.Items).Get(ctx, nil)
    if err != nil {
        // 库存更新失败，需要退款
        workflow.ExecuteActivity(ctx, RefundPayment, paymentID).Get(ctx, nil)
        return err
    }
    
    // 第四步：发货
    var trackingNumber string
    err = workflow.ExecuteActivity(ctx, ShipOrder, order).Get(ctx, &trackingNumber)
    if err != nil {
        logger.Error("Shipping failed", "error", err)
        return err
    }
    
    logger.Info("Order completed", 
        "OrderID", order.ID, 
        "PaymentID", paymentID,
        "TrackingNumber", trackingNumber)
    
    return nil
}
```

### 使用子工作流

```go
func ParentWorkflow(ctx workflow.Context, orderID string) error {
    // 启动子工作流
    childOptions := workflow.ChildWorkflowOptions{
        WorkflowID: "child-" + orderID,
    }
    ctx = workflow.WithChildOptions(ctx, childOptions)
    
    var result string
    err := workflow.ExecuteChildWorkflow(ctx, ChildWorkflow, orderID).Get(ctx, &result)
    if err != nil {
        return err
    }
    
    return nil
}

func ChildWorkflow(ctx workflow.Context, orderID string) (string, error) {
    // 子工作流逻辑
    return "completed", nil
}
```

### 并行执行

```go
func ParallelWorkflow(ctx workflow.Context, items []string) error {
    ao := workflow.ActivityOptions{
        StartToCloseTimeout: time.Minute,
    }
    ctx = workflow.WithActivityOptions(ctx, ao)
    
    // 并行启动所有活动
    futures := make([]workflow.Future, len(items))
    for i, item := range items {
        futures[i] = workflow.ExecuteActivity(ctx, ProcessItem, item)
    }
    
    // 等待所有活动完成
    for i, future := range futures {
        var result string
        if err := future.Get(ctx, &result); err != nil {
            return fmt.Errorf("item %d failed: %w", i, err)
        }
    }
    
    return nil
}
```

### 使用 Selector 实现复杂逻辑

```go
func SelectorWorkflow(ctx workflow.Context) error {
    ao := workflow.ActivityOptions{
        StartToCloseTimeout: time.Minute,
    }
    ctx = workflow.WithActivityOptions(ctx, ao)
    
    // 启动多个活动
    futureA := workflow.ExecuteActivity(ctx, ActivityA)
    futureB := workflow.ExecuteActivity(ctx, ActivityB)
    
    // 设置超时
    timer := workflow.NewTimer(ctx, 30*time.Second)
    
    // 使用 Selector 选择最先完成的结果
    selector := workflow.NewSelector(ctx)
    
    var resultA, resultB string
    
    selector.AddFuture(futureA, func(f workflow.Future) {
        f.Get(ctx, &resultA)
        workflow.GetLogger(ctx).Info("ActivityA completed")
    })
    
    selector.AddFuture(futureB, func(f workflow.Future) {
        f.Get(ctx, &resultB)
        workflow.GetLogger(ctx).Info("ActivityB completed")
    })
    
    selector.AddFuture(timer, func(f workflow.Future) {
        workflow.GetLogger(ctx).Info("Timeout reached")
    })
    
    // 选择第一个完成的结果
    selector.Select(ctx)
    
    return nil
}
```

---

## 最佳实践

### 1. 工作流设计原则

- **保持工作流简洁**：将业务逻辑放在活动中
- **遵循确定性约束**：确保重放时结果一致
- **合理设置超时**：避免工作流无限期运行
- **使用 Continue-As-New**：避免事件历史过长

### 2. 错误处理

```go
func RobustWorkflow(ctx workflow.Context, input string) error {
    ao := workflow.ActivityOptions{
        StartToCloseTimeout: time.Minute,
        RetryPolicy: &temporal.RetryPolicy{
            MaximumAttempts: 3,
        },
    }
    ctx = workflow.WithActivityOptions(ctx, ao)
    
    var result string
    err := workflow.ExecuteActivity(ctx, Activity, input).Get(ctx, &result)
    if err != nil {
        // 判断错误类型
        var applicationErr *temporal.ApplicationError
        if errors.As(err, &applicationErr) {
            // 应用错误，不可重试
            return applicationErr
        }
        // 其他错误，可能是暂时性的
        return err
    }
    
    return nil
}
```

### 3. 日志和指标

```go
func LoggingWorkflow(ctx workflow.Context, input string) error {
    logger := workflow.GetLogger(ctx)
    logger.Info("Workflow started", "input", input)
    
    // 记录指标
    workflow.MetricsCounter(ctx, "workflow_starts").Inc(1)
    
    // 执行活动
    var result string
    err := workflow.ExecuteActivity(ctx, Activity, input).Get(ctx, &result)
    if err != nil {
        logger.Error("Activity failed", "error", err)
        return err
    }
    
    logger.Info("Workflow completed", "result", result)
    return nil
}
```

---

## 相关资源

- [Temporal 官方文档 - Workflow](https://docs.temporal.io/workflows)
- [Go SDK Workflow 指南](https://docs.temporal.io/dev-guide/go/workflows)
- [确定性约束详解](https://docs.temporal.io/workflow-determinism)
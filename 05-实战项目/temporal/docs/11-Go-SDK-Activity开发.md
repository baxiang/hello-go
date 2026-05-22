# Go SDK - Activity 开发

本文档详细介绍 Temporal Go SDK 中 Activity 的定义和开发。

## 11.1 Activity 定义

### 基本结构

```go
package app

import (
    "context"
)

// MyActivity 是一个活动定义
// 参数：ctx - 活动上下文（可选），input - 自定义输入
// 返回：自定义输出和错误
func MyActivity(ctx context.Context, input string) (string, error) {
    return "result: " + input, nil
}
```

### Activity 签名规则

| 规则 | 说明 |
|------|------|
| 第一个参数 | 可选 `context.Context` |
| 返回值 | `(result, error)` 或仅 `error` |
| 参数/返回值 | 必须可序列化 |
| 命名 | 导出函数（首字母大写） |

### 支持的参数类型

```go
// 基本类型
func Activity1(ctx context.Context, s string, i int, f float64) error

// 切片和映射
func Activity2(ctx context.Context, items []string, config map[string]string) error

// 结构体
type Order struct {
    ID     string
    Amount float64
}

func Activity3(ctx context.Context, order Order) error

// 指针
func Activity4(ctx context.Context, order *Order) error

// 多参数
func Activity5(ctx context.Context, a string, b int, c bool) error
```

---

## 11.2 Activity Context

### 获取信息

```go
func MyActivity(ctx context.Context) error {
    info := activity.GetInfo(ctx)
    
    activityID := info.ActivityID
    workflowID := info.WorkflowExecution.ID
    taskQueue := info.TaskQueue
    
    return nil
}
```

### 获取 Logger

```go
func MyActivity(ctx context.Context) error {
    logger := activity.GetLogger(ctx)
    logger.Info("Activity 开始执行")
    
    return nil
}
```

---

## 11.3 心跳（Heartbeat）

用于报告长时间运行 Activity 的进度：

### 发送心跳

```go
func LongRunningActivity(ctx context.Context, items []string) error {
    for i, item := range items {
        // 检查是否需要取消
        if activity.IsHeartbeatSkipped(ctx) {
            return temporal.NewCanceledError()
        }
        
        // 发送心跳，包含进度
        activity.Heartbeat(ctx, i)
        
        // 处理 item
        processItem(item)
    }
    
    return nil
}
```

### 获取心跳详情

```go
func ResumableActivity(ctx context.Context, items []string) error {
    // 获取上次心跳的进度
    var lastIndex int
    if info := activity.GetInfo(ctx); info.HeartbeatDetails != nil {
        info.HeartbeatDetails.Unmarshal(&lastIndex)
    }
    
    // 从上次进度继续
    for i := lastIndex; i < len(items); i++ {
        activity.Heartbeat(ctx, i)
        processItem(items[i])
    }
    
    return nil
}
```

### 心跳超时配置

```go
// Workflow 中配置
ao := workflow.ActivityOptions{
    StartToCloseTimeout: time.Hour,
    HeartbeatTimeout:    30 * time.Second, // 心跳超时
}
```

---

## 11.4 错误处理

### 返回应用错误

```go
func MyActivity(ctx context.Context, orderID string) error {
    order, err := getOrder(orderID)
    if err != nil {
        return err
    }
    
    if order.Status == "cancelled" {
        // 返回应用错误，不重试
        return temporal.NewApplicationError("订单已取消", "OrderCancelled")
    }
    
    return nil
}
```

### 返回带详情的错误

```go
func MyActivity(ctx context.Context) error {
    return temporal.NewApplicationError(
        "处理失败",
        "ProcessingFailed",
        map[string]interface{}{
            "reason":  "insufficient_balance",
            "balance": 100,
            "required": 200,
        },
    )
}
```

### 错误类型判断

```go
func MyWorkflow(ctx workflow.Context) error {
    err := workflow.ExecuteActivity(ctx, MyActivity).Get(ctx, nil)
    if err != nil {
        var appErr *temporal.ApplicationError
        if errors.As(err, &appErr) {
            if appErr.Type() == "OrderCancelled" {
                // 特定错误处理
                return nil
            }
        }
        return err
    }
    return nil
}
```

---

## 11.5 重试策略

### 配置重试

```go
ao := workflow.ActivityOptions{
    StartToCloseTimeout: time.Minute,
    RetryPolicy: &temporal.RetryPolicy{
        InitialInterval:    time.Second,      // 首次重试间隔
        BackoffCoefficient: 2.0,              // 退避系数
        MaximumInterval:    time.Minute,      // 最大间隔
        MaximumAttempts:    5,                // 最大重试次数
        NonRetryableErrorTypes: []string{     // 不重试的错误类型
            "OrderCancelled",
            "InvalidInput",
        },
    },
}
```

### 重试策略参数详解

| 参数 | 默认值 | 说明 |
|------|--------|------|
| `InitialInterval` | 1s | 首次重试等待时间 |
| `BackoffCoefficient` | 2.0 | 每次重试间隔倍数 |
| `MaximumInterval` | 100x Initial | 最大等待时间 |
| `MaximumAttempts` | 0（无限） | 最大重试次数 |
| `NonRetryableErrorTypes` | [] | 不重试的错误类型 |

### 重试时间计算

```
第1次重试: InitialInterval = 1s
第2次重试: 1s * 2.0 = 2s
第3次重试: 2s * 2.0 = 4s
第4次重试: 4s * 2.0 = 8s
...
最大不超过 MaximumInterval
```

---

## 11.6 超时配置

### 超时类型

```go
ao := workflow.ActivityOptions{
    ScheduleToCloseTimeout: time.Minute * 10,  // 总超时（从调度到完成）
    ScheduleToStartTimeout: time.Minute,       // 调度超时（从调度到开始）
    StartToCloseTimeout:   time.Minute * 5,    // 执行超时（从开始到完成）
    HeartbeatTimeout:      time.Second * 30,   // 心跳超时
}
```

### 超时说明

```
┌─────────────────────────────────────────────────────────────┐
│                    ScheduleToCloseTimeout                   │
│  ┌─────────────────────┬────────────────────────────────┐  │
│  │ ScheduleToStart     │       StartToClose            │  │
│  │    Timeout          │        Timeout                │  │
│  │ ─────────────────►  │  ────────────────────────►   │  │
│  │                     │                               │  │
│  │  等待 Worker 接收   │    Activity 执行时间          │  │
│  └─────────────────────┴────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
       │                  │                    │
    调度时间            开始执行             执行完成
```

### 推荐配置

| 场景 | ScheduleToClose | StartToClose | Heartbeat |
|------|-----------------|--------------|-----------|
| 快速 API 调用 | 30s | 10s | - |
| 数据库操作 | 1min | 30s | - |
| 文件处理 | 10min | 5min | 30s |
| 长时间任务 | 1hour | 30min | 1min |

---

## 11.7 Local Activity

Local Activity 在 Workflow 进程内执行，适用于：

- 快速、确定性的操作
- 不需要重试策略
- 不需要心跳

### 定义 Local Activity

```go
func MyWorkflow(ctx workflow.Context) error {
    // Local Activity 选项
    lao := workflow.LocalActivityOptions{
        StartToCloseTimeout: time.Second * 10,
    }
    ctx = workflow.WithLocalActivityOptions(ctx, lao)
    
    // 执行 Local Activity
    var result string
    err := workflow.ExecuteLocalActivity(ctx, LocalActivity, "input").Get(ctx, &result)
    return err
}

// Local Activity 定义（更快的操作）
func LocalActivity(ctx context.Context, input string) (string, error) {
    return strings.ToUpper(input), nil
}
```

### Local Activity vs Remote Activity

| 特性 | Local Activity | Remote Activity |
|------|---------------|-----------------|
| 执行位置 | Workflow 进程内 | 独立 Worker |
| 延迟 | 更低 | 更高 |
| 重试策略 | 有限 | 完整支持 |
| 心跳 | 不支持 | 支持 |
| 适用场景 | 快速、确定性操作 | 外部调用、长时间操作 |

---

## 11.8 幂等性设计

Activity 应该设计为幂等的，因为可能会被多次执行：

```go
// ✅ 幂等的 Activity
func ProcessPaymentActivity(ctx context.Context, paymentID string, amount float64) error {
    // 使用唯一 ID 检查是否已处理
    if isProcessed(paymentID) {
        return nil // 已处理，直接返回成功
    }
    
    // 处理支付
    if err := chargePayment(paymentID, amount); err != nil {
        return err
    }
    
    // 标记为已处理
    markAsProcessed(paymentID)
    return nil
}
```

### 幂等性策略

| 策略 | 说明 |
|------|------|
| 唯一 ID | 使用业务唯一 ID 检查是否已处理 |
| 状态检查 | 检查目标状态是否已是期望状态 |
| 事务 | 使用数据库事务保证原子性 |
| 天然幂等 | GET 请求、设置相同值等 |

---

## 11.9 最佳实践

### 1. Activity 应该做一件事

```go
// ❌ 不好：Activity 做太多事情
func ProcessOrderActivity(ctx context.Context, order Order) error {
    validateOrder(order)
    chargePayment(order)
    sendEmail(order)
    updateInventory(order)
    return nil
}

// ✅ 好：拆分成多个 Activity
func ValidateOrderActivity(ctx context.Context, order Order) error
func ChargePaymentActivity(ctx context.Context, order Order) error
func SendEmailActivity(ctx context.Context, order Order) error
func UpdateInventoryActivity(ctx context.Context, order Order) error
```

### 2. 合理设置超时

```go
// 根据实际执行时间设置
ao := workflow.ActivityOptions{
    StartToCloseTimeout: expectedTime * 2, // 设置为预期的 2 倍
    HeartbeatTimeout:    expectedTime / 4, // 心跳为预期的 1/4
}
```

### 3. 提供有意义的错误

```go
func MyActivity(ctx context.Context, orderID string) error {
    order, err := getOrder(orderID)
    if err != nil {
        return fmt.Errorf("获取订单失败 [%s]: %w", orderID, err)
    }
    return nil
}
```

---

## 下一步

- [12-Go-SDK-高级特性](./12-Go-SDK-高级特性.md) - Signal、Query 等高级特性
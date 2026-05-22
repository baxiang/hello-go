# Go SDK - Workflow 开发

本文档详细介绍 Temporal Go SDK 中 Workflow 的定义和开发。

## 10.1 Workflow 定义

### 基本结构

```go
package app

import (
    "time"
    "go.temporal.io/sdk/workflow"
)

// MyWorkflow 是一个工作流定义
// 参数：ctx - 工作流上下文，input - 自定义输入
// 返回：自定义输出和错误
func MyWorkflow(ctx workflow.Context, input string) (string, error) {
    // 工作流逻辑
    return "result", nil
}
```

### Workflow 签名规则

| 规则 | 说明 |
|------|------|
| 第一个参数 | 必须是 `workflow.Context` |
| 返回值 | `(result, error)` 格式 |
| 参数/返回值 | 必须可序列化 |
| 命名 | 导出函数（首字母大写） |

---

## 10.2 Workflow Context

### 上下文功能

```go
func MyWorkflow(ctx workflow.Context, input string) error {
    // 获取 Logger
    logger := workflow.GetLogger(ctx)
    logger.Info("Workflow 开始")
    
    // 获取信息
    info := workflow.GetInfo(ctx)
    workflowID := info.WorkflowExecution.ID
    runID := info.WorkflowExecution.RunID
    namespace := info.Namespace
    
    return nil
}
```

### 设置 Activity 选项

```go
func MyWorkflow(ctx workflow.Context) error {
    ao := workflow.ActivityOptions{
        StartToCloseTimeout: 10 * time.Second,
        RetryPolicy: &temporal.RetryPolicy{
            InitialInterval:    time.Second,
            BackoffCoefficient: 2.0,
            MaximumInterval:    time.Minute,
            MaximumAttempts:    5,
        },
    }
    ctx = workflow.WithActivityOptions(ctx, ao)
    
    // 使用配置后的 ctx 执行 Activity
    return nil
}
```

### 设置 Local Activity 选项

```go
func MyWorkflow(ctx workflow.Context) error {
    lao := workflow.LocalActivityOptions{
        StartToCloseTimeout: 5 * time.Second,
    }
    ctx = workflow.WithLocalActivityOptions(ctx, lao)
    
    return nil
}
```

---

## 10.3 执行 Activity

### 同步执行

```go
func MyWorkflow(ctx workflow.Context, name string) (string, error) {
    var result string
    err := workflow.ExecuteActivity(ctx, MyActivity, name).Get(ctx, &result)
    if err != nil {
        return "", err
    }
    return result, nil
}
```

### 多个 Activity 串行执行

```go
func MyWorkflow(ctx workflow.Context) error {
    // 第一个 Activity
    var result1 string
    err := workflow.ExecuteActivity(ctx, Step1).Get(ctx, &result1)
    if err != nil {
        return err
    }
    
    // 第二个 Activity（依赖第一个的结果）
    var result2 string
    err = workflow.ExecuteActivity(ctx, Step2, result1).Get(ctx, &result2)
    if err != nil {
        return err
    }
    
    return nil
}
```

### 并行执行多个 Activity

```go
func MyWorkflow(ctx workflow.Context) error {
    // 启动多个 Activity
    future1 := workflow.ExecuteActivity(ctx, Activity1)
    future2 := workflow.ExecuteActivity(ctx, Activity2)
    future3 := workflow.ExecuteActivity(ctx, Activity3)
    
    // 等待所有完成
    var result1, result2, result3 string
    if err := future1.Get(ctx, &result1); err != nil {
        return err
    }
    if err := future2.Get(ctx, &result2); err != nil {
        return err
    }
    if err := future3.Get(ctx, &result3); err != nil {
        return err
    }
    
    return nil
}
```

### 使用 Selector 等待任意完成

```go
func MyWorkflow(ctx workflow.Context) error {
    future1 := workflow.ExecuteActivity(ctx, FastActivity)
    future2 := workflow.ExecuteActivity(ctx, SlowActivity)
    
    selector := workflow.NewSelector(ctx)
    
    var result string
    selector.AddFuture(future1, func(f workflow.Future) {
        f.Get(ctx, &result)
    })
    selector.AddFuture(future2, func(f workflow.Future) {
        f.Get(ctx, &result)
    })
    
    selector.Select(ctx) // 等待第一个完成
    
    return nil
}
```

---

## 10.4 定时器和延迟

### 简单延迟

```go
func MyWorkflow(ctx workflow.Context) error {
    // 等待 1 分钟
    workflow.Sleep(ctx, time.Minute)
    
    // 或使用 Timer
    timer := workflow.NewTimer(ctx, time.Minute)
    if err := timer.Get(ctx, nil); err != nil {
        return err
    }
    
    return nil
}
```

### 带超时的 Activity

```go
func MyWorkflow(ctx workflow.Context) error {
    ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
        StartToCloseTimeout: 10 * time.Second,
    })
    
    // 带超时控制
    future := workflow.ExecuteActivity(ctx, MyActivity)
    
    selector := workflow.NewSelector(ctx)
    selector.AddFuture(future, func(f workflow.Future) {
        var result string
        f.Get(ctx, &result)
    })
    selector.AddFuture(workflow.NewTimer(ctx, 5*time.Second), func(f workflow.Future) {
        // 超时处理
    })
    
    selector.Select(ctx)
    return nil
}
```

---

## 10.5 条件等待

### 等待多个条件

```go
func MyWorkflow(ctx workflow.Context) error {
    signalCh := workflow.GetSignalChannel(ctx, "my-signal")
    timerCh := workflow.NewTimer(ctx, time.Hour)
    
    var received bool
    selector := workflow.NewSelector(ctx)
    
    selector.AddReceive(signalCh, func(c workflow.ReceiveChannel, more bool) {
        var signal string
        c.Receive(ctx, &signal)
        received = true
    })
    
    selector.AddFuture(timerCh, func(f workflow.Future) {
        // 超时，未收到信号
        received = false
    })
    
    selector.Select(ctx)
    
    if !received {
        return fmt.Errorf("等待超时")
    }
    
    return nil
}
```

---

## 10.6 Child Workflow

### 启动子工作流

```go
func ParentWorkflow(ctx workflow.Context) error {
    childOpts := workflow.ChildWorkflowOptions{
        WorkflowID: "child-" + workflow.GetInfo(ctx).WorkflowExecution.ID,
    }
    ctx = workflow.WithChildOptions(ctx, childOpts)
    
    var result string
    err := workflow.ExecuteChildWorkflow(ctx, ChildWorkflow, "input").Get(ctx, &result)
    return err
}
```

### 等待子工作流完成

```go
func ParentWorkflow(ctx workflow.Context) error {
    future := workflow.ExecuteChildWorkflow(ctx, ChildWorkflow, "input")
    
    // 可以先做其他事情...
    
    var result string
    if err := future.Get(ctx, &result); err != nil {
        return err
    }
    
    return nil
}
```

---

## 10.7 确定性约束

### 必须遵守的规则

```go
// ❌ 禁止：使用 time.Now()
t := time.Now() // 非确定性！

// ✅ 正确：使用 workflow.Now()
t := workflow.Now(ctx)

// ❌ 禁止：使用 time.Sleep()
time.Sleep(time.Second) // 非确定性！

// ✅ 正确：使用 workflow.Sleep()
workflow.Sleep(ctx, time.Second)

// ❌ 禁止：使用随机数
rand.Intn(100) // 非确定性！

// ✅ 正确：使用确定性值或从输入获取

// ❌ 禁止：使用全局变量
var counter int // 不安全！

// ✅ 正确：使用 Workflow 状态
counter := 0
```

### 可用的确定性 API

| API | 用途 |
|-----|------|
| `workflow.Now()` | 获取当前时间 |
| `workflow.Sleep()` | 延迟 |
| `workflow.GetLogger()` | 日志 |
| `workflow.GetInfo()` | 工作流信息 |
| `workflow.SideEffect()` | 执行非确定性操作 |

---

## 10.8 SideEffect

用于执行非确定性操作但保存结果：

```go
func MyWorkflow(ctx workflow.Context) error {
    // 生成唯一 ID（非确定性操作）
    var uniqueID string
    workflow.SideEffect(ctx, func(ctx workflow.Context) interface{} {
        return uuid.New().String()
    }).Get(&uniqueID)
    
    // uniqueID 在重放时保持相同值
    workflow.GetLogger(ctx).Info("Unique ID", "id", uniqueID)
    
    return nil
}
```

---

## 10.9 最佳实践

### 1. 保持 Workflow 简洁

```go
// ✅ 好的做法：Workflow 只编排逻辑
func OrderWorkflow(ctx workflow.Context, orderID string) error {
    // 验证
    if err := workflow.ExecuteActivity(ctx, ValidateOrder, orderID).Get(ctx, nil); err != nil {
        return err
    }
    
    // 处理
    if err := workflow.ExecuteActivity(ctx, ProcessOrder, orderID).Get(ctx, nil); err != nil {
        return err
    }
    
    // 通知
    return workflow.ExecuteActivity(ctx, SendNotification, orderID).Get(ctx, nil)
}
```

### 2. 使用版本控制

```go
func MyWorkflow(ctx workflow.Context) error {
    v := workflow.GetVersion(ctx, "add-step", workflow.DefaultVersion, 1)
    if v == 1 {
        // 新增的步骤
        workflow.ExecuteActivity(ctx, NewStep).Get(ctx, nil)
    }
    return nil
}
```

### 3. 正确处理错误

```go
func MyWorkflow(ctx workflow.Context) error {
    var result string
    err := workflow.ExecuteActivity(ctx, MyActivity).Get(ctx, &result)
    if err != nil {
        var appErr *temporal.ApplicationError
        if errors.As(err, &appErr) {
            // 应用错误，可以重试或跳过
            return nil
        }
        return err
    }
    return nil
}
```

---

## 下一步

- [11-Go-SDK-Activity开发](./11-Go-SDK-Activity开发.md) - 学习 Activity 开发
- [12-Go-SDK-高级特性](./12-Go-SDK-高级特性.md) - Signal、Query 等高级特性
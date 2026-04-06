# Go SDK - 高级特性

本文档介绍 Temporal Go SDK 的高级特性，包括 Signal、Query、Interceptor 等。

## 12.1 Signal

Signal 用于从外部向运行中的 Workflow 发送消息。

### 定义 Signal 处理

```go
func OrderWorkflow(ctx workflow.Context, orderID string) error {
    // Signal 通道
    cancelCh := workflow.GetSignalChannel(ctx, "cancel-order")
    updateCh := workflow.GetSignalChannel(ctx, "update-order")
    
    // 状态
    var cancelled bool
    var updates []string
    
    // 处理循环
    for {
        selector := workflow.NewSelector(ctx)
        
        // 处理取消信号
        selector.AddReceive(cancelCh, func(c workflow.ReceiveChannel, more bool) {
            var reason string
            c.Receive(ctx, &reason)
            cancelled = true
            workflow.GetLogger(ctx).Info("订单取消", "reason", reason)
        })
        
        // 处理更新信号
        selector.AddReceive(updateCh, func(c workflow.ReceiveChannel, more bool) {
            var update string
            c.Receive(ctx, &update)
            updates = append(updates, update)
        })
        
        // 处理其他逻辑...
        selector.AddFuture(workflow.NewTimer(ctx, time.Second), func(f workflow.Future) {
            // 定时检查
        })
        
        selector.Select(ctx)
        
        if cancelled {
            return workflow.ExecuteActivity(ctx, CancelOrder, orderID).Get(ctx, nil)
        }
    }
}
```

### 发送 Signal

```go
// 从客户端发送
err := c.SignalWorkflow(ctx, workflowID, "", "cancel-order", "用户取消")

// 在 Activity 中发送（不推荐）
func NotifyActivity(ctx context.Context, workflowID string) error {
    // 需要 Client
    c.SignalWorkflow(ctx, workflowID, "", "notification-sent", true)
    return nil
}
```

### Signal 批量处理

```go
func MyWorkflow(ctx workflow.Context) error {
    signalCh := workflow.GetSignalChannel(ctx, "items")
    var items []string
    
    for {
        // 批量接收所有待处理的 Signal
        for signalCh.Len() > 0 {
            var item string
            signalCh.ReceiveAsync(&item)
            items = append(items, item)
        }
        
        // 处理批量
        if len(items) > 0 {
            workflow.ExecuteActivity(ctx, ProcessBatch, items).Get(ctx, nil)
            items = nil
        }
        
        workflow.Sleep(ctx, time.Minute)
    }
}
```

---

## 12.2 Query

Query 用于查询 Workflow 状态，不修改状态。

### 定义 Query Handler

```go
func OrderWorkflow(ctx workflow.Context, orderID string) error {
    // 状态
    status := "pending"
    var items []string
    
    // 注册 Query Handler
    err := workflow.SetQueryHandler(ctx, "status", func() (string, error) {
        return status, nil
    })
    if err != nil {
        return err
    }
    
    err = workflow.SetQueryHandler(ctx, "items", func() ([]string, error) {
        return items, nil
    })
    if err != nil {
        return err
    }
    
    // Workflow 逻辑更新 status...
    status = "processing"
    
    return nil
}
```

### 执行 Query

```go
// 查询状态
resp, err := c.QueryWorkflow(ctx, workflowID, "", "status")
var status string
resp.Get(&status)

// 查询 items
resp, err = c.QueryWorkflow(ctx, workflowID, "", "items")
var items []string
resp.Get(&items)
```

### Query 最佳实践

```go
// 返回结构化状态
type OrderState struct {
    Status    string
    Items     []string
    CreatedAt time.Time
    UpdatedAt time.Time
}

err := workflow.SetQueryHandler(ctx, "state", func() (OrderState, error) {
    return OrderState{
        Status:    status,
        Items:     items,
        CreatedAt: createdAt,
        UpdatedAt: workflow.Now(ctx),
    }, nil
})
```

---

## 12.3 Update

Update 是较新的特性，允许修改 Workflow 状态并返回结果。

### 定义 Update Handler

```go
func OrderWorkflow(ctx workflow.Context, orderID string) error {
    var items []string
    
    // 注册 Update Handler
    workflow.SetUpdateHandler(ctx, "add-item", func(ctx workflow.Context, item string) error {
        items = append(items, item)
        return nil
    })
    
    workflow.SetUpdateHandler(ctx, "remove-item", func(ctx workflow.Context, index int) error {
        if index >= 0 && index < len(items) {
            items = append(items[:index], items[index+1:]...)
        }
        return nil
    })
    
    // 等待完成信号
    workflow.GetSignalChannel(ctx, "complete").Receive(ctx, nil)
    return nil
}
```

### 执行 Update

```go
// 添加 item
err := c.UpdateWorkflow(ctx, workflowID, "", "add-item", "item-1")

// 删除 item
err = c.UpdateWorkflow(ctx, workflowID, "", "remove-item", 0)
```

---

## 12.4 Continue As New

长时间运行的 Workflow 可以使用 `Continue As New` 来重置历史：

```go
func LongRunningWorkflow(ctx workflow.Context, state State) error {
    // 处理一批工作
    for i := 0; i < 1000; i++ {
        workflow.ExecuteActivity(ctx, ProcessItem, i).Get(ctx, nil)
    }
    
    // 更新状态
    state.Processed += 1000
    
    // 历史过长时，Continue As New
    info := workflow.GetInfo(ctx)
    if info.GetCurrentHistoryLength() > 40000 {
        return workflow.NewContinueAsNewError(ctx, LongRunningWorkflow, state)
    }
    
    // 继续处理
    return nil
}
```

---

## 12.5 Interceptor

Interceptor 用于拦截 Workflow 和 Activity 的执行。

### 实现 Interceptor

```go
type MyInterceptor struct {
    interceptor.WorkflowInterceptorBase
}

func (i *MyInterceptor) InterceptWorkflow(
    ctx workflow.Context,
    in interceptor.WorkflowInput,
    next interceptor.WorkflowInboundInterceptor,
) (interceptor.WorkflowOutput, error) {
    // 前置处理
    workflow.GetLogger(ctx).Info("Workflow 开始", "type", in.WorkflowType)
    
    // 执行
    out, err := next.InterceptWorkflow(ctx, in)
    
    // 后置处理
    if err != nil {
        workflow.GetLogger(ctx).Error("Workflow 失败", "error", err)
    }
    
    return out, err
}
```

### 注册 Interceptor

```go
w := worker.New(c, taskQueue, worker.Options{
    Interceptors: []interceptor.WorkflowInterceptor{
        &MyInterceptor{},
    },
})
```

---

## 12.6 Metrics

Temporal SDK 内置 Prometheus 指标支持。

### 配置 Metrics

```go
import (
    "go.temporal.io/sdk/worker"
    "go.temporal.io/sdk/contrib/tally"
)

// 创建 Tally Scope
scope := tally.NewScope()

w := worker.New(c, taskQueue, worker.Options{
    MetricsHandler: tally.NewMetricsHandler(scope),
})
```

### 自定义指标

```go
func MyWorkflow(ctx workflow.Context) error {
    // 获取 Metrics Handler
    metrics := workflow.GetMetricsHandler(ctx)
    
    // 计数器
    metrics.Counter("workflow_executions").Inc(1)
    
    // 计时器
    timer := metrics.Timer("processing_time")
    timer.Start()
    defer timer.Stop()
    
    // Gauge
    metrics.Gauge("items_processed").Update(float64(itemsCount))
    
    return nil
}
```

---

## 12.7 数据转换

自定义数据序列化。

### 实现 Codec

```go
type JSONCodec struct{}

func (c *JSONCodec) Encode(value interface{}) ([]byte, error) {
    return json.Marshal(value)
}

func (c *JSONCodec) Decode(data []byte, valuePtr interface{}) error {
    return json.Unmarshal(data, valuePtr)
}
```

### 注册 Codec

```go
c, err := client.Dial(client.Options{
    DataConverter: converter.NewCompositeDataConverter(
        &JSONCodec{},
    ),
})
```

---

## 12.8 生产配置建议

### Worker 配置

```go
w := worker.New(c, taskQueue, worker.Options{
    // 并发控制
    MaxConcurrentWorkflowTaskExecutionSize:     100,
    MaxConcurrentActivityExecutionSize:         1000,
    MaxConcurrentLocalActivityExecutionSize:    1000,
    
    // 心跳
    DisableWorkflowWorker:                      false,
    DisableActivityWorker:                      false,
    
    // 标识
    Identity: hostname + "-" + workerID,
    
    // 资源限制
    MaxConcurrentWorkflowTaskPollers:           5,
    MaxConcurrentActivityTaskPollers:           10,
})
```

### 客户端配置

```go
c, err := client.Dial(client.Options{
    HostPort:  "localhost:7233",
    Namespace: "production",
    
    ConnectionOptions: client.ConnectionOptions{
        DialTimeout:           10 * time.Second,
        MaxConcurrentStreams:  1000,
    },
    
    Logger:    zap.NewStdoutLogger(),
    MetricsHandler: metricsHandler,
})
```

---

## 参考

- [Temporal Go SDK 文档](https://docs.temporal.io/dev-guide/go)
- [Go SDK API 参考](https://pkg.go.dev/go.temporal.io/sdk)
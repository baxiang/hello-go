# Temporal Go 客户端与使用

本节详细介绍 Temporal Go SDK 的使用，包括客户端连接、工作流和活动的定义、Worker 的配置以及各种高级特性。

## 2.1 安装与依赖

### 安装 SDK

```bash
go get go.temporal.io/sdk-go
```

### 项目初始化

```bash
# 创建项目
mkdir my-temporal-app && cd my-temporal-app
go mod init my-temporal-app

# 添加依赖
go get go.temporal.io/sdk-go
go get go.temporal.io/sdk-go/client
go get go.temporal.io/sdk-go/worker
go get go.temporal.io/sdk-go/workflow
```

---

## 2.2 客户端连接

### 2.2.1 基本连接

```go
package main

import (
    "context"
    "fmt"
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

    fmt.Println("已连接到 Temporal 服务器")
}
```

### 2.2.2 连接选项

```go
// 完整连接配置
c, err := client.Dial(client.Options{
    // 服务器地址
    HostPort: "localhost:7233",
    
    // 命名空间
    Namespace: "default",
    
    // TLS 配置
    TLSConfig: &tls.Config{
        InsecureSkipVerify: true,
    },
    
    // 连接超时
    ConnectionTimeout: 10 * time.Second,
    
    // 队列操作超时
    QueueOperationTimeout: 5 * time.Second,
    
    // 日志
    Logger: log.New(os.Stdout, "temporal", log.LstdFlags),
    
    // 拦截器
    Interceptors: []client.Interceptor{},
})
```

### 2.2.3 连接池

```go
// 创建多个客户端用于连接池
type ClientPool struct {
    clients []*client.Client
    index   int
    mu      sync.Mutex
}

func NewClientPool(size int, opts client.Options) (*ClientPool, error) {
    clients := make([]*client.Client, size)
    for i := 0; i < size; i++ {
        c, err := client.Dial(opts)
        if err != nil {
            return nil, err
        }
        clients[i] = c
    }
    
    return &ClientPool{clients: clients}, nil
}

func (p *ClientPool) Get() *client.Client {
    p.mu.Lock()
    defer p.mu.Unlock()
    
    c := p.clients[p.index]
    p.index = (p.index + 1) % len(p.clients)
    return c
}

func (p *ClientPool) Close() {
    for _, c := range p.clients {
        c.Close()
    }
}
```

---

## 2.3 工作流定义

### 2.3.1 基本工作流

```go
package main

import (
    "time"

    "go.temporal.io/sdk/workflow"
)

// MyWorkflow 工作流函数签名必须是 func(ctx workflow.Context, input InputType) (OutputType, error)
func MyWorkflow(ctx workflow.Context, name string) (string, error) {
    logger := workflow.GetLogger(ctx)
    
    logger.Info("工作流开始", "name", name)
    
    // 模拟处理
    time.Sleep(100 * time.Millisecond)
    
    result := "Hello, " + name
    logger.Info("工作流完成", "result", result)
    
    return result, nil
}
```

### 2.3.2 带参数的工作流

```go
// 定义输入输出类型
type OrderInput struct {
    OrderID   string
    CustomerID string
    Items     []Item
}

type Item struct {
    ProductID string
    Quantity int
    Price    float64
}

type OrderOutput struct {
    OrderID    string
    Status     string
    TotalAmount float64
}

// 工作流定义
func ProcessOrderWorkflow(ctx workflow.Context, input OrderInput) (*OrderOutput, error) {
    logger := workflow.GetLogger(ctx)
    
    // 验证订单
    if len(input.Items) == 0 {
        return nil, fmt.Errorf("订单为空")
    }
    
    // 计算总价
    var total float64
    for _, item := range input.Items {
        total += item.Price * float64(item.Quantity)
    }
    
    // 调用活动处理支付
    var paymentResult PaymentResult
    err := workflow.ExecuteActivity(ctx, ProcessPayment, PaymentRequest{
        OrderID: input.OrderID,
        Amount:  total,
    }).Get(ctx, &paymentResult)
    
    if err != nil {
        return nil, err
    }
    
    return &OrderOutput{
        OrderID:     input.OrderID,
        Status:      paymentResult.Status,
        TotalAmount: total,
    }, nil
}
```

### 2.3.3 工作流上下文

```go
func WorkflowWithContext(ctx workflow.Context) error {
    // 获取工作流信息
    info := workflow.GetInfo(ctx)
    
    logger := workflow.GetLogger(ctx)
    logger.Info("工作流信息",
        "WorkflowID", info.WorkflowExecution.ID,
        "RunID", info.WorkflowExecution.RunID,
        "TaskQueue", info.TaskQueueName,
    )
    
    // 获取历史数据
    var previousResult string
    if err := workflow.GetPreviousActivityResult(ctx, &previousResult); err != nil {
        logger.Info("没有之前的活动结果")
    }
    
    return nil
}
```

---

## 2.4 活动定义

### 2.4.1 基本活动

```go
package main

import (
    "context"
    "fmt"
    "log"

    "go.temporal.io/sdk/activity"
)

// 活动函数签名必须是 func(ctx context.Context, input InputType) (OutputType, error)
func ProcessPayment(ctx context.Context, req PaymentRequest) (*PaymentResult, error) {
    logger := activity.GetLogger(ctx)
    
    logger.Info("开始处理支付", "orderID", req.OrderID, "amount", req.Amount)
    
    // 业务逻辑
    // ...
    
    return &PaymentResult{
        Status: "success",
        TransactionID: "TXN" + req.OrderID,
    }, nil
}
```

### 2.4.2 带心跳的活动

```go
func LongRunningTask(ctx context.Context, taskID string) error {
    logger := activity.GetLogger(ctx)
    
    progress := 0
    for progress < 100 {
        // 检查取消
        if ctx.Err() != nil {
            logger.Info("活动被取消")
            return ctx.Err()
        }
        
        // 模拟处理
        processChunk(taskID, progress)
        progress += 10
        
        // 报告进度（心跳）
        activity.RecordHeartbeat(ctx, progress)
        logger.Info("进度", "progress", progress)
    }
    
    logger.Info("任务完成", "taskID", taskID)
    return nil
}

func processChunk(taskID string, progress int) {
    // 模拟耗时操作
}
```

### 2.4.3 幂等活动

```go
func IdempotentActivity(ctx context.Context, req Request) error {
    logger := activity.GetLogger(ctx)
    
    // 检查是否已完成
    if activity.HasHeartbeatDetails(ctx) {
        var completed bool
        if err := activity.GetHeartbeatDetails(ctx, &completed); err == nil && completed {
            logger.Info("活动已完成，跳过", "requestID", req.ID)
            return nil
        }
    }
    
    // 业务处理
    result, err := processRequest(req)
    if err != nil {
        return err
    }
    
    // 记录完成状态
    activity.RecordHeartbeat(ctx, true)
    
    logger.Info("活动完成", "requestID", req.ID, "result", result)
    return nil
}

func processRequest(req Request) (string, error) {
    // 处理逻辑
    return "processed", nil
}
```

---

## 2.5 Worker 配置

### 2.5.1 基本 Worker

```go
package main

import (
    "log"

    "go.temporal.io/sdk/client"
    "go.temporal.io/sdk/worker"
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
    
    // 创建 Worker
    w := worker.New(c, "my-task-queue", worker.Options{
        // Worker 配置
    })
    
    // 注册工作流
    w.RegisterWorkflow(MyWorkflow)
    w.RegisterWorkflow(ProcessOrderWorkflow)
    
    // 注册活动
    w.RegisterActivity(ProcessPayment)
    w.RegisterActivity(SendNotification)
    
    // 启动 Worker
    if err := w.Start(); err != nil {
        log.Fatalln("无法启动 Worker", err)
    }
    
    log.Println("Worker 已启动，按 Ctrl+C 退出")
    select {}
}
```

### 2.5.2 Worker 选项

```go
w := worker.New(c, "my-task-queue", worker.Options{
    // 并发数
    MaxConcurrentWorkflowTaskExecutionSize: 100,
    MaxConcurrentActivityExecutionSize: 50,
    MaxConcurrentLocalActivityExecutionSize: 20,
    
    // 任务轮询
    MaxConcurrentWorkflowTaskPollers: 2,
    MaxConcurrentActivityTaskPollers: 2,
    
    // 允许使用默认的本地活动执行器
    EnableLocalActivityWorker: true,
    
    // 拦截器
    WorkflowInterceptorChainFactories: []workflow.WorkflowInterceptorFactory{},
    ActivityInterceptorChainFactories: []activity.ActivityInterceptorFactory{},
    
    // 上下文传播
    ContextPropagators: []context.ContextPropagator{},
    
    // 错误处理
    PanicErrorHandler: func(err error) {
        log.Printf("工作流 panic: %v", err)
    },
})
```

### 2.5.3 多个 Worker

```go
// 为不同的任务队列创建不同的 Worker
func startWorkers(c client.Client) {
    // 订单处理 Worker
    orderWorker := worker.New(c, "order-task-queue", worker.Options{})
    orderWorker.RegisterWorkflow(OrderWorkflow)
    orderWorker.RegisterActivity(ProcessOrderActivity)
    orderWorker.Start()
    
    // 支付 Worker
    paymentWorker := worker.New(c, "payment-task-queue", worker.Options{})
    paymentWorker.RegisterActivity(ProcessPaymentActivity)
    paymentWorker.Start()
    
    // 通知 Worker
    notificationWorker := worker.New(c, "notification-task-queue", worker.Options{})
    notificationWorker.RegisterActivity(SendEmailActivity)
    notificationWorker.RegisterActivity(SendSMSActivity)
    notificationWorker.Start()
}
```

---

## 2.6 工作流执行

### 2.6.1 启动工作流

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    "go.temporal.io/sdk/client"
)

func main() {
    c, err := client.Dial(client.Options{HostPort: "localhost:7233"})
    if err != nil {
        log.Fatalln("无法创建客户端", err)
    }
    defer c.Close()
    
    // 启动工作流
    workflowID := "order-" + time.Now().Format("20060102150405")
    
    we, err := c.ExecuteWorkflow(
        context.Background(),
        client.StartWorkflowOptions{
            ID:        workflowID,
            TaskQueue: "my-task-queue",
        },
        ProcessOrderWorkflow,
        OrderInput{
            OrderID:   "ORD-001",
            CustomerID: "CUST-001",
            Items: []Item{
                {ProductID: "PROD-001", Quantity: 2, Price: 99.99},
            },
        },
    )
    if err != nil {
        log.Fatalln("无法启动工作流", err)
    }
    
    fmt.Printf("工作流已启动: %s\n", we.GetID())
    
    // 等待结果
    var result OrderOutput
    err = we.Get(context.Background(), &result)
    if err != nil {
        log.Fatalln("获取结果失败", err)
    }
    
    fmt.Printf("工作流完成: %+v\n", result)
}
```

### 2.6.2 信号工作流

```go
// 工作流端
func SignalWorkflow(ctx workflow.Context) error {
    logger := workflow.GetLogger(ctx)
    
    // 接收信号
    var signalData string
    workflow.GetSignalChannel(ctx, "my-signal").Receive(ctx, &signalData)
    
    logger.Info("收到信号", "data", signalData)
    
    return nil
}

// 客户端发送信号
err := c.SignalWorkflow(
    context.Background(),
    "workflow-id",
    "",
    "my-signal",
    "signal-data",
)
```

### 2.6.3 查询工作流

```go
// 工作流端
func QueryWorkflow(ctx workflow.Context) error {
    workflow.SetQueryHandler(ctx, "getStatus", func() (string, error) {
        return "processing", nil
    })
    
    workflow.SetQueryHandler(ctx, "getProgress", func() (int, error) {
        return 50, nil
    })
    
    // ...
}

// 客户端查询
resp, err := c.QueryWorkflow(
    context.Background(),
    "workflow-id",
    "",
    "getStatus",
)
```

### 2.6.4 取消工作流

```go
// 客户端取消
err := c.CancelWorkflow(
    context.Background(),
    "workflow-id",
    "",
)

// 工作流端处理取消
func CancelableWorkflow(ctx workflow.Context) error {
    logger := workflow.GetLogger(ctx)
    
    // 等待取消
    ctx, cancel := workflow.WithCancel(ctx)
    
    // 启动活动
    future := workflow.ExecuteActivity(ctx, LongTask)
    
    // 等待活动完成或取消
    select {
    case <-future.GetChannel(ctx):
        logger.Info("活动完成")
    case <-ctx.Done():
        logger.Info("工作流被取消")
        return ctx.Err()
    }
    
    return nil
}
```

---

## 2.7 活动执行

### 2.7.1 同步执行

```go
func SyncActivityWorkflow(ctx workflow.Context) error {
    // 同步执行活动，等待结果
    var result string
    err := workflow.ExecuteActivity(ctx, MyActivity, "input").Get(ctx, &result)
    if err != nil {
        return err
    }
    
    fmt.Println("活动结果:", result)
    return nil
}
```

### 2.7.2 异步执行

```go
func AsyncActivityWorkflow(ctx workflow.Context) error {
    // 异步启动活动
    future := workflow.ExecuteActivity(ctx, LongRunningActivity, "input")
    
    // 可以执行其他逻辑
    // ...
    
    // 等待活动完成
    return future.Get(ctx, nil)
}
```

### 2.7.3 并行执行

```go
func ParallelActivityWorkflow(ctx workflow.Context) error {
    // 并行执行多个活动
    futures := []workflow.Future{
        workflow.ExecuteActivity(ctx, ActivityA),
        workflow.ExecuteActivity(ctx, ActivityB),
        workflow.ExecuteActivity(ctx, ActivityC),
    }
    
    // 等待所有活动完成
    for _, future := range futures {
        if err := future.Get(ctx, nil); err != nil {
            return err
        }
    }
    
    return nil
}
```

### 2.7.4 活动链

```go
func ChainedActivityWorkflow(ctx workflow.Context) error {
    // 活动1
    var result1 string
    if err := workflow.ExecuteActivity(ctx, Activity1).Get(ctx, &result1); err != nil {
        return err
    }
    
    // 活动2（使用活动1的结果）
    var result2 string
    if err := workflow.ExecuteActivity(ctx, Activity2, result1).Get(ctx, &result2); err != nil {
        return err
    }
    
    // 活动3（使用活动2的结果）
    return workflow.ExecuteActivity(ctx, Activity3, result2).Get(ctx, nil)
}
```

---

## 2.8 定时器与等待

### 2.8.1 延迟

```go
func DelayWorkflow(ctx workflow.Context) error {
    // 等待 5 秒
    workflow.Sleep(ctx, 5*time.Second)
    
    // 执行后续逻辑
    return nil
}
```

### 2.8.2 定时器

```go
func TimerWorkflow(ctx workflow.Context) error {
    logger := workflow.GetLogger(ctx)
    
    // 创建定时器
    timer := workflow.NewTimer(ctx, 10*time.Second)
    
    // 等待定时器或信号
    selector := workflow.NewSelector(ctx)
    selector.AddFuture(timer, func(f workflow.Future) {
        f.Get(ctx, nil)
        logger.Info("定时器触发")
    })
    selector.AddReceive(workflow.GetSignalChannel(ctx, "stop"), func(c workflow.ReceiveChannel, more bool) {
        var signal string
        c.Receive(ctx, &signal)
        logger.Info("收到停止信号", "signal", signal)
    })
    
    selector.Select(ctx)
    
    return nil
}
```

### 2.8.3 条件等待

```go
func ConditionalWaitWorkflow(ctx workflow.Context) error {
    // 等待某个条件满足
    condition := false
    
    for !condition {
        // 检查条件
        var status string
        workflow.ExecuteActivity(ctx, CheckStatus).Get(ctx, &status)
        condition = status == "completed"
        
        if !condition {
            // 等待后重试
            workflow.Sleep(ctx, 30*time.Second)
        }
    }
    
    return nil
}
```

---

## 2.9 高级特性

### 2.9.1 上下文传播

```go
// 定义上下文键
type contextKey string

const (
    userIDKey contextKey = "userID"
    traceIDKey contextKey = "traceID"
)

// 创建传播器
propagator := NewContextPropagator()

// 工作流中设置上下文
func WorkflowWithContext(ctx workflow.Context) error {
    ctx = context.WithValue(ctx, userIDKey, "user-123")
    ctx = context.WithValue(ctx, traceIDKey, "trace-abc")
    
    // 调用活动时上下文会自动传播
    return workflow.ExecuteActivity(ctx, MyActivity).Get(ctx, nil)
}

// 活动中获取上下文
func MyActivity(ctx context.Context) error {
    userID := ctx.Value(userIDKey).(string)
    traceID := ctx.Value(traceIDKey).(string)
    
    fmt.Printf("userID: %s, traceID: %s\n", userID, traceID)
    return nil
}
```

### 2.9.2 计时器精度

```go
func HighPrecisionTimerWorkflow(ctx workflow.Context) error {
    // 使用精确计时器
    now := workflow.Now(ctx)
    target := now.Add(100 * time.Millisecond)
    
    // 创建精确计时器
    timer := workflow.NewTimer(ctx, target.Sub(now))
    
    // 等待计时器
    return timer.Get(ctx, nil)
}
```

### 2.9.3 继续作为新执行

```go
func ContinueAsNewWorkflow(ctx workflow.Context, counter int) error {
    logger := workflow.GetLogger(ctx)
    logger.Info("继续作为新执行", "counter", counter)
    
    if counter < 10 {
        // 继续作为新执行，增加计数器
        return workflow.NewContinueAsNewError(ctx, ContinueAsNewWorkflow, counter+1)
    }
    
    return nil
}

// 客户端启动
we, err := c.ExecuteWorkflow(ctx, client.StartWorkflowOptions{
    ID:        "continue-workflow",
    TaskQueue: "my-task-queue",
}, ContinueAsNewWorkflow, 0)
```

---

## 2.10 测试

### 2.10.1 工作流测试

```go
package main

import (
    "testing"

    "go.temporal.io/sdk/testsuite"
)

func TestMyWorkflow(t *testing.T) {
    env := &testsuite.WorkflowTestSuite{}
    
    // 注册工作流和活动
    env.RegisterWorkflow(MyWorkflow)
    env.RegisterActivity(MyActivity)
    
    // 创建环境
    s := env.NewTestWorkflowEnvironment()
    s.RegisterActivity(func() error {
        return nil
    })
    
    // 执行工作流
    s.ExecuteWorkflow(MyWorkflow, "test-input")
    
    // 验证结果
    s.GetWorkflowResult()
    
    // 验证活动调用
    activities := s.GetActivityCalls()
    if len(activities) != 1 {
        t.Errorf("预期 1 个活动调用，实际: %d", len(activities))
    }
}
```

### 2.10.2 活动测试

```go
func TestActivity(t *testing.T) {
    env := &testsuite.WorkflowTestSuite{}
    
    s := env.NewTestActivityEnvironment()
    s.RegisterActivity(MyActivity)
    
    // 执行活动
    result, err := s.ExecuteActivity(MyActivity, "input")
    if err != nil {
        t.Fatalf("活动执行失败: %v", err)
    }
    
    var output string
    result.Get(&output)
    
    if output != "expected-output" {
        t.Errorf("预期 'expected-output'，实际: %s", output)
    }
}
```

---

## 2.11 监控与调试

### 2.11.1 日志记录

```go
// 工作流日志
func LoggedWorkflow(ctx workflow.Context) error {
    logger := workflow.GetLogger(ctx)
    
    logger.Info("工作流开始")
    logger.Debug("调试信息", "key", "value")
    logger.Warn("警告信息")
    logger.Error("错误信息", "error", err)
    
    return nil
}

// 活动日志
func LoggedActivity(ctx context.Context) error {
    logger := activity.GetLogger(ctx)
    
    logger.Info("活动开始")
    return nil
}
```

### 2.11.2 指标收集

```go
// 配置指标
metricsHandler := client.MetricsHandler{
    Counter: func(name string, opts ...client.MetricsOption) client.MetricsCounter {
        return prometheus.NewCounterVec(prometheus.CounterOpts{
            Name: name,
        }, []string{"service", "operation"})
    },
}

c, err := client.Dial(client.Options{
    HostPort:  "localhost:7233",
    MetricsHandler: metricsHandler,
})
```

---

## 2.12 最佳实践

### 2.12.1 客户端使用

```go
// 推荐：使用单例客户端
var (
    clientOnce sync.Once
    temporalClient *client.Client
)

func GetTemporalClient() (*client.Client, error) {
    var err error
    clientOnce.Do(func() {
        temporalClient, err = client.Dial(client.Options{
            HostPort: "localhost:7233",
        })
    })
    return temporalClient, err
}
```

### 2.12.2 Worker 资源管理

```go
// 根据机器资源合理配置
w := worker.New(c, "task-queue", worker.Options{
    // CPU 核心数
    MaxConcurrentWorkflowTaskExecutionSize: runtime.NumCPU() * 10,
    MaxConcurrentActivityExecutionSize: runtime.NumCPU() * 10,
})
```

### 2.12.3 超时设置

```go
// 合理设置超时
activityOptions := workflow.ActivityOptions{
    StartToCloseTimeout: 5 * time.Minute,    // 活动执行超时
    ScheduleToStartTimeout: 1 * time.Minute, // 活动等待调度超时
    HeartbeatTimeout: 30 * time.Second,      // 心跳超时
}
```
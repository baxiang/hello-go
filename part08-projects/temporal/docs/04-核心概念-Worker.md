# 核心概念 - Worker（工作器）

Worker 是执行工作流和活动代码的进程。它从 Temporal Server 的任务队列中拉取任务并执行，然后将结果报告回 Server。本文档详细介绍 Temporal Worker 的核心概念和最佳实践。

---

## Worker Process（工作器进程）

Worker 进程是运行用户代码的实体，负责执行工作流和活动。

### Worker 架构

```
┌─────────────────────────────────────────────────────────────────────────┐
│                            Worker Process                                │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│  ┌───────────────────────────────────────────────────────────────────┐ │
│  │                        Worker Core                                │ │
│  │  ┌─────────────────┐        ┌─────────────────┐                  │ │
│  │  │ Workflow Worker │        │ Activity Worker │                  │ │
│  │  │   Executor      │        │   Executor      │                  │ │
│  │  └─────────────────┘        └─────────────────┘                  │ │
│  │           │                         │                             │ │
│  │           │                         │                             │ │
│  │  ┌─────────────────┐        ┌─────────────────┐                  │ │
│  │  │ Workflow        │        │ Activity        │                  │ │
│  │  │  Registry       │        │  Registry       │                  │ │
│  │  └─────────────────┘        └─────────────────┘                  │ │
│  │           │                         │                             │ │
│  │           └───────────┬─────────────┘                             │ │
│  │                       │                                           │ │
│  │              ┌────────▼────────┐                                  │ │
│  │              │   Task Poller   │                                  │ │
│  │              └────────┬────────┘                                  │ │
│  └───────────────────────┼───────────────────────────────────────────┘ │
│                          │                                              │
│                          │ Poll & Report                                │
│                          ▼                                              │
│              ┌──────────────────────┐                                   │
│              │  Temporal Server     │                                   │
│              │  (Matching Service)  │                                   │
│              └──────────────────────┘                                   │
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
```

### Worker 类型

Temporal 有两种 Worker 类型：

| Worker 类型 | 执行内容 | 特点 |
|-------------|----------|------|
| Workflow Worker | 执行工作流代码 | 确定性执行，受沙箱限制 |
| Activity Worker | 执行活动代码 | 可执行任意代码，无确定性限制 |

```go
func main() {
    c, _ := client.Dial(client.Options{})
    defer c.Close()

    w := worker.New(c, "my-task-queue", worker.Options{
        // Workflow Worker 配置
        MaxConcurrentWorkflowTaskExecutionSize: 100,
        
        // Activity Worker 配置
        MaxConcurrentActivityExecutionSize:     50,
    })
    
    // 注册工作流（由 Workflow Worker 执行）
    w.RegisterWorkflow(MyWorkflow)
    
    // 注册活动（由 Activity Worker 执行）
    w.RegisterActivity(MyActivity)
    
    // 启动 Worker
    w.Run(worker.InterruptCh())
}
```

### Worker 执行模型

```
┌─────────────────────────────────────────────────────────────────────────┐
│                     Worker Execution Model                               │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│  Workflow Worker                                                        │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │  Task Poller ──▶ Sandbox ──▶ Workflow Execution                 │   │
│  │                                                                  │   │
│  │  特点：                                                          │   │
│  │  - 确定性执行（Event History 驱动）                              │   │
│  │  - 沙箱隔离（检测非确定性操作）                                  │   │
│  │  - 快速执行（毫秒级）                                            │   │
│  │  - 状态无状态（每次从历史重建）                                  │   │
│  └─────────────────────────────────────────────────────────────────┘   │
│                                                                          │
│  Activity Worker                                                        │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │  Task Poller ──▶ Executor ──▶ Activity Execution                │   │
│  │                                                                  │   │
│  │  特点：                                                          │   │
│  │  - 可执行任意代码                                                │   │
│  │  - 有状态执行                                                    │   │
│  │  - 可长时间运行                                                  │   │
│  │  - 支持心跳和取消                                                │   │
│  └─────────────────────────────────────────────────────────────────┘   │
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
```

---

## Task Queue（任务队列）

任务队列是 Worker 与 Temporal Server 之间的通信桥梁。

### 任务队列架构

```
┌─────────────────────────────────────────────────────────────────────────┐
│                        Task Queue Architecture                           │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│  Temporal Server                                                        │
│  ┌───────────────────────────────────────────────────────────────────┐ │
│  │                        Matching Service                           │ │
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐              │ │
│  │  │  Task Queue │  │  Task Queue │  │  Task Queue │              │ │
│  │  │    "A"      │  │    "B"      │  │    "C"      │              │ │
│  │  └─────────────┘  └─────────────┘  └─────────────┘              │ │
│  │        │                │                │                        │ │
│  │        ▼                ▼                ▼                        │ │
│  │  [Workflow Tasks]  [Activity Tasks]  [Workflow Tasks]            │ │
│  │  [Activity Tasks]  [Activity Tasks]  [Activity Tasks]            │ │
│  │  [Activity Tasks]                                                │ │
│  └───────────────────────────────────────────────────────────────────┘ │
│                                                                          │
│  Worker Processes                                                       │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐                │
│  │ Worker A     │  │ Worker B     │  │ Worker C     │                │
│  │ Queue: "A"   │  │ Queue: "B"   │  │ Queue: "C"   │                │
│  └──────────────┘  └──────────────┘  └──────────────┘                │
│        │                  │                  │                          │
│        │                  │                  │                          │
│        ▼                  ▼                  ▼                          │
│  ┌───────────────────────────────────────────────────────────────────┐ │
│  │                    Task Dispatch                                  │ │
│  │  Worker A ──▶ Poll Queue "A" ──▶ Execute Task ──▶ Report Result  │ │
│  └───────────────────────────────────────────────────────────────────┘ │
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
```

### 任务类型

| 任务类型 | 说明 | 执行者 |
|----------|------|--------|
| WorkflowTask | 执行工作流的下一步 | Workflow Worker |
| ActivityTask | 执行活动代码 | Activity Worker |

### 创建任务队列

```go
func main() {
    c, _ := client.Dial(client.Options{
        HostPort: "localhost:7233",
    })
    defer c.Close()

    // 创建 Worker 并指定任务队列
    w := worker.New(c, "order-task-queue", worker.Options{})
    
    // 注册工作流和活动
    w.RegisterWorkflow(OrderProcessingWorkflow)
    w.RegisterActivity(ProcessPaymentActivity)
    w.RegisterActivity(UpdateInventoryActivity)
    
    // 启动 Worker
    w.Run(worker.InterruptCh())
}
```

### 多任务队列

```go
func main() {
    c, _ := client.Dial(client.Options{})
    defer c.Close()
    
    // 创建多个 Worker，监听不同的任务队列
    orderWorker := worker.New(c, "order-task-queue", worker.Options{})
    orderWorker.RegisterWorkflow(OrderWorkflow)
    orderWorker.RegisterActivity(OrderActivities{})
    
    paymentWorker := worker.New(c, "payment-task-queue", worker.Options{})
    paymentWorker.RegisterWorkflow(PaymentWorkflow)
    paymentWorker.RegisterActivity(PaymentActivities{})
    
    reportWorker := worker.New(c, "report-task-queue", worker.Options{})
    reportWorker.RegisterWorkflow(ReportWorkflow)
    reportWorker.RegisterActivity(ReportActivities{})
    
    // 并行启动所有 Worker
    var wg sync.WaitGroup
    
    wg.Add(1)
    go func() {
        defer wg.Done()
        orderWorker.Run(worker.InterruptCh())
    }()
    
    wg.Add(1)
    go func() {
        defer wg.Done()
        paymentWorker.Run(worker.InterruptCh())
    }()
    
    wg.Add(1)
    go func() {
        defer wg.Done()
        reportWorker.Run(worker.InterruptCh())
    }()
    
    wg.Wait()
}
```

### 工作流指定任务队列

```go
func main() {
    c, _ := client.Dial(client.Options{})
    defer c.Close()
    
    // 启动工作流，指定任务队列
    options := client.StartWorkflowOptions{
        ID:        "order-123",
        TaskQueue: "order-task-queue",
    }
    
    we, _ := c.ExecuteWorkflow(context.Background(), options, OrderWorkflow, order)
    we.Get(context.Background(), nil)
}
```

### 活动指定任务队列

```go
func OrderWorkflow(ctx workflow.Context, order Order) error {
    // 使用默认任务队列
    ao := workflow.ActivityOptions{
        StartToCloseTimeout: 5 * time.Minute,
    }
    ctx1 := workflow.WithActivityOptions(ctx, ao)
    workflow.ExecuteActivity(ctx1, ValidateOrder, order).Get(ctx1, nil)
    
    // 指定不同的任务队列
    ao2 := workflow.ActivityOptions{
        StartToCloseTimeout: 5 * time.Minute,
        TaskQueue:           "payment-task-queue",  // 专门的支付任务队列
    }
    ctx2 := workflow.WithActivityOptions(ctx, ao2)
    workflow.ExecuteActivity(ctx2, ProcessPayment, order).Get(ctx2, nil)
    
    return nil
}
```

### 任务队列最佳实践

```
任务队列命名建议：

1. 按业务领域划分
   - order-task-queue    （订单处理）
   - payment-task-queue  （支付处理）
   - notification-task-queue （通知服务）

2. 按优先级划分
   - high-priority-queue （高优先级）
   - normal-priority-queue （普通优先级）
   - low-priority-queue  （低优先级）

3. 按资源类型划分
   - cpu-intensive-queue （CPU 密集型）
   - io-intensive-queue  （I/O 密集型）
   - memory-intensive-queue （内存密集型）

4. 按团队划分
   - team-a-queue
   - team-b-queue
```

---

## Worker Options 配置

Worker Options 提供丰富的配置选项，用于控制 Worker 的行为。

### 基本配置

```go
options := worker.Options{
    // 任务队列（创建 Worker 时指定）
    
    // 并发控制
    MaxConcurrentWorkflowTaskExecutionSize:     1000,  // 最大并发工作流任务数
    MaxConcurrentWorkflowTaskPollers:           5,     // 工作流任务拉取器数量
    MaxConcurrentActivityExecutionSize:         1000,  // 最大并发活动任务数
    MaxConcurrentActivityTaskPollers:           10,    // 活动任务拉取器数量
    MaxConcurrentLocalActivityExecutionSize:    1000,  // 最大并发本地活动数
    
    // 资源限制
    WorkerActivitiesPerSecond:                  100.0, // 每秒最大活动执行数
    MaxTaskQueueActivitiesPerSecond:            100.0, // 任务队列每秒最大活动数
    
    // 超时设置
    WorkflowTaskTimeout:                        10 * time.Second,  // 工作流任务超时
    ActivityTaskTimeout:                        10 * time.Second,  // 活动任务超时
    
    // 重试配置
    StickyScheduleToStartTimeout:               10 * time.Second,  // Sticky 任务超时
    
    // Worker 标识
    Identity:                                   "worker-1",  // Worker 唯一标识
    
    // 其他配置
    DisableWorkflowWorker:                      false,  // 禁用 Workflow Worker
    DisableActivityWorker:                      false,  // 禁用 Activity Worker
    LocalActivityWorkerOnly:                    false,  // 仅执行本地活动
    MaxHeartbeatThrottleInterval:               30 * time.Second,  // 心跳限流间隔
    DisableStickyExecution:                     false,  // 禁用 Sticky 执行
}
```

### 并发控制详解

```go
// 高并发配置（适合大规模部署）
options := worker.Options{
    MaxConcurrentWorkflowTaskExecutionSize:     5000,
    MaxConcurrentWorkflowTaskPollers:           20,
    MaxConcurrentActivityExecutionSize:         10000,
    MaxConcurrentActivityTaskPollers:           50,
}

// 低资源配置（适合小型部署）
options := worker.Options{
    MaxConcurrentWorkflowTaskExecutionSize:     100,
    MaxConcurrentWorkflowTaskPollers:           2,
    MaxConcurrentActivityExecutionSize:         200,
    MaxConcurrentActivityTaskPollers:           5,
}

// 专用配置（只执行工作流或只执行活动）
workflowOnlyOptions := worker.Options{
    DisableActivityWorker:                      true,  // 只执行工作流
    MaxConcurrentWorkflowTaskExecutionSize:     1000,
}

activityOnlyOptions := worker.Options{
    DisableWorkflowWorker:                      true,  // 只执行活动
    MaxConcurrentActivityExecutionSize:         1000,
}
```

### Sticky 执行

Sticky 执行是指工作流倾向于在同一个 Worker 上执行，以提高性能。

```go
options := worker.Options{
    // 启用 Sticky 执行（默认启用）
    DisableStickyExecution:                     false,
    
    // Sticky 任务超时
    StickyScheduleToStartTimeout:               10 * time.Second,
}
```

```
┌─────────────────────────────────────────────────────────────────────────┐
│                        Sticky Execution Flow                             │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│  第一次执行：                                                            │
│  Workflow Started ──▶ Worker A ──▶ Activity Scheduled                   │
│                                                                          │
│  第二次执行（Activity 完成后）：                                         │
│  Activity Completed ──▶ Worker A（Sticky）──▶ Workflow Continues        │
│                                                                          │
│  优势：                                                                  │
│  - Worker A 已有工作流的缓存状态                                        │
│  - 无需重建状态，执行更快                                               │
│  - 减少网络传输                                                         │
│                                                                          │
│  适用场景：                                                              │
│  - 工作流执行频繁                                                       │
│  - 工作流状态复杂                                                       │
│                                                                          │
│  不适用场景：                                                            │
│  - Worker 可能随时崩溃                                                  │
│  - 需要高可用性                                                         │
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
```

### 会话配置（Session）

Session 用于将一组活动分配给同一个 Worker。

```go
options := worker.Options{
    MaxConcurrentSessionExecutionSize:          100,  // 最大并发会话数
}
```

---

## 资源限制

合理配置资源限制，避免 Worker 过载。

### 并发限制

```go
func main() {
    // 根据机器配置调整并发限制
    cpuCount := runtime.NumCPU()
    
    options := worker.Options{
        // Workflow 并发：每个 CPU 核心 100 个
        MaxConcurrentWorkflowTaskExecutionSize: cpuCount * 100,
        MaxConcurrentWorkflowTaskPollers:       cpuCount / 2,
        
        // Activity 并发：每个 CPU 核心 50 个
        MaxConcurrentActivityExecutionSize:     cpuCount * 50,
        MaxConcurrentActivityTaskPollers:       cpuCount,
        
        // 本地活动并发
        MaxConcurrentLocalActivityExecutionSize: cpuCount * 200,
    }
    
    w := worker.New(c, "task-queue", options)
    w.Run(worker.InterruptCh())
}
```

### 速率限制

```go
options := worker.Options{
    // Worker 级别速率限制
    WorkerActivitiesPerSecond:                  50.0,  // 每秒最多 50 个活动
    
    // 任务队列级别速率限制
    MaxTaskQueueActivitiesPerSecond:            100.0, // 任务队列每秒最多 100 个活动
}
```

### 动态调整

```go
func main() {
    c, _ := client.Dial(client.Options{})
    defer c.Close()
    
    // 创建 Worker
    w := worker.New(c, "task-queue", worker.Options{})
    w.RegisterWorkflow(MyWorkflow)
    w.RegisterActivity(MyActivity)
    
    // 启动 Worker
    stopCh := make(chan struct{})
    go w.Run(stopCh)
    
    // 监控系统负载
    go func() {
        for {
            time.Sleep(10 * time.Second)
            
            load := getSystemLoad()
            if load > 0.8 {
                // 高负载：减少并发
                w.SetConcurrentWorkflowTaskExecutionSize(100)
                w.SetConcurrentActivityExecutionSize(50)
            } else if load < 0.3 {
                // 低负载：增加并发
                w.SetConcurrentWorkflowTaskExecutionSize(1000)
                w.SetConcurrentActivityExecutionSize(500)
            }
        }
    }()
    
    // 等待停止信号
    <-worker.InterruptCh()
    close(stopCh)
}

func getSystemLoad() float64 {
    // 获取系统负载（示例）
    return 0.5
}
```

---

## 水平扩展

Worker 支持水平扩展，通过增加 Worker 实例来提高处理能力。

### 多 Worker 实例

```go
// 单机多 Worker
func main() {
    c, _ := client.Dial(client.Options{})
    defer c.Close()
    
    var wg sync.WaitGroup
    
    // 启动多个 Worker 实例
    for i := 0; i < 5; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            
            w := worker.New(c, "task-queue", worker.Options{
                Identity: fmt.Sprintf("worker-%d", id),
                MaxConcurrentActivityExecutionSize: 100,
            })
            w.RegisterWorkflow(MyWorkflow)
            w.RegisterActivity(MyActivity)
            w.Run(worker.InterruptCh())
        }(i)
    }
    
    wg.Wait()
}
```

### 多机部署

```
┌─────────────────────────────────────────────────────────────────────────┐
│                    Multi-Host Worker Deployment                          │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│                      Temporal Cluster                                    │
│  ┌───────────────────────────────────────────────────────────────────┐ │
│  │                      Matching Service                             │ │
│  │  ┌─────────────┐                                                  │ │
│  │  │  Task Queue │                                                  │ │
│  │  │  "orders"   │                                                  │ │
│  │  └─────────────┘                                                  │ │
│  │        │                                                          │ │
│  │        ▼                                                          │ │
│  │  [Task 1] [Task 2] [Task 3] [Task 4] [Task 5]                    │ │
│  └───────────────────────────────────────────────────────────────────┘ │
│                          │                                               │
│                          │                                               │
│          ┌───────────────┼───────────────┬───────────────┐              │
│          │               │               │               │              │
│          ▼               ▼               ▼               ▼              │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐   │
│  │  Host 1     │  │  Host 2     │  │  Host 3     │  │  Host 4     │   │
│  │  Worker A   │  │  Worker B   │  │  Worker C   │  │  Worker D   │   │
│  │  Queue:     │  │  Queue:     │  │  Queue:     │  │  Queue:     │   │
│  │  "orders"   │  │  "orders"   │  │  "orders"   │  │  "orders"   │   │
│  └─────────────┘  └─────────────┘  └─────────────┘  └─────────────┘   │
│                                                                          │
│  特点：                                                                  │
│  - 自动负载均衡                                                         │
│  - 高可用（任一 Host 故障不影响整体）                                   │
│  - 动态扩缩容                                                           │
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
```

### Kubernetes 部署示例

```yaml
# worker-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: temporal-worker
  namespace: temporal
spec:
  replicas: 10  # 10 个 Worker 实例
  selector:
    matchLabels:
      app: temporal-worker
  template:
    metadata:
      labels:
        app: temporal-worker
    spec:
      containers:
      - name: worker
        image: my-worker:latest
        env:
        - name: TEMPORAL_ADDRESS
          value: "temporal-frontend:7233"
        - name: TASK_QUEUE
          value: "order-task-queue"
        - name: MAX_CONCURRENT_ACTIVITIES
          value: "100"
        resources:
          requests:
            cpu: "1"
            memory: "512Mi"
          limits:
            cpu: "2"
            memory: "1Gi"
```

### 自动扩缩容（HPA）

```yaml
# worker-hpa.yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: temporal-worker-hpa
  namespace: temporal
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: temporal-worker
  minReplicas: 5
  maxReplicas: 50
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
  - type: Resource
    resource:
      name: memory
      target:
        type: Utilization
        averageUtilization: 80
```

---

## Go 代码示例

### 基本 Worker

```go
package main

import (
    "log"
    
    "go.temporal.io/sdk/client"
    "go.temporal.io/sdk/worker"
    "myapp/workflow"
    "myapp/activity"
)

func main() {
    // 创建 Temporal 客户端
    c, err := client.Dial(client.Options{
        HostPort: "localhost:7233",
        Namespace: "default",
    })
    if err != nil {
        log.Fatal("Failed to create client", err)
    }
    defer c.Close()
    
    // 创建 Worker
    w := worker.New(c, "order-task-queue", worker.Options{
        Identity: "order-worker-1",
        MaxConcurrentWorkflowTaskExecutionSize: 100,
        MaxConcurrentActivityExecutionSize: 50,
    })
    
    // 注册工作流
    w.RegisterWorkflow(workflow.OrderProcessingWorkflow)
    w.RegisterWorkflow(workflow.PaymentWorkflow)
    
    // 注册活动
    w.RegisterActivity(activity.ValidateOrderActivity)
    w.RegisterActivity(activity.ProcessPaymentActivity)
    w.RegisterActivity(activity.UpdateInventoryActivity)
    
    // 启动 Worker
    log.Println("Starting worker...")
    w.Run(worker.InterruptCh())
}
```

### 结构体活动注册

```go
package main

import (
    "log"
    
    "go.temporal.io/sdk/client"
    "go.temporal.io/sdk/worker"
    "go.temporal.io/sdk/workflow"
)

// 活动结构体
type OrderActivities struct {
    OrderService   OrderService
    PaymentService PaymentService
    InventoryService InventoryService
}

func (a *OrderActivities) ValidateOrder(ctx context.Context, order Order) error {
    return a.OrderService.Validate(order)
}

func (a *OrderActivities) ProcessPayment(ctx context.Context, order Order) (string, error) {
    return a.PaymentService.Process(order)
}

func (a *OrderActivities) UpdateInventory(ctx context.Context, order Order) error {
    return a.InventoryService.Update(order)
}

func main() {
    c, _ := client.Dial(client.Options{})
    defer c.Close()
    
    // 初始化活动结构体
    orderActivities := &OrderActivities{
        OrderService:   NewOrderService(),
        PaymentService: NewPaymentService(),
        InventoryService: NewInventoryService(),
    }
    
    w := worker.New(c, "order-task-queue", worker.Options{})
    
    // 注册工作流
    w.RegisterWorkflow(OrderProcessingWorkflow)
    
    // 注册活动结构体（所有方法都会注册）
    w.RegisterActivity(orderActivities)
    
    w.Run(worker.InterruptCh())
}
```

### 多任务队列 Worker

```go
package main

import (
    "log"
    "sync"
    
    "go.temporal.io/sdk/client"
    "go.temporal.io/sdk/worker"
)

func main() {
    c, _ := client.Dial(client.Options{})
    defer c.Close()
    
    var wg sync.WaitGroup
    
    // Worker 1: 处理订单
    wg.Add(1)
    go func() {
        defer wg.Done()
        w := worker.New(c, "order-task-queue", worker.Options{
            Identity: "order-worker",
            MaxConcurrentWorkflowTaskExecutionSize: 200,
            MaxConcurrentActivityExecutionSize: 100,
        })
        w.RegisterWorkflow(OrderWorkflow)
        w.RegisterActivity(&OrderActivities{})
        w.Run(worker.InterruptCh())
    }()
    
    // Worker 2: 处理支付
    wg.Add(1)
    go func() {
        defer wg.Done()
        w := worker.New(c, "payment-task-queue", worker.Options{
            Identity: "payment-worker",
            MaxConcurrentActivityExecutionSize: 50,
            DisableWorkflowWorker: true,  // 只执行活动
        })
        w.RegisterActivity(&PaymentActivities{})
        w.Run(worker.InterruptCh())
    }()
    
    // Worker 3: 处理通知
    wg.Add(1)
    go func() {
        defer wg.Done()
        w := worker.New(c, "notification-task-queue", worker.Options{
            Identity: "notification-worker",
            MaxConcurrentActivityExecutionSize: 200,
            DisableWorkflowWorker: true,
        })
        w.RegisterActivity(&NotificationActivities{})
        w.Run(worker.InterruptCh())
    }()
    
    log.Println("All workers started")
    wg.Wait()
}
```

### 高性能 Worker 配置

```go
package main

import (
    "runtime"
    
    "go.temporal.io/sdk/client"
    "go.temporal.io/sdk/worker"
)

func main() {
    c, _ := client.Dial(client.Options{})
    defer c.Close()
    
    cpuCount := runtime.NumCPU()
    
    // 高性能配置
    options := worker.Options{
        Identity: "high-performance-worker",
        
        // Workflow 并发配置
        MaxConcurrentWorkflowTaskExecutionSize: cpuCount * 200,
        MaxConcurrentWorkflowTaskPollers:       cpuCount * 2,
        
        // Activity 并发配置
        MaxConcurrentActivityExecutionSize:     cpuCount * 100,
        MaxConcurrentActivityTaskPollers:       cpuCount * 5,
        
        // 本地活动配置
        MaxConcurrentLocalActivityExecutionSize: cpuCount * 500,
        
        // Sticky 执行配置
        DisableStickyExecution:                 false,
        StickyScheduleToStartTimeout:           5 * time.Second,
        
        // 速率限制
        WorkerActivitiesPerSecond:              1000.0,
        MaxTaskQueueActivitiesPerSecond:        5000.0,
        
        // 心跳配置
        MaxHeartbeatThrottleInterval:           60 * time.Second,
    }
    
    w := worker.New(c, "high-performance-task-queue", options)
    w.RegisterWorkflow(MyWorkflow)
    w.RegisterActivity(MyActivity)
    
    w.Run(worker.InterruptCh())
}
```

### Worker 监控

```go
package main

import (
    "context"
    "log"
    "time"
    
    "go.temporal.io/sdk/client"
    "go.temporal.io/sdk/worker"
)

func main() {
    c, _ := client.Dial(client.Options{})
    defer c.Close()
    
    w := worker.New(c, "monitored-task-queue", worker.Options{
        Identity: "monitored-worker",
    })
    w.RegisterWorkflow(MyWorkflow)
    w.RegisterActivity(MyActivity)
    
    // 启动 Worker
    stopCh := make(chan struct{})
    go w.Run(stopCh)
    
    // 监控 Worker
    go monitorWorker(c, "monitored-task-queue")
    
    // 等待停止
    <-worker.InterruptCh()
    close(stopCh)
}

func monitorWorker(c client.Client, taskQueue string) {
    for {
        time.Sleep(10 * time.Second)
        
        // 查询任务队列信息
        resp, err := c.DescribeTaskQueue(context.Background(), taskQueue)
        if err != nil {
            log.Printf("Failed to describe task queue: %v", err)
            continue
        }
        
        log.Printf("Task Queue Status:")
        log.Printf("  - Workflow Tasks: %d (pending), %d (executing)",
            resp.WorkflowTaskQueueInfo.GetPending(),
            resp.WorkflowTaskQueueInfo.GetExecuting(),
        )
        log.Printf("  - Activity Tasks: %d (pending), %d (executing)",
            resp.ActivityTaskQueueInfo.GetPending(),
            resp.ActivityTaskQueueInfo.GetExecuting(),
        )
        log.Printf("  - Workers: %d", len(resp.Workers))
        
        for _, workerInfo := range resp.Workers {
            log.Printf("  - Worker %s: %s",
                workerInfo.Identity,
                workerInfo.TaskQueueTypes,
            )
        }
    }
}
```

### Worker 健康检查

```go
package main

import (
    "context"
    "net/http"
    "time"
    
    "go.temporal.io/sdk/client"
    "go.temporal.io/sdk/worker"
)

func main() {
    c, _ := client.Dial(client.Options{})
    defer c.Close()
    
    // 健康检查状态
    healthy := true
    
    w := worker.New(c, "task-queue", worker.Options{})
    w.RegisterWorkflow(MyWorkflow)
    w.RegisterActivity(MyActivity)
    
    // 启动 Worker
    stopCh := make(chan struct{})
    go func() {
        w.Run(stopCh)
        healthy = false
    }()
    
    // 健康检查 HTTP 服务
    go func() {
        http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
            if healthy {
                w.WriteHeader(http.StatusOK)
                w.Write([]byte("OK"))
            } else {
                w.WriteHeader(http.StatusServiceUnavailable)
                w.Write([]byte("Not OK"))
            }
        })
        http.ListenAndServe(":8080", nil)
    }()
    
    // 定期检查 Temporal 连接
    go func() {
        for {
            time.Sleep(30 * time.Second)
            ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
            _, err := c.Check(ctx)
            cancel()
            if err != nil {
                healthy = false
                log.Printf("Temporal connection failed: %v", err)
            } else {
                healthy = true
            }
        }
    }()
    
    <-worker.InterruptCh()
    close(stopCh)
}
```

---

## 最佳实践

### 1. Worker 配置原则

- **根据资源调整并发**：根据 CPU、内存配置并发数
- **合理设置 Poller 数量**：Poller 数量影响任务获取速度
- **启用 Sticky 执行**：提高工作流执行效率
- **设置速率限制**：避免过载

### 2. 资源规划

```go
// 资源规划建议
func getResourceBasedOptions() worker.Options {
    cpu := runtime.NumCPU()
    
    // Workflow Worker：CPU 轻量，但需要处理大量任务
    workflowConcurrency := cpu * 100
    workflowPollers := cpu
    
    // Activity Worker：根据活动类型调整
    // CPU 密集型活动
    cpuActivityConcurrency := cpu * 10
    // I/O 密集型活动（可以更高）
    ioActivityConcurrency := cpu * 50
    
    return worker.Options{
        MaxConcurrentWorkflowTaskExecutionSize: workflowConcurrency,
        MaxConcurrentWorkflowTaskPollers:       workflowPollers,
        MaxConcurrentActivityExecutionSize:     ioActivityConcurrency,
    }
}
```

### 3. 任务队列规划

```
任务队列规划建议：

1. 按业务隔离：
   - 不同业务使用不同任务队列
   - 避免 A 业务阻塞 B 业务

2. 按优先级隔离：
   - 高优先级队列：少量 Worker，高并发
   - 低优先级队列：多量 Worker，低并发

3. 按资源隔离：
   - CPU 密集型：单独任务队列
   - I/O 密集型：单独任务队列

4. 按团队隔离：
   - 团队 A：team-a-task-queue
   - 团队 B：team-b-task-queue
```

### 4. 高可用部署

```
高可用部署建议：

1. 多副本部署：
   - 至少 2 个 Worker 实例
   - 分布在不同主机

2. 健康检查：
   - HTTP 健康检查端点
   - 定期检查 Temporal 连接

3. 监控告警：
   - 监控 Worker 数量
   - 监控任务队列积压
   - 监控执行失败率

4. 自动扩缩容：
   - 基于 CPU/内存
   - 基于任务队列长度
```

### 5. 调试技巧

```go
// 启用调试日志
func main() {
    c, _ := client.Dial(client.Options{})
    defer c.Close()
    
    w := worker.New(c, "task-queue", worker.Options{
        Identity: "debug-worker",
    })
    
    // 注册时添加详细信息
    w.RegisterWorkflowWithOptions(
        MyWorkflow,
        workflow.RegisterOptions{
            Name: "my-workflow",
        },
    )
    
    w.Run(worker.InterruptCh())
}
```

---

## 相关资源

- [Temporal 官方文档 - Worker](https://docs.temporal.io/workers)
- [Go SDK Worker 指南](https://docs.temporal.io/dev-guide/go/workers)
- [Task Queues](https://docs.temporal.io/task-queues)
- [Worker Scaling](https://docs.temporal.io/scaling)
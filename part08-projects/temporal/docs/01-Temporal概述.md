# Temporal 概述

## 什么是 Temporal？

Temporal 是一个开源的工作流编排引擎，用于构建持久化、可扩展的应用程序。它源自 Uber 的 Cadence 项目，是一个用于构建可靠分布式系统的编程模型。

### 起源与发展

- **Uber Cadence**：Temporal 的前身，由 Uber 开发用于处理大规模业务流程编排
- **Temporal**：从 Cadence 分离出来的独立项目，提供了更现代化的架构和更丰富的功能
- **开源生态**：活跃的社区支持，多语言 SDK（Go、Java、Python、TypeScript、PHP、.NET）

### 核心理念

Temporal 的核心设计理念是**将业务逻辑与基础设施分离**：

```
┌─────────────────────────────────────────────────────────┐
│                    业务代码层                            │
│  ┌─────────────────────────────────────────────────┐   │
│  │   Workflow: 业务流程逻辑                         │   │
│  │   Activity: 具体操作                             │   │
│  └─────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────┘
                         ▲
                         │
┌─────────────────────────────────────────────────────────┐
│                  Temporal 平台层                        │
│  ┌─────────────────────────────────────────────────┐   │
│  │   持久化执行 │ 容错 │ 重试 │ 超时 │ 可见性       │   │
│  └─────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────┘
```

开发者只需要关注业务逻辑，Temporal 平台自动处理分布式系统的复杂性。

---

## 核心特性

| 特性 | 说明 |
|------|------|
| **持久化状态** | 工作流状态自动持久化，即使系统故障也不会丢失 |
| **可重试执行** | 活动失败时自动重试，无需手动编写重试逻辑 |
| **时间旅行** | 支持回溯和重新执行工作流 |
| **活动取消** | 支持优雅取消正在执行的工作流 |
| **可见性** | 提供完整的执行历史和调试工具 |
| **水平扩展** | 支持高并发和大规模工作流执行 |

### 1. 持久化执行

工作流的执行状态自动持久化到数据库中，这意味着：

- **故障恢复**：即使服务器崩溃，工作流也能从断点继续执行
- **长时间运行**：工作流可以运行数天、数月甚至数年
- **状态保存**：所有变量和执行位置都会被保存

```go
// 即使服务重启，这个工作流也能从断点继续
func OrderWorkflow(ctx workflow.Context, orderID string) error {
    // 第一步：处理支付
    err := workflow.ExecuteActivity(ctx, ProcessPayment, orderID).Get(ctx, nil)
    if err != nil {
        return err
    }
    
    // 服务崩溃重启后，从这里继续执行
    // 第二步：更新库存
    err = workflow.ExecuteActivity(ctx, UpdateInventory, orderID).Get(ctx, nil)
    if err != nil {
        return err
    }
    
    return nil
}
```

### 2. 内置容错

Temporal 提供多层次的容错机制：

| 层次 | 机制 | 说明 |
|------|------|------|
| 活动层 | 自动重试 | 失败后按策略自动重试 |
| 工作流层 | 状态恢复 | 从 Event History 恢复执行 |
| 服务层 | 高可用 | 多副本部署，自动故障转移 |

### 3. 完整可见性

```
┌─────────────────────────────────────────────────────────┐
│                    Temporal Web UI                      │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  │
│  │ 工作流列表   │  │ 执行历史     │  │ 堆栈跟踪     │  │
│  └──────────────┘  └──────────────┘  └──────────────┘  │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  │
│  │ 活动详情     │  │ 重试记录     │  │ 指标监控     │  │
│  └──────────────┘  └──────────────┘  └──────────────┘  │
└─────────────────────────────────────────────────────────┘
```

每个工作流的完整执行历史都被记录，包括：

- 所有活动的调用和结果
- 定时器的创建和触发
- 信号的接收和处理
- 错误和重试记录

---

## 系统架构

### 整体架构图

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
│                         │ (MySQL/Postgres) │                           │
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

### 核心组件详解

#### 1. Frontend Service（前端服务）

负责接收客户端请求，是集群的入口点。

**主要职责：**
- 接收客户端的 gRPC 请求
- 限流和认证
- 请求路由到其他服务

```go
// 客户端连接 Frontend Service
client, err := client.Dial(client.Options{
    HostPort: "localhost:7233",
})
```

#### 2. History Service（历史服务）

管理工作流的执行状态，是 Temporal 的核心组件。

**主要职责：**
- 维护工作流的执行状态
- 存储事件历史（Event History）
- 处理工作流的启动、执行、完成
- 支持分片以实现高并发

**Event History 示例：**
```
EventID: 1  EventType: WorkflowExecutionStarted
EventID: 2  EventType: ActivityTaskScheduled    ActivityID: 1
EventID: 3  EventType: ActivityTaskStarted      ActivityID: 1
EventID: 4  EventType: ActivityTaskCompleted    ActivityID: 1
EventID: 5  EventType: WorkflowExecutionCompleted
```

#### 3. Matching Service（匹配服务）

负责任务的调度和分发。

**主要职责：**
- 管理任务队列
- 将任务匹配给空闲的 Worker
- 实现负载均衡

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   Task      │     │   Task      │     │   Task      │
│   Queue 1   │     │   Queue 2   │     │   Queue 3   │
└──────┬──────┘     └──────┬──────┘     └──────┬──────┘
       │                   │                   │
       ▼                   ▼                   ▼
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   Worker    │     │   Worker    │     │   Worker    │
│     A       │     │     B       │     │     C       │
└─────────────┘     └─────────────┘     └─────────────┘
```

#### 4. Worker Service（工作服务）

运行用户的工作流和活动代码。

**主要职责：**
- 从任务队列拉取任务
- 执行工作流和活动代码
- 向 Temporal Server 报告执行结果

### 数据存储层

Temporal 使用持久化存储保存所有状态数据：

**存储内容：**
- 工作流执行状态
- 事件历史
- 任务队列信息
- 命名空间配置

**支持的数据库：**
| 数据库 | 适用场景 | 说明 |
|--------|----------|------|
| PostgreSQL | 生产环境 | 推荐使用，功能完善 |
| MySQL | 生产环境 | 广泛使用，稳定可靠 |
| Cassandra | 大规模部署 | 高写入性能，需要更多运维 |
| SQLite | 开发环境 | 轻量级，适合本地开发 |

---

## 与其他技术对比

### 对比表格

| 特性 | Temporal | AWS Step Functions | Camunda | Airflow |
|------|----------|---------------------|---------|----------|
| 编程模型 | 代码优先 | 声明式 | 声明式 | DAG |
| 持久化执行 | 原生支持 | AWS 托管 | 需要配置 | 需要配置 |
| 活动重试 | 自动 | 有限 | 有限 | 有限 |
| 长时间运行 | 支持 | 有超时限制 | 支持 | 有超时限制 |
| 分布式事务 | Saga 支持 | 有限 | Saga 支持 | 不支持 |
| 多语言支持 | 丰富 | AWS SDK | 丰富 | 丰富 |
| 本地开发 | 简单 | 需要 AWS | 中等 | 中等 |
| 学习曲线 | 中等 | 低 | 中等 | 中等 |

### vs AWS Step Functions

**Temporal 优势：**
- 代码即配置，更灵活
- 本地开发体验好
- 不锁定云服务商
- 更强大的重试机制

**Step Functions 优势：**
- 完全托管，无需运维
- 与 AWS 服务深度集成
- 可视化编辑器
- 按执行次数计费

### vs Camunda

**Temporal 优势：**
- 更现代的架构
- 代码优先，开发效率高
- 多语言 SDK 支持完善
- 开源社区活跃

**Camunda 优势：**
- BPMN 标准支持
- 可视化流程设计
- 成熟的企业支持
- 更好的流程分析工具

---

## 适用场景

### 1. 微服务编排

协调多个微服务之间的交互，实现复杂的业务流程。

```
订单处理流程：
┌──────────┐    ┌──────────┐    ┌──────────┐    ┌──────────┐
│ 订单服务 │───▶│ 支付服务 │───▶│ 库存服务 │───▶│ 物流服务 │
└──────────┘    └──────────┘    └──────────┘    └──────────┘
      │              │              │              │
      └──────────────┴──────────────┴──────────────┘
                          Temporal 协调
```

### 2. 业务流程自动化

自动化处理订单、审批、数据同步等业务流程。

```go
func ApprovalWorkflow(ctx workflow.Context, request ApprovalRequest) error {
    // 发起审批
    err := workflow.ExecuteActivity(ctx, SubmitApproval, request).Get(ctx, nil)
    if err != nil {
        return err
    }
    
    // 等待审批结果（可能需要数天）
    var approvalResult string
    workflow.Await(ctx, func() bool {
        // 等待外部信号
        return approvalResult != ""
    })
    
    // 处理审批结果
    return workflow.ExecuteActivity(ctx, ProcessResult, approvalResult).Get(ctx, nil)
}
```

### 3. 数据处理管道

ETL、批处理、数据同步等场景。

```go
func DataPipelineWorkflow(ctx workflow.Context, source, target string) error {
    // 抽取数据
    err := workflow.ExecuteActivity(ctx, ExtractData, source).Get(ctx, nil)
    if err != nil {
        return err
    }
    
    // 转换数据
    err = workflow.ExecuteActivity(ctx, TransformData).Get(ctx, nil)
    if err != nil {
        return err
    }
    
    // 加载数据
    return workflow.ExecuteActivity(ctx, LoadData, target).Get(ctx, nil)
}
```

### 4. 长时间运行流程

需要人工介入、等待外部回调的长时间流程。

```go
func LongRunningWorkflow(ctx workflow.Context) error {
    // 发送邮件通知
    workflow.ExecuteActivity(ctx, SendNotification).Get(ctx, nil)
    
    // 等待 7 天
    workflow.NewTimer(ctx, 7*24*time.Hour).Get(ctx, nil)
    
    // 检查是否有响应
    var response bool
    workflow.Await(ctx, func() bool {
        return response
    })
    
    // 继续后续流程
    return nil
}
```

### 5. 事件驱动架构

响应事件并执行复杂逻辑。

```go
func EventDrivenWorkflow(ctx workflow.Context) error {
    var events []Event
    
    // 持续监听事件
    for {
        selector := workflow.NewSelector(ctx)
        
        // 监听事件信号
        signalChannel := workflow.GetSignalChannel(ctx, "events")
        selector.AddReceive(signalChannel, func(c workflow.ReceiveChannel, more bool) {
            var event Event
            c.Receive(ctx, &event)
            events = append(events, event)
        })
        
        // 设置超时
        selector.AddFuture(workflow.NewTimer(ctx, time.Hour), func(f workflow.Future) {
            // 定期处理事件
            workflow.ExecuteActivity(ctx, ProcessEvents, events).Get(ctx, nil)
            events = nil
        })
        
        selector.Select(ctx)
    }
}
```

---

## 本地开发环境

### 方式一：Temporal CLI（推荐）

最简单的本地开发方式。

```bash
# 安装 Temporal CLI
# macOS
brew install temporal

# Linux
curl -sSL https://github.com/temporalio/cli/releases/latest/download/temporal_linux_amd64.tar.gz | tar -xz
sudo mv temporal /usr/local/bin/

# Windows
choco install temporal

# 启动开发服务器
temporal server start-dev

# 指定端口
temporal server start-dev --port 7234

# 指定数据库路径
temporal server start-dev --db-file /path/to/temporal.db

# 启动并打开 Web UI
temporal server start-dev --ui-port 8080
```

启动后可以访问：
- Temporal Server: `localhost:7233`
- Web UI: `http://localhost:8233`

### 方式二：Docker Compose

适合需要完整生产环境配置的场景。

```yaml
# docker-compose.yml
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
    depends_on:
      - temporal-postgresql
    volumes:
      - temporal-data:/var/lib/temporal

  temporal-postgresql:
    image: postgres:13
    environment:
      POSTGRES_USER: temporal
      POSTGRES_PASSWORD: temporal
    volumes:
      - postgres-data:/var/lib/postgresql/data

  temporal-web:
    image: temporalio/web:1.22.0
    ports:
      - "8080:8080"
    environment:
      - TEMPORAL_ADDRESS=temporal:7233
    depends_on:
      - temporal

volumes:
  temporal-data:
  postgres-data:
```

```bash
# 启动服务
docker-compose up -d

# 查看日志
docker-compose logs -f temporal

# 停止服务
docker-compose down

# 清理数据
docker-compose down -v
```

### 方式三：Kubernetes（生产环境）

使用 Helm 部署到 Kubernetes。

```bash
# 添加 Temporal Helm 仓库
helm repo add temporalio https://helm.temporal.io
helm repo update

# 安装 Temporal
helm install temporal temporalio/temporal \
  --namespace temporal \
  --create-namespace \
  --set server.replicaCount=1 \
  --set cassandra.config.cluster_size=1

# 查看服务
kubectl get svc -n temporal

# 端口转发
kubectl port-forward svc/temporal-frontend 7233:7233 -n temporal
```

### CLI 常用命令

```bash
# 查看命名空间
temporal namespace list

# 创建命名空间
temporal namespace create my-namespace

# 查看工作流执行
temporal workflow list

# 查看工作流详情
temporal workflow describe --workflow-id <id>

# 显示工作流历史
temporal workflow show --workflow-id <id>

# 取消工作流
temporal workflow cancel --workflow-id <id>

# 终止工作流
temporal workflow terminate --workflow-id <id>

# 查看任务队列
temporal task-queue describe --task-queue my-queue
```

---

## 快速开始示例

### 1. 定义工作流

```go
// workflow.go
package app

import (
    "time"
    "go.temporal.io/sdk/workflow"
)

func GreetingWorkflow(ctx workflow.Context, name string) (string, error) {
    ao := workflow.ActivityOptions{
        StartToCloseTimeout: 10 * time.Second,
    }
    ctx = workflow.WithActivityOptions(ctx, ao)

    var result string
    err := workflow.ExecuteActivity(ctx, SayHello, name).Get(ctx, &result)
    if err != nil {
        return "", err
    }

    return result, nil
}
```

### 2. 定义活动

```go
// activity.go
package app

import (
    "context"
    "fmt"
)

func SayHello(ctx context.Context, name string) (string, error) {
    return fmt.Sprintf("Hello, %s!", name), nil
}
```

### 3. 启动 Worker

```go
// worker.go
package main

import (
    "go.temporal.io/sdk/client"
    "go.temporal.io/sdk/worker"
    "myapp/app"
)

func main() {
    // 创建客户端
    c, _ := client.Dial(client.Options{
        HostPort: "localhost:7233",
    })
    defer c.Close()

    // 创建 Worker
    w := worker.New(c, "greeting-task-queue", worker.Options{})

    // 注册工作流和活动
    w.RegisterWorkflow(app.GreetingWorkflow)
    w.RegisterActivity(app.SayHello)

    // 启动 Worker
    w.Run(worker.InterruptCh())
}
```

### 4. 启动工作流

```go
// starter.go
package main

import (
    "context"
    "fmt"
    "time"
    "go.temporal.io/sdk/client"
    "myapp/app"
)

func main() {
    // 创建客户端
    c, _ := client.Dial(client.Options{
        HostPort: "localhost:7233",
    })
    defer c.Close()

    // 启动工作流
    options := client.StartWorkflowOptions{
        ID:        "greeting-workflow",
        TaskQueue: "greeting-task-queue",
    }

    we, _ := c.ExecuteWorkflow(context.Background(), options, app.GreetingWorkflow, "World")
    
    // 获取结果
    var result string
    we.Get(context.Background(), &result)
    
    fmt.Println(result) // Hello, World!
}
```

---

## 总结

Temporal 提供了一个强大的分布式工作流编排平台，其核心优势在于：

1. **简化分布式系统开发**：开发者只需关注业务逻辑
2. **内置容错能力**：自动处理故障恢复和重试
3. **完整的可见性**：提供详细的执行历史和监控
4. **多语言支持**：Go、Java、Python、TypeScript 等
5. **活跃的社区**：不断完善的开源生态

对于需要构建可靠、可扩展的分布式应用的团队来说，Temporal 是一个值得考虑的选择。
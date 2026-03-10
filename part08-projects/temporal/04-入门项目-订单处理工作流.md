# 入门项目：订单处理工作流

本项目实现一个基于 Temporal 的订单处理工作流，涵盖工作流定义、活动实现、Worker 配置和客户端调用。

## 4.1 项目概述

### 功能需求

1. 创建订单工作流
2. 验证订单
3. 预留库存
4. 处理支付
5. 发货
6. 发送通知

### 项目结构

```
order-workflow/
├── cmd/
│   ├── worker/main.go       # Worker 启动
│   └── starter/main.go      # 工作流启动
├── internal/
│   ├── workflow/            # 工作流定义
│   │   └── order.go
│   ├── activity/            # 活动定义
│   │   └── order.go
│   └── types/               # 类型定义
│       └── order.go
├── config.yaml
└── go.mod
```

## 4.2 类型定义

```go
// internal/types/order.go
package types

import "time"

// Order 订单
type Order struct {
    ID          string
    CustomerID  string
    Items       []OrderItem
    TotalAmount float64
    Status      string
    CreatedAt   time.Time
}

// OrderItem 订单商品
type OrderItem struct {
    ProductID   string
    ProductName string
    Quantity   int
    Price      float64
}

// PaymentRequest 支付请求
type PaymentRequest struct {
    OrderID   string
    Amount    float64
    Method    string
}

// PaymentResult 支付结果
type PaymentResult struct {
    Success        bool
    TransactionID  string
    Message        string
}

// InventoryRequest 库存请求
type InventoryRequest struct {
    ProductID string
    Quantity  int
}

// InventoryResult 库存结果
type InventoryResult struct {
    Success     bool
    Message     string
    Remaining   int
}

// NotificationRequest 通知请求
type NotificationRequest struct {
    CustomerID string
    Type       string
    Message    string
}
```

## 4.3 工作流定义

```go
// internal/workflow/order.go
package workflow

import (
    "fmt"
    "time"

	"go.temporal.io/sdk/workflow"

	"order-workflow/internal/types"
)

// OrderWorkflow 订单处理工作流
func OrderWorkflow(ctx workflow.Context, order types.Order) (string, error) {
	logger := workflow.GetLogger(ctx)

	logger.Info("订单处理工作流开始", "orderID", order.ID)

	// 设置活动选项
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 5 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumAttempts:    3,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	// 步骤 1: 验证订单
	logger.Info("步骤 1: 验证订单")
	err := workflow.ExecuteActivity(ctx, ValidateOrderActivity, order).Get(ctx, nil)
	if err != nil {
		logger.Error("订单验证失败", "error", err)
		return "", fmt.Errorf("订单验证失败: %w", err)
	}

	// 步骤 2: 预留库存
	logger.Info("步骤 2: 预留库存")
	for _, item := range order.Items {
		req := types.InventoryRequest{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
		}
		var invResult types.InventoryResult
		err := workflow.ExecuteActivity(ctx, ReserveInventoryActivity, req).Get(ctx, &invResult)
		if err != nil || !invResult.Success {
			logger.Error("库存预留失败", "productID", item.ProductID, "error", err)
			// 释放已预留的库存
			releaseInventory(ctx, order.Items)
			return "", fmt.Errorf("库存预留失败: %w", err)
		}
	}

	// 步骤 3: 处理支付
	logger.Info("步骤 3: 处理支付")
	paymentReq := types.PaymentRequest{
		OrderID:   order.ID,
		Amount:    order.TotalAmount,
		Method:    "credit_card",
	}
	var paymentResult types.PaymentResult
	err = workflow.ExecuteActivity(ctx, ProcessPaymentActivity, paymentReq).Get(ctx, &paymentResult)
	if err != nil || !paymentResult.Success {
		logger.Error("支付处理失败", "error", err)
		// 释放库存
		releaseInventory(ctx, order.Items)
		return "", fmt.Errorf("支付处理失败: %w", err)
	}

	// 步骤 4: 发货
	logger.Info("步骤 4: 发货")
	err = workflow.ExecuteActivity(ctx, ShipOrderActivity, order.ID).Get(ctx, nil)
	if err != nil {
		logger.Error("发货失败", "error", err)
		// 退款
		workflow.ExecuteActivity(ctx, RefundPaymentActivity, order.ID)
		// 释放库存
		releaseInventory(ctx, order.Items)
		return "", fmt.Errorf("发货失败: %w", err)
	}

	// 步骤 5: 发送通知
	logger.Info("步骤 5: 发送通知")
	notificationReq := types.NotificationRequest{
		CustomerID: order.CustomerID,
		Type:       "order_completed",
		Message:    fmt.Sprintf("订单 %s 已完成", order.ID),
	}
	err = workflow.ExecuteActivity(ctx, SendNotificationActivity, notificationReq).Get(ctx, nil)
	if err != nil {
		logger.Warn("通知发送失败", "error", err)
	}

	logger.Info("订单处理完成", "orderID", order.ID)
	return "completed", nil
}

// releaseInventory 释放库存
func releaseInventory(ctx workflow.Context, items []types.OrderItem) {
	for _, item := range items {
		req := types.InventoryRequest{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
		}
		workflow.ExecuteActivity(ctx, ReleaseInventoryActivity, req)
	}
}
```

## 4.4 活动定义

```go
// internal/activity/order.go
package activity

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.temporal.io/sdk/activity"

	"order-workflow/internal/types"
)

// ValidateOrderActivity 验证订单活动
func ValidateOrderActivity(ctx context.Context, order types.Order) error {
	logger := activity.GetLogger(ctx)
	logger.Info("验证订单", "orderID", order.ID)

	// 验证订单数据
	if order.ID == "" {
		return fmt.Errorf("订单ID不能为空")
	}
	if order.CustomerID == "" {
		return fmt.Errorf("客户ID不能为空")
	}
	if len(order.Items) == 0 {
		return fmt.Errorf("订单商品不能为空")
	}

	// 计算总价
	var total float64
	for _, item := range order.Items {
		total += item.Price * float64(item.Quantity)
	}

	if total != order.TotalAmount {
		return fmt.Errorf("订单金额不匹配: 期望 %.2f, 实际 %.2f", total, order.TotalAmount)
	}

	logger.Info("订单验证通过", "orderID", order.ID)
	return nil
}

// ReserveInventoryActivity 预留库存活动
func ReserveInventoryActivity(ctx context.Context, req types.InventoryRequest) (*types.InventoryResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("预留库存", "productID", req.ProductID, "quantity", req.Quantity)

	// 模拟库存预留
	// 实际应调用库存服务
	time.Sleep(100 * time.Millisecond)

	logger.Info("库存预留成功", "productID", req.ProductID)

	return &types.InventoryResult{
		Success:   true,
		Message:   "库存预留成功",
		Remaining: 100 - req.Quantity,
	}, nil
}

// ReleaseInventoryActivity 释放库存活动
func ReleaseInventoryActivity(ctx context.Context, req types.InventoryRequest) error {
	logger := activity.GetLogger(ctx)
	logger.Info("释放库存", "productID", req.ProductID, "quantity", req.Quantity)

	// 模拟库存释放
	time.Sleep(50 * time.Millisecond)

	return nil
}

// ProcessPaymentActivity 处理支付活动
func ProcessPaymentActivity(ctx context.Context, req types.PaymentRequest) (*types.PaymentResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("处理支付", "orderID", req.OrderID, "amount", req.Amount)

	// 模拟支付处理
	time.Sleep(200 * time.Millisecond)

	// 模拟支付成功
	logger.Info("支付处理成功", "orderID", req.OrderID)

	return &types.PaymentResult{
		Success:       true,
		TransactionID: fmt.Sprintf("TXN-%s-%d", req.OrderID, time.Now().Unix()),
		Message:       "支付成功",
	}, nil
}

// RefundPaymentActivity 退款活动
func RefundPaymentActivity(ctx context.Context, orderID string) error {
	logger := activity.GetLogger(ctx)
	logger.Info("退款", "orderID", orderID)

	// 模拟退款
	time.Sleep(150 * time.Millisecond)

	return nil
}

// ShipOrderActivity 发货活动
func ShipOrderActivity(ctx context.Context, orderID string) error {
	logger := activity.GetLogger(ctx)
	logger.Info("发货", "orderID", orderID)

	// 模拟发货
	time.Sleep(100 * time.Millisecond)

	logger.Info("发货成功", "orderID", orderID)
	return nil
}

// SendNotificationActivity 发送通知活动
func SendNotificationActivity(ctx context.Context, req types.NotificationRequest) error {
	logger := activity.GetLogger(ctx)
	logger.Info("发送通知", "customerID", req.CustomerID, "type", req.Type)

	// 模拟发送通知
	time.Sleep(50 * time.Millisecond)

	logger.Info("通知发送成功", "customerID", req.CustomerID)
	return nil
}
```

## 4.5 Worker 配置

```go
// cmd/worker/main.go
package main

import (
	"log"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	"order-workflow/internal/activity"
	"order-workflow/internal/workflow"
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
	w := worker.New(c, "order-task-queue", worker.Options{
		MaxConcurrentWorkflowTaskExecutionSize: 100,
		MaxConcurrentActivityExecutionSize:     50,
	})

	// 注册工作流
	w.RegisterWorkflow(workflow.OrderWorkflow)

	// 注册活动
	w.RegisterActivity(activity.ValidateOrderActivity)
	w.RegisterActivity(activity.ReserveInventoryActivity)
	w.RegisterActivity(activity.ReleaseInventoryActivity)
	w.RegisterActivity(activity.ProcessPaymentActivity)
	w.RegisterActivity(activity.RefundPaymentActivity)
	w.RegisterActivity(activity.ShipOrderActivity)
	w.RegisterActivity(activity.SendNotificationActivity)

	// 启动 Worker
	log.Println("启动 Worker...")
	if err := w.Start(); err != nil {
		log.Fatalln("无法启动 Worker", err)
	}

	log.Println("Worker 已启动，按 Ctrl+C 退出")
	select {}
}
```

## 4.6 工作流启动

```go
// cmd/starter/main.go
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.temporal.io/sdk/client"

	"order-workflow/internal/types"
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

	// 创建订单
	order := types.Order{
		ID:          "ORD-" + time.Now().Format("20060102150405"),
		CustomerID:  "CUST-001",
		TotalAmount: 299.99,
		Items: []types.OrderItem{
			{
				ProductID:   "PROD-001",
				ProductName: "iPhone 15",
				Quantity:    1,
				Price:       999.99,
			},
			{
				ProductID:   "PROD-002",
				ProductName: "AirPods Pro",
				Quantity:    1,
				Price:       249.99,
			},
			{
				ProductID:   "PROD-003",
				ProductName: "手机壳",
				Quantity:    2,
				Price:       24.99,
			},
		},
		CreatedAt: time.Now(),
	}

	// 修正总价
	order.TotalAmount = 999.99 + 249.99 + 24.99*2

	log.Printf("创建订单: %+v\n", order)

	// 启动工作流
	workflowID := "order-" + order.ID
	we, err := c.ExecuteWorkflow(
		context.Background(),
		client.StartWorkflowOptions{
			ID:        workflowID,
			TaskQueue: "order-task-queue",
		},
		workflow.OrderWorkflow,
		order,
	)
	if err != nil {
		log.Fatalln("无法启动工作流", err)
	}

	log.Printf("工作流已启动: %s\n", we.GetID())

	// 等待结果
	var result string
	err = we.Get(context.Background(), &result)
	if err != nil {
		log.Fatalln("工作流执行失败", err)
	}

	log.Printf("工作流完成: %s, 结果: %s\n", we.GetID(), result)
	fmt.Println("订单处理完成!")
}
```

## 4.7 配置文件

```yaml
# config.yaml
temporal:
  host: "localhost"
  port: 7233
  namespace: "default"

worker:
  task_queue: "order-task-queue"
  max_concurrent_workflow: 100
  max_concurrent_activity: 50

order:
  timeout: 300
  retry:
    max_attempts: 3
    initial_interval: 1
    backoff_coefficient: 2.0
```

## 4.8 运行项目

### 启动 Temporal 服务

```bash
# 使用 Docker Compose
docker-compose up -d

# 或使用 Temporal CLI
temporal server start-dev
```

### 启动 Worker

```bash
cd cmd/worker
go run main.go
```

### 启动工作流

```bash
cd cmd/starter
go run main.go
```

## 4.9 测试验证

### 查看工作流状态

```bash
temporal workflow list

# 输出示例
NAMESPACE  WORKFLOW ID             RUN ID           STATUS     START TIME
default    order-ORD-202401011200  a1b2c3d4-e5f6...  RUNNING    2024-01-01T12:00:00Z
```

### 查看工作流历史

```bash
temporal workflow show --workflow-id order-ORD-202401011200
```

### 测试取消

```bash
temporal workflow cancel --workflow-id order-ORD-202401011200
```

---

## 4.10 扩展练习

1. **添加订单状态查询**：实现查询订单当前状态的活动
2. **添加订单取消**：实现取消订单的补偿逻辑
3. **添加支付回调**：实现异步支付回调处理
4. **添加重试机制**：为各个活动配置不同的重试策略
5. **添加监控**：集成 Prometheus 监控
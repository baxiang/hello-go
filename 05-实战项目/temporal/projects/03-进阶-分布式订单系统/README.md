# 进阶项目：分布式订单处理系统

本项目实现一个基于 Temporal 的分布式订单处理系统，包含订单创建、支付处理、库存管理、订单状态追踪等完整功能。

## 5.1 项目概述

### 系统架构

```
┌─────────────────────────────────────────────────────────────────────────┐
│                         Order Processing System                          │
│                                                                          │
│  ┌──────────────┐    ┌──────────────┐    ┌──────────────┐              │
│  │   Starter    │───▶│   Worker     │───▶│  Temporal    │              │
│  │  (Client)    │    │  (Order)     │    │   Server     │              │
│  └──────────────┘    └──────────────┘    └──────┬───────┘              │
│                                                  │                       │
│  ┌──────────────────────────────────────────────────────────────────┐   │
│  │                      Worker Processes                            │   │
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐              │   │
│  │  │   Order     │  │  Payment    │  │ Inventory   │              │   │
│  │  │  Worker     │  │  Worker     │  │  Worker     │              │   │
│  │  └─────────────┘  └─────────────┘  └─────────────┘              │   │
│  └──────────────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────────────┘
```

### 项目结构

```
distributed-order/
├── cmd/
│   ├── worker/
│   │   ├── main.go
│   │   ├── order.go
│   │   ├── payment.go
│   │   └── inventory.go
│   ├── starter/
│   │   └── main.go
│   └── admin/
│       └── main.go
├── internal/
│   ├── workflow/
│   │   ├── order.go
│   │   ├── saga.go
│   │   └── batch.go
│   ├── activity/
│   │   ├── order.go
│   │   ├── payment.go
│   │   ├── inventory.go
│   │   └── notification.go
│   ├── types/
│   │   └── order.go
│   └── client/
│       └── client.go
├── config.yaml
└── go.mod
```

## 5.2 核心类型定义

```go
// internal/types/order.go
package types

import (
	"time"
)

type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "pending"
	OrderStatusValidated OrderStatus = "validated"
	OrderStatusPaid      OrderStatus = "paid"
	OrderStatusReserved  OrderStatus = "reserved"
	OrderStatusShipped   OrderStatus = "shipped"
	OrderStatusCompleted OrderStatus = "completed"
	OrderStatusCancelled OrderStatus = "cancelled"
	OrderStatusFailed    OrderStatus = "failed"
)

// Order 订单
type Order struct {
	ID           string                 `json:"id"`
	CustomerID   string                 `json:"customer_id"`
	Items        []OrderItem            `json:"items"`
	TotalAmount  float64                `json:"total_amount"`
	Status       OrderStatus            `json:"status"`
	PaymentID    string                 `json:"payment_id,omitempty"`
	ShippingID   string                 `json:"shipping_id,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
}

// OrderItem 订单商品
type OrderItem struct {
	ProductID    string  `json:"product_id"`
	ProductName  string  `json:"product_name"`
	Quantity     int     `json:"quantity"`
	Price        float64 `json:"price"`
	Subtotal     float64 `json:"subtotal"`
}

// Payment 支付信息
type Payment struct {
	ID             string    `json:"id"`
	OrderID        string    `json:"order_id"`
	Amount         float64   `json:"amount"`
	Method         string    `json:"method"`
	Status         string    `json:"status"`
	TransactionID  string    `json:"transaction_id"`
	PaidAt         time.Time `json:"paid_at,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
}

// Inventory 库存信息
type Inventory struct {
	ProductID   string `json:"product_id"`
	Available   int    `json:"available"`
	Reserved    int    `json:"reserved"`
}

// BatchOrder 批量订单
type BatchOrder struct {
	BatchID     string   `json:"batch_id"`
	Orders      []Order  `json:"orders"`
	Status      string   `json:"status"`
	Processed   int      `json:"processed"`
	Failed      int      `json:"failed"`
	CreatedAt   time.Time `json:"created_at"`
}
```

## 5.3 订单工作流

```go
// internal/workflow/order.go
package workflow

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/workflow"

	"distributed-order/internal/types"
)

// OrderProcessingWorkflow 订单处理工作流
func OrderProcessingWorkflow(ctx workflow.Context, order types.Order) (string, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("开始处理订单", "orderID", order.ID, "amount", order.TotalAmount)

	// 记录开始时间
	startTime := workflow.Now(ctx)

	// 活动选项
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
	logger.Info("验证订单")
	if err := workflow.ExecuteActivity(ctx, ValidateOrderActivity, order).Get(ctx, nil); err != nil {
		return handleFailure(ctx, order, "validation", err)
	}

	// 步骤 2: 预留库存（并行）
	logger.Info("预留库存")
	inventoryResults := make(map[string]bool)
	for _, item := range order.Items {
		future := workflow.ExecuteActivity(ctx, ReserveInventoryActivity, types.Inventory{
			ProductID: item.ProductID,
			Available: item.Quantity,
		})
		
		// 等待完成
		var result types.Inventory
		if err := future.Get(ctx, &result); err != nil {
			return handleFailure(ctx, order, "inventory", err)
		}
		inventoryResults[item.ProductID] = result.Available >= 0
	}

	// 检查库存预留结果
	for productID, success := range inventoryResults {
		if !success {
			// 释放已预留的库存
			releaseReservedInventory(ctx, order.Items)
			return handleFailure(ctx, order, "inventory", fmt.Errorf("商品 %s 库存不足", productID))
		}
	}

	// 步骤 3: 处理支付
	logger.Info("处理支付")
	var payment types.Payment
	err := workflow.ExecuteActivity(ctx, ProcessPaymentActivity, types.Payment{
		OrderID: order.ID,
		Amount:  order.TotalAmount,
		Method:  "credit_card",
	}).Get(ctx, &payment)
	
	if err != nil || payment.Status != "success" {
		// 释放库存
		releaseReservedInventory(ctx, order.Items)
		return handleFailure(ctx, order, "payment", err)
	}

	// 步骤 4: 发货
	logger.Info("发货")
	err = workflow.ExecuteActivity(ctx, ShipOrderActivity, order.ID).Get(ctx, nil)
	if err != nil {
		// 退款
		workflow.ExecuteActivity(ctx, RefundPaymentActivity, payment.TransactionID)
		// 释放库存
		releaseReservedInventory(ctx, order.Items)
		return handleFailure(ctx, order, "shipping", err)
	}

	// 步骤 5: 发送通知
	logger.Info("发送通知")
	workflow.ExecuteActivity(ctx, SendOrderNotificationActivity, types.Notification{
		CustomerID: order.CustomerID,
		Type:        "order_completed",
		OrderID:     order.ID,
		Message:     "您的订单已发货",
	}).Get(ctx, nil)

	// 计算耗时
	duration := workflow.Now(ctx).Sub(startTime)
	logger.Info("订单处理完成", 
		"orderID", order.ID, 
		"duration", duration.String())

	return "completed", nil
}

// handleFailure 处理失败
func handleFailure(ctx workflow.Context, order types.Order, stage string, err error) (string, error) {
	logger := workflow.GetLogger(ctx)
	logger.Error("订单处理失败",
		"orderID", order.ID,
		"stage", stage,
		"error", err)

	// 发送失败通知
	workflow.ExecuteActivity(ctx, SendOrderNotificationActivity, types.Notification{
		CustomerID: order.CustomerID,
		Type:       "order_failed",
		OrderID:    order.ID,
		Message:    fmt.Sprintf("订单处理失败: %s", err),
	})

	return "", fmt.Errorf("订单在 %s 阶段失败: %w", stage, err)
}

// releaseReservedInventory 释放预留库存
func releaseReservedInventory(ctx workflow.Context, items []types.OrderItem) {
	for _, item := range items {
		workflow.ExecuteActivity(ctx, ReleaseInventoryActivity, types.Inventory{
			ProductID: item.ProductID,
			Available: -item.Quantity,
		})
	}
}
```

## 5.4 Saga 工作流

```go
// internal/workflow/saga.go
package workflow

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/workflow"

	"distributed-order/internal/types"
)

// OrderSagaWorkflow 订单 Saga 工作流
func OrderSagaWorkflow(ctx workflow.Context, order types.Order) (string, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("开始 Saga 订单处理", "orderID", order.ID)

	// 定义补偿活动
	compensations := []func(){}

	// 活动选项
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 5 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumAttempts:    3,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	// 步骤 1: 创建订单
	logger.Info("创建订单")
	var orderID string
	err := workflow.ExecuteActivity(ctx, CreateOrderActivity, order).Get(ctx, &orderID)
	if err != nil {
		return "", err
	}
	order.ID = orderID
	compensations = append(compensations, func() {
		workflow.ExecuteActivity(ctx, CancelOrderActivity, orderID)
	})

	// 步骤 2: 预留库存
	logger.Info("预留库存")
	err = workflow.ExecuteActivity(ctx, ReserveInventoryBatchActivity, order.Items).Get(ctx, nil)
	if err != nil {
		executeCompensations(ctx, compensations)
		return "", err
	}
	compensations = append(compensations, func() {
		workflow.ExecuteActivity(ctx, ReleaseInventoryBatchActivity, order.Items)
	})

	// 步骤 3: 处理支付
	logger.Info("处理支付")
	var paymentID string
	err = workflow.ExecuteActivity(ctx, ProcessPaymentActivity, types.Payment{
		OrderID: orderID,
		Amount:  order.TotalAmount,
		Method:  "credit_card",
	}).Get(ctx, &paymentID)
	if err != nil {
		executeCompensations(ctx, compensations)
		return "", err
	}
	compensations = append(compensations, func() {
		workflow.ExecuteActivity(ctx, RefundPaymentActivity, paymentID)
	})

	// 步骤 4: 发货
	logger.Info("发货")
	err = workflow.ExecuteActivity(ctx, ShipOrderActivity, orderID).Get(ctx, nil)
	if err != nil {
		executeCompensations(ctx, compensations)
		return "", err
	}

	// 步骤 5: 发送通知
	logger.Info("发送通知")
	workflow.ExecuteActivity(ctx, SendOrderNotificationActivity, types.Notification{
		CustomerID: order.CustomerID,
		Type:        "order_completed",
		OrderID:     orderID,
		Message:    "订单处理完成",
	})

	logger.Info("Saga 订单处理完成", "orderID", orderID)
	return orderID, nil
}

// executeCompensations 执行补偿
func executeCompensations(ctx workflow.Context, compensations []func()) {
	logger := workflow.GetLogger(ctx)
	logger.Info("执行补偿")

	// 逆向执行补偿
	for i := len(compensations) - 1; i >= 0; i-- {
		compensations[i]()
	}
}
```

## 5.5 批量处理工作流

```go
// internal/workflow/batch.go
package workflow

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/workflow"

	"distributed-order/internal/types"
)

// BatchOrderProcessingWorkflow 批量订单处理工作流
func BatchOrderProcessingWorkflow(ctx workflow.Context, batch types.BatchOrder) (*types.BatchOrder, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("开始批量订单处理", "batchID", batch.BatchID, "count", len(batch.Orders))

	// 活动选项
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 30 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumAttempts:    2,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	// 使用信号处理每个订单
	orderChannel := workflow.NewSignalChannel(ctx, "order-channel")
	resultChannel := workflow.NewChannel(ctx)

	// 启动订单处理协程
	workflow.Go(ctx, func(ctx workflow.Context) {
		processed := 0
		failed := 0

		for _, order := range batch.Orders {
			// 处理订单
			var result string
			err := workflow.ExecuteActivity(ctx, ProcessSingleOrderActivity, order).Get(ctx, &result)
			
			if err != nil {
				logger.Error("订单处理失败", "orderID", order.ID, "error", err)
				failed++
			} else {
				processed++
			}

			// 发送进度更新
			resultChannel.Send(ctx, BatchProgress{
				Processed: processed,
				Failed:    failed,
				Total:     len(batch.Orders),
			})
		}

		// 关闭结果通道
		resultChannel.Close()
	})

	// 收集结果
	processed := 0
	failed := 0
	
	for resultChannel.Receive(ctx, nil) {
		// 可以在这里更新进度
		processed++
		if processed%10 == 0 {
			logger.Info("批量处理进度", "processed", processed, "total", len(batch.Orders))
		}
	}

	batch.Status = "completed"
	batch.Processed = processed
	batch.Failed = failed

	logger.Info("批量订单处理完成", 
		"batchID", batch.BatchID, 
		"processed", processed, 
		"failed", failed)

	return &batch, nil
}

// BatchProgress 批量进度
type BatchProgress struct {
	Processed int
	Failed    int
	Total     int
}
```

## 5.6 活动实现

```go
// internal/activity/order.go
package activity

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.temporal.io/sdk/activity"

	"distributed-order/internal/types"
)

// ValidateOrderActivity 验证订单活动
func ValidateOrderActivity(ctx context.Context, order types.Order) error {
	logger := activity.GetLogger(ctx)
	logger.Info("验证订单", "orderID", order.ID)

	// 验证逻辑
	if order.ID == "" {
		return fmt.Errorf("订单ID不能为空")
	}
	if order.CustomerID == "" {
		return fmt.Errorf("客户ID不能为空")
	}
	if len(order.Items) == 0 {
		return fmt.Errorf("订单商品不能为空")
	}

	// 验证商品数量和价格
	for _, item := range order.Items {
		if item.Quantity <= 0 {
			return fmt.Errorf("商品 %s 数量必须大于0", item.ProductID)
		}
		if item.Price <= 0 {
			return fmt.Errorf("商品 %s 价格必须大于0", item.ProductID)
		}
	}

	logger.Info("订单验证通过", "orderID", order.ID)
	return nil
}

// CreateOrderActivity 创建订单活动
func CreateOrderActivity(ctx context.Context, order types.Order) (string, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("创建订单", "orderID", order.ID)

	// 模拟创建订单
	time.Sleep(100 * time.Millisecond)

	// 生成订单ID
	orderID := fmt.Sprintf("ORD-%d", time.Now().UnixNano())
	
	logger.Info("订单创建成功", "orderID", orderID)
	return orderID, nil
}

// CancelOrderActivity 取消订单活动
func CancelOrderActivity(ctx context.Context, orderID string) error {
	logger := activity.GetLogger(ctx)
	logger.Info("取消订单", "orderID", orderID)

	// 模拟取消订单
	time.Sleep(50 * time.Millisecond)

	return nil
}

// ProcessSingleOrderActivity 处理单个订单活动
func ProcessSingleOrderActivity(ctx context.Context, order types.Order) (string, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("处理订单", "orderID", order.ID)

	// 验证订单
	if err := ValidateOrderActivity(ctx, order); err != nil {
		return "", err
	}

	// 预留库存
	for _, item := range order.Items {
		_, err := ReserveInventoryActivity(ctx, types.Inventory{
			ProductID: item.ProductID,
			Available: item.Quantity,
		})
		if err != nil {
			return "", err
		}
	}

	// 处理支付
	_, err := ProcessPaymentActivity(ctx, types.Payment{
		OrderID: order.ID,
		Amount:  order.TotalAmount,
		Method:  "credit_card",
	})
	if err != nil {
		return "", err
	}

	logger.Info("订单处理完成", "orderID", order.ID)
	return "completed", nil
}
```

```go
// internal/activity/payment.go
package activity

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.temporal.io/sdk/activity"

	"distributed-order/internal/types"
)

// ProcessPaymentActivity 处理支付活动
func ProcessPaymentActivity(ctx context.Context, payment types.Payment) (string, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("处理支付", "orderID", payment.OrderID, "amount", payment.Amount)

	// 模拟支付处理
	time.Sleep(200 * time.Millisecond)

	// 生成支付ID
	paymentID := fmt.Sprintf("PAY-%d", time.Now().UnixNano())

	logger.Info("支付处理成功", "paymentID", paymentID)
	
	return paymentID, nil
}

// RefundPaymentActivity 退款活动
func RefundPaymentActivity(ctx context.Context, paymentID string) error {
	logger := activity.GetLogger(ctx)
	logger.Info("退款", "paymentID", paymentID)

	// 模拟退款
	time.Sleep(150 * time.Millisecond)

	logger.Info("退款成功", "paymentID", paymentID)
	return nil
}
```

```go
// internal/activity/inventory.go
package activity

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.temporal.io/sdk/activity"

	"distributed-order/internal/types"
)

// ReserveInventoryActivity 预留库存活动
func ReserveInventoryActivity(ctx context.Context, inv types.Inventory) (*types.Inventory, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("预留库存", "productID", inv.ProductID, "quantity", inv.Available)

	// 模拟库存预留
	time.Sleep(100 * time.Millisecond)

	// 假设库存充足
	result := &types.Inventory{
		ProductID: inv.ProductID,
		Available: 100 - inv.Available,
		Reserved:  inv.Available,
	}

	logger.Info("库存预留成功", "productID", inv.ProductID)
	return result, nil
}

// ReleaseInventoryActivity 释放库存活动
func ReleaseInventoryActivity(ctx context.Context, inv types.Inventory) error {
	logger := activity.GetLogger(ctx)
	logger.Info("释放库存", "productID", inv.ProductID, "quantity", inv.Available)

	// 模拟释放库存
	time.Sleep(50 * time.Millisecond)

	return nil
}

// ReserveInventoryBatchActivity 批量预留库存活动
func ReserveInventoryBatchActivity(ctx context.Context, items []types.OrderItem) error {
	logger := activity.GetLogger(ctx)
	logger.Info("批量预留库存", "count", len(items))

	for _, item := range items {
		_, err := ReserveInventoryActivity(ctx, types.Inventory{
			ProductID: item.ProductID,
			Available: item.Quantity,
		})
		if err != nil {
			return err
		}
	}

	logger.Info("批量预留库存完成")
	return nil
}

// ReleaseInventoryBatchActivity 批量释放库存活动
func ReleaseInventoryBatchActivity(ctx context.Context, items []types.OrderItem) error {
	logger := activity.GetLogger(ctx)
	logger.Info("批量释放库存", "count", len(items))

	for _, item := range items {
		ReleaseInventoryActivity(ctx, types.Inventory{
			ProductID: item.ProductID,
			Available: -item.Quantity,
		})
	}

	return nil
}
```

```go
// internal/activity/notification.go
package activity

import (
	"context"
	"fmt"
	"log"

	"go.temporal.io/sdk/activity"

	"distributed-order/internal/types"
)

// Notification 通知
type Notification struct {
	CustomerID string
	Type       string
	OrderID    string
	Message    string
}

// SendOrderNotificationActivity 发送订单通知活动
func SendOrderNotificationActivity(ctx context.Context, notif Notification) error {
	logger := activity.GetLogger(ctx)
	logger.Info("发送通知", "customerID", notif.CustomerID, "type", notif.Type)

	// 模拟发送通知
	// 实际应调用邮件/短信服务

	fmt.Printf("通知: 客户 %s, 类型 %s, 消息 %s\n", 
		notif.CustomerID, notif.Type, notif.Message)

	logger.Info("通知发送成功", "customerID", notif.CustomerID)
	return nil
}
```

## 5.7 客户端封装

```go
// internal/client/client.go
package client

import (
	"context"
	"fmt"
	"time"

	"go.temporal.io/sdk/client"

	"distributed-order/internal/types"
)

// OrderClient 订单客户端
type OrderClient struct {
	c client.Client
}

// NewOrderClient 创建订单客户端
func NewOrderClient(opts client.Options) (*OrderClient, error) {
	c, err := client.Dial(opts)
	if err != nil {
		return nil, err
	}
	return &OrderClient{c: c}, nil
}

// CreateOrder 创建订单
func (oc *OrderClient) CreateOrder(ctx context.Context, order types.Order) (string, error) {
	workflowID := fmt.Sprintf("order-%s", order.ID)

	we, err := oc.c.ExecuteWorkflow(
		ctx,
		client.StartWorkflowOptions{
			ID:        workflowID,
			TaskQueue: "order-task-queue",
		},
		workflow.OrderProcessingWorkflow,
		order,
	)
	if err != nil {
		return "", err
	}

	return we.GetID(), nil
}

// CreateOrderWithSaga 使用 Saga 创建订单
func (oc *OrderClient) CreateOrderWithSaga(ctx context.Context, order types.Order) (string, error) {
	workflowID := fmt.Sprintf("saga-%s", order.ID)

	we, err := oc.c.ExecuteWorkflow(
		ctx,
		client.StartWorkflowOptions{
			ID:        workflowID,
			TaskQueue: "order-task-queue",
		},
		workflow.OrderSagaWorkflow,
		order,
	)
	if err != nil {
		return "", err
	}

	return we.GetID(), nil
}

// ProcessBatch 批量处理订单
func (oc *OrderClient) ProcessBatch(ctx context.Context, orders []types.Order) (string, error) {
	batch := types.BatchOrder{
		BatchID:   fmt.Sprintf("batch-%d", time.Now().Unix()),
		Orders:    orders,
		Status:    "processing",
		CreatedAt: time.Now(),
	}

	we, err := oc.c.ExecuteWorkflow(
		ctx,
		client.StartWorkflowOptions{
			ID:        batch.BatchID,
			TaskQueue: "batch-task-queue",
		},
		workflow.BatchOrderProcessingWorkflow,
		batch,
	)
	if err != nil {
		return "", err
	}

	return we.GetID(), nil
}

// GetOrderStatus 获取订单状态
func (oc *OrderClient) GetOrderStatus(ctx context.Context, workflowID string) (string, error) {
	we := oc.c.GetWorkflow(ctx, workflowID, "")
	
	var result string
	err := we.Get(ctx, &result)
	if err != nil {
		return "", err
	}

	return result, nil
}

// CancelOrder 取消订单
func (oc *OrderClient) CancelOrder(ctx context.Context, workflowID string) error {
	return oc.c.CancelWorkflow(ctx, workflowID, "")
}

// SignalOrder 发送信号
func (oc *OrderClient) SignalOrder(ctx context.Context, workflowID, signalName string, data interface{}) error {
	return oc.c.SignalWorkflow(ctx, workflowID "", signalName, data)
}

// Close 关闭客户端
func (oc *OrderClient) Close() error {
	return oc.c.Close()
}
```

## 5.8 Worker 配置

```go
// cmd/worker/main.go
package main

import (
	"log"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	"distributed-order/internal/activity"
	"distributed-order/internal/workflow"
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

	// 创建订单 Worker
	orderWorker := worker.New(c, "order-task-queue", worker.Options{
		MaxConcurrentWorkflowTaskExecutionSize: 100,
		MaxConcurrentActivityExecutionSize:     50,
	})
	orderWorker.RegisterWorkflow(workflow.OrderProcessingWorkflow)
	orderWorker.RegisterWorkflow(workflow.OrderSagaWorkflow)
	
	// 注册订单活动
	orderWorker.RegisterActivity(activity.ValidateOrderActivity)
	orderWorker.RegisterActivity(activity.CreateOrderActivity)
	orderWorker.RegisterActivity(activity.CancelOrderActivity)
	orderWorker.RegisterActivity(activity.ProcessSingleOrderActivity)
	orderWorker.RegisterActivity(activity.ProcessPaymentActivity)
	orderWorker.RegisterActivity(activity.RefundPaymentActivity)
	orderWorker.RegisterActivity(activity.ReserveInventoryActivity)
	orderWorker.RegisterActivity(activity.ReleaseInventoryActivity)
	orderWorker.RegisterActivity(activity.ReserveInventoryBatchActivity)
	orderWorker.RegisterActivity(activity.ReleaseInventoryBatchActivity)
	orderWorker.RegisterActivity(activity.ShipOrderActivity)
	orderWorker.RegisterActivity(activity.SendOrderNotificationActivity)

	// 创建批量处理 Worker
	batchWorker := worker.New(c, "batch-task-queue", worker.Options{
		MaxConcurrentWorkflowTaskExecutionSize: 10,
		MaxConcurrentActivityExecutionSize:    20,
	})
	batchWorker.RegisterWorkflow(workflow.BatchOrderProcessingWorkflow)

	// 启动 Workers
	log.Println("启动 Workers...")
	
	if err := orderWorker.Start(); err != nil {
		log.Fatalln("无法启动订单 Worker", err)
	}
	
	if err := batchWorker.Start(); err != nil {
		log.Fatalln("无法启动批量 Worker", err)
	}

	log.Println("Workers 已启动")
	select {}
}
```

## 5.9 运行项目

### 启动服务

```bash
# 启动 Temporal
temporal server start-dev

# 启动 Worker
go run cmd/worker/main.go
```

### 启动客户端

```go
// cmd/starter/main.go
package main

import (
	"context"
	"log"
	"time"

	"distributed-order/internal/client"
	"distributed-order/internal/types"
)

func main() {
	// 创建客户端
	oc, err := client.NewOrderClient(client.Options{
		HostPort: "localhost:7233",
	})
	if err != nil {
		log.Fatalln("无法创建客户端", err)
	}
	defer oc.Close()

	// 创建订单
	order := types.Order{
		ID:          "ORD-001",
		CustomerID:  "CUST-001",
		TotalAmount: 299.99,
		Items: []types.OrderItem{
			{ProductID: "PROD-001", ProductName: "商品1", Quantity: 2, Price: 99.99},
			{ProductID: "PROD-002", ProductName: "商品2", Quantity: 1, Price: 100.00},
		},
		CreatedAt: time.Now(),
	}

	// 启动工作流
	workflowID, err := oc.CreateOrder(context.Background(), order)
	if err != nil {
		log.Fatalln("无法启动工作流", err)
	}

	log.Printf("工作流已启动: %s", workflowID)
}
```

---

## 5.10 扩展练习

1. **添加支付回调**：实现异步支付回调处理
2. **添加订单查询**：实现订单状态查询功能
3. **添加定时任务**：实现定时清理过期订单
4. **添加监控告警**：集成 Prometheus 监控
5. **添加分布式追踪**：集成 OpenTelemetry
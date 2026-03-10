# 高级项目：微服务电商工作流平台

本项目实现一个基于 Temporal 的微服务电商工作流平台，包含订单、支付、库存、通知等多个服务的工作流编排。

## 6.1 项目概述

### 系统架构

```
┌─────────────────────────────────────────────────────────────────────────┐
│                         API Gateway (8080)                               │
└─────────────────────────────┬───────────────────────────────────────────┘
                              │
        ┌─────────────────────┼─────────────────────┐
        │                     │                     │
        ▼                     ▼                     ▼
┌───────────────┐   ┌───────────────┐   ┌───────────────┐
│ Order Service │   │Payment Service│   │ Stock Service │
│   (Worker)    │   │   (Worker)    │   │   (Worker)    │
│  Port: 8091   │   │  Port: 8092   │   │  Port: 8093   │
└───────┬───────┘   └───────┬───────┘   └───────┬───────┘
        │                   │                     │
        └───────────────────┼─────────────────────┘
                            │
                            ▼
                    ┌───────────────┐
                    │ Notification  │
                    │   Service     │
                    │   (Worker)    │
                    │  Port: 8094   │
                    └───────────────┘
                            │
                            ▼
                    ┌───────────────┐
                    │   Temporal    │
                    │   Server      │
                    │  Port: 7233   │
                    └───────────────┘
```

### 服务列表

| 服务 | 端口 | 任务队列 | 说明 |
|------|------|----------|------|
| order-service | 8091 | order-task-queue | 订单处理 |
| payment-service | 8092 | payment-task-queue | 支付处理 |
| stock-service | 8093 | stock-task-queue | 库存管理 |
| notification-service | 8094 | notification-task-queue | 通知服务 |

### 项目结构

```
ecommerce-workflow/
├── cmd/
│   ├── order-service/main.go
│   ├── payment-service/main.go
│   ├── stock-service/main.go
│   ├── notification-service/main.go
│   └── api/main.go
├── internal/
│   ├── workflow/
│   │   ├── order.go
│   │   ├── payment.go
│   │   └── saga.go
│   ├── activity/
│   │   ├── order.go
│   │   ├── payment.go
│   │   ├── stock.go
│   │   └── notification.go
│   ├── types/
│   │   └── types.go
│   └── client/
│       └── client.go
├── api/
│   └── proto/
│       └── v1/
│           └── order.proto
├── configs/
│   └── config.yaml
└── deployments/
    └── docker/
        └── docker-compose.yml
```

## 6.2 核心工作流

### 6.2.1 订单创建工作流

```go
// internal/workflow/order.go
package workflow

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/workflow"

	"ecommerce-workflow/internal/types"
)

// CreateOrderWorkflow 创建订单工作流
func CreateOrderWorkflow(ctx workflow.Context, order types.Order) (*types.OrderResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("创建订单工作流开始", "orderID", order.ID)

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

	// 步骤 1: 创建订单记录
	logger.Info("创建订单记录")
	var orderID string
	err := workflow.ExecuteActivity(ctx, CreateOrderActivity, order).Get(ctx, &orderID)
	if err != nil {
		return nil, fmt.Errorf("创建订单失败: %w", err)
	}
	order.ID = orderID

	// 步骤 2: 验证库存（调用库存服务）
	logger.Info("验证库存")
	var stockResult types.StockResult
	err = workflow.ExecuteActivity(ctx, CheckStockActivity, types.StockRequest{
		Items: order.Items,
	}).Get(ctx, &stockResult)
	if err != nil || !stockResult.Success {
		// 更新订单状态为失败
		workflow.ExecuteActivity(ctx, UpdateOrderStatusActivity, types.UpdateOrderRequest{
			OrderID: orderID,
			Status:  "stock_failed",
		})
		return nil, fmt.Errorf("库存验证失败: %w", err)
	}

	// 步骤 3: 预留库存
	logger.Info("预留库存")
	err = workflow.ExecuteActivity(ctx, ReserveStockActivity, types.StockRequest{
		OrderID: orderID,
		Items:   order.Items,
	}).Get(ctx, nil)
	if err != nil {
		// 释放已预留的库存
		workflow.ExecuteActivity(ctx, ReleaseStockActivity, orderID)
		return nil, fmt.Errorf("库存预留失败: %w", err)
	}

	// 步骤 4: 创建支付（调用支付服务）
	logger.Info("创建支付")
	var paymentResult types.PaymentResult
	err = workflow.ExecuteActivity(ctx, CreatePaymentActivity, types.PaymentRequest{
		OrderID:   orderID,
		Amount:    order.TotalAmount,
		CustomerID: order.CustomerID,
	}).Get(ctx, &paymentResult)
	if err != nil {
		// 释放库存
		workflow.ExecuteActivity(ctx, ReleaseStockActivity, orderID)
		return nil, fmt.Errorf("创建支付失败: %w", err)
	}

	// 步骤 5: 处理支付
	logger.Info("处理支付")
	var payResult types.PaymentResult
	err = workflow.ExecuteActivity(ctx, ProcessPaymentActivity, types.PaymentRequest{
		PaymentID: paymentResult.PaymentID,
		Amount:    order.TotalAmount,
	}).Get(ctx, &payResult)
	if err != nil || !payResult.Success {
		// 释放库存
		workflow.ExecuteActivity(ctx, ReleaseStockActivity, orderID)
		// 更新订单状态
		workflow.ExecuteActivity(ctx, UpdateOrderStatusActivity, types.UpdateOrderRequest{
			OrderID: orderID,
			Status:  "payment_failed",
		})
		return nil, fmt.Errorf("支付处理失败: %w", err)
	}

	// 步骤 6: 更新订单状态
	logger.Info("更新订单状态")
	err = workflow.ExecuteActivity(ctx, UpdateOrderStatusActivity, types.UpdateOrderRequest{
		OrderID:   orderID,
		Status:    "paid",
		PaymentID: payResult.PaymentID,
	}).Get(ctx, nil)
	if err != nil {
		logger.Error("更新订单状态失败", "error", err)
	}

	// 步骤 7: 发送通知
	logger.Info("发送通知")
	workflow.ExecuteActivity(ctx, SendNotificationActivity, types.NotificationRequest{
		CustomerID: order.CustomerID,
		Type:       "order_created",
		Title:      "订单创建成功",
		Content:    fmt.Sprintf("您的订单 %s 已创建，金额 %.2f", orderID, order.TotalAmount),
	})

	logger.Info("创建订单工作流完成", "orderID", orderID)
	
	return &types.OrderResult{
		OrderID: orderID,
		Status:  "created",
	}, nil
}

// CancelOrderWorkflow 取消订单工作流
func CancelOrderWorkflow(ctx workflow.Context, orderID, reason string) error {
	logger := workflow.GetLogger(ctx)
	logger.Info("取消订单工作流开始", "orderID", orderID)

	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 2 * time.Minute,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	// 获取订单信息
	var order types.Order
	err := workflow.ExecuteActivity(ctx, GetOrderActivity, orderID).Get(ctx, &order)
	if err != nil {
		return fmt.Errorf("获取订单失败: %w", err)
	}

	// 如果已支付，退款
	if order.Status == "paid" {
		logger.Info("处理退款")
		err = workflow.ExecuteActivity(ctx, RefundPaymentActivity, order.PaymentID).Get(ctx, nil)
		if err != nil {
			logger.Error("退款失败", "error", err)
			// 继续执行，不阻塞取消流程
		}
	}

	// 释放库存
	logger.Info("释放库存")
	workflow.ExecuteActivity(ctx, ReleaseStockActivity, orderID)

	// 更新订单状态
	logger.Info("更新订单状态")
	err = workflow.ExecuteActivity(ctx, UpdateOrderStatusActivity, types.UpdateOrderRequest{
		OrderID: orderID,
		Status:  "cancelled",
		Remark:  reason,
	}).Get(ctx, nil)

	// 发送取消通知
	workflow.ExecuteActivity(ctx, SendNotificationActivity, types.NotificationRequest{
		CustomerID: order.CustomerID,
		Type:       "order_cancelled",
		Title:      "订单已取消",
		Content:    fmt.Sprintf("您的订单 %s 已取消", orderID),
	})

	logger.Info("取消订单工作流完成", "orderID", orderID)
	return nil
}
```

### 6.2.2 支付工作流

```go
// internal/workflow/payment.go
package workflow

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/workflow"

	"ecommerce-workflow/internal/types"
)

// PaymentWorkflow 支付工作流
func PaymentWorkflow(ctx workflow.Context, req types.PaymentRequest) (*types.PaymentResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("支付工作流开始", "paymentID", req.PaymentID)

	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 5 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumAttempts:    3,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	// 步骤 1: 验证支付
	logger.Info("验证支付")
	err := workflow.ExecuteActivity(ctx, VerifyPaymentActivity, req).Get(ctx, nil)
	if err != nil {
		return &types.PaymentResult{
			Success: false,
			Message: "支付验证失败",
		}, err
	}

	// 步骤 2: 处理支付（调用第三方支付）
	logger.Info("处理支付")
	var payResult types.PaymentResult
	err = workflow.ExecuteActivity(ctx, ProcessPaymentActivity, req).Get(ctx, &payResult)
	if err != nil || !payResult.Success {
		// 发送支付失败通知
		workflow.ExecuteActivity(ctx, SendNotificationActivity, types.NotificationRequest{
			CustomerID: req.CustomerID,
			Type:       "payment_failed",
			Title:      "支付失败",
			Content:    "您的支付处理失败，请重试",
		})
		return &payResult, err
	}

	// 步骤 3: 更新订单状态
	logger.Info("更新订单状态")
	workflow.ExecuteActivity(ctx, UpdateOrderStatusActivity, types.UpdateOrderRequest{
		OrderID:   req.OrderID,
		Status:    "paid",
		PaymentID: payResult.PaymentID,
	})

	// 步骤 4: 发送支付成功通知
	logger.Info("发送通知")
	workflow.ExecuteActivity(ctx, SendNotificationActivity, types.NotificationRequest{
		CustomerID: req.CustomerID,
		Type:       "payment_success",
		Title:      "支付成功",
		Content:    fmt.Sprintf("您的订单已支付成功，金额 %.2f", req.Amount),
	})

	logger.Info("支付工作流完成", "paymentID", req.PaymentID)
	return &payResult, nil
}

// RefundWorkflow 退款工作流
func RefundWorkflow(ctx workflow.Context, req types.RefundRequest) (*types.RefundResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("退款工作流开始", "orderID", req.OrderID)

	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 3 * time.Minute,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	// 步骤 1: 获取支付信息
	var payment types.Payment
	err := workflow.ExecuteActivity(ctx, GetPaymentActivity, req.OrderID).Get(ctx, &payment)
	if err != nil {
		return nil, fmt.Errorf("获取支付信息失败: %w", err)
	}

	// 步骤 2: 执行退款
	logger.Info("执行退款")
	var refundResult types.RefundResult
	err = workflow.ExecuteActivity(ctx, ExecuteRefundActivity, types.RefundRequest{
		PaymentID:   payment.PaymentID,
		Amount:      req.Amount,
		Reason:      req.Reason,
	}).Get(ctx, &refundResult)
	if err != nil {
		return nil, fmt.Errorf("退款失败: %w", err)
	}

	// 步骤 3: 更新订单状态
	workflow.ExecuteActivity(ctx, UpdateOrderStatusActivity, types.UpdateOrderRequest{
		OrderID: req.OrderID,
		Status:  "refunded",
	})

	// 步骤 4: 释放库存
	workflow.ExecuteActivity(ctx, ReleaseStockActivity, req.OrderID)

	logger.Info("退款工作流完成", "orderID", req.OrderID)
	return &refundResult, nil
}
```

## 6.3 活动实现

### 6.3.1 订单活动

```go
// internal/activity/order.go
package activity

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.temporal.io/sdk/activity"

	"ecommerce-workflow/internal/types"
)

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

// GetOrderActivity 获取订单活动
func GetOrderActivity(ctx context.Context, orderID string) (*types.Order, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("获取订单", "orderID", orderID)

	// 模拟获取订单
	return &types.Order{
		ID:         orderID,
		CustomerID: "CUST-001",
		Status:     "pending",
		TotalAmount: 100.00,
	}, nil
}

// UpdateOrderStatusActivity 更新订单状态活动
func UpdateOrderStatusActivity(ctx context.Context, req types.UpdateOrderRequest) error {
	logger := activity.GetLogger(ctx)
	logger.Info("更新订单状态", "orderID", req.OrderID, "status", req.Status)

	// 模拟更新
	time.Sleep(50 * time.Millisecond)

	logger.Info("订单状态更新成功")
	return nil
}
```

### 6.3.2 支付活动

```go
// internal/activity/payment.go
package activity

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.temporal.io/sdk/activity"

	"ecommerce-workflow/internal/types"
)

// CreatePaymentActivity 创建支付活动
func CreatePaymentActivity(ctx context.Context, req types.PaymentRequest) (*types.PaymentResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("创建支付", "orderID", req.OrderID)

	// 模拟创建支付
	time.Sleep(100 * time.Millisecond)

	paymentID := fmt.Sprintf("PAY-%d", time.Now().UnixNano())

	logger.Info("支付创建成功", "paymentID", paymentID)
	return &types.PaymentResult{
		Success:   true,
		PaymentID: paymentID,
		Message:   "支付创建成功",
	}, nil
}

// ProcessPaymentActivity 处理支付活动
func ProcessPaymentActivity(ctx context.Context, req types.PaymentRequest) (*types.PaymentResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("处理支付", "paymentID", req.PaymentID, "amount", req.Amount)

	// 模拟支付处理
	time.Sleep(200 * time.Millisecond)

	// 模拟支付成功
	transactionID := fmt.Sprintf("TXN-%d", time.Now().UnixNano())

	logger.Info("支付处理成功", "transactionID", transactionID)
	return &types.PaymentResult{
		Success:       true,
		PaymentID:    req.PaymentID,
		TransactionID: transactionID,
		Message:      "支付成功",
	}, nil
}

// VerifyPaymentActivity 验证支付活动
func VerifyPaymentActivity(ctx context.Context, req types.PaymentRequest) error {
	logger := activity.GetLogger(ctx)
	logger.Info("验证支付", "orderID", req.OrderID)

	// 模拟验证
	time.Sleep(50 * time.Millisecond)

	return nil
}

// GetPaymentActivity 获取支付活动
func GetPaymentActivity(ctx context.Context, orderID string) (*types.Payment, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("获取支付", "orderID", orderID)

	return &types.Payment{
		PaymentID:   "PAY-001",
		OrderID:    orderID,
		Amount:      100.00,
		Status:      "success",
		TransactionID: "TXN-001",
	}, nil
}

// RefundPaymentActivity 退款活动
func RefundPaymentActivity(ctx context.Context, paymentID string) error {
	logger := activity.GetLogger(ctx)
	logger.Info("退款", "paymentID", paymentID)

	time.Sleep(150 * time.Millisecond)

	logger.Info("退款成功")
	return nil
}

// ExecuteRefundActivity 执行退款活动
func ExecuteRefundActivity(ctx context.Context, req types.RefundRequest) (*types.RefundResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("执行退款", "paymentID", req.PaymentID, "amount", req.Amount)

	time.Sleep(150 * time.Millisecond)

	return &types.RefundResult{
		Success:    true,
		RefundID:   fmt.Sprintf("REF-%d", time.Now().UnixNano()),
		Amount:     req.Amount,
		Message:    "退款成功",
	}, nil
}
```

### 6.3.3 库存活动

```go
// internal/activity/stock.go
package activity

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.temporal.io/sdk/activity"

	"ecommerce-workflow/internal/types"
)

// CheckStockActivity 检查库存活动
func CheckStockActivity(ctx context.Context, req types.StockRequest) (*types.StockResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("检查库存", "items", len(req.Items))

	// 模拟检查
	time.Sleep(100 * time.Millisecond)

	logger.Info("库存检查完成")
	return &types.StockResult{
		Success: true,
		Message: "库存充足",
	}, nil
}

// ReserveStockActivity 预留库存活动
func ReserveStockActivity(ctx context.Context, req types.StockRequest) error {
	logger := activity.GetLogger(ctx)
	logger.Info("预留库存", "orderID", req.OrderID, "items", len(req.Items))

	// 模拟预留
	time.Sleep(150 * time.Millisecond)

	logger.Info("库存预留成功")
	return nil
}

// ReleaseStockActivity 释放库存活动
func ReleaseStockActivity(ctx context.Context, orderID string) error {
	logger := activity.GetLogger(ctx)
	logger.Info("释放库存", "orderID", orderID)

	// 模拟释放
	time.Sleep(100 * time.Millisecond)

	logger.Info("库存释放成功")
	return nil
}
```

### 6.3.4 通知活动

```go
// internal/activity/notification.go
package activity

import (
	"context"
	"fmt"
	"log"

	"go.temporal.io/sdk/activity"

	"ecommerce-workflow/internal/types"
)

// SendNotificationActivity 发送通知活动
func SendNotificationActivity(ctx context.Context, req types.NotificationRequest) error {
	logger := activity.GetLogger(ctx)
	logger.Info("发送通知", "customerID", req.CustomerID, "type", req.Type)

	// 模拟发送通知
	fmt.Printf("通知: [%s] %s -> %s\n", req.Type, req.Title, req.Content)

	logger.Info("通知发送成功")
	return nil
}
```

## 6.4 类型定义

```go
// internal/types/types.go
package types

import "time"

// Order 订单
type Order struct {
	ID           string      `json:"id"`
	CustomerID   string      `json:"customer_id"`
	Items        []OrderItem `json:"items"`
	TotalAmount  float64     `json:"total_amount"`
	Status       string      `json:"status"`
	PaymentID    string      `json:"payment_id,omitempty"`
	Remark       string      `json:"remark,omitempty"`
	CreatedAt    time.Time   `json:"created_at"`
	UpdatedAt    time.Time   `json:"updated_at"`
}

// OrderItem 订单商品
type OrderItem struct {
	ProductID   string  `json:"product_id"`
	ProductName string  `json:"product_name"`
	Quantity    int     `json:"quantity"`
	Price       float64 `json:"price"`
	Subtotal    float64 `json:"subtotal"`
}

// OrderResult 订单结果
type OrderResult struct {
	OrderID string `json:"order_id"`
	Status  string `json:"status"`
}

// UpdateOrderRequest 更新订单请求
type UpdateOrderRequest struct {
	OrderID   string `json:"order_id"`
	Status    string `json:"status"`
	PaymentID string `json:"payment_id,omitempty"`
	Remark    string `json:"remark,omitempty"`
}

// Payment 支付
type Payment struct {
	PaymentID    string    `json:"payment_id"`
	OrderID      string    `json:"order_id"`
	Amount       float64   `json:"amount"`
	Status       string    `json:"status"`
	TransactionID string   `json:"transaction_id"`
	CreatedAt    time.Time `json:"created_at"`
}

// PaymentRequest 支付请求
type PaymentRequest struct {
	PaymentID   string  `json:"payment_id,omitempty"`
	OrderID     string  `json:"order_id"`
	Amount      float64 `json:"amount"`
	CustomerID  string  `json:"customer_id"`
}

// PaymentResult 支付结果
type PaymentResult struct {
	Success       bool   `json:"success"`
	PaymentID     string `json:"payment_id,omitempty"`
	TransactionID string `json:"transaction_id,omitempty"`
	Message       string `json:"message"`
}

// RefundRequest 退款请求
type RefundRequest struct {
	OrderID   string  `json:"order_id"`
	PaymentID string  `json:"payment_id"`
	Amount    float64 `json:"amount"`
	Reason    string  `json:"reason"`
}

// RefundResult 退款结果
type RefundResult struct {
	Success  bool    `json:"success"`
	RefundID string  `json:"refund_id"`
	Amount   float64 `json:"amount"`
	Message  string  `json:"message"`
}

// StockRequest 库存请求
type StockRequest struct {
	OrderID string      `json:"order_id,omitempty"`
	Items    []OrderItem `json:"items"`
}

// StockResult 库存结果
type StockResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// NotificationRequest 通知请求
type NotificationRequest struct {
	CustomerID string `json:"customer_id"`
	Type       string `json:"type"`
	Title      string `json:"title"`
	Content    string `json:"content"`
}
```

## 6.5 服务启动

### 6.5.1 订单服务

```go
// cmd/order-service/main.go
package main

import (
	"log"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	"ecommerce-workflow/internal/activity"
	"ecommerce-workflow/internal/workflow"
)

func main() {
	c, err := client.Dial(client.Options{HostPort: "localhost:7233"})
	if err != nil {
		log.Fatalln("无法创建客户端", err)
	}
	defer c.Close()

	w := worker.New(c, "order-task-queue", worker.Options{})

	// 注册工作流
	w.RegisterWorkflow(workflow.CreateOrderWorkflow)
	w.RegisterWorkflow(workflow.CancelOrderWorkflow)

	// 注册活动
	w.RegisterActivity(activity.CreateOrderActivity)
	w.RegisterActivity(activity.GetOrderActivity)
	w.RegisterActivity(activity.UpdateOrderStatusActivity)
	w.RegisterActivity(activity.CheckStockActivity)
	w.RegisterActivity(activity.ReserveStockActivity)
	w.RegisterActivity(activity.ReleaseStockActivity)
	w.RegisterActivity(activity.CreatePaymentActivity)
	w.RegisterActivity(activity.ProcessPaymentActivity)
	w.RegisterActivity(activity.SendNotificationActivity)

	log.Println("启动订单服务...")
	w.Start()
	log.Println("订单服务已启动")
	select {}
}
```

### 6.5.2 支付服务

```go
// cmd/payment-service/main.go
package main

import (
	"log"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	"ecommerce-workflow/internal/activity"
	"ecommerce-workflow/internal/workflow"
)

func main() {
	c, err := client.Dial(client.Options{HostPort: "localhost:7233"})
	if err != nil {
		log.Fatalln("无法创建客户端", err)
	}
	defer c.Close()

	w := worker.New(c, "payment-task-queue", worker.Options{})

	w.RegisterWorkflow(workflow.PaymentWorkflow)
	w.RegisterWorkflow(workflow.RefundWorkflow)

	w.RegisterActivity(activity.VerifyPaymentActivity)
	w.RegisterActivity(activity.ProcessPaymentActivity)
	w.RegisterActivity(activity.GetPaymentActivity)
	w.RegisterActivity(activity.RefundPaymentActivity)
	w.RegisterActivity(activity.ExecuteRefundActivity)
	w.RegisterActivity(activity.UpdateOrderStatusActivity)
	w.RegisterActivity(activity.SendNotificationActivity)

	log.Println("启动支付服务...")
	w.Start()
	log.Println("支付服务已启动")
	select {}
}
```

### 6.5.3 库存服务

```go
// cmd/stock-service/main.go
package main

import (
	"log"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	"ecommerce-workflow/internal/activity"
)

func main() {
	c, err := client.Dial(client.Options{HostPort: "localhost:7233"})
	if err != nil {
		log.Fatalln("无法创建客户端", err)
	}
	defer c.Close()

	w := worker.New(c, "stock-task-queue", worker.Options{})

	w.RegisterActivity(activity.CheckStockActivity)
	w.RegisterActivity(activity.ReserveStockActivity)
	w.RegisterActivity(activity.ReleaseStockActivity)

	log.Println("启动库存服务...")
	w.Start()
	log.Println("库存服务已启动")
	select {}
}
```

### 6.5.4 通知服务

```go
// cmd/notification-service/main.go
package main

import (
	"log"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	"ecommerce-workflow/internal/activity"
)

func main() {
	c, err := client.Dial(client.Options{HostPort: "localhost:7233"})
	if err != nil {
		log.Fatalln("无法创建客户端", err)
	}
	defer c.Close()

	w := worker.New(c, "notification-task-queue", worker.Options{})

	w.RegisterActivity(activity.SendNotificationActivity)

	log.Println("启动通知服务...")
	w.Start()
	log.Println("通知服务已启动")
	select {}
}
```

## 6.6 Docker 部署

```yaml
# deployments/docker/docker-compose.yml
version: '3.8'

services:
  temporal:
    image: temporalio/auto-setup:1.22.0
    ports:
      - "7233:7233"
    environment:
      - DB=postgresql
      - POSTGRES_USER=temporal
      - POSTGRES_PWD=temporal

  postgres:
    image: postgres:13
    environment:
      POSTGRES_USER: temporal
      POSTGRES_PASSWORD: temporal

  order-service:
    build: ./order-service
    ports:
      - "8091:8091"

  payment-service:
    build: ./payment-service
    ports:
      - "8092:8092"

  stock-service:
    build: ./stock-service
    ports:
      - "8093:8093"

  notification-service:
    build: ./notification-service
    ports:
      - "8094:8094"
```

---

## 6.7 扩展练习

1. **添加 API 网关**：实现统一的 HTTP 入口
2. **添加服务发现**：使用 Consul 或 etcd
3. **添加熔断器**：使用 Hystrix 模式
4. **添加限流**：实现令牌桶限流
5. **添加监控告警**：集成 Prometheus 和 Alertmanager
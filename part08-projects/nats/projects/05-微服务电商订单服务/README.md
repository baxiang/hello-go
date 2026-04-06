# 高级项目：微服务架构的电商订单服务

本项目实现一个基于 NATS 的微服务架构电商订单系统，包含订单服务、支付服务、库存服务、通知服务，使用 JetStream 实现服务间通信和事件驱动。

## 5.1 项目概述

### 系统架构

```
┌─────────────────────────────────────────────────────────────────┐
│                         API Gateway                              │
│                         (orders-api)                              │
└─────────────────────────────┬───────────────────────────────────┘
                              │
        ┌─────────────────────┼─────────────────────┐
        │                     │                     │
        ▼                     ▼                     ▼
┌───────────────┐   ┌───────────────┐   ┌───────────────┐
│  Order Service │   │Payment Service│   │ Stock Service │
│   (orders)    │   │  (payments)   │   │   (stocks)    │
└───────┬───────┘   └───────┬───────┘   └───────┬───────┘
        │                   │                     │
        └───────────────────┼─────────────────────┘
                            │
                            ▼
                    ┌───────────────┐
                    │ Notification   │
                    │   Service      │
                    └───────────────┘
```

### 服务列表

| 服务 | 端口 | 说明 |
|------|------|------|
| orders-api | 8080 | 订单 API 网关 |
| orders | 8081 | 订单服务 |
| payments | 8082 | 支付服务 |
| stocks | 8083 | 库存服务 |
| notifications | 8084 | 通知服务 |

### 消息流

1. 客户端 -> API Gateway -> Order Service (创建订单)
2. Order Service -> Stock Service (扣减库存)
3. Stock Service -> Payment Service (发起支付)
4. Payment Service -> Order Service (支付结果)
5. Order Service -> Notification Service (发送通知)

### 项目结构

```
ecommerce/
├── api/
│   └── proto/
│       └── order.proto
├── cmd/
│   ├── api-gateway/main.go
│   ├── orders/main.go
│   ├── payments/main.go
│   ├── stocks/main.go
│   └── notifications/main.go
├── internal/
│   ├── proto/
│   │   └── order.pb.go
│   ├── common/
│   │   ├── config.go
│   │   └── nats.go
│   ├── orders/
│   │   ├── service.go
│   │   ├── handler.go
│   │   └── repository.go
│   ├── payments/
│   │   └── service.go
│   ├── stocks/
│   │   └── service.go
│   └── notifications/
│       └── service.go
├── config.yaml
└── go.mod
```

## 5.2 公共组件

### 配置管理

```go
// internal/common/config.go
package common

import (
    "fmt"
    "time"

    "github.com/spf13/viper"
)

// Config 全局配置
type Config struct {
    NATS  NATSConfig  `mapstructure:"nats"`
    Order OrderConfig `mapstructure:"order"`
    HTTP  HTTPConfig  `mapstructure:"http"`
}

type NATSConfig struct {
    URL        string        `mapstructure:"url"`
    ClusterID  string        `mapstructure:"cluster_id"`
    ClientID   string        `mapstructure:"client_id"`
    Timeout    time.Duration `mapstructure:"timeout"`
}

type OrderConfig struct {
    StreamName    string `mapstructure:"stream_name"`
    Subjects      string `mapstructure:"subjects"`
    ConsumerGroup string `mapstructure:"consumer_group"`
}

type HTTPConfig struct {
    Host string `mapstructure:"host"`
    Port int    `mapstructure:"port"`
}

// Load 加载配置
func Load(path string) (*Config, error) {
    viper.SetConfigFile(path)
    viper.SetConfigType("yaml")

    if err := viper.ReadInConfig(); err != nil {
        return nil, fmt.Errorf("读取配置失败: %w", err)
    }

    var config Config
    if err := viper.Unmarshal(&config); err != nil {
        return nil, fmt.Errorf("解析配置失败: %w", err)
    }

    return &config, nil
}
```

### NATS 客户端

```go
// internal/common/nats.go
package common

import (
    "fmt"
    "sync"

    "github.com/nats-io/nats.go"
    "github.com/nats-io/nats.go/jetstream"
)

// NATSClient NATS 客户端封装
type NATSClient struct {
    nc *nats.Conn
    js jetstream.JetStream
    mu sync.RWMutex
}

// NewNATSClient 创建 NATS 客户端
func NewNATSClient(url string) (*NATSClient, error) {
    nc, err := nats.Connect(url,
        nats.Name("ecommerce-service"),
        nats.MaxReconnects(5),
        nats.ReconnectWait(time.Second),
    )
    if err != nil {
        return nil, fmt.Errorf("连接 NATS 失败: %w", err)
    }

    js, err := jetstream.New(nc)
    if err != nil {
        return nil, fmt.Errorf("创建 JetStream 失败: %w", err)
    }

    return &NATSClient{nc: nc, js: js}, nil
}

// GetConn 获取连接
func (c *NATSClient) GetConn() *nats.Conn {
    c.mu.RLock()
    defer c.mu.RUnlock()
    return c.nc
}

// GetJS 获取 JetStream
func (c *NATSClient) GetJS() jetstream.JetStream {
    c.mu.RLock()
    defer c.mu.RUnlock()
    return c.js
}

// Close 关闭连接
func (c *NATSClient) Close() {
    c.mu.Lock()
    defer c.mu.Unlock()
    if c.nc != nil {
        c.nc.Close()
    }
}

// CreateStream 创建 Stream
func (c *NATSClient) CreateStream(ctx context.Context, name string, subjects []string) error {
    _, err := c.js.Stream(ctx, name)
    if err == nil {
        return nil
    }

    _, err = c.js.CreateStream(ctx, jetstream.StreamConfig{
        Name:      name,
        Subjects:  subjects,
        Storage:   jetstream.FileStorage,
        Retention: jetstream.LimitsPolicy{
            MaxBytes: 1024 * 1024 * 1024,
            MaxAge:   time.Hour * 24 * 7,
        },
    })
    return err
}
```

## 5.3 订单服务

### 订单模型

```go
// internal/orders/model.go
package orders

import (
    "encoding/json"
    "time"
)

type OrderStatus string

const (
    OrderStatusPending   OrderStatus = "pending"
    OrderStatusPaid       OrderStatus = "paid"
    OrderStatusProcessing OrderStatus = "processing"
    OrderStatusShipped    OrderStatus = "shipped"
    OrderStatusCompleted  OrderStatus = "completed"
    OrderStatusCancelled  OrderStatus = "cancelled"
    OrderStatusFailed     OrderStatus = "failed"
)

type Order struct {
    ID          string                 `json:"id"`
    CustomerID  string                 `json:"customer_id"`
    Items       []OrderItem            `json:"items"`
    TotalAmount float64                `json:"total_amount"`
    Status      OrderStatus            `json:"status"`
    PaymentID   string                 `json:"payment_id,omitempty"`
    Metadata    map[string]interface{} `json:"metadata,omitempty"`
    CreatedAt   time.Time              `json:"created_at"`
    UpdatedAt   time.Time              `json:"updated_at"`
}

type OrderItem struct {
    ProductID   string  `json:"product_id"`
    ProductName string  `json:"product_name"`
    Quantity    int     `json:"quantity"`
    UnitPrice   float64 `json:"unit_price"`
    Subtotal    float64 `json:"subtotal"`
}

func (o *Order) ToJSON() ([]byte, error) {
    return json.Marshal(o)
}

func OrderFromJSON(data []byte) (*Order, error) {
    var order Order
    err := json.Unmarshal(data, &order)
    return &order, err
}
```

### 订单服务

```go
// internal/orders/service.go
package orders

import (
    "context"
    "fmt"
    "log"
    "sync"
    "time"

    "github.com/nats-io/nats.go"
    "github.com/nats-io/nats.go/jetstream"

    "ecommerce/internal/common"
)

type Service struct {
    nc  *nats.Conn
    js  jetstream.JetStream
    mu  sync.RWMutex
    orders map[string]*Order
}

func NewService(nc *nats.Conn, js jetstream.JetStream) *Service {
    return &Service{
        nc:     nc,
        js:     js,
        orders: make(map[string]*Order),
    }
}

// CreateOrder 创建订单
func (s *Service) CreateOrder(ctx context.Context, customerID string, items []OrderItem) (*Order, error) {
    // 计算总价
    var total float64
    for _, item := range items {
        item.Subtotal = float64(item.Quantity) * item.UnitPrice
        total += item.Subtotal
    }

    order := &Order{
        ID:          generateOrderID(),
        CustomerID:  customerID,
        Items:       items,
        TotalAmount: total,
        Status:      OrderStatusPending,
        CreatedAt:   time.Now(),
        UpdatedAt:   time.Now(),
    }

    // 保存订单
    s.mu.Lock()
    s.orders[order.ID] = order
    s.mu.Unlock()

    // 发布订单创建事件
    event := OrderEvent{
        Type:      "order.created",
        Order:     order,
        Timestamp: time.Now(),
    }
    s.publishEvent("orders.events", event)

    // 发送库存扣减请求
    stockReq := StockRequest{
        OrderID: order.ID,
        Items:   items,
    }
    s.publishRequest("stocks.deduct", stockReq)

    log.Printf("创建订单: %s, 金额: %.2f", order.ID, order.TotalAmount)
    return order, nil
}

// UpdateOrderStatus 更新订单状态
func (s *Service) UpdateOrderStatus(orderID string, status OrderStatus) error {
    s.mu.Lock()
    defer s.mu.Unlock()

    order, ok := s.orders[orderID]
    if !ok {
        return fmt.Errorf("订单不存在: %s", orderID)
    }

    order.Status = status
    order.UpdatedAt = time.Now()

    // 发布订单状态变更事件
    event := OrderEvent{
        Type:      "order.status_changed",
        Order:     order,
        Timestamp: time.Now(),
    }
    s.publishEvent("orders.events", event)

    log.Printf("更新订单状态: %s -> %s", orderID, status)
    return nil
}

// GetOrder 获取订单
func (s *Service) GetOrder(orderID string) (*Order, error) {
    s.mu.RLock()
    defer s.mu.RUnlock()

    order, ok := s.orders[orderID]
    if !ok {
        return nil, fmt.Errorf("订单不存在: %s", orderID)
    }
    return order, nil
}

// ListOrders 列出订单
func (s *Service) ListOrders(customerID string) []*Order {
    s.mu.RLock()
    defer s.mu.RUnlock()

    var result []*Order
    for _, order := range s.orders {
        if customerID == "" || order.CustomerID == customerID {
            result = append(result, order)
        }
    }
    return result
}

func (s *Service) publishEvent(subject string, event OrderEvent) error {
    data, _ := json.Marshal(event)
    _, err := s.js.Publish(context.Background(), subject, data)
    return err
}

func (s *Service) publishRequest(subject string, req interface{}) error {
    data, _ := json.Marshal(req)
    _, err := s.js.Publish(context.Background(), subject, data)
    return err
}

// Subscribe 订阅消息
func (s *Service) Subscribe() error {
    // 订阅支付结果
    _, err := s.js.Consume("ORDERS", "order-service",
        jetstream.ConsumeErrHandler(func(ctx jetstream.ConsumeContext, err error) {
            log.Printf("消费错误: %v", err)
        }),
    )
    if err != nil {
        return err
    }

    // 订阅库存结果
    sub, err := s.nc.Subscribe("stocks.result", func(m *nats.Msg) {
        var result StockResponse
        if err := json.Unmarshal(m.Data, &result); err != nil {
            log.Printf("解析库存响应失败: %v", err)
            return
        }

        if result.Success {
            s.UpdateOrderStatus(result.OrderID, OrderStatusPaid)
        } else {
            s.UpdateOrderStatus(result.OrderID, OrderStatusFailed)
        }
    })
    if err != nil {
        return err
    }

    log.Println("订单服务已订阅消息")
    return nil
}

// OrderEvent 订单事件
type OrderEvent struct {
    Type      string    `json:"type"`
    Order     *Order    `json:"order"`
    Timestamp time.Time `json:"timestamp"`
}

// StockRequest 库存请求
type StockRequest struct {
    OrderID string      `json:"order_id"`
    Items   []OrderItem `json:"items"`
}

// StockResponse 库存响应
type StockResponse struct {
    OrderID string `json:"order_id"`
    Success bool   `json:"success"`
    Message string `json:"message"`
}

func generateOrderID() string {
    return time.Now().Format("20060102150405") + "-" + randomString(8)
}

func randomString(n int) string {
    const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
    b := make([]byte, n)
    for i := range b {
        b[i] = letters[time.Now().UnixNano()%int64(len(letters))]
    }
    return string(b)
}
```

### HTTP 处理器

```go
// internal/orders/handler.go
package orders

import (
    "encoding/json"
    "log"
    "net/http"

    "github.com/gin-gonic/gin"
)

type Handler struct {
    service *Service
}

func NewHandler(service *Service) *Handler {
    return &Handler{service: service}
}

// CreateOrderRequest 创建订单请求
type CreateOrderRequest struct {
    CustomerID string      `json:"customer_id" binding:"required"`
    Items      []OrderItem `json:"items" binding:"required,min=1"`
}

// RegisterRoutes 注册路由
func (h *Handler) RegisterRoutes(r *gin.Engine) {
    orders := r.Group("/api/orders")
    {
        orders.POST("", h.CreateOrder)
        orders.GET("/:id", h.GetOrder)
        orders.GET("", h.ListOrders)
    }
}

// CreateOrder 创建订单
func (h *Handler) CreateOrder(c *gin.Context) {
    var req CreateOrderRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    order, err := h.service.CreateOrder(c.Request.Context(), req.CustomerID, req.Items)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusCreated, order)
}

// GetOrder 获取订单
func (h *Handler) GetOrder(c *gin.Context) {
    id := c.Param("id")
    order, err := h.service.GetOrder(id)
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, order)
}

// ListOrders 列出订单
func (h *Handler) ListOrders(c *gin.Context) {
    customerID := c.Query("customer_id")
    orders := h.service.ListOrders(customerID)
    c.JSON(http.StatusOK, orders)
}
```

### 主程序

```go
// cmd/orders/main.go
package main

import (
    "context"
    "log"
    "time"

    "ecommerce/internal/common"
    "ecommerce/internal/orders"

    "github.com/gin-gonic/gin"
)

func main() {
    // 加载配置
    config, err := common.Load("config.yaml")
    if err != nil {
        log.Fatal(err)
    }

    // 创建 NATS 客户端
    natsClient, err := common.NewNATSClient(config.NATS.URL)
    if err != nil {
        log.Fatal(err)
    }
    defer natsClient.Close()

    // 创建 Stream
    ctx := context.Background()
    natsClient.CreateStream(ctx, "ORDERS", []string{"orders.>", "stocks.>", "payments.>", "notifications.>"})

    // 创建订单服务
    orderService := orders.NewService(natsClient.GetConn(), natsClient.GetJS())
    orderService.Subscribe()

    // 创建 HTTP 处理器
    handler := orders.NewHandler(orderService)

    // 启动 HTTP 服务器
    r := gin.Default()
    handler.RegisterRoutes(r)

    addr := config.HTTP.Host + ":" + string(rune(config.HTTP.Port))
    log.Printf("订单服务启动: %s", addr)
    r.Run(addr)
}
```

## 5.4 库存服务

```go
// internal/stocks/service.go
package stocks

import (
    "context"
    "encoding/json"
    "log"
    "sync"
    "time"

    "github.com/nats-io/nats.go"
    "github.com/nats-io/nats.go/jetstream"

    "ecommerce/internal/orders"
)

type Service struct {
    nc  *nats.Conn
    js  jetstream.JetStream
    mu  sync.RWMutex
    stocks map[string]int  // productID -> quantity
}

func NewService(nc *nats.Conn, js jetstream.JetStream) *Service {
    svc := &Service{
        nc:     nc,
        js:     js,
        stocks: make(map[string]int),
    }
    
    // 初始化库存
    svc.stocks["P001"] = 100
    svc.stocks["P002"] = 50
    svc.stocks["P003"] = 200
    svc.stocks["P004"] = 500
    
    return svc
}

// Deduct 扣减库存
func (s *Service) Deduct(orderID string, items []orders.OrderItem) error {
    s.mu.Lock()
    defer s.mu.Unlock()

    // 检查库存
    for _, item := range items {
        current := s.stocks[item.ProductID]
        if current < item.Quantity {
            // 库存不足，发布失败消息
            s.publishResult(orderID, false, "库存不足")
            return nil
        }
    }

    // 扣减库存
    for _, item := range items {
        s.stocks[item.ProductID] -= item.Quantity
    }

    // 发布成功消息
    s.publishResult(orderID, true, "库存扣减成功")
    log.Printf("订单 %s 库存扣减成功", orderID)
    return nil
}

func (s *Service) publishResult(orderID string, success bool, message string) {
    result := orders.StockResponse{
        OrderID: orderID,
        Success: success,
        Message: message,
    }
    data, _ := json.Marshal(result)
    s.js.Publish(context.Background(), "stocks.result", data)
}

// Subscribe 订阅消息
func (s *Service) Subscribe() error {
    _, err := s.nc.Subscribe("stocks.deduct", func(m *nats.Msg) {
        var req orders.StockRequest
        if err := json.Unmarshal(m.Data, &req); err != nil {
            log.Printf("解析库存请求失败: %v", err)
            return
        }

        log.Printf("收到库存扣减请求: %s", req.OrderID)
        s.Deduct(req.OrderID, req.Items)
    })
    return err
}
```

## 5.5 支付服务

```go
// internal/payments/service.go
package payments

import (
    "context"
    "encoding/json"
    "log"
    "sync"
    "time"

    "github.com/nats-io/nats.go"
    "github.com/nats-io/nats.go/jetstream"
)

type PaymentStatus string

const (
    PaymentStatusPending   PaymentStatus = "pending"
    PaymentStatusSuccess  PaymentStatus = "success"
    PaymentStatusFailed   PaymentStatus = "failed"
)

type Payment struct {
    ID        string         `json:"id"`
    OrderID   string         `json:"order_id"`
    Amount    float64        `json:"amount"`
    Status    PaymentStatus  `json:"status"`
    CreatedAt time.Time      `json:"created_at"`
}

type Service struct {
    nc  *nats.Conn
    js  jetstream.JetStream
    mu  sync.RWMutex
    payments map[string]*Payment
}

func NewService(nc *nats.Conn, js jetstream.JetStream) *Service {
    return &Service{
        nc:       nc,
        js:       js,
        payments: make(map[string]*Payment),
    }
}

// ProcessPayment 处理支付
func (s *Service) ProcessPayment(orderID string, amount float64) error {
    s.mu.Lock()
    defer s.mu.Unlock()

    payment := &Payment{
        ID:        generatePaymentID(),
        OrderID:   orderID,
        Amount:    amount,
        Status:    PaymentStatusSuccess,  // 简化处理，总是成功
        CreatedAt: time.Now(),
    }

    s.payments[payment.ID] = payment

    // 发布支付结果
    event := PaymentEvent{
        Type:      "payment.completed",
        Payment:   payment,
        Timestamp: time.Now(),
    }
    data, _ := json.Marshal(event)
    s.js.Publish(context.Background(), "payments.events", data)

    log.Printf("支付处理完成: %s, 订单: %s, 金额: %.2f", payment.ID, orderID, amount)
    return nil
}

// Subscribe 订阅消息
func (s *Service) Subscribe() error {
    _, err := s.nc.Subscribe("stocks.result", func(m *nats.Msg) {
        var result struct {
            OrderID string `json:"order_id"`
            Success bool   `json:"success"`
        }
        if err := json.Unmarshal(m.Data, &result); err != nil {
            log.Printf("解析库存结果失败: %v", err)
            return
        }

        if result.Success {
            // 库存扣减成功，发起支付
            // 实际应该从订单服务获取金额
            s.ProcessPayment(result.OrderID, 1000.0)
        }
    })
    return err
}

type PaymentEvent struct {
    Type      string    `json:"type"`
    Payment   *Payment  `json:"payment"`
    Timestamp time.Time `json:"timestamp"`
}

func generatePaymentID() string {
    return "PAY" + time.Now().Format("20060102150405") + "-" + randomString(6)
}

func randomString(n int) string {
    const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
    b := make([]byte, n)
    for i := range b {
        b[i] = letters[time.Now().UnixNano()%int64(len(letters))]
    }
    return string(b)
}
```

## 5.6 通知服务

```go
// internal/notifications/service.go
package notifications

import (
    "context"
    "encoding/json"
    "log"

    "github.com/nats-io/nats.go"
    "github.com/nats-io/nats.go/jetstream"
)

type NotificationType string

const (
    NotificationTypeEmail    NotificationType = "email"
    NotificationTypeSMS      NotificationType = "sms"
    NotificationTypePush      NotificationType = "push"
)

type Notification struct {
    ID         string           `json:"id"`
    Type       NotificationType `json:"type"`
    Recipient  string           `json:"recipient"`
    Subject    string           `json:"subject"`
    Content    string           `json:"content"`
    OrderID    string           `json:"order_id"`
}

type Service struct {
    nc *nats.Conn
    js jetstream.JetStream
}

func NewService(nc *nats.Conn, js jetstream.JetStream) *Service {
    return &Service{nc: nc, js: js}
}

// Send 发送通知
func (s *Service) Send(notification *Notification) error {
    // 实际应该调用邮件/短信服务
    log.Printf("发送通知: 类型=%s, 收件人=%s, 主题=%s",
        notification.Type, notification.Recipient, notification.Subject)
    return nil
}

// Subscribe 订阅消息
func (s *Service) Subscribe() error {
    // 订阅订单事件
    _, err := s.nc.Subscribe("orders.events", func(m *nats.Msg) {
        var event struct {
            Type string `json:"type"`
            Order struct {
                ID        string `json:"id"`
                CustomerID string `json:"customer_id"`
                Status    string `json:"status"`
            } `json:"order"`
        }
        if err := json.Unmarshal(m.Data, &event); err != nil {
            log.Printf("解析订单事件失败: %v", err)
            return
        }

        // 根据事件类型发送通知
        switch event.Type {
        case "order.created":
            s.Send(&Notification{
                Type:      NotificationTypeEmail,
                Recipient: event.Order.CustomerID + "@example.com",
                Subject:   "订单创建成功",
                Content:   "您的订单 " + event.Order.ID + " 已创建",
                OrderID:   event.Order.ID,
            })
        case "order.status_changed":
            s.Send(&Notification{
                Type:      NotificationTypeEmail,
                Recipient: event.Order.CustomerID + "@example.com",
                Subject:   "订单状态更新",
                Content:   "您的订单 " + event.Order.ID + " 状态已更新为 " + event.Order.Status,
                OrderID:   event.Order.ID,
            })
        }
    })
    return err
}
```

## 5.7 API Gateway

```go
// cmd/api-gateway/main.go
package main

import (
    "log"
    "net/http"
    "time"

    "github.com/gin-gonic/gin"
)

func main() {
    r := gin.Default()

    // 代理到订单服务
    r.POST("/api/orders", func(c *gin.Context) {
        // 转发到订单服务
        c.JSON(http.StatusOK, gin.H{
            "message": "订单服务响应",
            "time":    time.Now(),
        })
    })

    r.GET("/api/orders/:id", func(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{
            "id":   c.Param("id"),
            "time": time.Now(),
        })
    })

    // 健康检查
    r.GET("/health", func(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{"status": "ok"})
    })

    log.Println("API Gateway 启动: 8080")
    r.Run(":8080")
}
```

## 5.8 配置文件

```yaml
# config.yaml
nats:
  url: "nats://localhost:4222"
  cluster_id: "ecommerce"
  timeout: 10s

order:
  stream_name: "ORDERS"
  subjects:
    - "orders.>"
    - "stocks.>"
    - "payments.>"
    - "notifications.>"
  consumer_group: "order-service"

http:
  host: "0.0.0.0"
  port: 8081
```

## 5.9 运行项目

### 启动 NATS 服务器

```bash
docker run -d --name nats-server -p 4222:4222 -p 8222:8222 nats:latest -js
```

### 启动各个服务

```bash
# 终端 1: 订单服务
cd cmd/orders
go run main.go

# 终端 2: 库存服务
cd cmd/stocks
go run main.go

# 终端 3: 支付服务
cd cmd/payments
go run main.go

# 终端 4: 通知服务
cd cmd/notifications
go run main.go

# 终端 5: API Gateway
cd cmd/api-gateway
go run main.go
```

### 测试流程

```bash
# 创建订单
curl -X POST http://localhost:8080/api/orders \
  -H "Content-Type: application/json" \
  -d '{
    "customer_id": "C001",
    "items": [
      {"product_id": "P001", "product_name": "iPhone 15", "quantity": 1, "unit_price": 999.0}
    ]
  }'

# 查看订单
curl http://localhost:8080/api/orders/{order_id}
```

---

## 5.10 扩展练习

1. **添加 gRPC**：使用 gRPC 替代 HTTP 进行服务间通信
2. **添加服务注册**：使用 NATS Service Gateway 实现服务发现
3. **添加分布式事务**：使用 Saga 模式实现跨服务事务
4. **添加监控**：集成 Prometheus 监控各服务
5. **添加链路追踪**：使用 OpenTelemetry 实现分布式追踪
6. **添加认证授权**：实现 JWT 认证
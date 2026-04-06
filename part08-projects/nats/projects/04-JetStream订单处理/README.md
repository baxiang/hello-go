# 进阶项目：基于 JetStream 的订单处理系统

本项目实现一个基于 NATS JetStream 的持久化订单处理系统，支持消息持久化、消息确认、消费者组和消息重试。

## 4.1 项目概述

### 功能需求

1. 使用 JetStream 持久化订单消息
2. 实现消息的可靠投递和确认
3. 实现消费者组负载均衡
4. 实现消息重试机制
5. 实现消息去重
6. 实现订单状态管理

### 项目结构

```
order-processing/
├── cmd/
│   ├── producer/main.go       # 订单生产者
│   ├── consumer/main.go       # 订单消费者
│   └── admin/main.go          # 管理工具
├── internal/
│   ├── order/
│   │   ├── order.go           # 订单模型
│   │   └── repository.go     # 订单仓库
│   ├── stream/
│   │   └── manager.go         # Stream 管理
│   ├── consumer/
│   │   └── processor.go       # 消费者处理
│   └── metrics/
│       └── metrics.go          # 指标监控
├── config.yaml
└── go.mod
```

## 4.2 订单模型

```go
// internal/order/order.go
package order

import (
    "encoding/json"
    "errors"
    "time"
)

type OrderStatus string

const (
	StatusPending   OrderStatus = "pending"
	StatusPaid      OrderStatus = "paid"
	StatusProcessing OrderStatus = "processing"
	StatusCompleted OrderStatus = "completed"
	StatusCancelled OrderStatus = "cancelled"
	StatusFailed    OrderStatus = "failed"
)

// Order 订单
type Order struct {
    ID          string      `json:"id"`
    CustomerID  string      `json:"customer_id"`
    ProductID   string      `json:"product_id"`
    ProductName string      `json:"product_name"`
    Quantity    int         `json:"quantity"`
    Amount      float64     `json:"amount"`
    Status      OrderStatus `json:"status"`
    CreatedAt   time.Time   `json:"created_at"`
    UpdatedAt   time.Time   `json:"updated_at"`
    PaidAt      *time.Time  `json:"paid_at,omitempty"`
}

// NewOrder 创建新订单
func NewOrder(customerID, productID, productName string, quantity int, amount float64) *Order {
    now := time.Now()
    return &Order{
        ID:          generateOrderID(),
        CustomerID:  customerID,
        ProductID:   productID,
        ProductName: productName,
        Quantity:    quantity,
        Amount:      amount,
        Status:      StatusPending,
        CreatedAt:   now,
        UpdatedAt:   now,
    }
}

// Validate 验证订单
func (o *Order) Validate() error {
    if o.CustomerID == "" {
        return errors.New("客户ID不能为空")
    }
    if o.ProductID == "" {
        return errors.New("商品ID不能为空")
    }
    if o.Quantity <= 0 {
        return errors.New("数量必须大于0")
    }
    if o.Amount <= 0 {
        return errors.New("金额必须大于0")
    }
    return nil
}

// ToJSON 序列化为 JSON
func (o *Order) ToJSON() ([]byte, error) {
    return json.Marshal(o)
}

// FromJSON 从 JSON 反序列化
func FromJSON(data []byte) (*Order, error) {
    var order Order
    err := json.Unmarshal(data, &order)
    return &order, err
}

// UpdateStatus 更新订单状态
func (o *Order) UpdateStatus(status OrderStatus) {
    o.Status = status
    o.UpdatedAt = time.Now()
    if status == StatusPaid {
        now := time.Now()
        o.PaidAt = &now
    }
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

## 4.3 Stream 管理

```go
// internal/stream/manager.go
package stream

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/nats-io/nats.go"
    "github.com/nats-io/nats.go/jetstream"
)

// Manager JetStream Stream 管理器
type Manager struct {
    js jetstream.JetStream
    nc *nats.Conn
}

// New 创建 Stream 管理器
func New(url string) (*Manager, error) {
    nc, err := nats.Connect(url)
    if err != nil {
        return nil, fmt.Errorf("连接 NATS 失败: %w", err)
    }

    js, err := jetstream.New(nc)
    if err != nil {
        return nil, fmt.Errorf("创建 JetStream 失败: %w", err)
    }

    return &Manager{js: js, nc: nc}, nil
}

// Close 关闭连接
func (m *Manager) Close() {
    if m.nc != nil {
        m.nc.Close()
    }
}

// CreateOrderStream 创建订单 Stream
func (m *Manager) CreateOrderStream(ctx context.Context) error {
    // 检查是否已存在
    stream, err := m.js.Stream(ctx, "ORDERS")
    if err == nil {
        info, _ := stream.CachedInfo()
        log.Printf("Stream ORDERS 已存在，消息数: %d", info.State.Msgs)
        return nil
    }

    // 创建 Stream
    stream, err = m.js.CreateStream(ctx, jetstream.StreamConfig{
        Name:     "ORDERS",
        Subjects: []string{"orders.>"},
        Storage: jetstream.FileStorage,
        
        // 保留策略
        Retention: jetstream.LimitsPolicy{
            MaxBytes: 1024 * 1024 * 1024,  // 1GB
            MaxAge:   time.Hour * 24 * 7,  // 7 天
            MaxMsgs:  100000,
        },
        
        // 丢弃策略
        Discard: jetstream.DiscardOld,
        
        // 副本数
        Replicas: 1,
        
        // 允许直接访问
        AllowDirect: true,
    })
    if err != nil {
        return fmt.Errorf("创建 Stream 失败: %w", err)
    }

    log.Printf("创建 Stream 成功: %s", stream.CachedInfo().Config.Name)
    return nil
}

// CreateConsumer 创建消费者
func (m *Manager) CreateConsumer(ctx context.Context, name, subject string) (jetstream.Consumer, error) {
    consumer, err := m.js.CreateConsumer(ctx, "ORDERS", jetstream.ConsumerConfig{
        Name:          name,
        Durable:       true,
        FilterSubject: subject,
        DeliverPolicy: jetstream.DeliverAll,
        
        // 确认策略
        AckPolicy:   jetstream.AckExplicit,
        AckWait:     30 * time.Second,
        MaxDeliver:  3,
        
        // 背压策略
        MaxAckPending: 100,
        
        // 重试策略
        BackOff: []time.Duration{
            time.Second,
            time.Second * 5,
            time.Second * 30,
        },
        
        // 过滤器
        FilterSubject: subject,
    })
    if err != nil {
        return nil, fmt.Errorf("创建消费者失败: %w", err)
    }

    return consumer, nil
}

// GetStreamInfo 获取 Stream 信息
func (m *Manager) GetStreamInfo(ctx context.Context) (*jetstream.StreamInfo, error) {
    stream, err := m.js.Stream(ctx, "ORDERS")
    if err != nil {
        return nil, err
    }
    return stream.CachedInfo(), nil
}

// ListConsumers 列出所有消费者
func (m *Manager) ListConsumers(ctx context.Context) ([]jetstream.ConsumerInfo, error) {
    return m.js.ListConsumers(ctx, "ORDERS")
}

// PurgeStream 清空 Stream
func (m *Manager) PurgeStream(ctx context.Context) error {
    stream, err := m.js.Stream(ctx, "ORDERS")
    if err != nil {
        return err
    }
    return stream.Purge(ctx)
}
```

## 4.4 订单生产者

```go
// cmd/producer/main.go
package main

import (
    "context"
    "fmt"
    "log"
    "os"
    "os/signal"
    "syscall"
    "time"

    "order-processing/internal/order"
    "order-processing/internal/stream"

    "github.com/nats-io/nats.go/jetstream"
)

type Producer struct {
    js    jetstream.JetStream
    ctx   context.Context
}

func NewProducer(url string) (*Producer, error) {
    mgr, err := stream.New(url)
    if err != nil {
        return nil, err
    }

    js, err := jetstream.New(mgr.GetNATSConn())
    if err != nil {
        return nil, err
    }

    return &Producer{
        js:  js,
        ctx: context.Background(),
    }, nil
}

func (p *Producer) PublishOrder(order *order.Order) error {
    data, err := order.ToJSON()
    if err != nil {
        return err
    }

    // 使用消息 ID 实现幂等
    msgID := fmt.Sprintf("%s-%d", order.ID, time.Now().UnixNano())

    _, err = p.js.Publish(p.ctx, "orders.created", data,
        jetstream.WithMsgID(msgID),
        jetstream.WithHeader(nats.Header{
            "Content-Type": []string{"application/json"},
            "Order-Status": []string{string(order.Status)},
        }),
    )
    return err
}

func (p *Producer) PublishOrderUpdate(order *order.Order) error {
    data, err := order.ToJSON()
    if err != nil {
        return err
    }

    subject := fmt.Sprintf("orders.%s", order.ID)
    _, err = p.js.Publish(p.ctx, subject, data)
    return err
}

func main() {
    producer, err := NewProducer(nats.DefaultURL)
    if err != nil {
        log.Fatal(err)
    }

    // 创建 Stream
    if err := producer.js.CreateStream(producer.ctx, jetstream.StreamConfig{
        Name:     "ORDERS",
        Subjects: []string{"orders.>"},
        Storage:  jetstream.FileStorage,
    }); err != nil {
        log.Printf("创建 Stream: %v", err)
    }

    log.Println("生产者已启动")

    // 产品列表
    products := []struct {
        id, name string
        price    float64
    }{
        {"P001", "iPhone 15", 999.0},
        {"P002", "MacBook Pro", 1999.0},
        {"P003", "iPad Air", 599.0},
        {"P004", "AirPods Pro", 249.0},
    }

    // 监听信号
    sigCh := make(chan os.Signal, 1)
    signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

    ticker := time.NewTicker(2 * time.Second)
    defer ticker.Stop()

    i := 0
    for {
        select {
        case <-ticker.C:
            product := products[i%len(products)]
            ord := order.NewOrder(
                fmt.Sprintf("C%03d", i%10+1),
                product.id,
                product.name,
                1,
                product.price,
            )

            if err := producer.PublishOrder(ord); err != nil {
                log.Printf("发布订单失败: %v", err)
            } else {
                log.Printf("已发布订单: %s, %s, %.2f", ord.ID, ord.ProductName, ord.Amount)
            }
            i++

        case <-sigCh:
            log.Println("收到退出信号")
            return
        }
    }
}
```

## 4.5 订单消费者

```go
// internal/consumer/processor.go
package consumer

import (
    "context"
    "fmt"
    "log"
    "sync"
    "time"

    "github.com/nats-io/nats.go/jetstream"

    "order-processing/internal/order"
)

// Processor 订单处理器
type Processor struct {
    name     string
    consumer jetstream.Consumer
    wg       sync.WaitGroup
    running  bool
}

// New 创建处理器
func New(name string, consumer jetstream.Consumer) *Processor {
    return &Processor{
        name:     name,
        consumer: consumer,
    }
}

// Start 启动处理
func (p *Processor) Start(ctx context.Context, workers int) {
    p.running = true
    
    for i := 0; i < workers; i++ {
        p.wg.Add(1)
        go p.worker(ctx, i)
    }
    
    log.Printf("[%s] 启动了 %d 个 worker", p.name, workers)
}

// Stop 停止处理
func (p *Processor) Stop() {
    p.running = false
    p.wg.Wait()
    log.Printf("[%s] 已停止", p.name)
}

func (p *Processor) worker(ctx context.Context, id int) {
    defer p.wg.Done()

    for p.running {
        select {
        case <-ctx.Done():
            return
        default:
        }

        // 拉取消息
        msgs, err := p.consumer.Fetch(1, jetstream.FetchMaxWait(5*time.Second))
        if err != nil {
            if err == context.Canceled {
                return
            }
            log.Printf("[%s Worker-%d] 获取消息失败: %v", p.name, id, err)
            time.Sleep(time.Second)
            continue
        }

        for msg := range msgs {
            if err := p.processMessage(msg); err != nil {
                log.Printf("[%s Worker-%d] 处理失败: %v", p.name, id, err)
            }
        }
    }
}

func (p *Processor) processMessage(msg jetstream.Msg) error {
    // 解析订单
    ord, err := order.FromJSON(msg.Data())
    if err != nil {
        log.Printf("[%s] 解析订单失败: %v", p.name, err)
        msg.Term()  // 终止消息，不再重试
        return err
    }

    log.Printf("[%s] 收到订单: %s, 状态: %s", p.name, ord.ID, ord.Status)

    // 模拟业务处理
    time.Sleep(500 * time.Millisecond)

    // 更新订单状态
    switch ord.Status {
    case order.StatusPending:
        ord.UpdateStatus(order.StatusPaid)
    case order.StatusPaid:
        ord.UpdateStatus(order.StatusProcessing)
    case order.StatusProcessing:
        ord.UpdateStatus(order.StatusCompleted)
    }

    log.Printf("[%s] 处理完成: %s -> %s", p.name, ord.ID, ord.Status)

    // 确认消息
    return msg.Ack()
}

// GetMetrics 获取处理指标
func (p *Processor) GetMetrics() map[string]interface{} {
    info, _ := p.consumer.CachedInfo()
    return map[string]interface{}{
        "name":           p.name,
        "pending":        info.NumPending,
        "delivered":      info.NumDelivered,
        "ack_wait":       info.AckWait.Nanoseconds(),
        "max_deliver":    info.MaxDeliver,
    }
}
```

```go
// cmd/consumer/main.go
package main

import (
    "context"
    "log"
    "os"
    "os/signal"
    "syscall"
    "time"

    "order-processing/internal/consumer"
    "order-processing/internal/stream"

    "github.com/nats-io/nats.go"
)

func main() {
    // 创建 Stream 管理器
    mgr, err := stream.New(nats.DefaultURL)
    if err != nil {
        log.Fatal(err)
    }
    defer mgr.Close()

    ctx := context.Background()

    // 创建 Stream
    if err := mgr.CreateOrderStream(ctx); err != nil {
        log.Printf("创建 Stream: %v", err)
    }

    // 创建消费者
    consumerName := "order-processor"
    consumer, err := mgr.CreateConsumer(ctx, consumerName, "orders.created")
    if err != nil {
        log.Printf("创建消费者: %v", err)
    }

    // 创建处理器
    proc := consumer.NewProcessor(consumerName, consumer)
    
    // 启动处理
    ctx, cancel := context.WithCancel(context.Background())
    proc.Start(ctx, 3)

    // 监听信号
    sigCh := make(chan os.Signal, 1)
    signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

    // 定期打印指标
    go func() {
        ticker := time.NewTicker(10 * time.Second)
        defer ticker.Stop()
        for {
            select {
            case <-ticker.C:
                metrics := proc.GetMetrics()
                log.Printf("消费者指标: %+v", metrics)
            case <-ctx.Done():
                return
            }
        }
    }()

    <-sigCh
    log.Println("收到退出信号，正在关闭...")
    cancel()
    proc.Stop()
    log.Println("已关闭")
}
```

## 4.6 管理工具

```go
// cmd/admin/main.go
package main

import (
    "context"
    "fmt"
    "log"
    "os"

    "order-processing/internal/stream"

    "github.com/nats-io/nats.go"
)

func main() {
    mgr, err := stream.New(nats.DefaultURL)
    if err != nil {
        log.Fatal(err)
    }
    defer mgr.Close()

    ctx := context.Background()

    // 获取 Stream 信息
    info, err := mgr.GetStreamInfo(ctx)
    if err != nil {
        log.Printf("获取 Stream 信息失败: %v", err)
        os.Exit(1)
    }

    fmt.Println("=== Stream 信息 ===")
    fmt.Printf("名称: %s\n", info.Config.Name)
    fmt.Printf("主题: %v\n", info.Config.Subjects)
    fmt.Printf("消息数: %d\n", info.State.Msgs)
    fmt.Printf("字节数: %d\n", info.State.Bytes)
    fmt.Printf("消费者数: %d\n", info.State.Consumers)

    // 列出消费者
    consumers, err := mgr.ListConsumers(ctx)
    if err != nil {
        log.Printf("列出消费者失败: %v", err)
    } else {
        fmt.Println("\n=== 消费者列表 ===")
        for _, c := range consumers {
            fmt.Printf("名称: %s, 待处理: %d, 已投递: %d\n",
                c.Name, c.NumPending, c.NumDelivered)
        }
    }
}
```

## 4.7 配置文件

```yaml
# config.yaml
nats:
  url: "nats://localhost:4222"
  timeout: 10s

stream:
  name: "ORDERS"
  subjects:
    - "orders.>"
  storage: "file"
  max_bytes: 1073741824  # 1GB
  max_age: 604800        # 7 天
  max_msgs: 100000

consumer:
  name: "order-processor"
  workers: 3
  prefetch: 10
  ack_wait: 30s
  max_deliver: 3
  max_ack_pending: 100

retry:
  backoff:
    - 1s
    - 5s
    - 30s
```

## 4.8 运行项目

### 启动 NATS 服务器

```bash
docker run -d --name nats-server -p 4222:4222 -p 8222:8222 nats:latest -js
```

### 启动生产者

```bash
cd cmd/producer
go run main.go
```

### 启动消费者

```bash
cd cmd/consumer
go run main.go
```

### 启动管理工具

```bash
cd cmd/admin
go run main.go
```

## 4.9 测试验证

### 查看 Stream 状态

```bash
go run cmd/admin/main.go

# 输出示例
=== Stream 信息 ===
名称: ORDERS
主题: [orders.>]
消息数: 156
字节数: 24567
消费者数: 1

=== 消费者列表 ===
名称: order-processor, 待处理: 0, 已投递: 156
```

### 查看监控面板

```bash
# 访问 NATS 监控页面
http://localhost:8222/streamz
```

---

## 4.10 扩展练习

1. **添加数据库持久化**：将订单存储到 PostgreSQL
2. **添加消息转换**：支持 Protobuf 格式
3. **添加分布式事务**：实现订单创建和支付的分布式事务
4. **添加死信队列**：处理失败的消息
5. **添加监控告警**：集成 Prometheus 和 Alertmanager
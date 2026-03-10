# 入门项目：NATS 消息订阅发布系统

本项目实现一个基于 NATS 的消息订阅发布系统，包含发布者、订阅者和队列组功能。

## 3.1 项目概述

### 功能需求

1. 实现消息发布者 (Publisher)
2. 实现消息订阅者 (Subscriber)
3. 支持主题通配符订阅
4. 实现请求/响应模式
5. 实现队列组负载均衡

### 项目结构

```
nats-pubsub/
├── cmd/
│   ├── publisher/main.go      # 发布者
│   ├── subscriber/main.go     # 订阅者
│   └── requester/main.go      # 请求者
├── internal/
│   ├── publisher/publisher.go # 发布者逻辑
│   ├── subscriber/subscriber.go
│   └── message/message.go     # 消息定义
├── go.mod
└── config.yaml
```

## 3.2 消息定义

```go
// internal/message/message.go
package message

import (
    "encoding/json"
    "time"
)

// Order 订单消息
type Order struct {
    ID        string    `json:"id"`
    Product   string    `json:"product"`
    Amount    float64   `json:"amount"`
    Status    string    `json:"status"`
    CreatedAt time.Time `json:"created_at"`
}

// NewOrder 创建新订单
func NewOrder(product string, amount float64) *Order {
    return &Order{
        ID:        generateID(),
        Product:   product,
        Amount:    amount,
        Status:    "pending",
        CreatedAt: time.Now(),
    }
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

// generateID 生成唯一 ID
func generateID() string {
    return time.Now().Format("20060102150405") + "-" + randomString(6)
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

## 3.3 发布者实现

```go
// internal/publisher/publisher.go
package publisher

import (
    "fmt"
    "log"
    "time"

    "github.com/nats-io/nats.go"
    "nats-pubsub/internal/message"
)

// Publisher NATS 发布者
type Publisher struct {
    nc *nats.Conn
}

// New 创建发布者
func New(url string) (*Publisher, error) {
    nc, err := nats.Connect(url)
    if err != nil {
        return nil, fmt.Errorf("连接 NATS 失败: %w", err)
    }
    return &Publisher{nc: nc}, nil
}

// Close 关闭连接
func (p *Publisher) Close() {
    if p.nc != nil {
        p.nc.Close()
    }
}

// PublishOrder 发布订单消息
func (p *Publisher) PublishOrder(subject string, order *message.Order) error {
    data, err := order.ToJSON()
    if err != nil {
        return fmt.Errorf("序列化消息失败: %w", err)
    }

    err = p.nc.Publish(subject, data)
    if err != nil {
        return fmt.Errorf("发布消息失败: %w", err)
    }

    log.Printf("已发布订单 [%s] 到主题 %s", order.ID, subject)
    return nil
}

// PublishWithHeader 带消息头发布
func (p *Publisher) PublishWithHeader(subject string, order *message.Order) error {
    data, err := order.ToJSON()
    if err != nil {
        return err
    }

    msg := &nats.Msg{
        Subject: subject,
        Header: nats.Header{
            "Content-Type":   []string{"application/json"},
            "Correlation-ID": []string{order.ID},
            "Timestamp":      []string{time.Now().Format(time.RFC3339)},
        },
        Data: data,
    }

    return p.nc.PublishMsg(msg)
}

// PublishAsync 异步发布
func (p *Publisher) PublishAsync(subject string, order *message.Order) error {
    data, err := order.ToJSON()
    if err != nil {
        return err
    }

    _, err = p.nc.PublishAsync(subject, data)
    return err
}

// Flush 刷新缓冲区
func (p *Publisher) Flush() error {
    return p.nc.Flush()
}

// GetStats 获取连接状态
func (p *Publisher) GetStats() map[string]interface{} {
    return map[string]interface{}{
        "connected":    p.nc.IsConnected(),
        "server":       p.nc.ConnectedUrl(),
        "max_payload":  p.nc.MaxPayload(),
    }
}
```

```go
// cmd/publisher/main.go
package main

import (
    "fmt"
    "log"
    "os"
    "os/signal"
    "syscall"
    "time"

    "nats-pubsub/internal/message"
    "nats-publisher"
)

func main() {
    // 创建发布者
    pub, err := publisher.New(nats.DefaultURL)
    if err != nil {
        log.Fatal(err)
    }
    defer pub.Close()

    log.Printf("发布者已启动，连接到 %s", pub.GetStats()["server"])

    // 监听系统信号
    sigCh := make(chan os.Signal, 1)
    signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

    ticker := time.NewTicker(2 * time.Second)
    defer ticker.Stop()

    productNames := []string{"iPhone 15", "MacBook Pro", "iPad Air", "AirPods Pro"}
    i := 0

    for {
        select {
        case <-ticker.C:
            product := productNames[i%len(productNames)]
            amount := 999.0 + float64(i)*100

            order := message.NewOrder(product, amount)
            
            // 发布到订单主题
            subject := "orders.created"
            if err := pub.PublishOrder(subject, order); err != nil {
                log.Printf("发布失败: %v", err)
            }

            // 同时发布带消息头的版本
            pub.PublishWithHeader("orders.created.with-header", order)

            i++

        case <-sigCh:
            log.Println("收到退出信号，正在关闭...")
            pub.Flush()
            return
        }
    }
}
```

## 3.4 订阅者实现

```go
// internal/subscriber/subscriber.go
package subscriber

import (
    "encoding/json"
    "fmt"
    "log"
    "time"

    "github.com/nats-io/nats.go"
    "nats-pubsub/internal/message"
)

// Subscriber NATS 订阅者
type Subscriber struct {
    nc     *nats.Conn
    sub    *nats.Subscription
    name   string
}

// New 创建订阅者
func New(url, name string) (*Subscriber, error) {
    nc, err := nats.Connect(url)
    if err != nil {
        return nil, fmt.Errorf("连接 NATS 失败: %w", err)
    }
    return &Subscriber{nc: nc, name: name}, nil
}

// Close 关闭连接
func (s *Subscriber) Close() {
    if s.sub != nil {
        s.sub.Unsubscribe()
    }
    if s.nc != nil {
        s.nc.Close()
    }
}

// Subscribe 订阅主题
func (s *Subscriber) Subscribe(subject string, handler func(*message.Order) error) error {
    sub, err := s.nc.Subscribe(subject, func(m *nats.Msg) {
        order, err := message.FromJSON(m.Data)
        if err != nil {
            log.Printf("[%s] 解析消息失败: %v", s.name, err)
            m.Term()
            return
        }

        // 调用处理函数
        if err := handler(order); err != nil {
            log.Printf("[%s] 处理消息失败: %v", s.name, err)
            m.Nak()
            return
        }

        m.Ack()
        log.Printf("[%s] 已处理订单: %s", s.name, order.ID)
    })
    if err != nil {
        return err
    }
    s.sub = sub
    return nil
}

// SubscribeWithOptions 带选项订阅
func (s *Subscriber) SubscribeWithOptions(subject string, handler func(*message.Order) error) error {
    sub, err := s.nc.Subscribe(subject, func(m *nats.Msg) {
        start := time.Now()
        
        order, err := message.FromJSON(m.Data)
        if err != nil {
            log.Printf("[%s] 解析消息失败: %v", s.name, err)
            m.Term()
            return
        }

        if err := handler(order); err != nil {
            log.Printf("[%s] 处理消息失败: %v", s.name, err)
            m.Nak()
            return
        }

        m.Ack()
        log.Printf("[%s] 已处理订单: %s, 耗时: %v", s.name, order.ID, time.Since(start))
    },
        nats.Durable(s.name),
        nats.AckExplicit(),
    )
    if err != nil {
        return err
    }
    s.sub = sub
    return nil
}

// SubscribeWildcard 订阅通配符主题
func (s *Subscriber) SubscribeWildcard(subject string, handler func(string, *message.Order) error) error {
    sub, err := s.nc.Subscribe(subject, func(m *nats.Msg) {
        order, err := message.FromJSON(m.Data)
        if err != nil {
            log.Printf("[%s] 解析消息失败: %v", s.name, err)
            return
        }

        if err := handler(m.Subject, order); err != nil {
            log.Printf("[%s] 处理消息失败: %v", s.name, err)
        }
    })
    if err != nil {
        return err
    }
    s.sub = sub
    return nil
}

// QueueSubscribe 队列订阅
func (s *Subscriber) QueueSubscribe(subject, queue string, handler func(*message.Order) error) error {
    sub, err := s.nc.QueueSubscribe(subject, queue, func(m *nats.Msg) {
        order, err := message.FromJSON(m.Data)
        if err != nil {
            log.Printf("[%s] 解析消息失败: %v", s.name, err)
            return
        }

        if err := handler(order); err != nil {
            log.Printf("[%s] 处理消息失败: %v", s.name, err)
        }
    })
    if err != nil {
        return err
    }
    s.sub = sub
    return nil
}
```

```go
// cmd/subscriber/main.go
package main

import (
    "fmt"
    "log"
    "time"

    "nats-pubsub/internal/message"
    "nats-pubsub/internal/subscriber"
)

func main() {
    // 创建订阅者
    sub, err := subscriber.New(nats.DefaultURL, "order-processor")
    if err != nil {
        log.Fatal(err)
    }
    defer sub.Close()

    // 处理函数
    handler := func(order *message.Order) error {
        // 模拟业务处理
        fmt.Printf("收到订单: ID=%s, Product=%s, Amount=%.2f\n",
            order.ID, order.Product, order.Amount)
        
        // 模拟处理延迟
        time.Sleep(100 * time.Millisecond)
        
        order.Status = "processed"
        return nil
    }

    // 订阅订单创建主题
    if err := sub.SubscribeWithOptions("orders.created", handler); err != nil {
        log.Fatal(err)
    }

    log.Println("订阅者已启动，等待消息...")

    // 保持运行
    select {}
}
```

## 3.5 请求/响应实现

```go
// cmd/requester/main.go
package main

import (
    "encoding/json"
    "fmt"
    "log"
    "time"

    "github.com/nats-io/nats.go"
)

type OrderQuery struct {
    OrderID string `json:"order_id"`
}

type OrderResponse struct {
    OrderID  string  `json:"order_id"`
    Product  string  `json:"product"`
    Amount   float64 `json:"amount"`
    Status   string  `json:"status"`
    Found    bool    `json:"found"`
}

func main() {
    nc, err := nats.Connect(nats.DefaultURL)
    if err != nil {
        log.Fatal(err)
    }
    defer nc.Close()

    // 模拟订单数据
    orders := map[string]OrderResponse{
        "1001": {OrderID: "1001", Product: "iPhone 15", Amount: 999.0, Status: "pending", Found: true},
        "1002": {OrderID: "1002", Product: "MacBook Pro", Amount: 1999.0, Status: "completed", Found: true},
    }

    // 注册请求处理器
    nc.Subscribe("orders.get", func(m *nats.Msg) {
        var query OrderQuery
        if err := json.Unmarshal(m.Data, &query); err != nil {
            log.Printf("解析请求失败: %v", err)
            return
        }

        log.Printf("收到查询请求: %s", query.OrderID)

        // 查找订单
        order, ok := orders[query.OrderID]
        if !ok {
            order = OrderResponse{OrderID: query.OrderID, Found: false}
        }

        // 响应
        resp, _ := json.Marshal(order)
        m.Respond(resp)
    })

    // 发送请求
    query := OrderQuery{OrderID: "1001"}
    data, _ := json.Marshal(query)

    log.Printf("发送查询请求: %s", query.OrderID)
    
    resp, err := nc.Request("orders.get", data, 5*time.Second)
    if err != nil {
        log.Printf("请求失败: %v", err)
        return
    }

    var order OrderResponse
    json.Unmarshal(resp.Data, &order)
    fmt.Printf("收到响应: %+v\n", order)
}
```

## 3.6 配置文件

```yaml
# config.yaml
nats:
  url: "nats://localhost:4222"
  name: "nats-pubsub"
  timeout: 10s
  max_reconnects: 5

publisher:
  batch_size: 100
  flush_interval: 1s

subscriber:
  queue_name: "order-workers"
  prefetch: 10

topics:
  orders_created: "orders.created"
  orders_updated: "orders.updated"
  orders_cancelled: "orders.cancelled"
```

## 3.7 运行项目

### 启动 NATS 服务器

```bash
# 使用 Docker
docker run -d --name nats-server -p 4222:4222 nats:latest -js

# 或本地运行
nats-server -js
```

### 启动发布者

```bash
cd cmd/publisher
go run main.go
```

### 启动订阅者

```bash
cd cmd/subscriber
go run main.go
```

### 启动多个订阅者（队列组）

```bash
# 终端 1
cd cmd/subscriber
go run main.go

# 终端 2
cd cmd/subscriber
go run main.go
```

### 启动请求者

```bash
cd cmd/requester
go run main.go
```

## 3.8 测试验证

### 测试发布/订阅

```bash
# 发布测试消息
go run cmd/publisher/main.go

# 订阅者输出
2024/01/01 12:00:00 订阅者已启动，等待消息...
2024/01/01 12:00:02 已发布订单 [20240101120002-abc123] 到主题 orders.created
2024/01/01 12:00:02 [order-processor] 收到订单: ID=20240101120002-abc123, Product=iPhone 15, Amount=999.00
2024/01/01 12:00:02 [order-processor] 已处理订单: 20240101120002-abc123
```

### 测试队列组

```bash
# 启动两个订阅者，观察负载均衡
# 消息会被交替分配给不同的订阅者
```

---

## 3.9 扩展练习

1. **添加日志记录**：使用 zerolog 或 zap 记录详细日志
2. **添加指标监控**：集成 Prometheus 监控消息处理
3. **添加重试机制**：实现失败消息的重试逻辑
4. **添加消息转换**：支持 Protobuf 消息格式
5. **添加配置管理**：使用 Viper 读取配置文件
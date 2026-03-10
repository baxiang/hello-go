# NATS 核心概念与基础

NATS 是一个轻量级、高性能、云原生的消息系统，由 Apcera 开发并维护。它以其简单性、高吞吐量和可靠性著称，广泛用于微服务架构、事件驱动系统和物联网场景。

## 1.1 NATS 简介

### 核心特性

| 特性 | 说明 |
|------|------|
| **轻量级** | 二进制消息格式，协议简单，客户端库极小 |
| **高性能** | 单节点可达百万级消息/秒 |
| **多模式** | 支持发布/订阅、请求/响应、队列组 |
| **持久化** | JetStream 提供消息持久化和 Exactly-Once 语义 |
| **云原生** | 原生支持 Kubernetes，自动服务发现 |
| **多语言** | 支持 40+ 编程语言客户端 |

### NATS 与其他消息队列对比

| 特性 | NATS | RabbitMQ | Kafka |
|------|------|----------|-------|
| 协议 | NATS | AMQP | 自定义 |
| 延迟 | 极低 (~1ms) | 低 | 中等 |
| 持久化 | JetStream | 支持 | 支持 |
| 消息顺序 | 支持 | 支持 | 支持 |
| 集群模式 | 原生 | 支持 | 支持 |
| 客户端大小 | 极小 | 中等 | 大 |

### 安装 NATS Server

```bash
# 使用 Docker 运行
docker run -d --name nats-server -p 4222:4222 nats:latest

# 使用 Homebrew (macOS)
brew install nats-server
nats-server

# 使用 Go 安装
go install github.com/nats-io/nats-server/v2@latest

# 编译运行
nats-server
```

---

## 1.2 NATS Go 客户端

### 安装

```bash
go get github.com/nats-io/nats.go
```

### 基础连接

```go
package main

import (
    "fmt"
    "log"

    "github.com/nats-io/nats.go"
)

func main() {
    // 连接到默认服务器 (localhost:4222)
    nc, err := nats.Connect(nats.DefaultURL)
    if err != nil {
        log.Fatal(err)
    }
    defer nc.Close()

    fmt.Println("已连接到 NATS 服务器")
}
```

### 连接选项

```go
// 多种连接方式
nc, err := nats.Connect("nats://localhost:4222")

// 使用选项
nc, err := nats.Connect(
    "nats://localhost:4222",
    nats.Name("MyClient"),
    nats.MaxReconnects(5),
    nats.ReconnectWait(time.Second),
    nats.Timeout(10 * time.Second),
)

// 认证连接
nc, err := nats.Connect(
    "nats://user:password@localhost:4222",
)

// 使用 Token 认证
nc, err := nats.Connect(
    "nats://mytoken@localhost:4222",
)

// 启用 TLS
nc, err := nats.Connect(
    "tls://localhost:4222",
    nats.ClientCert("client.crt", "client.key"),
    nats.RootCAs("ca.crt"),
)

// 连接到集群
nc, err := nats.Connect(
    "nats://server1:4222,nats://server2:4222,nats://server3:4222",
)
```

### 连接状态监听

```go
// 订阅连接状态
nc.AddStatusListener(func(conn *nats.Conn, evt nats.Status) {
    fmt.Printf("连接状态变化: %s\n", evt)
}, nats.DISCONNECTED, nats.RECONNECTING, nats.CONNECTED)

// 处理断开连接
if nc.LastError() != nil {
    fmt.Printf("连接错误: %s\n", nc.LastError())
}
```

---

## 1.3 发布/订阅模式

### 基本发布/订阅

```go
package main

import (
    "fmt"
    "log"

    "github.com/nats-io/nats.go"
)

// 发布者
func publish(nc *nats.Conn, subject, message string) error {
    return nc.Publish(subject, []byte(message))
}

// 订阅者
func subscribe(nc *nats.Conn, subject string) {
    nc.Subscribe(subject, func(m *nats.Msg) {
        fmt.Printf("收到消息 [%s]: %s\n", m.Subject, string(m.Data))
        // 手动响应
        m.Respond([]byte("ACK"))
    })
}

func main() {
    nc, err := nats.Connect(nats.DefaultURL)
    if err != nil {
        log.Fatal(err)
    }
    defer nc.Close()

    // 订阅主题
    subscribe(nc, "events.orders")

    // 保持订阅通道活跃
    nc.Flush()

    // 发布消息
    for i := 1; i <= 5; i++ {
        msg := fmt.Sprintf("订单 #%d 已创建", i)
        if err := publish(nc, "events.orders", msg); err != nil {
            log.Printf("发布失败: %v", err)
        } else {
            fmt.Printf("已发布: %s\n", msg)
        }
    }

    // 等待消息处理
    nc.Flush()
}
```

### 主题通配符

```go
// * 匹配单个词 (不含点号)
nc.Subscribe("orders.*", func(m *nats.Msg) {
    fmt.Printf("收到订单事件: %s\n", string(m.Data))
})

// > 匹配多层主题
nc.Subscribe("orders.>", func(m *nats.Msg) {
    fmt.Printf("收到所有订单相关消息: %s\n", m.Subject)
})

// 组合使用
nc.Subscribe("orders.*.created", func(m *nats.Msg) {
    fmt.Printf("收到订单创建消息: %s\n", string(m.Data))
})
```

### 异步订阅

```go
// 异步订阅 - 返回订阅对象
sub, err := nc.Subscribe("events.orders", func(m *nats.Msg) {
    fmt.Printf("收到消息: %s\n", string(m.Data))
})
if err != nil {
    log.Fatal(err)
}

// 手动管理订阅
sub.Unsubscribe()           // 取消订阅
sub.AutoUnsubscribe(10)    // 收到 10 条消息后自动取消
```

### 消息结构

```go
// NATS 消息结构
type Msg struct {
    Subject string        // 主题
    Reply   string        // 回复地址 (用于请求/响应)
    Data    []byte        // 消息内容
    Header  nats.Header   // 消息头 (NATS 2.0+)
}

// 使用消息头
nc.Subscribe("events.orders", func(m *nats.Msg) {
    // 读取消息头
    if m.Header == nil {
        m.Header = nats.Header{}
    }
    contentType := m.Header.Get("Content-Type")
    correlationID := m.Header.Get("Correlation-ID")
    
    fmt.Printf("Content-Type: %s, Correlation-ID: %s\n", contentType, correlationID)
})

// 发布带消息头的消息
nc.PublishMsg(&nats.Msg{
    Subject: "events.orders",
    Header: nats.Header{
        "Content-Type":     []string{"application/json"},
        "Correlation-ID":   []string{"uuid-12345"},
    },
    Data: []byte(`{"order_id": 1001}`),
})
```

---

## 1.4 请求/响应模式

### 基本请求/响应

```go
package main

import (
    "fmt"
    "log"
    "time"

    "github.com/nats-io/nats.go"
)

// 请求服务
func requestHandler(nc *nats.Conn, subject string) {
    nc.Subscribe(subject, func(m *nats.Msg) {
        fmt.Printf("收到请求: %s\n", string(m.Data))
        // 处理请求并响应
        response := fmt.Sprintf(`{"status": "ok", "echo": "%s"}`, string(m.Data))
        m.Respond([]byte(response))
    })
}

// 发送请求
func request(nc *nats.Conn, subject, payload string) (string, error) {
    msg, err := nc.Request(subject, []byte(payload), time.Second)
    if err != nil {
        return "", err
    }
    return string(msg.Data), nil
}

func main() {
    nc, err := nats.Connect(nats.DefaultURL)
    if err != nil {
        log.Fatal(err)
    }
    defer nc.Close()

    // 注册请求处理器
    requestHandler(nc, "orders.get")

    // 发送请求
    response, err := request(nc, "orders.get", `{"order_id": 1001}`)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("响应: %s\n", response)
}
```

### 异步请求

```go
// 异步请求 - 不等待响应
nc.PublishRequest("orders.get", "reply.subject", []byte(`{"order_id": 1001}`))

// 带回调的异步请求
nc.RequestAsync("orders.get", "reply.subject", []byte(`{"order_id": 1001}`), 
    func(conn *nats.Conn, msg *nats.Msg, err error) {
        if err != nil {
            fmt.Printf("请求失败: %v\n", err)
            return
        }
        fmt.Printf("异步响应: %s\n", string(msg.Data))
    },
)
```

### 请求超时和重试

```go
// 设置超时
msg, err := nc.Request("orders.get", []byte(""), 5*time.Second)

// 使用上下文
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

msg, err := nc.RequestWithContext(ctx, "orders.get", []byte("request data"))
if err != nil {
    if ctx.Err() == context.DeadlineExceeded {
        fmt.Println("请求超时")
    }
}
```

---

## 1.5 队列组

### 基本队列组

```go
package main

import (
    "fmt"
    "log"
    "sync"
    "time"

    "github.com/nats-io/nats.go"
)

func main() {
    nc, err := nats.Connect(nats.DefaultURL)
    if err != nil {
        log.Fatal(err)
    }
    defer nc.Close()

    var wg sync.WaitGroup
    
    // 创建多个队列消费者
    for i := 1; i <= 3; i++ {
        wg.Add(1)
        workerID := i
        go func() {
            defer wg.Done()
            // 队列组名相同，消息只会被一个消费者接收
            sub, err := nc.QueueSubscribe("orders.queue", "order-workers", 
                func(m *nats.Msg) {
                    fmt.Printf("Worker %d 收到消息: %s\n", workerID, string(m.Data))
                    m.Ack()
                })
            if err != nil {
                log.Printf("订阅失败: %v", err)
            }
            
            // 保持订阅活跃
            time.Sleep(10 * time.Second)
        }()
    }

    // 等待消费者启动
    time.Sleep(500 * time.Millisecond)

    // 发布多条消息
    for i := 1; i <= 6; i++ {
        nc.Publish("orders.queue", []byte(fmt.Sprintf("订单 #%d", i)))
    }
    nc.Flush()

    wg.Wait()
}
```

### 队列组配置

```go
// 带选项的队列订阅
sub, err := nc.QueueSubscribe(
    "orders.queue",
    "order-workers",
    func(m *nats.Msg) {
        fmt.Printf("收到: %s\n", string(m.Data))
    },
    nats.Durable("order-worker-1"),  // 持久订阅
    nats.MaxDeliver(3),              // 最大重试次数
    nats.AckExplicit(),              // 显式确认
)
if err != nil {
    log.Fatal(err)
}

// 手动管理
sub.Unsubscribe()
sub.NextMsg(100 * time.Millisecond)  // 阻塞获取下一条消息
```

---

## 1.6 错误处理

### 连接错误

```go
nc, err := nats.Connect(nats.DefaultURL)
if err != nil {
    switch {
    case errors.Is(err, nats.ErrNoServers):
        fmt.Println("没有可用的服务器")
    case errors.Is(err, nats.ErrTimeout):
        fmt.Println("连接超时")
    case errors.Is(err, nats.ErrAuthorization):
        fmt.Println("认证失败")
    default:
        fmt.Printf("连接错误: %v\n", err)
    }
}

// 监听错误
nc.SetErrorHandler(func(conn *nats.Conn, err error) {
    fmt.Printf("NATS 错误: %v\n", err)
})
```

### 消息处理错误

```go
nc.Subscribe("orders.created", func(m *nats.Msg) {
    // 使用 panic/recover 处理
    defer func() {
        if r := recover(); r != nil {
            fmt.Printf("处理 panic: %v\n", r)
            // 不确认消息，让其重新投递
        }
    }()
    
    // 处理消息
    var order Order
    if err := json.Unmarshal(m.Data, &order); err != nil {
        // 解析失败，确认跳过
        m.Term()
        return
    }
    
    // 业务处理...
    m.Ack()
})
```

---

## 1.7 监控和管理

### NATS 监控端点

```bash
# 启动带监控的服务器
nats-server -m 8222

# 访问监控端点
curl http://localhost:8222/healthz
curl http://localhost:8222/varz
curl http://localhost:8222/connz
curl http://localhost:8222/subz
curl http://localhost:8222/streamsz
```

### 使用 Go 客户端监控

```go
nc, _ := nats.Connect(nats.DefaultURL)

// 获取连接信息
info := nc.ConnectedUrl()
serverInfo := nc.ServerInfo()
fmt.Printf("已连接到: %s, 服务器版本: %s\n", info, serverInfo.Version)
```

---

## 1.8 安全配置

### 用户认证

```bash
# 启动带用户认证的服务器
nats-server --user myuser --pass mypassword
```

### Token 认证

```bash
# 启动带 Token 认证
nats-server --auth mytoken
```

### TLS 加密

```bash
# 启动带 TLS
nats-server --tlsverify --tlscert "server.crt" --tlskey "server.key"
```

### 在 Go 客户端使用 TLS

```go
// 加载客户端证书
cert, err := tls.LoadX509KeyPair("client.crt", "client.key")
if err != nil {
    log.Fatal(err)
}

config := &tls.Config{
    Certificates: []tls.Certificate{cert},
    ServerName:   "localhost",
}

nc, err := nats.Connect(
    "tls://localhost:4222",
    nats.Secure(config),
)
```

---

## 1.9 最佳实践

### 1. 连接管理

```go
// 推荐：使用连接池或单例
type NATSClient struct {
    nc *nats.Conn
    js jetstream.JetStream
}

var client *NATSClient
var once sync.Once

func GetNATSClient() (*NATSClient, error) {
    var err error
    once.Do(func() {
        nc, e := nats.Connect(nats.DefaultURL)
        if e != nil {
            err = e
            return
        }
        js, e := jetstream.New(nc)
        if e != nil {
            err = e
            return
        }
        client = &NATSClient{nc: nc, js: js}
    })
    return client, err
}
```

### 2. 优雅关闭

```go
func main() {
    nc, _ := nats.Connect(nats.DefaultURL)
    defer nc.Close()

    // 等待信号
    sigCh := make(chan os.Signal, 1)
    signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
    
    <-sigCh
    
    // 优雅关闭：等待正在处理的消息
    nc.Drain()
}
```

### 3. 主题命名规范

```
# 推荐的主题命名
orders.created        # 事件
orders.updated        # 事件
orders.cancelled      # 事件

orders.get            # 请求
orders.create         # 请求
orders.update         # 请求

notifications.email   # 通知
notifications.sms     # 通知
notifications.push    # 通知
```

---

## 1.10 相关资源

- [NATS 官方文档](https://docs.nats.io/)
- [NATS Go 客户端](https://github.com/nats-io/nats.go)
- [NATS JetStream](https://docs.nats.io/using-nats/jetstream)
- [NATS 示例](https://github.com/nats-io/nats.go/tree/main/examples)
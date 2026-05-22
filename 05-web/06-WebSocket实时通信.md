# 5.6 WebSocket 实时通信

## WebSocket 协议基础

### 什么是 WebSocket

WebSocket 是一种在单个 TCP 连接上进行全双工通信的协议，适合实时通信场景。

```
HTTP vs WebSocket:

HTTP:
  客户端             服务器
    | --请求-->        |
    | <--响应--        |
    | --请求-->        |
    | <--响应--        |
  (每次请求都需要建立连接)

WebSocket:
  客户端             服务器
    | --握手 (HTTP)-->  |
    | <--升级协议--     |
    | <====TCP 连接====> |
    |  (双向通信)       |
    | <--推送消息--     |
    | --发送消息-->     |
  (连接保持，双向通信)
```

### WebSocket 握手过程

```
1. 客户端发送 HTTP 请求 (Upgrade)
   GET /ws HTTP/1.1
   Host: example.com
   Upgrade: websocket
   Connection: Upgrade
   Sec-WebSocket-Key: dGhlIHNhbXBsZSBub25jZQ==
   Sec-WebSocket-Version: 13

2. 服务器响应 (101 Switching Protocols)
   HTTP/1.1 101 Switching Protocols
   Upgrade: websocket
   Connection: Upgrade
   Sec-WebSocket-Accept: s3pPLMBiTxaQ9kYGzzhZRbK+xOo=
```

---

## gorilla/websocket 使用

### 安装

```bash
go get github.com/gorilla/websocket
```

### 服务端示例

```go
package main

import (
    "fmt"
    "log"
    "net/http"

    "github.com/gorilla/websocket"
)

// 配置 Upgrader
var upgrader = websocket.Upgrader{
    ReadBufferSize:  1024,
    WriteBufferSize: 1024,
    // 允许所有来源 (生产环境应该限制)
    CheckOrigin: func(r *http.Request) bool {
        return true
    },
}

// WebSocket 连接
type Client struct {
    conn *websocket.Conn
    send chan []byte
}

func (c *Client) readPump() {
    defer c.conn.Close()

    for {
        _, message, err := c.conn.ReadMessage()
        if err != nil {
            if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
                log.Printf("error: %v", err)
            }
            break
        }
        // 处理收到的消息
        log.Printf("收到：%s", message)
    }
}

func (c *Client) writePump() {
    defer c.conn.Close()

    for {
        select {
        case message, ok := <-c.send:
            if !ok {
                // 关闭连接
                c.conn.WriteMessage(websocket.CloseMessage, []byte{})
                return
            }

            w, err := c.conn.NextWriter(websocket.TextMessage)
            if err != nil {
                return
            }
            w.Write(message)

            if err := w.Close(); err != nil {
                return
            }
        }
    }
}

// WebSocket 处理函数
func wsHandler(w http.ResponseWriter, r *http.Request) {
    // 升级为 WebSocket 连接
    conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        log.Println(err)
        return
    }

    client := &Client{
        conn: conn,
        send: make(chan []byte, 256),
    }

    go client.writePump()
    go client.readPump()
}

func main() {
    http.HandleFunc("/ws", wsHandler)
    fmt.Println("WebSocket 服务器启动：ws://localhost:8080/ws")
    log.Fatal(http.ListenAndServe(":8080", nil))
}
```

### 客户端示例

```go
package main

import (
    "fmt"
    "log"
    "time"

    "github.com/gorilla/websocket"
)

func main() {
    // 连接 WebSocket 服务器
    conn, _, err := websocket.DefaultDialer.Dial("ws://localhost:8080/ws", nil)
    if err != nil {
        log.Fatal("连接失败:", err)
    }
    defer conn.Close()

    // 发送消息
    go func() {
        for {
            err := conn.WriteMessage(websocket.TextMessage, []byte("Hello Server!"))
            if err != nil {
                log.Println("发送失败:", err)
                return
            }
            time.Sleep(time.Second)
        }
    }()

    // 接收消息
    for {
        _, message, err := conn.ReadMessage()
        if err != nil {
            log.Println("接收失败:", err)
            return
        }
        fmt.Println("收到:", string(message))
    }
}
```

---

## 聊天室实现

### 完整的聊天室示例

```go
package main

import (
    "fmt"
    "log"
    "net/http"
    "sync"

    "github.com/gorilla/websocket"
)

// 消息
type Message struct {
    Type string `json:"type"`  // join, chat, leave
    User string `json:"user"`
    Data string `json:"data"`
}

// 客户端
type Client struct {
    hub  *Hub
    conn *websocket.Conn
    send chan []byte
}

// 中心 (Hub) 管理所有客户端
type Hub struct {
    mu         sync.RWMutex
    clients    map[*Client]bool
    broadcast  chan []byte
    register   chan *Client
    unregister chan *Client
}

func NewHub() *Hub {
    return &Hub{
        clients:    make(map[*Client]bool),
        broadcast:  make(chan []byte),
        register:   make(chan *Client),
        unregister: make(chan *Client),
    }
}

func (h *Hub) run() {
    for {
        select {
        case client := <-h.register:
            h.mu.Lock()
            h.clients[client] = true
            h.mu.Unlock()
            log.Printf("客户端加入，总数：%d", len(h.clients))

        case client := <-h.unregister:
            h.mu.Lock()
            if _, ok := h.clients[client]; ok {
                delete(h.clients, client)
                close(client.send)
            }
            h.mu.Unlock()
            log.Printf("客户端离开，总数：%d", len(h.clients))

        case message := <-h.broadcast:
            h.mu.RLock()
            for client := range h.clients {
                select {
                case client.send <- message:
                default:
                    close(client.send)
                    delete(h.clients, client)
                }
            }
            h.mu.RUnlock()
        }
    }
}

var hub = NewHub()

func (c *Client) readPump() {
    defer func() {
        c.hub.unregister <- c
        c.conn.Close()
    }()

    for {
        _, message, err := c.conn.ReadMessage()
        if err != nil {
            break
        }
        // 广播消息给所有客户端
        c.hub.broadcast <- message
    }
}

func (c *Client) writePump() {
    defer func() {
        c.conn.Close()
    }()

    for {
        select {
        case message, ok := <-c.send:
            if !ok {
                c.conn.WriteMessage(websocket.CloseMessage, []byte{})
                return
            }
            c.conn.WriteMessage(websocket.TextMessage, message)
        }
    }
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
    conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        log.Println(err)
        return
    }

    client := &Client{
        hub:  hub,
        conn: conn,
        send: make(chan []byte, 256),
    }

    client.hub.register <- client

    go client.writePump()
    go client.readPump()
}

func main() {
    go hub.run()

    http.HandleFunc("/ws", wsHandler)
    fmt.Println("聊天室服务器启动：ws://localhost:8080/ws")
    log.Fatal(http.ListenAndServe(":8080", nil))
}
```

### 前端 HTML 示例

```html
<!DOCTYPE html>
<html>
<head>
    <title>WebSocket 聊天室</title>
</head>
<body>
    <div id="messages"></div>
    <input id="input" placeholder="输入消息...">
    <button onclick="send()">发送</button>

    <script>
        const ws = new WebSocket('ws://localhost:8080/ws');
        const messages = document.getElementById('messages');
        const input = document.getElementById('input');

        ws.onmessage = function(event) {
            const msg = document.createElement('div');
            msg.textContent = event.data;
            messages.appendChild(msg);
        };

        function send() {
            ws.send(input.value);
            input.value = '';
        }

        input.addEventListener('keypress', function(e) {
            if (e.key === 'Enter') send();
        });
    </script>
</body>
</html>
```

---

## WebSocket 心跳机制

### 心跳检测实现

```go
package main

import (
    "log"
    "net/http"
    "time"

    "github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
    CheckOrigin: func(r *http.Request) bool { return true },
}

const (
    // 心跳间隔
    pingPeriod = 30 * time.Second
    // 写超时
    writeWait = 10 * time.Second
    // 读取超时 (pong 等待时间)
    pongWait = 60 * time.Second
)

func wsHandler(w http.ResponseWriter, r *http.Request) {
    conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        log.Println(err)
        return
    }

    // 设置读取超时
    conn.SetReadLimit(65536)
    conn.SetReadDeadline(time.Now().Add(pongWait))

    // 设置 Pong 处理器
    conn.SetPongHandler(func(string) error {
        conn.SetReadDeadline(time.Now().Add(pongWait))
        return nil
    })

    // 发送 Ping
    go func() {
        ticker := time.NewTicker(pingPeriod)
        defer ticker.Stop()

        for range ticker.C {
            conn.SetWriteDeadline(time.Now().Add(writeWait))
            if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
                log.Println("发送 Ping 失败:", err)
                return
            }
        }
    }()

    // 读取消息
    for {
        _, _, err := conn.ReadMessage()
        if err != nil {
            break
        }
    }

    conn.Close()
}

func main() {
    http.HandleFunc("/ws", wsHandler)
    log.Fatal(http.ListenAndServe(":8080", nil))
}
```

---

## 连接管理

### 连接池实现

```go
package main

import (
    "sync"
    "time"

    "github.com/gorilla/websocket"
)

type Connection struct {
    Conn      *websocket.Conn
    CreatedAt time.Time
    LastActive time.Time
    Metadata   map[string]interface{}
}

type ConnectionPool struct {
    mu          sync.RWMutex
    connections map[string]*Connection
}

func NewConnectionPool() *ConnectionPool {
    return &ConnectionPool{
        connections: make(map[string]*Connection),
    }
}

func (p *ConnectionPool) Add(id string, conn *websocket.Conn) {
    p.mu.Lock()
    defer p.mu.Unlock()

    p.connections[id] = &Connection{
        Conn:       conn,
        CreatedAt:  time.Now(),
        LastActive: time.Now(),
        Metadata:   make(map[string]interface{}),
    }
}

func (p *ConnectionPool) Get(id string) (*Connection, bool) {
    p.mu.RLock()
    defer p.mu.RUnlock()

    conn, ok := p.connections[id]
    return conn, ok
}

func (p *ConnectionPool) Remove(id string) {
    p.mu.Lock()
    defer p.mu.Unlock()

    if conn, ok := p.connections[id]; ok {
        conn.Conn.Close()
        delete(p.connections, id)
    }
}

func (p *ConnectionPool) Broadcast(message []byte) {
    p.mu.RLock()
    defer p.mu.RUnlock()

    for _, conn := range p.connections {
        conn.Conn.WriteMessage(websocket.TextMessage, message)
    }
}

func (p *ConnectionPool) Count() int {
    p.mu.RLock()
    defer p.mu.RUnlock()
    return len(p.connections)
}

// 清理不活跃的连接
func (p *ConnectionPool) Cleanup(timeout time.Duration) {
    p.mu.Lock()
    defer p.mu.Unlock()

    now := time.Now()
    for id, conn := range p.connections {
        if now.Sub(conn.LastActive) > timeout {
            conn.Conn.Close()
            delete(p.connections, id)
        }
    }
}
```

---

## Redis Pub/Sub 集成

### 水平扩展方案

```go
package main

import (
    "encoding/json"
    "log"

    "github.com/go-redis/redis/v8"
    "github.com/gorilla/websocket"
    "golang.org/x/context"
)

type RedisHub struct {
    redisClient *redis.Client
    ctx         context.Context
    clients     map[*Client]bool
    mu          sync.RWMutex
}

func NewRedisHub(addr string) *RedisHub {
    client := redis.NewClient(&redis.Options{
        Addr: addr,
    })

    return &RedisHub{
        redisClient: client,
        ctx:         context.Background(),
        clients:     make(map[*Client]bool),
    }
}

func (h *RedisHub) Subscribe(channel string) {
    pubsub := h.redisClient.Subscribe(h.ctx, channel)

    _, err := pubsub.Receive(h.ctx)
    if err != nil {
        log.Println("订阅失败:", err)
        return
    }

    ch := pubsub.Channel()
    for msg := range ch {
        h.Broadcast([]byte(msg.Payload))
    }
}

func (h *RedisHub) Publish(channel string, message []byte) error {
    return h.redisClient.Publish(h.ctx, channel, string(message)).Err()
}

func (h *RedisHub) Broadcast(message []byte) {
    h.mu.RLock()
    defer h.mu.RUnlock()

    for client := range h.clients {
        select {
        case client.send <- message:
        default:
            h.RemoveClient(client)
        }
    }
}

func (h *RedisHub) AddClient(client *Client) {
    h.mu.Lock()
    defer h.mu.Unlock()
    h.clients[client] = true
}

func (h *RedisHub) RemoveClient(client *Client) {
    h.mu.Lock()
    defer h.mu.Unlock()
    delete(h.clients, client)
    close(client.send)
}
```

---

## WebSocket 检查清单

```
[ ] 配置 CheckOrigin 函数限制来源
[ ] 实现心跳机制检测连接状态
[ ] 设置合理的读写超时
[ ] 使用中心 (Hub) 管理所有连接
[ ] 处理连接关闭情况
[ ] 广播时处理慢客户端
[ ] 考虑使用 Redis Pub/Sub 扩展
[ ] 添加连接认证
[ ] 限制消息大小
[ ] 使用 TLS (wss://)
[ ] 监控连接数
[ ] 优雅关闭连接

安全注意事项:
1. 验证 Origin 防止 CSRF 攻击
2. 实现认证机制 (JWT、Session)
3. 限制单个 IP 的连接数
4. 验证消息内容和格式
5. 使用 wss:// 加密通信
```

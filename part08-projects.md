# 第八部分：实战项目

## 8.1 入门级项目

### 命令行工具 (CLI)

```go
// cmd/todo/main.go
package main

import (
    "bufio"
    "fmt"
    "os"
    "strconv"
)

// Todo 任务结构
type Todo struct {
    ID      int
    Title   string
    Done    bool
}

// TodoList 任务列表
type TodoList struct {
    items []Todo
    nextID int
}

func NewTodoList() *TodoList {
    return &TodoList{items: make([]Todo, 0), nextID: 1}
}

func (t *TodoList) Add(title string) {
    t.items = append(t.items, Todo{
        ID:    t.nextID,
        Title: title,
        Done:  false,
    })
    t.nextID++
    fmt.Println("✓ 添加成功")
}

func (t *TodoList) List() {
    if len(t.items) == 0 {
        fmt.Println("暂无任务")
        return
    }
    for _, item := range t.items {
        status := " "
        if item.Done {
            status = "✓"
        }
        fmt.Printf("[%s] %d. %s\n", status, item.ID, item.Title)
    }
}

func (t *TodoList) Done(id int) bool {
    for i, item := range t.items {
        if item.ID == id {
            t.items[i].Done = true
            return true
        }
    }
    return false
}

func (t *TodoList) Remove(id int) bool {
    for i, item := range t.items {
        if item.ID == id {
            t.items = append(t.items[:i], t.items[i+1:]...)
            return true
        }
    }
    return false
}

func printUsage() {
    fmt.Println("用法：todo <命令> [参数]")
    fmt.Println("命令:")
    fmt.Println("  add <任务>    添加任务")
    fmt.Println("  list          列出任务")
    fmt.Println("  done <ID>     完成任务")
    fmt.Println("  remove <ID>   删除任务")
}

func main() {
    if len(os.Args) < 2 {
        printUsage()
        os.Exit(1)
    }

    list := NewTodoList()
    command := os.Args[1]

    switch command {
    case "add":
        if len(os.Args) < 3 {
            fmt.Println("请提供任务标题")
            os.Exit(1)
        }
        list.Add(os.Args[2])

    case "list":
        list.List()

    case "done":
        if len(os.Args) < 3 {
            fmt.Println("请提供任务 ID")
            os.Exit(1)
        }
        id, _ := strconv.Atoi(os.Args[2])
        if !list.Done(id) {
            fmt.Println("任务不存在")
        } else {
            fmt.Println("✓ 任务已完成")
        }

    case "remove":
        if len(os.Args) < 3 {
            fmt.Println("请提供任务 ID")
            os.Exit(1)
        }
        id, _ := strconv.Atoi(os.Args[2])
        if !list.Remove(id) {
            fmt.Println("任务不存在")
        } else {
            fmt.Println("✓ 任务已删除")
        }

    default:
        printUsage()
        os.Exit(1)
    }
}
```

### Todo List API

```go
// cmd/api/main.go
package main

import (
    "net/http"
    "sync"
    "github.com/gin-gonic/gin"
)

type Todo struct {
    ID    int    `json:"id"`
    Title string `json:"title"`
    Done  bool   `json:"done"`
}

type TodoStore struct {
    mu     sync.RWMutex
    todos  map[int]Todo
    nextID int
}

func NewTodoStore() *TodoStore {
    return &TodoStore{
        todos:  make(map[int]Todo),
        nextID: 1,
    }
}

func (s *TodoStore) Create(title string) Todo {
    s.mu.Lock()
    defer s.mu.Unlock()

    todo := Todo{ID: s.nextID, Title: title, Done: false}
    s.todos[s.nextID] = todo
    s.nextID++
    return todo
}

func (s *TodoStore) GetAll() []Todo {
    s.mu.RLock()
    defer s.mu.RUnlock()

    todos := make([]Todo, 0, len(s.todos))
    for _, t := range s.todos {
        todos = append todos, t)
    }
    return todos
}

func (s *TodoStore) GetByID(id int) (Todo, bool) {
    s.mu.RLock()
    defer s.mu.RUnlock()
    todo, ok := s.todos[id]
    return todo, ok
}

func (s *TodoStore) Update(id int, done bool) (Todo, bool) {
    s.mu.Lock()
    defer s.mu.Unlock()

    todo, ok := s.todos[id]
    if !ok {
        return Todo{}, false
    }
    todo.Done = done
    s.todos[id] = todo
    return todo, true
}

func (s *TodoStore) Delete(id int) bool {
    s.mu.Lock()
    defer s.mu.Unlock()

    if _, ok := s.todos[id]; !ok {
        return false
    }
    delete(s.todos, id)
    return true
}

var store = NewTodoStore()

func main() {
    r := gin.Default()

    // GET /todos - 获取所有任务
    r.GET("/todos", func(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{"data": store.GetAll()})
    })

    // POST /todos - 创建任务
    r.POST("/todos", func(c *gin.Context) {
        var req struct {
            Title string `json:"title" binding:"required"`
        }
        if err := c.ShouldBindJSON(&req); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
            return
        }
        todo := store.Create(req.Title)
        c.JSON(http.StatusCreated, gin.H{"data": todo})
    })

    // GET /todos/:id - 获取单个任务
    r.GET("/todos/:id", func(c *gin.Context) {
        id := mustParseID(c.Param("id"))
        todo, ok := store.GetByID(id)
        if !ok {
            c.JSON(http.StatusNotFound, gin.H{"error": "任务不存在"})
            return
        }
        c.JSON(http.StatusOK, gin.H{"data": todo})
    })

    // PATCH /todos/:id - 更新任务
    r.PATCH("/todos/:id", func(c *gin.Context) {
        id := mustParseID(c.Param("id"))
        var req struct {
            Done bool `json:"done"`
        }
        if err := c.ShouldBindJSON(&req); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
            return
        }
        todo, ok := store.Update(id, req.Done)
        if !ok {
            c.JSON(http.StatusNotFound, gin.H{"error": "任务不存在"})
            return
        }
        c.JSON(http.StatusOK, gin.H{"data": todo})
    })

    // DELETE /todos/:id - 删除任务
    r.DELETE("/todos/:id", func(c *gin.Context) {
        id := mustParseID(c.Param("id"))
        if !store.Delete(id) {
            c.JSON(http.StatusNotFound, gin.H{"error": "任务不存在"})
            return
        }
        c.JSON(http.StatusNoContent, nil)
    })

    r.Run(":8080")
}

func mustParseID(s string) int {
    id, _ := strconv.Atoi(s)
    return id
}
```

### 博客系统后端

```go
// 项目结构
/*
blog/
├── cmd/
│   └── api/
│       └── main.go
├── internal/
│   ├── handler/
│   │   └── post.go
│   ├── service/
│   │   └── post.go
│   ├── repository/
│   │   └── post.go
│   └── models/
│       └── post.go
└── go.mod
*/

// internal/models/post.go
package models

import "time"

type Post struct {
    ID        int       `json:"id"`
    Title     string    `json:"title"`
    Content   string    `json:"content"`
    AuthorID  int       `json:"author_id"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}

// internal/repository/post.go
package repository

import (
    "database/sql"
    "github.com/blog/internal/models"
)

type PostRepository struct {
    db *sql.DB
}

func NewPostRepository(db *sql.DB) *PostRepository {
    return &PostRepository{db: db}
}

func (r *PostRepository) Create(post *models.Post) error {
    query := `INSERT INTO posts (title, content, author_id) VALUES (?, ?, ?)`
    result, err := r.db.Exec(query, post.Title, post.Content, post.AuthorID)
    if err != nil {
        return err
    }
    id, _ := result.LastInsertId()
    post.ID = int(id)
    return nil
}

func (r *PostRepository) GetByID(id int) (*models.Post, error) {
    query := `SELECT id, title, content, author_id, created_at, updated_at FROM posts WHERE id = ?`
    var post models.Post
    err := r.db.QueryRow(query, id).Scan(
        &post.ID, &post.Title, &post.Content, &post.AuthorID, &post.CreatedAt, &post.UpdatedAt,
    )
    if err != nil {
        return nil, err
    }
    return &post, nil
}

func (r *PostRepository) List(page, pageSize int) ([]*models.Post, error) {
    query := `SELECT id, title, content, author_id, created_at, updated_at FROM posts ORDER BY created_at DESC LIMIT ? OFFSET ?`
    rows, err := r.db.Query(query, pageSize, (page-1)*pageSize)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var posts []*models.Post
    for rows.Next() {
        var post models.Post
        rows.Scan(&post.ID, &post.Title, &post.Content, &post.AuthorID, &post.CreatedAt, &post.UpdatedAt)
        posts = append(posts, &post)
    }
    return posts, rows.Err()
}

// internal/handler/post.go
package handler

import (
    "net/http"
    "github.com/gin-gonic/gin"
    "github.com/blog/internal/models"
    "github.com/blog/internal/repository"
)

type PostHandler struct {
    repo *repository.PostRepository
}

func NewPostHandler(repo *repository.PostRepository) *PostHandler {
    return &PostHandler{repo: repo}
}

func (h *PostHandler) Create(c *gin.Context) {
    var req struct {
        Title   string `json:"title" binding:"required"`
        Content string `json:"content" binding:"required"`
    }
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    post := &models.Post{
        Title:    req.Title,
        Content:  req.Content,
        AuthorID: 1, // 从认证上下文获取
    }

    if err := h.repo.Create(post); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusCreated, gin.H{"data": post})
}

func (h *PostHandler) GetByID(c *gin.Context) {
    id := mustParseID(c.Param("id"))
    post, err := h.repo.GetByID(id)
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "文章不存在"})
        return
    }
    c.JSON(http.StatusOK, gin.H{"data": post})
}

func (h *PostHandler) List(c *gin.Context) {
    page := parseInt(c.DefaultQuery("page", "1"), 1)
    pageSize := parseInt(c.DefaultQuery("page_size", "10"), 10)

    posts, err := h.repo.List(page, pageSize)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{"data": posts})
}
```

---

## 8.2 进阶级项目

### 分布式任务队列

```go
// 项目结构
/*
taskqueue/
├── cmd/
│   ├── server/
│   │   └── main.go       # 服务器
│   └── worker/
│       └── main.go       # 工作节点
├── internal/
│   ├── queue/
│   │   └── queue.go      # 队列实现
│   ├── task/
│   │   └── task.go       # 任务定义
│   └── storage/
│       └── redis.go      # Redis 存储
└── go.mod
*/

// internal/task/task.go
package task

import (
    "encoding/json"
    "time"
    "github.com/google/uuid"
)

type TaskStatus int

const (
    TaskPending TaskStatus = iota
    TaskRunning
    TaskCompleted
    TaskFailed
)

type Task struct {
    ID        string                 `json:"id"`
    Type      string                 `json:"type"`
    Payload   map[string]interface{} `json:"payload"`
    Status    TaskStatus             `json:"status"`
    Priority  int                    `json:"priority"`
    Retry     int                    `json:"retry"`
    MaxRetry  int                    `json:"max_retry"`
    Error     string                 `json:"error,omitempty"`
    CreatedAt time.Time              `json:"created_at"`
    StartedAt *time.Time             `json:"started_at,omitempty"`
    EndedAt   *time.Time             `json:"ended_at,omitempty"`
}

func NewTask(taskType string, payload map[string]interface{}) *Task {
    return &Task{
        ID:        uuid.New().String(),
        Type:      taskType,
        Payload:   payload,
        Status:    TaskPending,
        Priority:  0,
        Retry:     0,
        MaxRetry:  3,
        CreatedAt: time.Now(),
    }
}

func (t *Task) ToJSON() ([]byte, error) {
    return json.Marshal(t)
}

func TaskFromJSON(data []byte) (*Task, error) {
    var task Task
    err := json.Unmarshal(data, &task)
    return &task, err
}

// internal/queue/queue.go
package queue

import (
    "context"
    "github.com/redis/go-redis/v9"
    "github.com/taskqueue/internal/task"
    "sort"
)

type Queue struct {
    client *redis.Client
    name   string
}

func NewQueue(client *redis.Client, name string) *Queue {
    return &Queue{client: client, name: name}
}

// Enqueue 添加任务到队列
func (q *Queue) Enqueue(ctx context.Context, t *task.Task) error {
    data, err := t.ToJSON()
    if err != nil {
        return err
    }
    return q.client.ZAdd(ctx, q.name, redis.Z{
        Score:  float64(-t.Priority),  // 优先级高的在前
        Member: string(data),
    }).Err()
}

// Dequeue 从队列获取任务
func (q *Queue) Dequeue(ctx context.Context) (*task.Task, error) {
    // 使用 ZPOPMIN 获取优先级最高的任务
    result := q.client.ZPopMin(ctx, q.name)
    vals, err := result.Result()
    if err != nil {
        return nil, err
    }
    if len(vals) == 0 {
        return nil, nil
    }

    return task.TaskFromJSON([]byte(vals[0].Member))
}

// Requeue 重新入队 (失败重试)
func (q *Queue) Requeue(ctx context.Context, t *task.Task) error {
    t.Retry++
    if t.Retry > t.MaxRetry {
        t.Status = task.TaskFailed
        return nil
    }
    t.Status = task.TaskPending
    return q.Enqueue(ctx, t)
}

// Size 队列大小
func (q *Queue) Size(ctx context.Context) (int64, error) {
    return q.client.ZCard(ctx, q.name).Result()
}

// cmd/server/main.go
package main

import (
    "context"
    "net/http"
    "github.com/gin-gonic/gin"
    "github.com/redis/go-redis/v9"
    "github.com/taskqueue/internal/queue"
    "github.com/taskqueue/internal/task"
)

var q *queue.Queue

func main() {
    client := redis.NewClient(&redis.Options{
        Addr: "localhost:6379",
    })
    q = queue.NewQueue(client, "tasks")

    r := gin.Default()

    // POST /tasks - 提交任务
    r.POST("/tasks", func(c *gin.Context) {
        var req struct {
            Type    string                 `json:"type" binding:"required"`
            Payload map[string]interface{} `json:"payload"`
            Priority int                   `json:"priority"`
        }
        if err := c.ShouldBindJSON(&req); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
            return
        }

        t := task.NewTask(req.Type, req.Payload)
        t.Priority = req.Priority

        if err := q.Enqueue(c.Request.Context(), t); err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }

        c.JSON(http.StatusCreated, gin.H{"task_id": t.ID})
    })

    // GET /tasks/stats - 队列统计
    r.GET("/tasks/stats", func(c *gin.Context) {
        size, _ := q.Size(c.Request.Context())
        c.JSON(http.StatusOK, gin.H{"pending": size})
    })

    r.Run(":8080")
}

// cmd/worker/main.go
package main

import (
    "context"
    "log"
    "time"
    "github.com/redis/go-redis/v9"
    "github.com/taskqueue/internal/queue"
    "github.com/taskqueue/internal/task"
)

var taskHandlers = map[string]func(*task.Task) error{
    "email": handleEmail,
    "image": handleImage,
}

func main() {
    client := redis.NewClient(&redis.Options{
        Addr: "localhost:6379",
    })
    q := queue.NewQueue(client, "tasks")

    log.Println("Worker started")

    for {
        task, err := q.Dequeue(context.Background())
        if err != nil {
            log.Printf("Dequeue error: %v", err)
            time.Sleep(time.Second)
            continue
        }

        if task == nil {
            time.Sleep(100 * time.Millisecond)
            continue
        }

        if err := processTask(task); err != nil {
            task.Error = err.Error()
            q.Requeue(context.Background(), task)
        }
    }
}

func processTask(t *task.Task) error {
    handler, ok := taskHandlers[t.Type]
    if !ok {
        return fmt.Errorf("unknown task type: %s", t.Type)
    }
    return handler(t)
}

func handleEmail(t *task.Task) error {
    // 发送邮件逻辑
    log.Printf("Sending email: %+v", t.Payload)
    return nil
}

func handleImage(t *task.Task) error {
    // 处理图片逻辑
    log.Printf("Processing image: %+v", t.Payload)
    return nil
}
```

### 短链服务

```go
// internal/models/url.go
package models

import "time"

type ShortURL struct {
    ID        int       `json:"id"`
    ShortCode string    `json:"short_code"`
    LongURL   string    `json:"long_url"`
    Clicks    int       `json:"clicks"`
    ExpiresAt *time.Time `json:"expires_at,omitempty"`
    CreatedAt time.Time `json:"created_at"`
}

// internal/handler/shorturl.go
package handler

import (
    "math/rand"
    "net/http"
    "time"
    "github.com/gin-gonic/gin"
)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func generateShortCode(length int) string {
    rand.Seed(time.Now().UnixNano())
    b := make([]byte, length)
    for i := range b {
        b[i] = charset[rand.Intn(len(charset))]
    }
    return string(b)
}

type ShortURLHandler struct {
    store ShortURLStore
}

func NewShortURLHandler(store ShortURLStore) *ShortURLHandler {
    return &ShortURLHandler{store: store}
}

func (h *ShortURLHandler) Create(c *gin.Context) {
    var req struct {
        URL     string `json:"url" binding:"required,url"`
        Expires int    `json:"expires"` // 过期时间 (秒)
    }
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // 生成短码
    shortCode := generateShortCode(6)

    var expiresAt *time.Time
    if req.Expires > 0 {
        t := time.Now().Add(time.Duration(req.Expires) * time.Second)
        expiresAt = &t
    }

    url := &models.ShortURL{
        ShortCode: shortCode,
        LongURL:   req.URL,
        Clicks:    0,
        ExpiresAt: expiresAt,
        CreatedAt: time.Now(),
    }

    if err := h.store.Create(url); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusCreated, gin.H{
        "short_code": shortCode,
        "short_url":  "http://localhost:8080/" + shortCode,
        "long_url":   req.URL,
    })
}

func (h *ShortURLHandler) Redirect(c *gin.Context) {
    shortCode := c.Param("code")

    url, err := h.store.GetByShortCode(shortCode)
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "短链接不存在"})
        return
    }

    // 检查过期
    if url.ExpiresAt != nil && time.Now().After(*url.ExpiresAt) {
        c.JSON(http.StatusGone, gin.H{"error": "短链接已过期"})
        return
    }

    // 增加点击
    h.store.IncrementClicks(url.ID)

    c.Redirect(http.StatusMovedPermanently, url.LongURL)
}

// 使用：
// POST /api/short {"url": "https://example.com/very/long/url"}
// GET /abc123  -> 重定向到原 URL
```

---

## 8.3 高级项目

### 微服务电商平台

```go
/*
ecommerce/
├── api-gateway/           # API 网关
├── user-service/          # 用户服务
├── product-service/       # 商品服务
├── order-service/         # 订单服务
├── payment-service/       # 支付服务
├── shared/                # 共享代码
│   ├── models/
│   ├── middleware/
│   └── grpc/
└── deployments/
    ├── docker/
    └── k8s/

技术栈:
- gRPC: 服务间通信
- JWT: 认证
- Prometheus: 监控
- Jaeger: 链路追踪
- Kubernetes: 编排
*/
```

### 实时数据推送系统

```go
// WebSocket 实时推送
package main

import (
    "log"
    "net/http"
    "sync"
    "github.com/gorilla/websocket"
    "github.com/gin-gonic/gin"
)

type Client struct {
    conn *websocket.Conn
    send chan []byte
}

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

func (h *Hub) Run() {
    for {
        select {
        case client := <-h.register:
            h.mu.Lock()
            h.clients[client] = true
            h.mu.Unlock()

        case client := <-h.unregister:
            h.mu.Lock()
            if _, ok := h.clients[client]; ok {
                delete(h.clients, client)
                close(client.send)
            }
            h.mu.Unlock()

        case message := <-h.broadcast:
            h.mu.RLock()
            for client := range h.clients {
                select {
                case client.send <- message:
                default:
                    go func(c *Client) {
                        h.mu.Lock()
                        delete(h.clients, c)
                        h.mu.Unlock()
                        close(c.send)
                    }(client)
                }
            }
            h.mu.RUnlock()
        }
    }
}

func (h *Hub) Broadcast(message []byte) {
    h.broadcast <- message
}

var upgrader = websocket.Upgrader{
    ReadBufferSize:  1024,
    WriteBufferSize: 1024,
    CheckOrigin: func(r *http.Request) bool {
        return true
    },
}

func main() {
    hub := NewHub()
    go hub.Run()

    r := gin.Default()

    // WebSocket 连接
    r.GET("/ws", func(c *gin.Context) {
        conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
        if err != nil {
            log.Println(err)
            return
        }

        client := &Client{
            conn: conn,
            send: make(chan []byte, 256),
        }

        hub.register <- client

        go func() {
            defer func() {
                hub.unregister <- client
                conn.Close()
            }()

            for {
                _, message, err := conn.ReadMessage()
                if err != nil {
                    break
                }
                // 处理客户端消息
                log.Printf("Received: %s", message)
            }
        }()

        go func() {
            for message := range client.send {
                conn.WriteMessage(websocket.TextMessage, message)
            }
        }()
    })

    // 广播 API
    r.POST("/broadcast", func(c *gin.Context) {
        var req struct {
            Message string `json:"message"`
        }
        if err := c.ShouldBindJSON(&req); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
            return
        }
        hub.Broadcast([]byte(req.Message))
        c.JSON(http.StatusOK, gin.H{"status": "ok"})
    })

    r.Run(":8080")
}
```

---

## 8.4 开源贡献

### 阅读优秀开源项目

推荐项目:

1. **Gin** (Web 框架)
   - https://github.com/gin-gonic/gin
   - 学习：中间件、路由、绑定验证

2. **Viper** (配置管理)
   - https://github.com/spf13/viper
   - 学习：配置加载、热更新

3. **Cobra** (CLI 框架)
   - https://github.com/spf13/cobra
   - 学习：命令行解析、命令树

4. **GORM** (ORM)
   - https://github.com/go-gorm/gorm
   - 学习：反射、SQL 生成

5. **Etcd** (分布式 KV)
   - https://github.com/etcd-io/etcd
   - 学习：Raft 共识、分布式系统

### 提交 PR 参与开源项目

```bash
# ========== 贡献流程 ==========

# 1. Fork 项目
# 在 GitHub 上点击 Fork

# 2. Clone 到本地
git clone https://github.com/YOUR_USERNAME/project.git
cd project

# 3. 添加 upstream remote
git remote add upstream https://github.com/ORIGINAL_OWNER/project.git

# 4. 创建分支
git checkout -b feature/your-feature

# 5. 开发并提交
# ... 编写代码 ...
git add .
git commit -m "feat: add new feature"

# 6. 同步 upstream 变更
git fetch upstream
git rebase upstream/main

# 7. 推送到自己的仓库
git push origin feature/your-feature

# 8. 创建 Pull Request
# 在 GitHub 上创建 PR
```

### 发布自己的开源库

```bash
# ========== 发布步骤 ==========

# 1. 创建 GitHub 仓库
# - 创建仓库 mylib
# - 添加 LICENSE (推荐 MIT/Apache 2.0)
# - 添加 README.md

# 2. 初始化 Go 模块
cd mylib
go mod init github.com/username/mylib

# 3. 编写代码
/*
// mylib.go
package mylib

// Add 返回两数之和
func Add(a, b int) int {
    return a + b
}
*/

# 4. 编写测试
/*
// mylib_test.go
package mylib

import "testing"

func TestAdd(t *testing.T) {
    if Add(2, 3) != 5 {
        t.Error("Add failed")
    }
}
*/

# 5. 添加文档
// 使用 godoc 格式注释

# 6. 打标签发布
git add .
git commit -m "Initial release"
git tag v1.0.0
git push origin main v1.0.0

# 7. 其他人可以使用
# go get github.com/username/mylib@v1.0.0
```

---

## 完整学习路线总结

| 阶段 | 内容 | 目标 |
|------|------|------|
| 入门 (2-3 周) | 语法基础、核心特性 | 能编写简单 Go 程序 |
| 进阶 (3-4 周) | 并发编程、标准库 | 掌握 Go 核心优势 |
| 实战 (4-6 周) | Web 开发、工程实践 | 能独立开发项目 |
| 提高 (4 周+) | 高级主题 | 理解底层原理 |
| 精通 (持续) | 实战项目、开源贡献 | 成为 Go 专家 |

---

## 第八部分完

至此，Go 语言学习大纲全部内容完成！

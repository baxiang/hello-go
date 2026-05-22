# go-zero 核心概念与基础

go-zero 是一个集成了各种工程实践的微服务框架，包含 API 定义、代码生成、熔断、限流、自动熔断等功能。它以其简洁的 API 设计和强大的工程能力著称。

## 1.1 go-zero 简介

### 什么是 go-zero？

go-zero 是一个稳定、可扩展的微服务框架，由 kevinten10 开发并维护。它提供了完整的微服务开发工具链，包括 API 网关、服务注册、配置管理、熔断、限流、日志、监控等功能。

### 核心特性

| 特性 | 说明 |
|------|------|
| **API 优先** | 通过 YAML/JSON 定义 API，自动生成代码 |
| **代码生成** | 根据 API 定义自动生成 CRUD 代码 |
| **内置熔断** | 基于 Sina/Circle 的熔断器实现 |
| **内置限流** | 支持令牌桶限流算法 |
| **自动缓存** | 支持内存缓存和 Redis 缓存 |
| **链路追踪** | 集成 Prometheus 和 OpenTelemetry |
| **高并发** | 基于 Go 的高并发模型 |

### go-zero 与其他框架对比

| 特性 | go-zero | gin | kratos | echo |
|------|---------|-----|--------|------|
| API 定义 | YAML/JSON | 无 | Protobuf | 无 |
| 代码生成 | 自动生成 | 无 | 插件生成 | 无 |
| 熔断限流 | 内置 | 需集成 | 需集成 | 需集成 |
| 服务注册 | 内置 | 需集成 | 内置 | 需集成 |
| 学习成本 | 低 | 极低 | 中等 | 低 |
| 适用场景 | 微服务 | API 服务 | 微服务 | API 服务 |

### 适用场景

- 微服务架构
- RESTful API 开发
- 高并发 API 服务
- 企业级应用开发

---

## 1.2 安装与配置

### 安装 goctl

```bash
# 使用 Go 安装
go install github.com/zeromicro/go-zero/tools/goctl@latest

# 验证安装
goctl version
```

### 安装 Redis（可选）

go-zero 的限流和缓存功能需要 Redis：

```bash
# macOS
brew install redis

# 启动 Redis
redis-server

# 或使用 Docker
docker run -d -p 6379:6379 redis:7-alpine
```

---

## 1.3 核心概念

### 1.3.1 API 定义

go-zero 使用 YAML 格式定义 API：

```yaml
type (
  # 请求类型定义
  AddRequest {
    A int `json:"a"`
    B int `json:"b"`
  }

  # 响应类型定义
  AddResponse {
    Sum int `json:"sum"`
  }
)

# 服务定义
service calculator-api {
  # 路由定义
  @handler AddHandler
  post /add (AddRequest) returns (AddResponse)
}
```

### 1.3.2 服务类型

go-zero 支持两种服务类型：

1. **API 服务**：HTTP API 网关
2. **RPC 服务**：gRPC 服务

```yaml
# API 服务
service user-api {
  @handler GetUserHandler
  get /user/:id returns (User)
}

# RPC 服务
service user-rpc {
  rpc getUser(GetUserRequest) returns (User)
}
```

### 1.3.3 处理器（Handler）

处理器是业务逻辑的核心：

```go
func (l *AddLogic) Add() (int, error) {
    return l.req.A + l.req.B, nil
}
```

### 1.3.4 逻辑层（Logic）

逻辑层封装业务逻辑：

```go
type AddLogic struct {
    ctx    context.Context
    svcCtx *svc.ServiceContext
    req    types.AddRequest
}
```

### 1.3.5 服务上下文（ServiceContext）

服务上下文包含共享资源：

```go
type ServiceContext struct {
    UserModel model.UserModel
    Redis     *redis.Redis
}
```

---

## 1.4 快速开始

### 1.4.1 创建项目

```bash
# 创建项目目录
mkdir myapp && cd myapp

# 初始化 Go 模块
go mod init myapp

# 添加 go-zero 依赖
go get github.com/zeromicro/go-zero
go get github.com/zeromicro/go-zero/tools/goctl
```

### 1.4.2 定义 API

创建 `user.api` 文件：

```yaml
type (
  # 用户信息
  User {
    Id       int64  `json:"id"`
    Username string `json:"username"`
    Email    string `json:"email"`
    Phone    string `json:"phone"`
  }

  # 创建用户请求
  CreateUserRequest {
    Username string `json:"username"`
    Email    string `json:"email"`
    Phone    string `json:"phone"`
    Password string `json:"password"`
  }

  # 用户列表请求
  ListUserRequest {
    Page     int `json:"page,default=1"`
    PageSize int `json:"page_size,default=10"`
  }

  # 用户列表响应
  ListUserResponse {
    Total int64   `json:"total"`
    Users []User  `json:"users"`
  }
)

# 用户服务
service user-api {
  @handler CreateUserHandler
  post /api/user/create (CreateUserRequest) returns (User)

  @handler GetUserHandler
  get /api/user/:id returns (User)

  @handler ListUserHandler
  get /api/user/list (ListUserRequest) returns (ListUserResponse)

  @handler DeleteUserHandler
  delete /api/user/:id returns (string)
}
```

### 1.4.3 生成代码

```bash
# 生成 API 代码
goctl api go -api user.api -dir . -style gozero

# 生成 RPC 代码
goctl rpc proto -src user.proto -dir .
```

### 1.4.4 完整示例

```go
// main.go
package main

import (
    "flag"
    "fmt"

    "github.com/zeromicro/go-zero/core/conf"
    "github.com/zeromicro/go-zero/rest"
    "github.com/zeromicro/go-zero/rest/httpx"
)

var configFile = flag.String("f", "user-api.yaml", "the config file")

type Config struct {
    rest.RestConf
    Database struct {
        DataSource string
    }
    Redis struct {
        Host string
        Pass string
    }
}

func main() {
    flag.Parse()

    var c Config
    conf.MustLoad(*flag.Parse(), &c)

    server := rest.MustNewServer(c.RestConf)
    defer server.Stop()

    server.AddRoute([]rest.Route{
        {Method: http.MethodGet, Path: "/api/user/:id", Handler: getUserHandler},
        {Method: http.MethodPost, Path: "/api/user/create", Handler: createUserHandler},
    })

    fmt.Println("Server started at", c.Host+":"+c.Port)
    server.Start()
}

// Handler functions
func getUserHandler(w http.ResponseWriter, r *http.Request) {
    id := httpx.GetParam(r, "id")
    // 业务逻辑
    httpx.OkJson(w, map[string]interface{}{
        "id":       id,
        "username": "user1",
        "email":    "user1@example.com",
    })
}

func createUserHandler(w http.ResponseWriter, r *http.Request) {
    var req struct {
        Username string `json:"username"`
        Email    string `json:"email"`
    }
    httpx.Parse(r, &req)
    // 业务逻辑
    httpx.OkJson(w, map[string]interface{}{
        "id":       1,
        "username": req.Username,
        "email":    req.Email,
    })
}
```

---

## 1.5 核心组件

### 1.5.1 配置文件

```yaml
# user-api.yaml
Name: user-api
Host: 0.0.0.0
Port: 8080

Log:
  Mode: console
  Level: info

# 数据库配置
Database:
  Type: mysql
  DataSource: root:root123@tcp(localhost:3306)/user_db?charset=utf8mb4&parseTime=True&loc=Local

# Redis 配置
Redis:
  Host: localhost:6379
  Pass: ""
  Type: node

# 熔断配置
Cbreaker:
  Threshold: 1000

# 限流配置
Shaper:
  - Strategy: token
    Capacity: 1000
    Rate: 1000
```

### 1.5.2 日志配置

```go
// 自定义日志
logx.MustSetup(logx.Config{
    Mode:        "console",  // console, file, volume
    Level:       "info",
    ServiceName: "user-api",
})

// 使用日志
logx.Infof("User created: %d", userID)
logx.Error("Failed to create user: %v", err)
```

### 1.5.3 错误处理

```go
import "github.com/zeromicro/go-zero/core/errorx"

// 错误定义
var (
    ErrUserNotFound   = errorx.New("user not found")
    ErrInvalidParam   = errorx.New("invalid parameter")
    ErrUnauthorized   = errorx.New("unauthorized")
)

// 在 Handler 中使用
if user == nil {
    return nil, errorx.New("user not found")
}

// 返回错误
httpx.Error(w, errorx.New("internal error"))
```

---

## 1.6 数据库操作

### 1.6.1 使用 GORM

```go
import "gorm.io/gorm"

// 定义模型
type User struct {
    gorm.Model
    Username string `json:"username"`
    Email    string `json:"email"`
    Phone    string `json:"phone"`
    Password string `json:"-"`
}

// 创建
func (m *UserModel) Create(ctx context.Context, user *User) error {
    return m.db.WithContext(ctx).Create(user).Error
}

// 查询
func (m *UserModel) FindOne(ctx context.Context, id int64) (*User, error) {
    var user User
    err := m.db.WithContext(ctx).First(&user, id).Error
    if err != nil {
        return nil, err
    }
    return &user, nil
}

// 更新
func (m *UserModel) Update(ctx context.Context, user *User) error {
    return m.db.WithContext(ctx).Save(user).Error
}

// 删除
func (m *UserModel) Delete(ctx context.Context, id int64) error {
    return m.db.WithContext(ctx).Delete(&User{}, id).Error
}

// 列表
func (m *UserModel) FindAll(ctx context.Context, page, pageSize int) ([]*User, int64, error) {
    var users []*User
    var total int64

    query := m.db.WithContext(ctx).Model(&User{})
    query.Count(&total)

    offset := (page - 1) * pageSize
    err := query.Offset(offset).Limit(pageSize).Find(&users).Error

    return users, total, err
}
```

### 1.6.2 使用 go-zero 的 SQLx

```go
import (
    "github.com/zeromicro/go-zero/core/stores/sqlx"
)

// 创建连接
conn := sqlx.NewMysql("root:root123@tcp(localhost:3306)/user_db")

// 执行查询
var result struct {
    Id       int64  `json:"id"`
    Username string `json:"username"`
}

err := conn.QueryRow(&result, "SELECT id, username FROM users WHERE id = ?", id)
```

---

## 1.7 缓存

### 1.7.1 Redis 缓存

```go
import (
    "context"
    "time"

    "github.com/zeromicro/go-zero/core/stores/cache"
    "github.com/zeromicro/go-zero/core/stores/redis"
)

// 创建 Redis 客户端
rds := redis.New("localhost:6379")

// 设置值
err := rds.Set("key", "value")
err := rds.Setex("key", "value", 3600)

// 获取值
val, err := rds.Get("key")

// 删除值
err := rds.Del("key")

// 过期时间
err := rds.Expire("key", time.Hour)
```

### 1.7.2 缓存装饰器

```go
import "github.com/zeromicro/go-zero/core/stores/cache"

// 创建缓存
c, err := cache.NewCache(
    cache.NewRedis(rds),
    cache.WithExpiry(time.Hour),
)

// 使用缓存
var user User
err := c.Take(&user, func(v interface{}) error {
    return db.First(v, id).Error
}, "user:%d", id)
```

---

## 1.8 限流与熔断

### 1.8.1 限流

```go
import "github.com/zeromicro/go-zero/core/limit"

// 令牌桶限流
tokenLimiter := limit.NewTokenLimiter(100, 100, redis.New("localhost:6379"))

// 获取令牌
if !tokenLimiter.Allow() {
    return errors.New("rate limit exceeded")
}
```

### 1.8.2 熔断

```go
import "github.com/zeromicro/go-zero/core/breaker"

// 创建熔断器
br := breaker.GetBreaker("service-name")

// 执行操作
err := br.Do(func() error {
    return callService()
})
```

### 1.8.3 配置文件限流

```yaml
Shaper:
  - Strategy: token
    Capacity: 1000
    Rate: 1000
    Wait: 1000
```

---

## 1.9 中间件

### 1.9.1 全局中间件

```go
// 日志中间件
func loggingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        next.ServeHTTP(w, r)
        logx.Infof("%s %s %v", r.Method, r.URL.Path, time.Since(start))
    })
}

// 认证中间件
func authMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        token := r.Header.Get("Authorization")
        if token == "" {
            http.Error(w, "unauthorized", http.StatusUnauthorized)
            return
        }
        next.ServeHTTP(w, r)
    })
}

// 注册中间件
server.Use(loggingMiddleware)
server.Use(authMiddleware)
```

### 1.9.2 路由中间件

```go
// 单个路由使用中间件
server.AddRoute([]rest.Route{
    {
        Method:  http.MethodGet,
        Path:    "/api/user/:id",
        Handler: getUserHandler,
        Middlewares: []rest.Middleware{
            authMiddleware,
        },
    },
})
```

---

## 1.10 最佳实践

### 1.10.1 项目结构

```
myapp/
├── api/
│   └── user.api           # API 定义
├── cmd/
│   └── user-api/
│       └── main.go        # 入口文件
├── configs/
│   └── user-api.yaml      # 配置文件
├── internal/
│   ├── config/
│   │   └── config.go      # 配置结构
│   ├── handler/
│   │   ├── user.go        # Handler
│   │   └── routes.go      # 路由
│   ├── logic/
│   │   └── user.go        # 业务逻辑
│   ├── model/
│   │   └── user.go        # 数据模型
│   ├── svc/
│   │   └── service.go     # 服务上下文
│   └── types/
│       └── types.go       # 类型定义
├── go.mod
└── go.sum
```

### 1.10.2 错误处理

```go
// 统一错误响应
func Error(w http.ResponseWriter, r *http.Request, err error) {
    var (
        code int
        msg  string
    )

    switch {
    case errors.Is(err, ErrUserNotFound):
        code = 404
        msg = "用户不存在"
    case errors.Is(err, ErrInvalidParam):
        code = 400
        msg = "参数错误"
    default:
        code = 500
        msg = "内部错误"
    }

    httpx.Error(w, errorx.New(msg))
}
```

### 1.10.3 日志规范

```go
// 结构化日志
logx.Infow("create user",
    "user_id", user.ID,
    "username", user.Username,
)

// 错误日志
logx.Errorw("create user failed",
    "error", err.Error(),
    "username", req.Username,
)
```

---

## 1.11 相关资源

- [go-zero 官方文档](https://go-zero.dev/)
- [go-zero GitHub](https://github.com/zeromicro/go-zero)
- [go-zero 示例](https://github.com/zeromicro/zero-examples)
- [goctl 文档](https://go-zero.dev/cn/docs/goctl/goctl)
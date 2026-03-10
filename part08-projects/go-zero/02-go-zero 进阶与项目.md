# go-zero 进阶与项目

本节介绍 go-zero 的进阶特性，包括服务注册、RPC 调用、分布式事务、缓存策略等，并提供完整的项目示例。

## 2.1 服务注册与发现

### 2.1.1 基于 Consul 的服务注册

```go
// config.yaml
Name: user-api
Host: 0.0.0.0
Port: 8080

Consul:
  Host: localhost:8500
  Key: user-api
  Port: 8500
```

```go
// main.go
package main

import (
    "flag"

    "github.com/zeromicro/go-zero/core/discovery"
    "github.com/zeromicro/go-zero/core/conf"
)

var configFile = flag.String("f", "user-api.yaml", "the config file")

type Config struct {
    rest.RestConf
    Consul discovery.ConsulConf `json:",optional"`
}

func main() {
    flag.Parse()

    var c Config
    conf.MustLoad(*flag.Parse(), &c)

    // 启动服务注册
    if c.Consul.Host != "" {
        discovery.RegisterService(c.Consul)
    }

    // 启动服务
    // ...
}
```

### 2.1.2 服务发现

```go
import "github.com/zeromicro/go-zero/core/discovery"

// 发现服务
endpoints, err := discovery.GetNodes(discovery.NewEtcd(
    []string{"localhost:2379"},
), "user-rpc")
```

---

## 2.2 RPC 调用

### 2.2.1 定义 RPC 服务

创建 `user.proto`：

```protobuf
syntax = "proto3";

package user;

option go_package = "./user";

service User {
  rpc GetUser(GetUserRequest) returns (User);
  rpc CreateUser(CreateUserRequest) returns (User);
  rpc DeleteUser(DeleteUserRequest) returns (google.protobuf.Empty);
}

message GetUserRequest {
  int64 id = 1;
}

message CreateUserRequest {
  string username = 1;
  string email = 2;
  string phone = 3;
  string password = 4;
}

message DeleteUserRequest {
  int64 id = 1;
}

message User {
  int64 id = 1;
  string username = 2;
  string email = 3;
  string phone = 4;
  int32 status = 5;
  string created_at = 6;
  string updated_at = 7;
}
```

### 2.2.2 生成 RPC 代码

```bash
# 生成 RPC 代码
goctl rpc proto -src user.proto -dir ./rpc -style gozero
```

### 2.2.3 RPC 客户端

```go
import "github.com/zeromicro/go-zero/zrpc"

type Config struct {
    UserRpc zrpc.RpcClientConf `json:",optional"`
}

// 创建 RPC 客户端
client := zrpc.MustNewClient(c.UserRpc)

// 获取 User 服务
userClient := user.NewUser(client.Conn(), nil)

// 调用 RPC
resp, err := userClient.GetUser(context.Background(), &user.GetUserRequest{
    Id: 1,
})
```

---

## 2.3 分布式事务

### 2.3.1 TCC 模式

```go
import "github.com/zeromicro/go-zero/core/transaction"

// 定义 TCC 事务
type OrderTCC struct{}

// Try 预留资源
func (t *OrderTCC) Try(ctx context.Context, req OrderRequest) error {
    // 1. 预留库存
    err := reserveStock(req.Items)
    if err != nil {
        return err
    }

    // 2. 冻结金额
    err = freezeAmount(req.UserID, req.TotalAmount)
    if err != nil {
        releaseStock(req.Items)
        return err
    }

    return nil
}

// Confirm 确认执行
func (t *OrderTCC) Confirm(ctx context.Context, req OrderRequest) error {
    // 1. 扣减库存
    err := deductStock(req.Items)
    if err != nil {
        return err
    }

    // 2. 扣减金额
    err = chargeAmount(req.UserID, req.TotalAmount)
    if err != nil {
        // 补偿
        restoreStock(req.Items)
        return err
    }

    return nil
}

// Cancel 取消
func (t *OrderTCC) Cancel(ctx context.Context, req OrderRequest) error {
    // 1. 释放库存
    releaseStock(req.Items)

    // 2. 释放金额
    unfreezeAmount(req.UserID, req.TotalAmount)

    return nil
}

// 执行事务
txn := transaction.NewTCC(&OrderTCC{})
err := txn.Execute(ctx, req)
```

### 2.3.2 消息事务

```go
import "github.com/zeromicro/go-zero/core/stores/sqlx"

type OrderMessage struct {
    ID        int64
    Status    string
    Payload   string
}

// 发送消息
func sendMessage(orderID int64, action string) error {
    msg := OrderMessage{
        ID:      orderID,
        Status:  "pending",
        Payload: fmt.Sprintf(`{"order_id":%d,"action":"%s"}`, orderID, action),
    }

    // 保存到消息表
    _, err := db.Exec("INSERT INTO messages (id, status, payload) VALUES (?, ?, ?)",
        msg.ID, msg.Status, msg.Payload)

    return err
}

// 确认消息
func confirmMessage(orderID int64) error {
    _, err := db.Exec("UPDATE messages SET status='confirmed' WHERE id=?", orderID)
    return err
}
```

---

## 2.4 缓存策略

### 2.4.1 多级缓存

```go
import (
    "context"
    "time"

    "github.com/zeromicro/go-zero/core/stores/cache"
    "github.com/zeromicro/go-zero/core/stores/redis"
)

// 本地缓存
localCache, _ := cache.NewCache(
    cache.NewMemory(),
    cache.WithExpiry(time.Minute * 5),
)

// Redis 缓存
redisCache, _ := cache.NewCache(
    cache.NewRedis(redis.New("localhost:6379")),
    cache.WithExpiry(time.Hour),
)

// 多级缓存
multiCache := cache.NewMultiCache(localCache, redisCache)

// 使用
var user User
err := multiCache.Take(&user, func(v interface{}) error {
    return db.First(v, id).Error
}, "user:%d", id)
```

### 2.4.2 缓存模式

```go
// Cache-Aside 模式
func GetUser(id int64) (*User, error) {
    // 1. 先查缓存
    key := fmt.Sprintf("user:%d", id)
    cached, err := redis.Get(key)
    if err == nil {
        var user User
        json.Unmarshal([]byte(cached), &user)
        return &user, nil
    }

    // 2. 缓存未命中，查数据库
    var user User
    err = db.First(&user, id).Error
    if err != nil {
        return nil, err
    }

    // 3. 写入缓存
    data, _ := json.Marshal(user)
    redis.Setex(key, string(data), 3600)

    return &user, nil
}

// Write-Through 模式
func CreateUser(user *User) error {
    // 1. 写入数据库
    err := db.Create(user).Error
    if err != nil {
        return err
    }

    // 2. 写入缓存
    key := fmt.Sprintf("user:%d", user.ID)
    data, _ := json.Marshal(user)
    redis.Setex(key, string(data), 3600)

    return nil
}
```

---

## 2.5 限流与熔断

### 2.5.1 手动限流

```go
import (
    "github.com/zeromicro/go-zero/core/limit"
    "github.com/zeromicro/go-zero/core/stores/redis"
)

// 令牌桶限流
tokenLimiter := limit.NewTokenLimiter(
    1000,  // 容量
    1000,  // 速率
    redis.New("localhost:6379"),
)

// 在请求处理中使用
func handleRequest(w http.ResponseWriter, r *http.Request) {
    if !tokenLimiter.Allow() {
        httpx.Error(w, errorx.New("rate limit exceeded"))
        return
    }
    // 处理请求
}
```

### 2.5.2 手动熔断

```go
import "github.com/zeromicro/go-zero/core/breaker"

// 获取熔断器
br := breaker.GetBreaker("user-service")

// 执行受保护的操作
err := br.Do(func() error {
    return callRemoteService()
}, func(err error) bool {
    // 判断是否需要重试
    return err != nil
})
```

### 2.5.3 HTTP 客户端熔断

```go
import "github.com/zeromicro/go-zero/core/breaker"

type HttpClient struct {
    client  *http.Client
    breaker *breaker.Breaker
}

func NewHttpClient() *HttpClient {
    return &HttpClient{
        client:  &http.Client{},
        breaker: breaker.GetBreaker("http-client"),
    }
}

func (c *HttpClient) Get(url string) ([]byte, error) {
    var result []byte
    err := c.breaker.Do(func() error {
        resp, err := c.client.Get(url)
        if err != nil {
            return err
        }
        defer resp.Body.Close()

        result, err = io.ReadAll(resp.Body)
        return err
    })

    return result, err
}
```

---

## 2.6 项目示例：用户服务

### 2.6.1 项目结构

```
user-service/
├── api/
│   └── user.api
├── cmd/
│   └── user-api/
│       └── main.go
├── configs/
│   └── user-api.yaml
├── internal/
│   ├── config/
│   │   └── config.go
│   ├── handler/
│   │   ├── user.go
│   │   └── routes.go
│   ├── logic/
│   │   └── user.go
│   ├── model/
│   │   └── user.go
│   ├── svc/
│   │   └── service.go
│   └── types/
│       └── types.go
├── go.mod
└── go.sum
```

### 2.6.2 API 定义

```yaml
# user.api
type (
  User {
    Id        int64  `json:"id"`
    Username  string `json:"username"`
    Email     string `json:"email"`
    Phone     string `json:"phone"`
    Nickname  string `json:"nickname"`
    Avatar    string `json:"avatar"`
    Status    int    `json:"status"`
    CreatedAt string `json:"created_at"`
    UpdatedAt string `json:"updated_at"`
  }

  CreateUserRequest {
    Username string `json:"username"`
    Email    string `json:"email"`
    Phone    string `json:"phone"`
    Password string `json:"password"`
    Nickname string `json:"nickname"`
  }

  UpdateUserRequest {
    Id       int64  `json:"id"`
    Email    string `json:"email"`
    Phone    string `json:"phone"`
    Nickname string `json:"nickname"`
    Avatar   string `json:"avatar"`
  }

  GetUserRequest {
    Id int64 `json:"id"`
  }

  ListUserRequest {
    Page     int `json:"page,default=1"`
    PageSize int `json:"page_size,default=10"`
    Keyword  string `json:"keyword,optional"`
  }

  ListUserResponse {
    Total int64  `json:"total"`
    Users []User `json:"users"`
  }

  DeleteUserRequest {
    Id int64 `json:"id"`
  }

  LoginRequest {
    Username string `json:"username"`
    Password string `json:"password"`
  }

  LoginResponse {
    Token string `json:"token"`
    User  User   `json:"user"`
  }
)

service user-api {
  @handler CreateUserHandler
  post /api/user/create (CreateUserRequest) returns (User)

  @handler UpdateUserHandler
  put /api/user/update (UpdateUserRequest) returns (User)

  @handler GetUserHandler
  get /api/user/get (GetUserRequest) returns (User)

  @handler ListUserHandler
  get /api/user/list (ListUserRequest) returns (ListUserResponse)

  @handler DeleteUserHandler
  delete /api/user/delete (DeleteUserRequest) returns (string)

  @handler LoginHandler
  post /api/user/login (LoginRequest) returns (LoginResponse)
}
```

### 2.6.3 配置文件

```yaml
# configs/user-api.yaml
Name: user-api
Host: 0.0.0.0
Port: 8080

Log:
  Mode: console
  Level: info

Database:
  Type: mysql
  DataSource: root:root123@tcp(localhost:3306)/user_db?charset=utf8mb4&parseTime=True&loc=Local

Redis:
  Host: localhost:6379
  Pass: ""
  Type: node

Auth:
  Secret: your-secret-key
  TokenExpire: 86400
```

### 2.6.4 配置结构

```go
// internal/config/config.go
package config

import (
    "github.com/zeromicro/go-zero/rest"
    "github.com/zeromicro/go-zero/core/stores/sqlx"
)

type Config struct {
    rest.RestConf
    Database struct {
        Type     string
        DataSource string
    }
    Redis struct {
        Host string
        Pass string
        Type string
    }
    Auth struct {
        Secret     string
        TokenExpire int
    }
}
```

### 2.6.5 数据模型

```go
// internal/model/user.go
package model

import (
    "time"

    "github.com/zeromicro/go-zero/core/stores/sqlx"
)

var (
    ErrNotFound = sqlx.ErrNotFound
)

type User struct {
    Id        int64     `json:"id"`
    Username  string    `json:"username"`
    Email     string    `json:"email"`
    Phone     string    `json:"phone"`
    Password  string    `json:"-"`
    Nickname  string    `json:"nickname"`
    Avatar    string    `json:"avatar"`
    Status    int       `json:"status"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}

type UserModel interface {
    Insert(user *User) (int64, error)
    FindOne(id int64) (*User, error)
    FindOneByUsername(username string) (*User, error)
    Update(user *User) error
    Delete(id int64) error
    List(keyword string, page, pageSize int) ([]*User, int64, error)
}

type userModel struct {
    conn sqlx.SqlConn
}

func NewUserModel(conn sqlx.SqlConn) UserModel {
    return &userModel{conn: conn}
}

func (m *userModel) Insert(user *User) (int64, error) {
    query := `INSERT INTO users (username, email, phone, password, nickname, avatar, status) 
              VALUES (?, ?, ?, ?, ?, ?, ?)`
    result, err := m.conn.Exec(query, user.Username, user.Email, user.Phone, 
        user.Password, user.Nickname, user.Avatar, user.Status)
    if err != nil {
        return 0, err
    }
    return result.LastInsertId()
}

func (m *userModel) FindOne(id int64) (*User, error) {
    query := `SELECT id, username, email, phone, password, nickname, avatar, status, created_at, updated_at 
              FROM users WHERE id = ?`
    var user User
    err := m.conn.QueryRow(&user, query, id)
    return &user, err
}

func (m *userModel) FindOneByUsername(username string) (*User, error) {
    query := `SELECT id, username, email, phone, password, nickname, avatar, status, created_at, updated_at 
              FROM users WHERE username = ?`
    var user User
    err := m.conn.QueryRow(&user, query, username)
    return &user, err
}

func (m *userModel) Update(user *User) error {
    query := `UPDATE users SET email=?, phone=?, nickname=?, avatar=?, status=?, updated_at=? WHERE id=?`
    _, err := m.conn.Exec(query, user.Email, user.Phone, user.Nickname, 
        user.Avatar, user.Status, time.Now(), user.Id)
    return err
}

func (m *userModel) Delete(id int64) error {
    query := `DELETE FROM users WHERE id = ?`
    _, err := m.conn.Exec(query, id)
    return err
}

func (m *userModel) List(keyword string, page, pageSize int) ([]*User, int64, error) {
    var users []*User
    var total int64

    where := ""
    if keyword != "" {
        where = " WHERE username LIKE '%" + keyword + "%' OR email LIKE '%" + keyword + "%'"
    }

    countQuery := "SELECT COUNT(*) FROM users" + where
    err := m.conn.QueryRow(&total, countQuery)
    if err != nil {
        return nil, 0, err
    }

    query := "SELECT id, username, email, phone, password, nickname, avatar, status, created_at, updated_at FROM users" + where
    query += " LIMIT " + string(rune(pageSize)) + " OFFSET " + string(rune((page-1)*pageSize))

    err = m.conn.QueryRows(&users, query)
    return users, total, err
}
```

### 2.6.6 服务上下文

```go
// internal/svc/service.go
package svc

import (
    "github.com/zeromicro/go-zero/core/stores/redis"
    "github.com/zeromicro/go-zero/core/stores/sqlx"

    "user-service/internal/config"
    "user-service/internal/model"
)

type ServiceContext struct {
    UserModel model.UserModel
    Redis     *redis.Redis
    Config    *config.Config
}

func NewServiceContext(c *config.Config) *ServiceContext {
    conn := sqlx.NewMysql(c.Database.DataSource)
    return &ServiceContext{
        UserModel: model.NewUserModel(conn),
        Redis:     redis.New(c.Redis.Host),
        Config:    c,
    }
}
```

### 2.6.7 业务逻辑

```go
// internal/logic/user.go
package logic

import (
    "context"
    "errors"
    "time"

    "github.com/golang-jwt/jwt/v4"
    "github.com/zeromicro/go-zero/core/errorx"
    "github.com/zeromicro/go-zero/core/logx"

    "user-service/internal/svc"
    "user-service/internal/types"
)

type UserLogic struct {
    ctx    context.Context
    svcCtx *svc.ServiceContext
}

func NewUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UserLogic {
    return &UserLogic{
        ctx:    ctx,
        svcCtx: svcCtx,
    }
}

func (l *UserLogic) CreateUser(req *types.CreateUserRequest) (*types.User, error) {
    // 检查用户名是否存在
    existing, err := l.svcCtx.UserModel.FindOneByUsername(req.Username)
    if err == nil && existing != nil {
        return nil, errors.New("用户名已存在")
    }

    // 创建用户
    user := &types.User{
        Username: req.Username,
        Email:    req.Email,
        Phone:    req.Phone,
        Nickname: req.Nickname,
        Status:   1,
    }

    id, err := l.svcCtx.UserModel.Insert(user)
    if err != nil {
        logx.Error("创建用户失败", "error", err)
        return nil, errorx.New("创建用户失败")
    }

    user.Id = id
    return user, nil
}

func (l *UserLogic) GetUser(req *types.GetUserRequest) (*types.User, error) {
    user, err := l.svcCtx.UserModel.FindOne(req.Id)
    if err != nil {
        return nil, errorx.New("用户不存在")
    }

    return user, nil
}

func (l *UserLogic) ListUser(req *types.ListUserRequest) (*types.ListUserResponse, error) {
    users, total, err := l.svcCtx.UserModel.List(req.Keyword, req.Page, req.PageSize)
    if err != nil {
        return nil, errorx.New("获取用户列表失败")
    }

    return &types.ListUserResponse{
        Total: total,
        Users: users,
    }, nil
}

func (l *UserLogic) UpdateUser(req *types.UpdateUserRequest) (*types.User, error) {
    existing, err := l.svcCtx.UserModel.FindOne(req.Id)
    if err != nil {
        return nil, errorx.New("用户不存在")
    }

    existing.Email = req.Email
    existing.Phone = req.Phone
    existing.Nickname = req.Nickname
    existing.Avatar = req.Avatar

    err = l.svcCtx.UserModel.Update(existing)
    if err != nil {
        return nil, errorx.New("更新用户失败")
    }

    return existing, nil
}

func (l *UserLogic) DeleteUser(req *types.DeleteUserRequest) (string, error) {
    err := l.svcCtx.UserModel.Delete(req.Id)
    if err != nil {
        return "", errorx.New("删除用户失败")
    }

    return "删除成功", nil
}

func (l *UserLogic) Login(req *types.LoginRequest) (*types.LoginResponse, error) {
    user, err := l.svcCtx.UserModel.FindOneByUsername(req.Username)
    if err != nil {
        return nil, errorx.New("用户名或密码错误")
    }

    // 验证密码（实际应使用 bcrypt）
    if user.Password != req.Password {
        return nil, errorx.New("用户名或密码错误")
    }

    // 生成 JWT Token
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
        "user_id":  user.Id,
        "username": user.Username,
        "exp":      time.Now().Add(time.Hour * 24).Unix(),
    })
    tokenString, _ := token.SignedString([]byte(l.svcCtx.Config.Auth.Secret))

    return &types.LoginResponse{
        Token: tokenString,
        User:  user,
    }, nil
}
```

### 2.6.8 Handler

```go
// internal/handler/user.go
package handler

import (
    "net/http"

    "github.com/zeromicro/go-zero/rest/httpx"

    "user-service/internal/logic"
    "user-service/internal/svc"
    "user-service/internal/types"
)

func CreateUserHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        var req types.CreateUserRequest
        if err := httpx.Parse(r, &req); err != nil {
            httpx.Error(w, err)
            return
        }

        l := logic.NewUserLogic(r.Context(), svcCtx)
        resp, err := l.CreateUser(&req)
        if err != nil {
            httpx.Error(w, err)
        } else {
            httpx.OkJson(w, resp)
        }
    }
}

func GetUserHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        var req types.GetUserRequest
        if err := httpx.Parse(r, &req); err != nil {
            httpx.Error(w, err)
            return
        }

        l := logic.NewUserLogic(r.Context(), svcCtx)
        resp, err := l.GetUser(&req)
        if err != nil {
            httpx.Error(w, err)
        } else {
            httpx.OkJson(w, resp)
        }
    }
}

func ListUserHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        var req types.ListUserRequest
        if err := httpx.Parse(r, &req); err != nil {
            httpx.Error(w, err)
            return
        }

        l := logic.NewUserLogic(r.Context(), svcCtx)
        resp, err := l.ListUser(&req)
        if err != nil {
            httpx.Error(w, err)
        } else {
            httpx.OkJson(w, resp)
        }
    }
}

func UpdateUserHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        var req types.UpdateUserRequest
        if err := httpx.Parse(r, &req); err != nil {
            httpx.Error(w, err)
            return
        }

        l := logic.NewUserLogic(r.Context(), svcCtx)
        resp, err := l.UpdateUser(&req)
        if err != nil {
            httpx.Error(w, err)
        } else {
            httpx.OkJson(w, resp)
        }
    }
}

func DeleteUserHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        var req types.DeleteUserRequest
        if err := httpx.Parse(r, &req); err != nil {
            httpx.Error(w, err)
            return
        }

        l := logic.NewUserLogic(r.Context(), svcCtx)
        resp, err := l.DeleteUser(&req)
        if err != nil {
            httpx.Error(w, err)
        } else {
            httpx.OkJson(w, resp)
        }
    }
}

func LoginHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        var req types.LoginRequest
        if err := httpx.Parse(r, &req); err != nil {
            httpx.Error(w, err)
            return
        }

        l := logic.NewUserLogic(r.Context(), svcCtx)
        resp, err := l.Login(&req)
        if err != nil {
            httpx.Error(w, err)
        } else {
            httpx.OkJson(w, resp)
        }
    }
}
```

### 2.6.9 路由配置

```go
// internal/handler/routes.go
package handler

import (
    "net/http"

    "github.com/zeromicro/go-zero/rest"

    "user-service/internal/svc"
)

func RegisterHandlers(server *rest.Server, serverCtx *svc.ServiceContext) {
    server.AddRoutes(
        []rest.Route{
            {
                Method:  http.MethodPost,
                Path:    "/api/user/create",
                Handler: CreateUserHandler(serverCtx),
            },
            {
                Method:  http.MethodGet,
                Path:    "/api/user/get",
                Handler: GetUserHandler(serverCtx),
            },
            {
                Method:  http.MethodGet,
                Path:    "/api/user/list",
                Handler: ListUserHandler(serverCtx),
            },
            {
                Method:  http.MethodPut,
                Path:    "/api/user/update",
                Handler: UpdateUserHandler(serverCtx),
            },
            {
                Method:  http.MethodDelete,
                Path:    "/api/user/delete",
                Handler: DeleteUserHandler(serverCtx),
            },
            {
                Method:  http.MethodPost,
                Path:    "/api/user/login",
                Handler: LoginHandler(serverCtx),
            },
        },
    )
}
```

### 2.6.10 入口文件

```go
// cmd/user-api/main.go
package main

import (
    "flag"

    "github.com/zeromicro/go-zero/core/conf"
    "github.com/zeromicro/go-zero/core/logx"
    "github.com/zeromicro/go-zero/rest"

    "user-service/internal/config"
    "user-service/internal/handler"
    "user-service/internal/svc"
)

var configFile = flag.String("f", "configs/user-api.yaml", "the config file")

func main() {
    flag.Parse()

    var c config.Config
    conf.MustLoad(*flag.Parse(), &c)

    logx.MustSetup(logx.Config{
        Mode: c.Log.Mode,
        Level: c.Log.Level,
    })

    server := rest.MustNewServer(c.RestConf)
    defer server.Stop()

    handler.RegisterHandlers(server, svc.NewServiceContext(&c))

    logx.Infof("Starting server at %s:%d...", c.Host, c.Port)
    server.Start()
}
```

---

## 2.7 运行项目

### 2.7.1 生成代码

```bash
# 进入项目目录
cd user-service

# 生成代码
goctl api go -api api/user.api -dir . -style gozero
```

### 2.7.2 启动服务

```bash
# 启动服务
go run cmd/user-api/main.go -f configs/user-api.yaml
```

### 2.7.3 测试 API

```bash
# 创建用户
curl -X POST http://localhost:8080/api/user/create \
  -H "Content-Type: application/json" \
  -d '{"username":"test","email":"test@example.com","password":"123456"}'

# 获取用户
curl http://localhost:8080/api/user/get?id=1

# 用户列表
curl http://localhost:8080/api/user/list?page=1&page_size=10

# 登录
curl -X POST http://localhost:8080/api/user/login \
  -H "Content-Type: application/json" \
  -d '{"username":"test","password":"123456"}'
```

---

## 2.8 扩展练习

1. **添加 Redis 缓存**：为用户查询添加缓存
2. **添加 JWT 认证**：保护需要认证的接口
3. **添加分页**：完善分页逻辑
4. **添加日志**：集成结构化日志
5. **添加监控**：集成 Prometheus
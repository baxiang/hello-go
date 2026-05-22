# Wire 依赖注入

## 简介

Wire 是 Go 的编译时依赖注入工具，特点：
- **编译时检查**：运行时零开销
- **类型安全**：编译错误而非运行时错误
- **代码生成**：生成可读的 Go 代码
- **易于重构**：显式依赖关系

## 安装

```bash
# 安装 wire 工具
go install github.com/google/wire/cmd/wire@latest

# 验证安装
wire version
```

## 基础概念

### 核心概念

```
Provider:    提供依赖的函数
Injector:    需要注入依赖的函数
Set:         Provider 的集合
ProviderSet: 一组 Provider 的集合
```

### 示例：数据库连接

```go
// database.go
package main

import (
    "database/sql"
    _ "github.com/go-sql-driver/mysql"
)

// Provider: 提供数据库连接
func ProvideMySQLDSN() string {
    return "root:password@tcp(localhost:3306)/mydb?parseTime=true"
}

func ProvideDB(dsn string) (*sql.DB, error) {
    db, err := sql.Open("mysql", dsn)
    if err != nil {
        return nil, err
    }
    return db, nil
}
```

```go
// wire.go
//go:build wireinject
// +build wireinject

package main

import (
    "github.com/google/wire"
)

// Injector: 声明需要注入的依赖
func InitializeApp() (*App, error) {
    wire.Build(
        ProvideMySQLDSN,
        ProvideDB,
        NewApp,
    )
    return nil, nil
}
```

```go
// app.go
package main

import "database/sql"

type App struct {
    db *sql.DB
}

func NewApp(db *sql.DB) *App {
    return &App{db: db}
}

func (a *App) Run() error {
    return a.db.Ping()
}
```

```go
// main.go
package main

func main() {
    app := InitializeApp()
    if err := app.Run(); err != nil {
        panic(err)
    }
}
```

### 生成代码

```bash
# 生成 wire_gen.go
wire gen ./...

# 运行程序
go run .
```

## Wire 语法详解

### Provider

```go
// 简单 Provider
func ProvideConfig() *Config {
    return &Config{
        Debug: true,
        Port:  8080,
    }
}

// 带依赖的 Provider
func ProvideLogger(config *Config) *Logger {
    return NewLogger(config.Debug)
}

// 带多个返回值 (包括 error)
func ProvideDB(dsn string) (*sql.DB, error) {
    return sql.Open("mysql", dsn)
}

// 具名 Provider (使用接口)
type Database interface {
    Query(string) ([]Row, error)
}

func ProvideMySQL(dsn string) (Database, error) {
    return NewMySQL(dsn)
}
```

### ProviderSet

```go
// 将相关 Provider 分组
var DatabaseSet = wire.NewSet(
    ProvideMySQLDSN,
    ProvideDB,
)

var LoggerSet = wire.NewSet(
    ProvideConfig,
    ProvideLogger,
)

// 在 Injector 中使用
func InitializeApp() (*App, error) {
    wire.Build(
        DatabaseSet,
        LoggerSet,
        NewApp,
    )
    return nil, nil
}
```

### 绑定接口

```go
// repository.go
type UserRepository interface {
    FindByID(id int) (*User, error)
    Save(user *User) error
}

type MySQLUserRepository struct {
    db *sql.DB
}

func NewMySQLUserRepository(db *sql.DB) *MySQLUserRepository {
    return &MySQLUserRepository{db: db}
}

func (r *MySQLUserRepository) FindByID(id int) (*User, error) {
    // 实现...
}

// wire.go
var RepositorySet = wire.NewSet(
    NewMySQLUserRepository,
    wire.Bind(new(UserRepository), new(*MySQLUserRepository)),
)
```

### 使用 value

```go
// 直接提供值
func InitializeApp() (*App, error) {
    wire.Build(
        wire.Value("localhost:8080"),  // 直接提供字符串
        wire.Value(30*time.Second),     // 提供 duration
        NewApp,
    )
    return nil, nil
}

// 使用 struct{} 提供空值
func InitializeApp() (*App, error) {
    wire.Build(
        wire.Struct(new(Config), "*"),  // 提供零值结构体
        NewApp,
    )
    return nil, nil
}
```

### 清理函数

```go
// Provider 返回清理函数
func ProvideDB(dsn string) (*sql.DB, func(), error) {
    db, err := sql.Open("mysql", dsn)
    if err != nil {
        return nil, nil, err
    }

    // 清理函数
    cleanup := func() {
        db.Close()
    }

    return db, cleanup, nil
}

// 生成的代码会自动处理清理
func InitializeApp() (*App, func(), error) {
    db, cleanup, err := ProvideDB(dsn)
    if err != nil {
        return nil, nil, err
    }

    app := NewApp(db)

    // 返回清理函数
    return app, func() {
        cleanup()
    }, nil
}
```

## 完整项目实践

### 项目结构

```
project/
├── cmd/
│   └── server/
│       ├── main.go
│       ├── wire.go
│       └── wire_gen.go
├── internal/
│   ├── config/
│   ├── database/
│   ├── repository/
│   ├── service/
│   └── handler/
└── go.mod
```

### 配置层

```go
// internal/config/config.go
package config

type Config struct {
    Server   ServerConfig
    Database DatabaseConfig
    Redis    RedisConfig
}

type ServerConfig struct {
    Port int
}

type DatabaseConfig struct {
    DSN     string
    MaxIdle int
    MaxOpen int
}

type RedisConfig struct {
    Addr     string
    Password string
    DB       int
}

func ProvideConfig() *Config {
    return &Config{
        Server: ServerConfig{
            Port: 8080,
        },
        Database: DatabaseConfig{
            DSN:     "root:secret@tcp(localhost:3306)/mydb?parseTime=true",
            MaxIdle: 10,
            MaxOpen: 100,
        },
        Redis: RedisConfig{
            Addr:     "localhost:6379",
            Password: "",
            DB:       0,
        },
    }
}
```

### 数据库层

```go
// internal/database/mysql.go
package database

import (
    "database/sql"
    "github.com/go-sql-driver/mysql"
    "time"
)

func ProvideMySQLDB(cfg *DatabaseConfig) (*sql.DB, error) {
    db, err := sql.Open("mysql", cfg.DSN)
    if err != nil {
        return nil, err
    }

    db.SetMaxIdleConns(cfg.MaxIdle)
    db.SetMaxOpenConns(cfg.MaxOpen)
    db.SetConnMaxLifetime(time.Hour)

    // 测试连接
    if err := db.Ping(); err != nil {
        return nil, err
    }

    return db, nil
}

// 清理函数
func CloseDB(db *sql.DB) {
    db.Close()
}

var MySQLSet = wire.NewSet(
    ProvideMySQLDB,
    wire.Bind(new(DB), new(*sql.DB)),
)
```

### Redis 层

```go
// internal/database/redis.go
package database

import (
    "github.com/go-redis/redis/v8"
    "context"
)

type RedisClient struct {
    *redis.Client
}

func ProvideRedisClient(cfg *RedisConfig) (*RedisClient, error) {
    client := redis.NewClient(&redis.Options{
        Addr:     cfg.Addr,
        Password: cfg.Password,
        DB:       cfg.DB,
    })

    if err := client.Ping(context.Background()).Err(); err != nil {
        return nil, err
    }

    return &RedisClient{client}, nil
}

var RedisSet = wire.NewSet(
    ProvideRedisClient,
)
```

### Repository 层

```go
// internal/repository/user.go
package repository

import (
    "database/sql"
    "context"
)

type User struct {
    ID    int64
    Name  string
    Email string
}

type UserRepository interface {
    FindByID(ctx context.Context, id int64) (*User, error)
    Save(ctx context.Context, user *User) error
}

type mysqlUserRepository struct {
    db *sql.DB
}

func NewUserRepository(db *sql.DB) *mysqlUserRepository {
    return &mysqlUserRepository{db: db}
}

func (r *mysqlUserRepository) FindByID(ctx context.Context, id int64) (*User, error) {
    row := r.db.QueryRowContext(ctx, "SELECT id, name, email FROM users WHERE id = ?", id)

    var user User
    err := row.Scan(&user.ID, &user.Name, &user.Email)
    if err != nil {
        return nil, err
    }
    return &user, nil
}

func (r *mysqlUserRepository) Save(ctx context.Context, user *User) error {
    _, err := r.db.ExecContext(ctx,
        "INSERT INTO users (name, email) VALUES (?, ?)",
        user.Name, user.Email)
    return err
}

var UserRepositorySet = wire.NewSet(
    NewUserRepository,
    wire.Bind(new(UserRepository), new(*mysqlUserRepository)),
)
```

### Service 层

```go
// internal/service/user.go
package service

import (
    "context"
    "github.com/example/project/internal/repository"
    "github.com/example/project/internal/database"
)

type UserService struct {
    userRepo repository.UserRepository
    redis    *database.RedisClient
}

func NewUserService(userRepo repository.UserRepository, redis *database.RedisClient) *UserService {
    return &UserService{
        userRepo: userRepo,
        redis:    redis,
    }
}

func (s *UserService) GetUser(ctx context.Context, id int64) (*repository.User, error) {
    // 尝试从缓存获取
    // ...

    // 从数据库获取
    return s.userRepo.FindByID(ctx, id)
}

var UserServiceSet = wire.NewSet(
    NewUserService,
)
```

### Handler 层

```go
// internal/handler/user.go
package handler

import (
    "net/http"
    "github.com/example/project/internal/service"
)

type UserHandler struct {
    userService *service.UserService
}

func NewUserHandler(userService *service.UserService) *UserHandler {
    return &UserHandler{userService: userService}
}

func (h *UserHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    // 处理请求
}

var UserHandlerSet = wire.NewSet(
    NewUserHandler,
)
```

### Wire 配置

```go
// cmd/server/wire.go
//go:build wireinject
// +build wireinject

package main

import (
    "github.com/google/wire"
    "github.com/example/project/internal/config"
    "github.com/example/project/internal/database"
    "github.com/example/project/internal/repository"
    "github.com/example/project/internal/service"
    "github.com/example/project/internal/handler"
)

func InitializeApp() (*App, func(), error) {
    wire.Build(
        // 配置
        config.ProvideConfig,

        // 基础设施
        database.MySQLSet,
        database.RedisSet,

        // Repository
        repository.UserRepositorySet,

        // Service
        service.UserServiceSet,

        // Handler
        handler.UserHandlerSet,

        // 应用
        NewApp,
    )
    return nil, nil, nil
}
```

### 主程序

```go
// cmd/server/main.go
package main

import (
    "context"
    "log"
    "net/http"
    "os"
    "os/signal"
    "syscall"
)

type App struct {
    handler *handler.UserHandler
    server  *http.Server
}

func NewApp(h *handler.UserHandler) *App {
    return &App{
        handler: h,
        server:  &http.Server{Addr: ":8080", Handler: h},
    }
}

func (a *App) Run() error {
    return a.server.ListenAndServe()
}

func (a *App) Shutdown(ctx context.Context) error {
    return a.server.Shutdown(ctx)
}

func main() {
    // 初始化应用
    app, cleanup, err := InitializeApp()
    if err != nil {
        log.Fatal(err)
    }
    defer cleanup()

    // 启动服务器
    go func() {
        if err := app.Run(); err != nil && err != http.ErrServerClosed {
            log.Fatal(err)
        }
    }()

    // 等待中断信号
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit

    // 优雅关闭
    if err := app.Shutdown(context.Background()); err != nil {
        log.Fatal(err)
    }
}
```

### 生成和运行

```bash
# 生成依赖注入代码
wire gen ./cmd/server

# 查看生成的代码
cat cmd/server/wire_gen.go

# 运行
go run ./cmd/server
```

## 条件依赖

```go
// 根据环境选择不同的实现
type Mailer interface {
    Send(to, subject, body string) error
}

type SMTPMailer struct{ /* ... */ }
type SendGridMailer struct{ /* ... */ }

func ProvideSMTPMailer(cfg *Config) *SMTPMailer {
    // SMTP 实现
}

func ProvideSendGridMailer(cfg *Config) *SendGridMailer {
    // SendGrid API 实现
}

// 使用 build tags 选择
// wire_dev.go
//go:build wireinject
// +build wireinject

func InitializeApp() (*App, error) {
    wire.Build(
        ProvideSMTPMailer,  // 开发环境
        wire.Bind(new(Mailer), new(*SMTPMailer)),
        NewApp,
    )
    return nil, nil
}

// wire_prod.go
//go:build wireinject && prod
// +build wireinject,prod

func InitializeApp() (*App, error) {
    wire.Build(
        ProvideSendGridMailer,  // 生产环境
        wire.Bind(new(Mailer), new(*SendGridMailer)),
        NewApp,
    )
    return nil, nil
}
```

## Wire 检查清单

```
[ ] 使用接口定义依赖
[ ] 使用 ProviderSet 组织代码
[ ] 使用 wire.Bind 绑定接口
[ ] 为资源提供清理函数
[ ] 按层组织 Provider
[ ] 使用 build tags 区分环境
[ ] 提交生成的 wire_gen.go
[ ] 避免循环依赖
```

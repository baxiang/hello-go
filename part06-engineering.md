# 第六部分：工程实践

## 6.1 项目结构

### 标准项目布局

```
project/
├── cmd/                        # 应用程序入口
│   ├── myapp/                  # 主应用
│   │   └── main.go
│   └── migration/              # 数据库迁移工具
│       └── main.go
├── internal/                   # 私有包 (不可被外部导入)
│   ├── config/                 # 配置加载与解析
│   │   ├── config.go
│   │   └── config_test.go
│   ├── handler/                # HTTP 处理器
│   │   ├── user_handler.go
│   │   └── middleware.go
│   ├── service/                # 业务逻辑层
│   │   ├── user_service.go
│   │   └── auth_service.go
│   ├── repository/             # 数据访问层
│   │   ├── user_repo.go
│   │   └── db.go
│   └── models/                 # 数据模型
│       └── user.go
├── pkg/                        # 公共包 (可被外部导入)
│   ├── logger/                 # 日志包
│   ├── utils/                  # 工具函数
│   └── middleware/             # 通用中间件
├── api/                        # API 定义
│   ├── openapi.yaml
│   └── proto/                  # protobuf 定义
│       └── user.proto
├── configs/                    # 配置文件
│   ├── config.yaml
│   └── config.local.yaml
├── scripts/                    # 脚本文件
│   ├── build.sh
│   └── deploy.sh
├── deployments/                # 部署配置
│   ├── Dockerfile
│   └── k8s/
│       ├── deployment.yaml
│       └── service.yaml
├── test/                       # 测试文件
│   ├── integration/
│   └── e2e/
├── docs/                       # 文档
│   └── README.md
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

### cmd 目录示例

```go
// cmd/myapp/main.go
package main

import (
    "context"
    "flag"
    "fmt"
    "log"
    "os"
    "os/signal"
    "syscall"
    "time"

    "example.com/myapp/internal/config"
    "example.com/myapp/internal/handler"
    "example.com/myapp/internal/service"
    "example.com/myapp/pkg/logger"
)

var version = "dev"

func main() {
    // ========== 命令行参数 ==========
    var configFile string
    flag.StringVar(&configFile, "config", "configs/config.yaml", "配置文件路径")
    flag.BoolVar(&showVersion, "version", false, "显示版本")
    flag.Parse()

    if showVersion {
        fmt.Println(version)
        return
    }

    // ========== 初始化日志 ==========
    log := logger.NewLogger()
    log.Info("应用启动", "version", version)

    // ========== 加载配置 ==========
    cfg, err := config.Load(configFile)
    if err != nil {
        log.Error("加载配置失败", "error", err)
        os.Exit(1)
    }

    // ========== 初始化服务 ==========
    userService := service.NewUserService(cfg.Database)
    authService := service.NewAuthService(cfg.JWT)

    // ========== 初始化 HTTP 处理器 ==========
    userHandler := handler.NewUserHandler(userService)

    // ========== 启动服务器 ==========
    server := handler.NewServer(cfg.Server, userHandler)

    // ========== 优雅关闭 ==========
    ctx, stop := signal.NotifyContext(context.Background(),
        os.Interrupt, syscall.SIGTERM)
    defer stop()

    go func() {
        <-ctx.Done()
        log.Info("收到退出信号，开始优雅关闭")

        shutdownCtx, cancel := context.WithTimeout(
            context.Background(), 30*time.Second)
        defer cancel()

        if err := server.Shutdown(shutdownCtx); err != nil {
            log.Error("服务器关闭失败", "error", err)
        }
    }()

    // ========== 启动监听 ==========
    if err := server.Start(); err != nil {
        log.Error("服务器启动失败", "error", err)
        os.Exit(1)
    }

    log.Info("应用已停止")
}
```

### 配置管理 (viper)

```go
// internal/config/config.go
package config

import (
    "github.com/spf13/viper"
)

// Config 应用配置
type Config struct {
    Server   ServerConfig   `mapstructure:"server"`
    Database DatabaseConfig `mapstructure:"database"`
    JWT      JWTConfig      `mapstructure:"jwt"`
    Log      LogConfig      `mapstructure:"log"`
}

type ServerConfig struct {
    Port         int    `mapstructure:"port"`
    ReadTimeout  int    `mapstructure:"read_timeout"`
    WriteTimeout int    `mapstructure:"write_timeout"`
}

type DatabaseConfig struct {
    Host     string `mapstructure:"host"`
    Port     int    `mapstructure:"port"`
    User     string `mapstructure:"user"`
    Password string `mapstructure:"password"`
    DBName   string `mapstructure:"dbname"`
    SSLMode  string `mapstructure:"sslmode"`
    MaxOpen  int    `mapstructure:"max_open"`
    MaxIdle  int    `mapstructure:"max_idle"`
}

type JWTConfig struct {
    Secret     string `mapstructure:"secret"`
    ExpireHour int    `mapstructure:"expire_hour"`
}

type LogConfig struct {
    Level  string `mapstructure:"level"`
    Format string `mapstructure:"format"`
}

// Load 加载配置
func Load(configPath string) (*Config, error) {
    v := viper.New()

    // 配置文件路径
    v.SetConfigFile(configPath)

    // 环境变量前缀
    v.SetEnvPrefix("APP")
    v.AutomaticEnv()

    // 读取配置
    if err := v.ReadInConfig(); err != nil {
        return nil, err
    }

    // 解析配置
    var cfg Config
    if err := v.Unmarshal(&cfg); err != nil {
        return nil, err
    }

    // 验证配置
    if err := validate(&cfg); err != nil {
        return nil, err
    }

    return &cfg, nil
}

// validate 验证配置
func validate(cfg *Config) error {
    if cfg.Server.Port <= 0 || cfg.Server.Port > 65535 {
        return fmt.Errorf("invalid server port: %d", cfg.Server.Port)
    }
    if cfg.Database.Host == "" {
        return fmt.Errorf("database host is required")
    }
    return nil
}

// GetDSN 获取数据库连接字符串
func (c *DatabaseConfig) GetDSN() string {
    return fmt.Sprintf(
        "host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
        c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode,
    )
}
```

### 配置文件示例

```yaml
# configs/config.yaml
server:
  port: 8080
  read_timeout: 30
  write_timeout: 30

database:
  host: localhost
  port: 5432
  user: postgres
  password: secret
  dbname: myapp
  sslmode: disable
  max_open: 100
  max_idle: 10

jwt:
  secret: your-secret-key
  expire_hour: 24

log:
  level: info
  format: json
```

---

## 6.2 代码规范

### Effective Go 要点

```go
// ========== 1. 命名规范 ==========

// 包名：简短、小写、无下划线
package handler      // 好
package handlers     // 好
package UserHandler  // 不好

// 变量名：驼峰命名
var userID int       // 导出
var localCount int   // 私有

// 常量名：驼峰或全大写
const MaxRetry = 3
const APIVersion = "v1"

// 函数名：驼峰命名
func GetUser() {}     // 导出
func getUser() {}     // 私有


// ========== 2. 注释规范 ==========

// 包注释必须在 package 之前
// Package handler provides HTTP handlers for the API
package handler

// 导出函数必须有注释
// GetUser retrieves a user by ID
func GetUser(id int) (*User, error) {
    // ...
}


// ========== 3. 错误处理 ==========

// 错误是值，应像其他值一样处理
if err != nil {
    return err
}

// 错误消息小写开头，不带句号
return errors.New("invalid input")

// 错误包装提供上下文
if err != nil {
    return fmt.Errorf("parse config: %w", err)
}


// ========== 4. 并发安全 ==========

// 文档说明并发安全性
// Counter is a thread-safe counter
type Counter struct {
    mu    sync.Mutex
    value int
}


// ========== 5. 接口设计 ==========

// 小接口，单一职责
type Reader interface {
    Read(p []byte) (n int, err error)
}

// 接受接口，返回结构体
func Process(r io.Reader) (*Result, error) {
    // ...
}
```

### golint 与 golangci-lint

```bash
# ========== golangci-lint 安装 ==========
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# ========== 基本使用 ==========
golangci-lint run                    # 运行所有 linter
golangci-lint run ./...              # 检查所有包
golangci-lint run --fast             # 只运行快速的 linter
golangci-lint run -E gocritic        # 启用特定 linter
golangci-lint run --disable-all -E gofmt  # 只启用 gofmt

# ========== 配置文件 .golangci.yml ==========
# .golangci.yml
linters:
  enable:
    - gofmt
    - govet
    - gosimple
    - staticcheck
    - unused
    - errcheck
    - gocritic
    - gosec
    - misspell
    - prealloc

linters-settings:
  gofmt:
    simplify: true
  gocritic:
    enabled-checks:
      - hugeParam
      - rangeValCopy
  misspell:
    locale: US

issues:
  exclude-rules:
    - path: _test\.go
      linters:
        - errcheck
        - gosec
  max-issues-per-linter: 0
  max-same-issues: 0

run:
  timeout: 5m
  tests: true


# ========== CI/CD 集成 ==========
# GitHub Actions 示例
# .github/workflows/lint.yml
# name: Lint
# on: [push, pull_request]
# jobs:
#   lint:
#     runs-on: ubuntu-latest
#     steps:
#       - uses: actions/checkout@v3
#       - uses: actions/setup-go@v4
#         with:
#           go-version: '1.21'
#       - name: golangci-lint
#         uses: golangci/golangci-lint-action@v3
```

### 代码风格指南

```go
// ========== 1. 简化 if 语句 ==========
// 不好
if x == true {
    return true
}

// 好
if x {
    return true
}


// ========== 2. 使用 strings.Contains 代替 strings.Index ==========
// 不好
if strings.Index(s, substr) >= 0 {
    // ...
}

// 好
if strings.Contains(s, substr) {
    // ...
}


// ========== 3. 避免不必要的嵌套 ==========
// 不好
func process() error {
    if cond1 {
        if cond2 {
            // 逻辑
        }
    }
    return nil
}

// 好
func process() error {
    if !cond1 {
        return nil
    }
    if !cond2 {
        return nil
    }
    // 逻辑
    return nil
}


// ========== 4. 提前返回 (Guard Clause) ==========
// 不好
func validate(u *User) error {
    if u != nil {
        if u.Name != "" {
            if u.Age >= 0 {
                return nil
            }
        }
    }
    return errors.New("invalid user")
}

// 好
func validate(u *User) error {
    if u == nil {
        return errors.New("user is nil")
    }
    if u.Name == "" {
        return errors.New("name is required")
    }
    if u.Age < 0 {
        return errors.New("age must be non-negative")
    }
    return nil
}


// ========== 5. 使用 make 预设容量 ==========
// 不好
var result []int
for i := 0; i < n; i++ {
    result = append(result, i)
}

// 好
result := make([]int, 0, n)
for i := 0; i < n; i++ {
    result = append(result, i)
}


// ========== 6. 避免裸返回 ==========
// 不好
func parse() (name string, err error) {
    name, err = doParse()
    return  // 裸返回
}

// 好
func parse() (string, error) {
    return doParse()
}


// ========== 7. 零值可用 ==========
// 设计结构体时，确保零值可用
type Counter struct {
    mu    sync.Mutex
    value int
}
// 零值 Counter 可直接使用


// ========== 8. 上下文传递 ==========
// 不好
func process(data string) {
    // 没有上下文
}

// 好
func process(ctx context.Context, data string) {
    // 支持取消和超时
}
```

---

## 6.3 依赖管理

### Go Modules 进阶

```bash
# ========== 理解 go.mod ==========
/*
module example.com/myapp

go 1.21

require (
    github.com/gin-gonic/gin v1.9.1
)

// 间接依赖
require github.com/gin-contrib/sse v0.1.0 // indirect

// 替换 (本地开发/测试 fork)
replace github.com/old/pkg => github.com/new/pkg v1.0.0
replace example.com/myapp => ./

// 排除有问题的版本
exclude github.com/problematic/pkg v1.0.0

// 撤回已发布版本 (semver)
retract v1.0.0
retract [v1.0.0, v1.1.0]
*/


# ========== 依赖版本选择 ==========
# 主版本 < 2
go get github.com/pkg@v1.2.3

# 主版本 >= 2 (需要修改导入路径)
go get github.com/pkg/v2@v2.0.0
# 导入：import "github.com/pkg/v2"


# ========== 依赖升级策略 ==========
# 升级到最新版
go get github.com/pkg@latest

# 升级到特定版本
go get github.com/pkg@v1.2.3

# 升级所有依赖
go get -u ./...

# 升级次要版本和补丁
go get -u=patch ./...


# ========== 清理依赖 ==========
# 移除未使用的依赖
go mod tidy

# 验证依赖完整性
go mod verify

# 清理模块缓存
go clean -modcache


# ========== 查看依赖信息 ==========
go list -m all                           # 列出所有依赖
go list -m -versions github.com/pkg      # 查看可用版本
go mod graph                             # 依赖图
go mod why github.com/pkg                # 为什么需要这个依赖
go mod edit -json                        # JSON 格式 go.mod
```

### Vendor 模式

```bash
# ========== 使用 vendor ==========
# 将依赖复制到项目内
go mod vendor

# 使用 vendor 目录构建
go build -mod=vendor

# 测试时使用 vendor
go test -mod=vendor ./...


# ========== vendor 优势 ==========
# 1. 离线构建
# 2. 依赖固定
# 3. 审查依赖代码


# ========== vendor 劣势 ==========
# 1. 增加仓库大小
# 2. 需要手动更新
```

---

## 6.4 构建与部署

### 交叉编译

```bash
# ========== 查看支持的平台 ==========
go tool dist list


# ========== 交叉编译 ==========
# Linux AMD64
GOOS=linux GOARCH=amd64 go build -o app-linux

# Windows AMD64
GOOS=windows GOARCH=amd64 go build -o app.exe

# macOS ARM64
GOOS=darwin GOARCH=arm64 go build -o app-macos-arm

# ========== 构建优化 ==========
# 减小二进制大小
go build -ldflags="-s -w" -o app

# 添加版本信息
go build -ldflags="-s -w -X main.version=1.0.0 -X main.buildTime=$(date)"


# ========== Makefile 示例 ==========
# Makefile
.PHONY: build clean test

VERSION := 1.0.0
LDFLAGS := -s -w -X main.version=$(VERSION)

build:
    go build -ldflags="$(LDFLAGS)" -o bin/app ./cmd/myapp

build-all:
    GOOS=linux GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o bin/app-linux-amd64 ./cmd/myapp
    GOOS=darwin GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o bin/app-darwin-amd64 ./cmd/myapp
    GOOS=darwin GOARCH=arm64 go build -ldflags="$(LDFLAGS)" -o bin/app-darwin-arm64 ./cmd/myapp
    GOOS=windows GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o bin/app-windows-amd64.exe ./cmd/myapp

clean:
    rm -rf bin/

test:
    go test -race -cover ./...
```

### Docker 容器化

```dockerfile
# ========== 多阶段构建 ==========
# Dockerfile

# ========== 第一阶段：构建 ==========
FROM golang:1.21-alpine AS builder

# 安装必要工具
RUN apk add --no-cache git ca-certificates

# 设置工作目录
WORKDIR /build

# 复制 go.mod 和 go.sum
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w" \
    -o /app/myapp \
    ./cmd/myapp


# ========== 第二阶段：运行 ==========
FROM scratch

# 从 builder 复制 CA 证书
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# 从 builder 复制应用
COPY --from=builder /app/myapp /app/myapp

# 暴露端口
EXPOSE 8080

# 运行应用
ENTRYPOINT ["/app/myapp"]


# ========== 使用 alpine 作为基础镜像 ==========
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app
COPY myapp .

EXPOSE 8080
CMD ["./myapp"]


# ========== docker-compose ==========
# docker-compose.yml
version: '3.8'

services:
  app:
    build:
      context: .
      dockerfile: Dockerfile
    ports:
      - "8080:8080"
    environment:
      - APP_ENV=production
      - DB_HOST=postgres
    depends_on:
      - postgres
    restart: unless-stopped

  postgres:
    image: postgres:15-alpine
    environment:
      POSTGRES_USER: myapp
      POSTGRES_PASSWORD: secret
      POSTGRES_DB: myapp
    volumes:
      - postgres_data:/var/lib/postgresql/data
    ports:
      - "5432:5432"

volumes:
  postgres_data:
```

### CI/CD 流程

```yaml
# ========== GitHub Actions ==========
# .github/workflows/ci.yml
name: CI

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'
          cache: true

      - name: Download dependencies
        run: go mod download

      - name: Run tests
        run: go test -race -coverprofile=coverage.txt -covermode=atomic ./...

      - name: Upload coverage
        uses: codecov/codecov-action@v3

  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'

      - name: golangci-lint
        uses: golangci/golangci-lint-action@v3

  build:
    runs-on: ubuntu-latest
    needs: [test, lint]
    steps:
      - uses: actions/checkout@v3

      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'

      - name: Build
        run: go build -v ./...

      - name: Build Docker image
        run: docker build -t myapp:${{ github.sha }} .

      - name: Push to registry
        if: github.ref == 'refs/heads/main'
        run: |
          docker tag myapp:${{ github.sha }} registry.com/myapp:latest
          docker push registry.com/myapp:latest


# ========== GitLab CI ==========
# .gitlab-ci.yml
stages:
  - test
  - lint
  - build
  - deploy

variables:
  GO_VERSION: "1.21"

test:
  stage: test
  image: golang:${GO_VERSION}
  script:
    - go test -race -coverprofile=coverage.txt ./...
  artifacts:
    reports:
      coverage_report:
        coverage_format: cobertura
        path: coverage.xml

lint:
  stage: lint
  image: golangci/golangci-lint:latest
  script:
    - golangci-lint run

build:
  stage: build
  image: docker:latest
  services:
    - docker:dind
  script:
    - docker build -t myapp:${CI_COMMIT_SHA} .
    - docker push myapp:${CI_COMMIT_SHA}
  only:
    - main

deploy:
  stage: deploy
  image: bitnami/kubectl:latest
  script:
    - kubectl set image deployment/myapp myapp=myapp:${CI_COMMIT_SHA}
  only:
    - main
```

---

## 6.5 监控与可观测性

### 指标收集 (Prometheus)

```go
package main

import (
    "net/http"
    "time"
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
    "github.com/prometheus/client_golang/prometheus/promhttp"
    "github.com/gin-gonic/gin"
)

// ========== 定义指标 ==========
var (
    httpRequestsTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "http_requests_total",
            Help: "Total number of HTTP requests",
        },
        []string{"method", "endpoint", "status"},
    )

    httpRequestDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "http_request_duration_seconds",
            Help:    "HTTP request duration in seconds",
            Buckets: prometheus.DefBuckets,
        },
        []string{"method", "endpoint"},
    )

    activeConnections = promauto.NewGauge(
        prometheus.GaugeOpts{
            Name: "active_connections",
            Help: "Number of active connections",
        },
    )
)

// ========== 监控中间件 ==========
func PrometheusMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()

        activeConnections.Inc()
        defer activeConnections.Dec()

        c.Next()

        duration := time.Since(start).Seconds()

        httpRequestsTotal.WithLabelValues(
            c.Request.Method,
            c.FullPath(),
            fmt.Sprintf("%d", c.Writer.Status()),
        ).Inc()

        httpRequestDuration.WithLabelValues(
            c.Request.Method,
            c.FullPath(),
        ).Observe(duration)
    }
}

func main() {
    r := gin.Default()

    // Prometheus 指标端点
    r.GET("/metrics", gin.WrapH(promhttp.Handler()))

    // 应用路由
    r.Use(PrometheusMiddleware())
    r.GET("/api/users", getUsers)

    r.Run(":8080")
}

func getUsers(c *gin.Context) {
    c.JSON(200, gin.H{"users": []string{}})
}
```

### 链路追踪 (OpenTelemetry)

```go
package main

import (
    "context"
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/exporters/jaeger"
    "go.opentelemetry.io/otel/sdk/resource"
    sdktrace "go.opentelemetry.io/otel/sdk/trace"
    semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

// ========== 初始化 Jaeger Exporter ==========
func initTracer() (*sdktrace.TracerProvider, error) {
    exporter, err := jaeger.New(
        jaeger.WithCollectorEndpoint(
            jaeger.WithEndpoint("http://localhost:14268/api/traces"),
        ),
    )
    if err != nil {
        return nil, err
    }

    tp := sdktrace.NewTracerProvider(
        sdktrace.WithBatcher(exporter),
        sdktrace.WithResource(resource.NewWithAttributes(
            semconv.SchemaURL,
            semconv.ServiceName("myapp"),
            semconv.ServiceVersion("1.0.0"),
        )),
    )

    return tp, nil
}

// ========== 使用 Trace ==========
func handleRequest(ctx context.Context) {
    tracer := otel.Tracer("myapp")

    ctx, span := tracer.Start(ctx, "handleRequest")
    defer span.End()

    // 业务逻辑
    process(ctx)
}

// ========== Gin 集成 ==========
// 使用 go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin
import "go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"

func main() {
    r := gin.Default()
    r.Use(otelgin.Middleware("myapp"))
    r.Run(":8080")
}
```

### 结构化日志 (zap/zerolog)

```go
// ========== 使用 zap ==========
package main

import (
    "go.uber.org/zap"
    "go.uber.org/zap/zapcore"
)

func main() {
    // ========== 生产环境配置 ==========
    config := zap.NewProductionConfig()
    config.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
    config.OutputPaths = []string{"stdout", "/var/log/app.log"}

    logger, _ := config.Build()
    defer logger.Sync()

    // ========== 使用日志 ==========
    logger.Info("用户登录",
        zap.String("username", "alice"),
        zap.Int("user_id", 123),
    )

    logger.Error("数据库错误",
        zap.String("query", "SELECT * FROM users"),
        zap.Error(err),
    )

    // ========== 开发环境配置 ==========
    devLogger, _ := zap.NewDevelopment()
    devLogger.Debug("调试信息")
}

// ========== 使用 zerolog ==========
package main

import (
    "os"
    "github.com/rs/zerolog"
    "github.com/rs/zerolog/log"
)

func main() {
    // ========== 初始化 ==========
    zerolog.TimeFieldFormat = zerolog.TimeFormatISO8601

    log.Logger = log.Output(
        zerolog.ConsoleWriter{Out: os.Stdout},
    ).Level(zerolog.InfoLevel)

    // ========== 使用日志 ==========
    log.Info().
        Str("username", "alice").
        Int("user_id", 123).
        Msg("用户登录")

    log.Error().
        Err(err).
        Str("query", sql).
        Msg("数据库错误")

    // ========== 带上下文的日志 ==========
    contextLogger := log.With().
        Str("request_id", requestID).
        Logger()

    contextLogger.Info().Msg("处理请求")
}
```

### 健康检查端点

```go
package main

import (
    "context"
    "database/sql"
    "net/http"
    "time"
    "github.com/gin-gonic/gin"
)

// ========== 健康检查响应 ==========
type HealthResponse struct {
    Status    string            `json:"status"`
    Timestamp time.Time         `json:"timestamp"`
    Checks    map[string]string `json:"checks"`
}

// ========== 健康检查服务 ==========
type HealthService struct {
    db *sql.DB
}

func NewHealthService(db *sql.DB) *HealthService {
    return &HealthService{db: db}
}

func (s *HealthService) Check(ctx context.Context) HealthResponse {
    response := HealthResponse{
        Status:    "healthy",
        Timestamp: time.Now(),
        Checks:    make(map[string]string),
    }

    // 数据库检查
    if s.db != nil {
        if err := s.db.PingContext(ctx); err != nil {
            response.Checks["database"] = "unhealthy"
            response.Status = "unhealthy"
        } else {
            response.Checks["database"] = "healthy"
        }
    }

    // 可添加其他检查：Redis, MQ, 外部 API 等

    return response
}

// ========== 健康检查 Handler ==========
func HealthHandler(hs *HealthService) gin.HandlerFunc {
    return func(c *gin.Context) {
        ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
        defer cancel()

        health := hs.Check(ctx)

        status := http.StatusOK
        if health.Status != "healthy" {
            status = http.StatusServiceUnavailable
        }

        c.JSON(status, health)
    }
}

// ========== 存活/就绪检查 ==========
// Liveness: 应用是否运行
// Readiness: 应用是否准备好处理请求

func LivenessHandler(c *gin.Context) {
    c.JSON(http.StatusOK, gin.H{"status": "alive"})
}

func ReadinessHandler(hs *HealthService) gin.HandlerFunc {
    return func(c *gin.Context) {
        ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
        defer cancel()

        health := hs.Check(ctx)

        if health.Status != "healthy" {
            c.JSON(http.StatusServiceUnavailable, gin.H{"status": "not ready"})
            return
        }

        c.JSON(http.StatusOK, gin.H{"status": "ready"})
    }
}

func main() {
    r := gin.Default()

    hs := NewHealthService(db)

    // Kubernetes 探针
    r.GET("/healthz", LivenessHandler)      // liveness probe
    r.GET("/readyz", ReadinessHandler(hs))  // readiness probe
    r.GET("/health", HealthHandler(hs))     // 详细健康信息

    r.Run(":8080")
}
```

---

## 第六部分完

接下来可以继续学习第七部分：高级主题

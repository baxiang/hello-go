# 五、HTTP Server 详解

本章将详细介绍 Kratos 中 HTTP Server 的创建、启动和配置。

## 5.1 HTTP Server 创建

### 5.1.1 创建示例

```go
// 创建 HTTP Server
server := http.NewServer(
    http.Address(":8080"),
    http.Middleware(
        recovery.Recovery(),  // 错误恢复中间件
        logging.Server(),      // 日志中间件
    ),
)
```

### 5.1.2 Server 结构

```go
// http/server.go
type Server struct {
    // 网络配置
    network string
    address string
    timeout time.Duration
    
    // 中间件
    middlewares []Middleware
    
    // 路由
    router *Router
    
    // HTTP 服务器
    server *http.Server
    
    // 日志
    logger log.Logger
}
```

**图解 HTTP Server**：

```
┌─────────────────────────────────────────────────────────────┐
│                    HTTPServer 结构                          │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  network: "tcp"                                            │
│  address: ":8080"                                          │
│  timeout: 1s                                               │
│                                                             │
│  ┌─────────────────────────────────────────────────────┐   │
│  │  middlewares: []Middleware                           │   │
│  │  ┌──────────┐ ┌──────────┐ ┌──────────┐           │   │
│  │  │Recovery  │ │Logging   │ │Tracing   │  ...      │   │
│  │  └──────────┘ └──────────┘ └──────────┘           │   │
│  └─────────────────────────────────────────────────────┘   │
│                                                             │
│  ┌─────────────────────────────────────────────────────┐   │
│  │  router: *Router                                    │   │
│  │  ┌─────────────────────────────────────────────┐  │   │
│  │  │  trees: map[string]*node                       │  │   │
│  │  │  GET  → /api/user/*                          │  │   │
│  │  │  POST → /api/user/*                          │  │   │
│  │  │  PUT  → /api/user/*                          │  │   │
│  │  │  DELETE → /api/user/*                        │  │   │
│  │  └─────────────────────────────────────────────┘  │   │
│  └─────────────────────────────────────────────────────┘   │
│                                                             │
│  ┌─────────────────────────────────────────────────────┐   │
│  │  server: *http.Server                               │   │
│  │  Addr: ":8080"                                     │   │
│  │  Handler: router                                    │   │
│  │  ReadTimeout: 1s                                    │   │
│  │  WriteTimeout: 1s                                   │   │
│  └─────────────────────────────────────────────────────┘   │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

## 5.2 HTTP Server 启动

### 5.2.1 Start() 源码

```go
// http/server.go
func (s *Server) Start() error {
    // 1. 创建路由
    s.router = NewRouter(s.opts)
    
    // 2. 注册路由（通过 Option 传入）
    for _, f := range s.opt.funcs {
        f(s.router)
    }
    
    // 3. 创建 HTTP 服务器
    s.server = &http.Server{
        Addr:         s.address,
        Handler:       s.router,
        ReadTimeout:   s.timeout,
        WriteTimeout:  s.timeout,
    }
    
    // 4. 监听端口
    ln, err := net.Listen(s.network, s.address)
    if err != nil {
        return err
    }
    
    // 5. 启动服务
    go func() {
        if err := s.server.Serve(ln); err != nil {
            // 处理错误
        }
    }()
    
    return nil
}
```

### 5.2.2 启动流程图解

```
server.Start()
        │
        ▼
┌───────────────────┐
│  1. 创建路由      │  s.router = NewRouter()
└───────────────────┘
        │
        ▼
┌───────────────────┐
│  2. 注册路由      │  for f in funcs: f(router)
└───────────────────┘
        │
        ▼
┌───────────────────┐
│  3. 创建 Server  │  s.server = &http.Server{...}
└───────────────────┘
        │
        ▼
┌───────────────────┐
│  4. 监听端口      │  net.Listen("tcp", ":8080")
└───────────────────┘
        │
        ▼
┌───────────────────┐
│  5. 启动服务      │  go server.Serve(ln)
└───────────────────┘
```

## 5.3 Server 选项

### 5.3.1 常用选项

```go
// 网络配置
http.Network("tcp")
http.Address(":8080")
http.Timeout(time.Second)

// 中间件
http.Middleware(
    recovery.Recovery(),
    logging.Server(),
    tracing.Server(),
)

// 路由选项
http.Filter(func(h http.Handler) http.Handler {
    return h
})

// CORS
http.CORS(
    func(header http.Header) {
        header.Set("Access-Control-Allow-Origin", "*")
    },
)
```

### 5.3.2 选项模式

Kratos 使用选项模式来配置 Server：

```go
// 选项定义
type Option func(*Server)

// 示例选项
func Address(addr string) Option {
    return func(s *Server) {
        s.address = addr
    }
}

func Middleware(m ...Middleware) Option {
    return func(s *Server) {
        s.middlewares = append(s.middlewares, m...)
    }
}

// 使用选项
http.NewServer(
    Address(":8080"),
    Middleware(recovery.Recovery()),
)
```

## 5.4 本章小结

```
┌─────────────────────────────────────────────────────────────┐
│                      本章总结                                │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  ✓ HTTP Server 是 Kratos 的核心组件之一                     │
│                                                             │
│  ✓ 包含路由、中间件、HTTP 服务器等组件                      │
│                                                             │
│  ✓ 使用选项模式进行配置                                     │
│                                                             │
│  ✓ 启动流程：创建路由 → 注册路由 → 启动监听                 │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

---

## 下章预告

下一章我们将学习 **路由系统详解**。

敬请期待！
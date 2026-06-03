# Select 企业生产级实战案例

> 本章收集了来自 Kubernetes、etcd、Docker、微服务等真实生产环境中的 select 使用案例，帮助你理解如何在企业级项目中正确使用 select。

---

## 目录

- [1. 超时重试机制](#1-超时重试机制)
- [2. 服务健康检查](#2-服务健康检查)
- [3. 连接池管理](#3-连接池管理)
- [4. 限流器实现](#4-限流器实现)
- [5. 优雅关闭](#5-优雅关闭)
- [6. 多数据源同步](#6-多数据源同步)
- [7. 任务调度器](#7-任务调度器)
- [8. 监控告警系统](#8-监控告警系统)

---

## 1. 超时重试机制

### 1.1 场景描述

在微服务调用中，网络请求可能失败，需要实现带超时和重试的机制。

### 1.2 生产级代码

```go
package httpclient

import (
    "context"
    "errors"
    "fmt"
    "net/http"
    "time"
)

// HTTPClient 带重试的 HTTP 客户端
type HTTPClient struct {
    client      *http.Client
    maxRetries  int
    timeout     time.Duration
    backoff     time.Duration
}

// NewHTTPClient 创建客户端
func NewHTTPClient(timeout time.Duration, maxRetries int) *HTTPClient {
    return &HTTPClient{
        client: &http.Client{
            Timeout: timeout,
        },
        maxRetries: maxRetries,
        timeout:    timeout,
        backoff:    100 * time.Millisecond,
    }
}

// Do 执行 HTTP 请求（带重试）
func (c *HTTPClient) Do(ctx context.Context, req *http.Request) (*http.Response, error) {
    var lastErr error
    
    for attempt := 0; attempt < c.maxRetries; attempt++ {
        // 创建带超时的上下文
        reqCtx, cancel := context.WithTimeout(ctx, c.timeout)
        
        // 执行请求
        respCh := make(chan *http.Response, 1)
        errCh := make(chan error, 1)
        
        go func() {
            resp, err := c.client.Do(req.WithContext(reqCtx))
            if err != nil {
                errCh <- err
            } else {
                respCh <- resp
            }
        }()
        
        // 使用 select 等待结果或超时
        select {
        case resp := <-respCh:
            cancel()
            // 请求成功
            if resp.StatusCode < 500 {
                return resp, nil
            }
            // 服务器错误，需要重试
            resp.Body.Close()
            lastErr = fmt.Errorf("服务器错误：%d", resp.StatusCode)
            
        case err := <-errCh:
            cancel()
            lastErr = err
            
        case <-ctx.Done():
            cancel()
            return nil, ctx.Err()
            
        case <-reqCtx.Done():
            cancel()
            lastErr = fmt.Errorf("请求超时")
        }
        
        // 重试前等待（指数退避）
        if attempt < c.maxRetries-1 {
            backoff := c.backoff * time.Duration(1<<uint(attempt))
            select {
            case <-time.After(backoff):
            case <-ctx.Done():
                return nil, ctx.Err()
            }
        }
    }
    
    return nil, fmt.Errorf("重试 %d 次后失败：%w", c.maxRetries, lastErr)
}
```

### 1.3 使用示例

```go
client := NewHTTPClient(5*time.Second, 3)

req, _ := http.NewRequest("GET", "http://api.example.com/users", nil)

resp, err := client.Do(context.Background(), req)
if err != nil {
    log.Printf("请求失败：%v", err)
    return
}
defer resp.Body.Close()
```

---

## 2. 服务健康检查

### 2.1 场景描述

微服务架构中，需要定期检查各个服务的健康状态。

### 2.2 生产级代码（参考 Kubernetes）

```go
package health

import (
    "context"
    "sync"
    "time"
)

// HealthChecker 健康检查器
type HealthChecker struct {
    services  map[string]*Service
    interval  time.Duration
    timeout   time.Duration
    resultCh  chan HealthResult
    done      chan struct{}
    wg        sync.WaitGroup
}

// Service 服务信息
type Service struct {
    Name      string
    Endpoint  string
    LastCheck time.Time
    Healthy   bool
}

// HealthResult 健康检查结果
type HealthResult struct {
    ServiceName string
    Healthy     bool
    Latency     time.Duration
    Error       error
}

// NewHealthChecker 创建健康检查器
func NewHealthChecker(interval, timeout time.Duration) *HealthChecker {
    return &HealthChecker{
        services: make(map[string]*Service),
        interval: interval,
        timeout:  timeout,
        resultCh: make(chan HealthResult, 100),
        done:     make(chan struct{}),
    }
}

// AddService 添加服务
func (hc *HealthChecker) AddService(name, endpoint string) {
    hc.services[name] = &Service{
        Name:     name,
        Endpoint: endpoint,
    }
}

// Start 启动健康检查
func (hc *HealthChecker) Start(ctx context.Context) {
    ticker := time.NewTicker(hc.interval)
    defer ticker.Stop()
    
    for {
        select {
        case <-ticker.C:
            // 定期检查所有服务
            for _, svc := range hc.services {
                hc.wg.Add(1)
                go hc.checkService(svc)
            }
            
        case result := <-hc.resultCh:
            // 处理检查结果
            hc.processResult(result)
            
        case <-ctx.Done():
            // 收到退出信号
            close(hc.done)
            hc.wg.Wait()
            return
        }
    }
}

// checkService 检查单个服务
func (hc *HealthChecker) checkService(svc *Service) {
    defer hc.wg.Done()
    
    start := time.Now()
    
    // 创建带超时的上下文
    ctx, cancel := context.WithTimeout(context.Background(), hc.timeout)
    defer cancel()
    
    // 模拟健康检查（实际是 HTTP 请求）
    done := make(chan error, 1)
    go func() {
        // 实际代码：resp, err := http.Get(svc.Endpoint + "/health")
        time.Sleep(50 * time.Millisecond) // 模拟
        done <- nil
    }()
    
    select {
    case err := <-done:
        latency := time.Since(start)
        hc.resultCh <- HealthResult{
            ServiceName: svc.Name,
            Healthy:     err == nil,
            Latency:     latency,
            Error:       err,
        }
        
    case <-ctx.Done():
        hc.resultCh <- HealthResult{
            ServiceName: svc.Name,
            Healthy:     false,
            Error:       ctx.Err(),
        }
    }
    
    // 更新服务状态
    svc.LastCheck = time.Now()
}

// processResult 处理检查结果
func (hc *HealthChecker) processResult(result HealthResult) {
    if svc, ok := hc.services[result.ServiceName]; ok {
        svc.Healthy = result.Healthy
        
        if !result.Healthy {
            // 发送告警
            log.Printf("⚠️  服务 %s 不健康：%v", result.ServiceName, result.Error)
        }
    }
}

// GetHealthyServices 获取健康的服务列表
func (hc *HealthChecker) GetHealthyServices() []string {
    var healthy []string
    for name, svc := range hc.services {
        if svc.Healthy {
            healthy = append(healthy, name)
        }
    }
    return healthy
}
```

---

## 3. 连接池管理

### 3.1 场景描述

数据库连接池需要管理有限数量的连接，支持获取、归还和超时。

### 3.2 生产级代码

```go
package pool

import (
    "context"
    "errors"
    "sync"
    "time"
)

// Connection 连接接口
type Connection interface {
    Close() error
    IsClosed() bool
}

// ConnectionPool 连接池
type ConnectionPool struct {
    connections chan Connection
    factory     func() (Connection, error)
    maxSize     int
    timeout     time.Duration
    mu          sync.Mutex
    closed      bool
}

// NewConnectionPool 创建连接池
func NewConnectionPool(
    maxSize int,
    timeout time.Duration,
    factory func() (Connection, error),
) *ConnectionPool {
    pool := &ConnectionPool{
        connections: make(chan Connection, maxSize),
        maxSize:     maxSize,
        timeout:     timeout,
        factory:     factory,
    }
    
    // 预创建连接
    for i := 0; i < maxSize; i++ {
        if conn, err := factory(); err == nil {
            pool.connections <- conn
        }
    }
    
    return pool
}

// Get 获取连接（带超时）
func (p *ConnectionPool) Get(ctx context.Context) (Connection, error) {
    p.mu.Lock()
    if p.closed {
        p.mu.Unlock()
        return nil, errors.New("连接池已关闭")
    }
    p.mu.Unlock()
    
    select {
    case conn := <-p.connections:
        // 检查连接是否有效
        if conn.IsClosed() {
            // 创建新连接
            if newConn, err := p.factory(); err == nil {
                return newConn, nil
            }
            // 重试获取
            return p.Get(ctx)
        }
        return conn, nil
        
    case <-ctx.Done():
        return nil, ctx.Err()
        
    case <-time.After(p.timeout):
        return nil, errors.New("获取连接超时")
    }
}

// Put 归还连接
func (p *ConnectionPool) Put(conn Connection) error {
    p.mu.Lock()
    defer p.mu.Unlock()
    
    if p.closed || conn.IsClosed() {
        return conn.Close()
    }
    
    select {
    case p.connections <- conn:
        return nil
    default:
        // 连接池已满，关闭连接
        return conn.Close()
    }
}

// Close 关闭连接池
func (p *ConnectionPool) Close() error {
    p.mu.Lock()
    defer p.mu.Unlock()
    
    if p.closed {
        return nil
    }
    p.closed = true
    
    close(p.connections)
    
    // 关闭所有连接
    for conn := range p.connections {
        conn.Close()
    }
    
    return nil
}

// Stats 连接池统计
func (p *ConnectionPool) Stats() map[string]int {
    return map[string]int{
        "available": len(p.connections),
        "max":       p.maxSize,
        "in_use":    p.maxSize - len(p.connections),
    }
}
```

### 3.3 使用示例

```go
// 创建数据库连接池
pool := NewConnectionPool(
    10,              // 最大连接数
    5*time.Second,   // 获取超时
    func() (Connection, error) {
        return sql.Open("mysql", dsn)
    },
)

// 获取连接
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

conn, err := pool.Get(ctx)
if err != nil {
    log.Printf("获取连接失败：%v", err)
    return
}
defer pool.Put(conn)

// 使用连接...
```

---

## 4. 限流器实现

### 4.1 场景描述

API 网关需要限制请求速率，防止系统过载。

### 4.2 生产级代码（令牌桶算法）

```go
package ratelimit

import (
    "context"
    "sync"
    "time"
)

// RateLimiter 限流器
type RateLimiter struct {
    tokens     chan struct{}
    rate       int
    burst      int
    mu         sync.Mutex
    closed     bool
}

// NewRateLimiter 创建限流器
func NewRateLimiter(rate, burst int) *RateLimiter {
    rl := &RateLimiter{
        tokens: make(chan struct{}, burst),
        rate:   rate,
        burst:  burst,
    }
    
    // 初始化令牌桶
    for i := 0; i < burst; i++ {
        rl.tokens <- struct{}{}
    }
    
    // 启动令牌补充协程
    go rl.refill()
    
    return rl
}

// refill 定期补充令牌
func (rl *RateLimiter) refill() {
    ticker := time.NewTicker(time.Second / time.Duration(rl.rate))
    defer ticker.Stop()
    
    for range ticker.C {
        rl.mu.Lock()
        if rl.closed {
            rl.mu.Unlock()
            return
        }
        rl.mu.Unlock()
        
        select {
        case rl.tokens <- struct{}{}:
            // 添加令牌成功
        default:
            // 令牌桶已满
        }
    }
}

// Allow 允许请求通过
func (rl *RateLimiter) Allow() bool {
    select {
    case <-rl.tokens:
        return true
    default:
        return false
    }
}

// Wait 等待令牌（阻塞）
func (rl *RateLimiter) Wait(ctx context.Context) error {
    select {
    case <-rl.tokens:
        return nil
    case <-ctx.Done():
        return ctx.Err()
    }
}

// WaitTimeout 等待令牌（带超时）
func (rl *RateLimiter) WaitTimeout(timeout time.Duration) error {
    select {
    case <-rl.tokens:
        return nil
    case <-time.After(timeout):
        return errors.New("等待令牌超时")
    }
}

// Close 关闭限流器
func (rl *RateLimiter) Close() {
    rl.mu.Lock()
    defer rl.mu.Unlock()
    rl.closed = true
}
```

### 4.3 使用示例

```go
// 创建限流器：每秒 100 个请求，突发 200
limiter := NewRateLimiter(100, 200)

// HTTP 中间件
func RateLimitMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if !limiter.Allow() {
            http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
            return
        }
        next.ServeHTTP(w, r)
    })
}
```

---

## 5. 优雅关闭

### 5.1 场景描述

服务需要优雅地处理关闭信号，完成正在处理的请求后再退出。

### 5.2 生产级代码

```go
package server

import (
    "context"
    "log"
    "net/http"
    "os"
    "os/signal"
    "sync"
    "syscall"
    "time"
)

// Server 服务器
type Server struct {
    httpServer *http.Server
    wg         sync.WaitGroup
    done       chan struct{}
}

// NewServer 创建服务器
func NewServer(addr string, handler http.Handler) *Server {
    return &Server{
        httpServer: &http.Server{
            Addr:         addr,
            Handler:      handler,
            ReadTimeout:  15 * time.Second,
            WriteTimeout: 15 * time.Second,
            IdleTimeout:  60 * time.Second,
        },
        done: make(chan struct{}),
    }
}

// Start 启动服务器
func (s *Server) Start() error {
    // 启动 HTTP 服务器
    go func() {
        log.Printf("服务器启动在 %s", s.httpServer.Addr)
        if err := s.httpServer.ListenAndServe(); err != http.ErrServerClosed {
            log.Printf("HTTP 服务器错误：%v", err)
        }
        close(s.done)
    }()
    
    return nil
}

// Shutdown 优雅关闭
func (s *Server) Shutdown(timeout time.Duration) error {
    log.Println("开始优雅关闭...")
    
    // 创建带超时的上下文
    ctx, cancel := context.WithTimeout(context.Background(), timeout)
    defer cancel()
    
    // 等待正在处理的请求完成
    done := make(chan struct{})
    go func() {
        s.wg.Wait()
        close(done)
    }()
    
    // 使用 select 等待或超时
    select {
    case <-done:
        log.Println("所有请求处理完成")
    case <-ctx.Done():
        log.Println("等待请求完成超时")
    }
    
    // 关闭 HTTP 服务器
    if err := s.httpServer.Shutdown(ctx); err != nil {
        return err
    }
    
    <-s.done
    log.Println("服务器已关闭")
    return nil
}

// WaitForSignal 等待退出信号
func (s *Server) WaitForSignal() os.Signal {
    sigCh := make(chan os.Signal, 1)
    signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
    return <-sigCh
}

// Run 运行服务器（带信号处理）
func (s *Server) Run() error {
    if err := s.Start(); err != nil {
        return err
    }
    
    // 等待退出信号
    sig := s.WaitForSignal()
    log.Printf("收到信号：%v", sig)
    
    // 优雅关闭
    return s.Shutdown(30 * time.Second)
}
```

### 5.3 使用示例

```go
handler := http.NewServeMux()
handler.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("OK"))
})

server := NewServer(":8080", handler)

if err := server.Run(); err != nil {
    log.Fatalf("服务器错误：%v", err)
}
```

---

## 6. 多数据源同步

### 6.1 场景描述

需要从多个数据源（数据库、缓存、外部 API）获取数据，哪个先返回就用哪个。

### 6.2 生产级代码

```go
package data

import (
    "context"
    "errors"
    "sync"
    "time"
)

// DataSource 数据源接口
type DataSource interface {
    Name() string
    Fetch(ctx context.Context, key string) (string, error)
}

// DataSync 数据同步器
type DataSync struct {
    sources []DataSource
    timeout time.Duration
}

// NewDataSync 创建数据同步器
func NewDataSync(timeout time.Duration, sources ...DataSource) *DataSync {
    return &DataSync{
        sources: sources,
        timeout: timeout,
    }
}

// FetchFirst 获取第一个成功的数据
func (ds *DataSync) FetchFirst(ctx context.Context, key string) (string, string, error) {
    ctx, cancel := context.WithTimeout(ctx, ds.timeout)
    defer cancel()
    
    resultCh := make(chan struct {
        source string
        data   string
        err    error
    }, len(ds.sources))
    
    // 从所有数据源并行获取
    var wg sync.WaitGroup
    for _, source := range ds.sources {
        wg.Add(1)
        go func(s DataSource) {
            defer wg.Done()
            data, err := s.Fetch(ctx, key)
            resultCh <- struct {
                source string
                data   string
                err    error
            }{s.Name(), data, err}
        }(source)
    }
    
    // 等待第一个成功结果
    var lastErr error
    received := 0
    
    for {
        select {
        case result := <-resultCh:
            received++
            if result.err == nil {
                return result.source, result.data, nil
            }
            lastErr = result.err
            
            // 所有数据源都失败
            if received == len(ds.sources) {
                return "", "", lastErr
            }
            
        case <-ctx.Done():
            return "", "", ctx.Err()
        }
    }
}

// FetchAll 获取所有数据源的数据
func (ds *DataSync) FetchAll(ctx context.Context, key string) (map[string]string, []error) {
    ctx, cancel := context.WithTimeout(ctx, ds.timeout)
    defer cancel()
    
    resultCh := make(chan struct {
        source string
        data   string
        err    error
    }, len(ds.sources))
    
    // 从所有数据源并行获取
    for _, source := range ds.sources {
        go func(s DataSource) {
            data, err := s.Fetch(ctx, key)
            resultCh <- struct {
                source string
                data   string
                err    error
            }{s.Name(), data, err}
        }(source)
    }
    
    results := make(map[string]string)
    var errs []error
    
    for i := 0; i < len(ds.sources); i++ {
        select {
        case result := <-resultCh:
            if result.err == nil {
                results[result.source] = result.data
            } else {
                errs = append(errs, result.err)
            }
        case <-ctx.Done():
            errs = append(errs, ctx.Err())
            break
        }
    }
    
    return results, errs
}
```

### 6.3 使用示例

```go
// 创建数据源
sources := []DataSource{
    NewRedisSource(redisClient),
    NewDatabaseSource(db),
    NewAPISource(apiClient),
}

// 创建同步器
sync := NewDataSync(100*time.Millisecond, sources...)

// 获取数据（哪个快用哪个）
source, data, err := sync.FetchFirst(context.Background(), "user:123")
if err != nil {
    log.Printf("获取数据失败：%v", err)
    return
}
log.Printf("从 %s 获取数据：%s", source, data)
```

---

## 7. 任务调度器

### 7.1 场景描述

需要定时执行任务，支持动态添加、删除任务。

### 7.2 生产级代码

```go
package scheduler

import (
    "context"
    "sync"
    "time"
)

// Task 任务
type Task struct {
    ID       string
    Interval time.Duration
    Fn       func()
}

// Scheduler 任务调度器
type Scheduler struct {
    tasks     map[string]*Task
    taskCh    chan *Task
    removeCh  chan string
    done      chan struct{}
    wg        sync.WaitGroup
    mu        sync.RWMutex
}

// NewScheduler 创建调度器
func NewScheduler() *Scheduler {
    s := &Scheduler{
        tasks:    make(map[string]*Task),
        taskCh:   make(chan *Task, 100),
        removeCh: make(chan string, 100),
        done:     make(chan struct{}),
    }
    
    // 启动调度协程
    s.wg.Add(1)
    go s.run()
    
    return s
}

// Add 添加任务
func (s *Scheduler) Add(task *Task) {
    s.taskCh <- task
}

// Remove 移除任务
func (s *Scheduler) Remove(taskID string) {
    s.removeCh <- taskID
}

// run 调度主循环
func (s *Scheduler) run() {
    defer s.wg.Done()
    
    // 定时器映射
    timers := make(map[string]*time.Timer)
    
    for {
        select {
        case task := <-s.taskCh:
            // 添加新任务
            s.mu.Lock()
            if oldTask, ok := s.tasks[task.ID]; ok {
                // 停止旧定时器
                if timer, ok := timers[task.ID]; ok {
                    timer.Stop()
                }
                _ = oldTask
            }
            s.tasks[task.ID] = task
            s.mu.Unlock()
            
            // 创建新定时器
            timers[task.ID] = time.NewTimer(task.Interval)
            
            // 启动任务协程
            go s.runTask(task, timers[task.ID])
            
        case taskID := <-s.removeCh:
            // 移除任务
            s.mu.Lock()
            delete(s.tasks, taskID)
            s.mu.Unlock()
            
            // 停止定时器
            if timer, ok := timers[taskID]; ok {
                timer.Stop()
                delete(timers, taskID)
            }
            
        case <-s.done:
            // 关闭调度器
            for _, timer := range timers {
                timer.Stop()
            }
            return
        }
    }
}

// runTask 运行任务
func (s *Scheduler) runTask(task *Task, timer *time.Timer) {
    for {
        select {
        case <-timer.C:
            // 执行任务
            task.Fn()
            // 重置定时器
            timer.Reset(task.Interval)
            
        case <-s.done:
            return
        }
    }
}

// Close 关闭调度器
func (s *Scheduler) Close() {
    close(s.done)
    s.wg.Wait()
}
```

### 7.3 使用示例

```go
scheduler := NewScheduler()

// 添加定时任务
scheduler.Add(&Task{
    ID:       "cleanup",
    Interval: time.Hour,
    Fn: func() {
        log.Println("执行清理任务")
    },
})

scheduler.Add(&Task{
    ID:       "report",
    Interval: 24 * time.Hour,
    Fn: func() {
        log.Println("生成日报")
    },
})

// 运行一段时间后移除任务
time.Sleep(time.Hour)
scheduler.Remove("cleanup")

// 关闭调度器
defer scheduler.Close()
```

---

## 8. 监控告警系统

### 8.1 场景描述

监控系统需要从多个指标源收集数据，超过阈值时发送告警。

### 8.2 生产级代码

```go
package monitor

import (
    "context"
    "log"
    "sync"
    "time"
)

// Metric 指标
type Metric struct {
    Name      string
    Value     float64
    Threshold float64
    Timestamp time.Time
}

// Alert 告警
type Alert struct {
    Metric    string
    Value     float64
    Threshold float64
    Message   string
}

// Monitor 监控器
type Monitor struct {
    metricsCh   chan Metric
    alertCh     chan Alert
    alertFunc   func(Alert)
    interval    time.Duration
    done        chan struct{}
    wg          sync.WaitGroup
    thresholds  map[string]float64
    mu          sync.RWMutex
}

// NewMonitor 创建监控器
func NewMonitor(interval time.Duration, alertFunc func(Alert)) *Monitor {
    m := &Monitor{
        metricsCh:  make(chan Metric, 1000),
        alertCh:    make(chan Alert, 100),
        alertFunc:  alertFunc,
        interval:   interval,
        done:       make(chan struct{}),
        thresholds: make(map[string]float64),
    }
    
    // 启动指标收集
    m.wg.Add(1)
    go m.collectMetrics()
    
    // 启动告警处理
    m.wg.Add(1)
    go m.processAlerts()
    
    return m
}

// SetThreshold 设置阈值
func (m *Monitor) SetThreshold(metricName string, threshold float64) {
    m.mu.Lock()
    defer m.mu.Unlock()
    m.thresholds[metricName] = threshold
}

// ReportMetric 上报指标
func (m *Monitor) ReportMetric(name string, value float64) {
    m.metricsCh <- Metric{
        Name:      name,
        Value:     value,
        Threshold: m.getThreshold(name),
        Timestamp: time.Now(),
    }
}

// getThreshold 获取阈值
func (m *Monitor) getThreshold(name string) float64 {
    m.mu.RLock()
    defer m.mu.RUnlock()
    if t, ok := m.thresholds[name]; ok {
        return t
    }
    return 0
}

// collectMetrics 收集指标
func (m *Monitor) collectMetrics() {
    defer m.wg.Done()
    
    ticker := time.NewTicker(m.interval)
    defer ticker.Stop()
    
    for {
        select {
        case metric := <-m.metricsCh:
            // 检查是否超过阈值
            if metric.Value > metric.Threshold && metric.Threshold > 0 {
                m.alertCh <- Alert{
                    Metric:    metric.Name,
                    Value:     metric.Value,
                    Threshold: metric.Threshold,
                    Message:   metric.Name + " 超过阈值",
                }
            }
            
        case <-ticker.C:
            // 定期处理
            log.Println("监控检查...")
            
        case <-m.done:
            return
        }
    }
}

// processAlerts 处理告警
func (m *Monitor) processAlerts() {
    defer m.wg.Done()
    
    // 告警去重
    lastAlert := make(map[string]time.Time)
    
    for {
        select {
        case alert := <-m.alertCh:
            // 告警去重（5 分钟内不重复告警）
            if lastTime, ok := lastAlert[alert.Metric]; ok {
                if time.Since(lastTime) < 5*time.Minute {
                    continue
                }
            }
            
            // 发送告警
            if m.alertFunc != nil {
                m.alertFunc(alert)
            }
            lastAlert[alert.Metric] = time.Now()
            
        case <-m.done:
            return
        }
    }
}

// Close 关闭监控器
func (m *Monitor) Close() {
    close(m.done)
    m.wg.Wait()
}
```

### 8.3 使用示例

```go
// 创建监控器
monitor := NewMonitor(time.Second, func(alert Alert) {
    log.Printf("⚠️  告警：%s = %.2f (阈值：%.2f)",
        alert.Metric, alert.Value, alert.Threshold)
    
    // 发送短信/邮件/钉钉
    // sendSMS(...)
})

// 设置阈值
monitor.SetThreshold("cpu_usage", 80.0)
monitor.SetThreshold("memory_usage", 90.0)
monitor.SetThreshold("request_latency", 1000.0)

// 上报指标
go func() {
    for {
        monitor.ReportMetric("cpu_usage", 75.5)
        monitor.ReportMetric("memory_usage", 85.2)
        time.Sleep(time.Second)
    }
}()

// 运行
time.Sleep(time.Minute)
monitor.Close()
```

---

## 9. 本章小结

```
┌─────────────────────────────────────────────────────────────┐
│                  企业级 select 使用要点                      │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  ✓ 超时控制：始终使用 context 或 time.After               │
│                                                             │
│  ✓ 资源清理：使用 defer 确保资源释放                       │
│                                                             │
│  ✓ 优雅关闭：监听退出信号，完成正在处理的任务              │
│                                                             │
│  ✓ 错误处理：记录详细错误信息，便于排查                    │
│                                                             │
│  ✓ 并发安全：使用 mutex 保护共享数据                       │
│                                                             │
│  ✓ 性能考虑：合理设置 channel 缓冲区大小                   │
│                                                             │
│  ✓ 监控告警：关键路径添加监控和告警                        │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

---

[上一章：← Select 底层原理](./04-Select底层原理.md) | [下一章：同步原语详解 →](./06-同步原语详解.md)
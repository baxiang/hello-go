# Exporter 开发

Exporter 将第三方系统的指标转换为 Prometheus 格式。

## Exporter 类型

- **直接 instrumentation**：应用内置指标
- **独立 Exporter**：独立进程抓取指标

## Go Exporter 开发

### 基本结构

```go
package main

import (
    "net/http"
    
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
    // 注册默认指标
    prometheus.MustRegister(prometheus.NewBuildInfoCollector())
    
    // 暴露指标
    http.Handle("/metrics", promhttp.Handler())
    http.ListenAndServe(":8080", nil)
}
```

### Counter 示例

```go
var (
    httpRequestsTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "http_requests_total",
            Help: "Total number of HTTP requests",
        },
        []string{"method", "path", "status"},
    )
)

func init() {
    prometheus.MustRegister(httpRequestsTotal)
}

func handler(w http.ResponseWriter, r *http.Request) {
    // 处理请求...
    
    // 记录指标
    httpRequestsTotal.WithLabelValues(
        r.Method,
        r.URL.Path,
        "200",
    ).Inc()
    
    w.Write([]byte("OK"))
}
```

### Gauge 示例

```go
var (
    activeConnections = prometheus.NewGauge(prometheus.GaugeOpts{
        Name: "active_connections",
        Help: "Number of active connections",
    })
)

func handleConnection(conn net.Conn) {
    activeConnections.Inc()
    defer activeConnections.Dec()
    
    // 处理连接...
}
```

### Histogram 示例

```go
var (
    requestDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "http_request_duration_seconds",
            Help:    "HTTP request duration in seconds",
            Buckets: []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
        },
        []string{"method", "path"},
    )
)

func handler(w http.ResponseWriter, r *http.Request) {
    start := time.Now()
    
    // 处理请求...
    
    duration := time.Since(start).Seconds()
    requestDuration.WithLabelValues(
        r.Method,
        r.URL.Path,
    ).Observe(duration)
}
```

### Summary 示例

```go
var (
    responseSize = prometheus.NewSummary(prometheus.SummaryOpts{
        Name:       "http_response_size_bytes",
        Help:       "HTTP response size in bytes",
        Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
    })
)
```

## 自定义 Collector

```go
type MyCollector struct {
    metric *prometheus.Desc
}

func NewMyCollector() *MyCollector {
    return &MyCollector{
        metric: prometheus.NewDesc(
            "my_metric_total",
            "My custom metric",
            []string{"label1"},
            nil,
        ),
    }
}

func (c *MyCollector) Describe(ch chan<- *prometheus.Desc) {
    ch <- c.metric
}

func (c *MyCollector) Collect(ch chan<- prometheus.Metric) {
    // 收集指标值
    value := getMetricValue()
    
    ch <- prometheus.MustNewConstMetric(
        c.metric,
        prometheus.CounterValue,
        value,
        "label_value",
    )
}

// 注册
prometheus.MustRegister(NewMyCollector())
```

## 完整 Web 应用示例

```go
package main

import (
    "net/http"
    "time"
    
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
    httpRequestsTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "http_requests_total",
            Help: "Total HTTP requests",
        },
        []string{"method", "path", "status"},
    )
    
    httpRequestDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "http_request_duration_seconds",
            Help:    "HTTP request duration",
            Buckets: prometheus.DefBuckets,
        },
        []string{"method", "path"},
    )
)

func init() {
    prometheus.MustRegister(httpRequestsTotal)
    prometheus.MustRegister(httpRequestDuration)
}

func instrumentMiddleware(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        
        // 包装 ResponseWriter 以获取状态码
        wrapped := &responseWriter{ResponseWriter: w}
        
        next(wrapped, r)
        
        duration := time.Since(start).Seconds()
        
        httpRequestDuration.WithLabelValues(
            r.Method,
            r.URL.Path,
        ).Observe(duration)
        
        httpRequestsTotal.WithLabelValues(
            r.Method,
            r.URL.Path,
            string(rune(wrapped.statusCode)),
        ).Inc()
    }
}

func main() {
    http.HandleFunc("/api", instrumentMiddleware(apiHandler))
    http.Handle("/metrics", promhttp.Handler())
    
    http.ListenAndServe(":8080", nil)
}
```

## 最佳实践

1. **命名规范**：
   - 使用 `_total` 后缀表示 Counter
   - 使用单位后缀（`_seconds`, `_bytes`）
   - 使用基本单位（秒、字节）

2. **标签使用**：
   - 避免高基数标签（user_id, request_id）
   - 使用有意义的标签名

3. **Histogram vs Summary**：
   - Histogram：可聚合，适合服务级别
   - Summary：精确分位数，适合应用级别

4. **文档完整**：为每个指标添加 Help

5. **避免阻塞**：Collect 方法不应阻塞

参考：https://prometheus.io/docs/instrumenting/writing_exporters/
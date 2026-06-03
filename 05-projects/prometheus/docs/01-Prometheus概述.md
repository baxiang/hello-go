# Prometheus 概述

Prometheus 是开源的监控和告警系统，具有强大的数据模型和查询语言。

## 核心特性

- **多维数据模型**：指标名称和键值对标签
- **PromQL**：强大的查询语言
- **拉取模式**：主动抓取指标
- **服务发现**：自动发现监控目标

## 架构

```
Prometheus Server
     │
     ├─ Pull Metrics
     │
┌────┴────┬─────────┬─────────┐
│         │         │         │
App1      App2      Node     MySQL
Exporter  Exporter  Exporter  Exporter
```

## 快速开始

### Docker 部署

```bash
docker run -d --name prometheus \
  -p 9090:9090 \
  prom/prometheus:v2.51.0
```

### prometheus.yml

```yaml
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: 'prometheus'
    static_configs:
      - targets: ['localhost:9090']
```

### Go 应用暴露指标

```go
package main

import (
    "net/http"
    "github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
    http.Handle("/metrics", promhttp.Handler())
    http.ListenAndServe(":8080", nil)
}
```

## 数据模型

```
指标名称{标签1="值1", 标签2="值2"} 值

示例：
http_requests_total{method="GET", status="200"} 1234
```

## 版本信息

| 组件 | 版本 |
|------|------|
| Prometheus | 2.51.0 |
| Go client | 1.19.1 |

在下一章中，我们将学习数据模型和PromQL。

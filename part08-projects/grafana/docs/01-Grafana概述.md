# Grafana 概述

Grafana 是开源的数据可视化和监控平台，支持多种数据源。

## 核心特性

- **多数据源**：Prometheus、InfluxDB、MySQL 等
- **可视化**：丰富的图表和面板类型
- **告警**：灵活的告警规则和通知
- **模板**：可重用的仪表板

## 快速开始

### Docker 部署

```bash
docker run -d --name grafana \
  -p 3000:3000 \
  -e "GF_SECURITY_ADMIN_PASSWORD=admin" \
  grafana/grafana:10.4.0
```

### 访问

- URL: http://localhost:3000
- 用户名: admin
- 密码: admin

### 配置数据源

1. 添加数据源 → Prometheus
2. URL: http://prometheus:9090
3. Save & Test

## 创建面板

### Query

```promql
rate(http_requests_total[5m])
```

### 可视化类型

- Time series：时间序列图
- Stat：统计值
- Table：表格
- Gauge：仪表盘

## 版本信息

| 组件 | 版本 |
|------|------|
| Grafana | 10.4.0 |

在下一章中，我们将学习数据源配置。

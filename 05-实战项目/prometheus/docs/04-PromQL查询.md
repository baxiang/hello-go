# PromQL 查询

PromQL（Prometheus Query Language）是 Prometheus 的查询语言。

## 基本查询

### 瞬时查询

```promql
# 查询指标
http_requests_total

# 查询特定标签
http_requests_total{method="GET"}

# 多标签匹配
http_requests_total{method="GET", status="200"}

# 正则匹配
http_requests_total{method=~"GET|POST"}

# 排除匹配
http_requests_total{status!="500"}
```

### 范围查询

```promql
# 5分钟内的数据
http_requests_total[5m]

# 1小时内的数据
http_requests_total[1h]
```

## 时间偏移

```promql
# 5分钟前的数据
http_requests_total offset 5m

# 昨天的数据
http_requests_total offset 1d
```

## 聚合操作

### 常用聚合

```promql
# 求和
sum(http_requests_total)

# 按标签分组求和
sum by (method) (http_requests_total)

# 平均值
avg(http_requests_total)

# 最大/最小值
max(http_requests_total)
min(http_requests_total)

# 计数
count(http_requests_total)
```

### 聚合示例

```promql
# 按实例分组统计 QPS
sum by (instance) (rate(http_requests_total[5m]))

# 统计状态码分布
sum by (status) (http_requests_total)
```

## 数学运算

```promql
# 加法
http_requests_total + 100

# 减法
http_requests_total - 100

# 乘法
http_requests_total * 2

# 除法（计算错误率）
http_errors_total / http_requests_total * 100
```

## 比较运算

```promql
# 大于
http_requests_total > 100

# 小于
http_requests_total < 100

# 等于
http_requests_total == 100

# 范围
http_requests_total > 100 and http_requests_total < 1000
```

## 内置函数

### rate（速率）

计算每秒增长率。

```promql
# 每秒请求数
rate(http_requests_total[5m])

# QPS
rate(http_requests_total[1m])
```

### irate（瞬时速率）

计算瞬时增长率。

```promql
irate(http_requests_total[5m])
```

### increase（增量）

计算范围内的增量。

```promql
# 1小时内的请求数
increase(http_requests_total[1h])
```

### 其他常用函数

```promql
# 绝对值
abs(http_requests_total)

# 向上取整
ceil(http_requests_total)

# 向下取整
floor(http_requests_total)

# 变化趋势
changes(http_requests_total[1h])

# 一天内最大值
max_over_time(http_requests_total[1d])

# 一天内最小值
min_over_time(http_requests_total[1d])

# 一天内平均值
avg_over_time(http_requests_total[1d])
```

## 高级查询

### 多指标运算

```promql
# 错误率
rate(http_errors_total[5m]) / rate(http_requests_total[5m]) * 100

# 内存使用率
(node_memory_MemTotal_bytes - node_memory_MemFree_bytes) / node_memory_MemTotal_bytes * 100

# CPU 使用率
100 - (avg by (instance) (irate(node_cpu_seconds_total{mode="idle"}[5m])) * 100)
```

### 子查询

```promql
# 过去1小时内，每5分钟的最大值
max_over_time(rate(http_requests_total[5m])[1h:5m])
```

### 预测

```promql
# 预测4小时后的值
predict_linear(http_requests_total[1h], 4*3600)
```

## 实用查询示例

### HTTP 指标

```promql
# QPS
sum(rate(http_requests_total[5m]))

# 错误率
sum(rate(http_requests_total{status=~"5.."}[5m])) / sum(rate(http_requests_total[5m])) * 100

# P95 延迟
histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))

# 响应时间
rate(http_request_duration_seconds_sum[5m]) / rate(http_request_duration_seconds_count[5m])
```

### 系统指标

```promql
# CPU 使用率
100 - (avg(irate(node_cpu_seconds_total{mode="idle"}[5m])) * 100)

# 内存使用率
(node_memory_MemTotal_bytes - node_memory_MemAvailable_bytes) / node_memory_MemTotal_bytes * 100

# 磁盘使用率
(node_filesystem_size_bytes - node_filesystem_free_bytes) / node_filesystem_size_bytes * 100

# 网络流量
rate(node_network_receive_bytes_total[5m])
```

## 最佳实践

1. **使用 rate 而非 irate**：rate 更平滑，适合告警
2. **合理选择时间窗口**：通常使用 5m 或 1m
3. **使用标签分组**：按业务维度聚合
4. **避免高基数标签**：如 user_id、request_id

参考：https://prometheus.io/docs/prometheus/latest/querying/
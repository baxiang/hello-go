# Prometheus 学习系列

本系列文档系统性地介绍 Prometheus 监控系统，从基础概念到生产实践。

## 学习路径

### 第一阶段：基础概念

| 文档 | 描述 | 预计时间 |
|------|------|---------|
| 01-Prometheus概述 | Prometheus 简介、架构、应用场景 | 40 分钟 |
| 02-数据模型 | 指标类型、标签、时间序列 | 45 分钟 |
| 03-Go客户端集成 | 客户端库使用、指标暴露 | 50 分钟 |

### 第二阶段：进阶特性

| 文档 | 描述 | 预计时间 |
|------|------|---------|
| 04-PromQL查询 | 查询语法、函数、操作符 | 60 分钟 |
| 05-告警规则 | Alertmanager、告警配置 | 50 分钟 |
| 06-服务发现 | 自动发现、动态配置 | 45 分钟 |

### 第三阶段：生产实践

| 文档 | 描述 | 预计时间 |
|------|------|---------|
| 07-Exporter开发 | 自定义 Exporter、最佳实践 | 50 分钟 |
| 08-高可用部署 | 联邦、集群、存储 | 45 分钟 |
| 09-性能优化 | 查询优化、存储优化 | 50 分钟 |

## 前置知识

- Go 语言基础
- 监控概念
- Linux 基本操作

## 学习建议

1. 理解 Prometheus 数据模型
2. 按顺序学习理论文档
3. 完成对应 hands-on 练习
4. 参考 official 文档：https://prometheus.io/docs/

## 相关资源

- [Prometheus 官方文档](https://prometheus.io/docs/)
- [Prometheus GitHub](https://github.com/prometheus/prometheus)
- [Go client](https://github.com/prometheus/client_golang)
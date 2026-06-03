# Prometheus 学习笔记

本目录包含完整的 Prometheus 学习资料，包括理论文档、实战项目和代码示例。

## 目录结构

```
prometheus/
├── docs/                    # 理论文档
│   ├── README.md           # 系列导航和学习路径
│   ├── 01-Prometheus概述.md
│   ├── 02-数据模型.md
│   └── ...
├── projects/               # 实战项目
│   ├── 01-getting-started-helloworld/
│   ├── 02-advanced-app-monitoring/
│   └── ...
├── hands-on/               # 动手练习代码
└── example/                # 完整示例应用
```

## 学习路径

### 入门路径
docs/01-Prometheus概述 → docs/02-数据模型 → docs/03-Go客户端集成

### 进阶路径
docs/04-PromQL查询 → docs/05-告警规则 → docs/06-服务发现

### 实战路径
完成理论学习后 → projects/01-getting-started-helloworld → projects/02-advanced-app-monitoring → ...

## 本地开发环境

### 启动 Prometheus

```bash
# 使用 Docker
docker run -d --name prometheus \
  -p 9090:9090 \
  -v /path/to/prometheus.yml:/etc/prometheus/prometheus.yml \
  prom/prometheus:v2.51.0

# 或下载二进制文件
# https://github.com/prometheus/prometheus/releases
```

### 连接地址

- Prometheus UI: http://localhost:9090
- Metrics 端点: http://localhost:9090/metrics

## 快速开始

1. 阅读 [docs/README.md](./docs/README.md) 了解完整学习路径
2. 从 [docs/01-Prometheus概述.md](./docs/01-Prometheus概述.md) 开始学习
3. 使用 hands-on/ 目录中的代码进行练习

## 版本信息

| 组件 | 版本 |
|------|------|
| Prometheus | 2.51.0 |
| Go client | 1.19.1 |
| Go | 1.21+ |
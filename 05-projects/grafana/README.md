# Grafana 学习笔记

本目录包含完整的 Grafana 学习资料，包括理论文档、实战项目和代码示例。

## 目录结构

```
grafana/
├── docs/                    # 理论文档
│   ├── README.md           # 系列导航和学习路径
│   ├── 01-Grafana概述.md
│   ├── 02-数据源配置.md
│   └── ...
├── projects/               # 实战项目
│   ├── 01-getting-started-helloworld/
│   ├── 02-advanced-monitoring-dashboard/
│   └── ...
├── hands-on/               # 动手练习代码
└── example/                # 完整示例应用
```

## 学习路径

### 入门路径
docs/01-Grafana概述 → docs/02-数据源配置 → docs/03-面板创建

### 进阶路径
docs/04-告警配置 → docs/05-仪表板设计 → docs/06-权限管理

### 实战路径
完成理论学习后 → projects/01-getting-started-helloworld → projects/02-advanced-monitoring-dashboard → ...

## 本地开发环境

### 启动 Grafana

```bash
# 使用 Docker
docker run -d --name grafana \
  -p 3000:3000 \
  -e "GF_SECURITY_ADMIN_PASSWORD=admin" \
  grafana/grafana:10.4.0

# 或下载二进制文件
# https://grafana.com/grafana/download
```

### 连接地址

- Grafana UI: http://localhost:3000
- 默认账号: admin / admin

## 快速开始

1. 阅读 [docs/README.md](./docs/README.md) 了解完整学习路径
2. 从 [docs/01-Grafana概述.md](./docs/01-Grafana概述.md) 开始学习
3. 使用 hands-on/ 目录中的代码进行练习

## 版本信息

| 组件 | 版本 |
|------|------|
| Grafana | 10.4.0 |
| Go | 1.21+ |
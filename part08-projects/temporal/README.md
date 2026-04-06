# Temporal 学习笔记

本目录包含完整的 Temporal 学习资料，包括理论文档、实战项目和代码示例。

## 目录结构

```
temporal/
├── docs/                    # 理论文档（10篇系统性学习指南）
│   ├── README.md           # 系列导航和学习路径
│   ├── 01-Temporal概述.md
│   ├── 02-核心概念-Workflow.md
│   ├── 03-核心概念-Activity.md
│   ├── 04-核心概念-Worker.md
│   ├── 05-信号查询与更新.md
│   ├── 06-错误处理与重试策略.md
│   ├── 07-工作流版本控制.md
│   ├── 08-测试与调试.md
│   ├── 09-Go-SDK实战.md
│   └── 10-生产部署.md
├── projects/               # 实战项目（4个渐进式项目）
│   ├── 01-入门-HelloWorld/
│   ├── 02-入门-订单处理/
│   ├── 03-进阶-分布式订单系统/
│   └── 04-高级-电商工作流平台/
├── hands-on/               # 动手练习代码
└── example/                # 完整示例应用
```

## 学习路径

### 入门路径
docs/01-Temporal概述 → docs/02-核心概念-Workflow → docs/03-核心概念-Activity

### 进阶路径
docs/05-信号查询与更新 → docs/06-错误处理与重试策略 → docs/07-工作流版本控制

### 实战路径
完成理论学习后 → projects/01-入门-HelloWorld → projects/02-入门-订单处理 → ...

## 本地开发环境

### 启动 Temporal 服务

```bash
# 使用 Temporal CLI（推荐）
temporal server start-dev

# 或使用 Docker
docker run -d --name temporal \
  -p 7233:7233 \
  -p 8233:8233 \
  temporalio/server:1.26.1
```

### 连接地址

- Temporal 服务端口: localhost:7233
- Web UI 端口: http://localhost:8233

## 快速开始

1. 阅读 [docs/README.md](./docs/README.md) 了解完整学习路径
2. 从 [docs/01-Temporal概述.md](./docs/01-Temporal概述.md) 开始学习
3. 使用 hands-on/ 目录中的代码进行练习

## 版本信息

| 组件 | 版本 |
|------|------|
| Temporal Server | 1.26.1 |
| Go SDK | 1.31.0 |
| Go | 1.21+ |
# Temporal 技术指南系列

> 本系列是一份面向工程师的 Temporal 系统性学习指南，从基础概念到生产实践，覆盖 Workflow、Activity、Worker、错误处理、版本控制以及 Go 客户端编程。

---

## 目录

| 序号 | 文章 | 主题 | 难度 |
|------|------|------|------|
| 01 | [Temporal 概述](./01-Temporal概述.md) | 平台介绍、架构、适用场景 | ★☆☆ |
| 02 | [核心概念-Workflow](./02-核心概念-Workflow.md) | Workflow 定义、执行、生命周期 | ★★☆ |
| 03 | [核心概念-Activity](./03-核心概念-Activity.md) | Activity 定义、执行、重试策略 | ★★☆ |
| 04 | [核心概念-Worker](./04-核心概念-Worker.md) | Worker 配置、Task Queue | ★★☆ |
| 05 | [信号查询与更新](./05-信号查询与更新.md) | Signal、Query、Update | ★★★ |
| 06 | [错误处理与重试策略](./06-错误处理与重试策略.md) | 错误类型、重试、心跳、Saga | ★★★ |
| 07 | [工作流版本控制](./07-工作流版本控制.md) | Patching、Worker Versioning | ★★★ |
| 08 | [测试与调试](./08-测试与调试.md) | 单元测试、Web UI、CLI | ★★☆ |
| 09 | [Go-SDK基础](./09-Go-SDK基础.md) | 安装、连接、基本结构 | ★★☆ |
| 10 | [Go-SDK-Workflow开发](./10-Go-SDK-Workflow开发.md) | Workflow 定义、Context、最佳实践 | ★★★ |
| 11 | [Go-SDK-Activity开发](./11-Go-SDK-Activity开发.md) | Activity 定义、心跳、幂等性 | ★★★ |
| 12 | [Go-SDK-高级特性](./12-Go-SDK-高级特性.md) | Signal、Query、Interceptor | ★★★ |
| 13 | [生产部署](./13-生产部署.md) | Temporal Cloud、自托管、高可用 | ★★★ |

---

## 推荐学习路径

### 路径 A：快速入门（2-3 小时）

```
01-Temporal概述 → 02-核心概念-Workflow → 03-核心概念-Activity → 04-核心概念-Worker
```

适合：刚接触 Temporal、希望快速理解其定位与基础使用方式的工程师。

### 路径 B：进阶开发（基于路径 A，再加 3-4 小时）

```
→ 05-信号查询与更新 → 06-错误处理与重试策略 → 07-工作流版本控制
```

适合：需要在项目中深度使用 Temporal 的后端工程师。

### 路径 C：生产运维（基于路径 B，再加 2-3 小时）

```
→ 08-测试与调试 → 09-Go-SDK实战 → 10-生产部署
```

适合：负责 Temporal 应用部署、维护和监控的工程师。

---

## 各章内容简介

### 01 - Temporal 概述

介绍 Temporal 的起源（Uber Cadence）、核心特性（持久化执行、容错、可见性）、系统架构（Cluster、Worker、Client），与 AWS Step Functions、Camunda 等技术的对比，以及适用场景。

### 02 - 核心概念-Workflow

深入讲解 Workflow 的定义、类型、执行模型、Event History、确定性约束，以及长时间运行的工作流和 Schedule/Cron Job。

### 03 - 核心概念-Activity

讲解 Activity 的定义、执行、重试策略、心跳机制、幂等性要求，以及 Local Activity 与 Remote Activity 的区别。

### 04 - 核心概念-Worker

介绍 Worker Process、Task Queue、Worker Options 配置、资源限制和水平扩展策略。

### 05 - 信号查询与更新

讲解 Signal（外部消息）、Query（状态查询）、Update（状态更新）三种交互方式的使用场景和代码示例。

### 06 - 错误处理与重试策略

深入讲解错误类型、重试策略配置、心跳与进度报告、Saga 模式实现，以及超时配置最佳实践。

### 07 - 工作流版本控制

讲解为什么需要版本控制、Patching 方法、Worker Versioning，以及版本控制最佳实践。

### 08 - 测试与调试

介绍 Workflow 单元测试、Activity Mock、时间模拟、Temporal Web UI 使用和 CLI 调试命令。

### 09 - Go-SDK实战

使用官方 Go SDK 进行完整的代码实战，涵盖客户端连接、Workflow 定义、Activity 定义、Worker 配置、Signal/Query 操作等。

### 10 - 生产部署

介绍 Temporal Cloud、自托管部署方案、高可用配置、监控与告警、安全配置。

---

## 快速概念速查

### 核心术语对照

| Temporal 术语 | 含义 | 类比 |
|--------------|------|------|
| Workflow | 业务流程定义 | 函数/流程 |
| Activity | 单一操作单元 | 函数调用 |
| Worker | 执行代码的进程 | 服务实例 |
| Task Queue | 任务队列 | 消息队列 |
| Signal | 外部消息 | 事件 |
| Query | 状态查询 | HTTP GET |
| Event History | 事件历史 | 审计日志 |

### 执行语义速查

| 概念 | 说明 |
|------|------|
| 持久化执行 | Workflow 状态自动持久化 |
| 确定性 | Workflow 代码必须确定性 |
| 重试 | Activity 失败自动重试 |
| 心跳 | 长时间 Activity 进度报告 |

---

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

---

## 版本兼容性说明

本指南基于以下版本编写：

| 组件 | 版本 | 说明 |
|------|------|------|
| Temporal Server | 1.26.1 | 当前稳定版 |
| Go SDK | 1.31.0 | 推荐使用最新版 |
| Go | 1.21+ | 客户端最低要求 |

---

## 参考资源

### 官方资源

- [Temporal 官方文档](https://docs.temporal.io)
- [Temporal Go SDK](https://github.com/temporalio/sdk-go)
- [Temporal CLI](https://github.com/temporalio/cli)
- [Learn Temporal](https://learn.temporal.io)

### 学习资源

- [Temporal 101 课程](https://learn.temporal.io/courses/temporal_101/)
- [Temporal 102 课程](https://learn.temporal.io/courses/temporal_102/)
- [Temporal YouTube 频道](https://www.youtube.com/c/Temporalio)

### 社区资源

- [Temporal Slack](https://temporal.io/slack)
- [Temporal 社区论坛](https://community.temporal.io)
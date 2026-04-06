# Temporal Go SDK 动手练习

本目录包含一系列渐进式的练习，帮助您掌握 Temporal Go SDK 开发。

## 练习列表

### 01-workflow-basics - 工作流基础

学习内容：
- 创建简单的 Workflow
- 配置 Activity 选项
- 处理 Workflow 返回值
- 理解 Workflow 上下文

运行：
```bash
make run-01
# 或
go run ./01-workflow-basics/main.go
```

### 02-activity-patterns - Activity 模式

学习内容：
- 单个 Activity 调用
- 多个 Activity 顺序执行
- 多个 Activity 并行执行
- Activity 本地重试

运行：
```bash
make run-02
# 或
go run ./02-activity-patterns/main.go
```

### 03-error-handling - 错误处理

学习内容：
- Activity 错误处理
- 自定义错误类型
- 重试策略配置
- 超时处理

运行：
```bash
make run-03
# 或
go run ./03-error-handling/main.go
```

### 04-signals-queries - 信号与查询

学习内容：
- 发送信号到 Workflow
- 查询 Workflow 状态
- 使用选择器处理多个信号
- 更新 Workflow 状态

运行：
```bash
make run-04
# 或
go run ./04-signals-queries/main.go
```

## 环境要求

1. **Go 1.21+**

2. **Temporal 服务器**（任选一种）:
   
   方式一：Docker（推荐）
   ```bash
   docker run -d --name temporal -p 7233:7233 temporalio/auto-setup:latest
   ```
   
   方式二：本地开发服务器
   ```bash
   go install github.com/temporalio/cli@latest
   temporal server start-dev
   ```

3. **Temporal Web UI**（可选）:
   ```bash
   docker run -d --name temporal-ui -p 8080:8080 --link temporal temporalio/ui:latest
   ```

## 快速开始

1. 初始化依赖：
   ```bash
   cd hands-on
   go mod tidy
   ```

2. 启动 Temporal 服务器

3. 选择一个练习运行：
   ```bash
   make run-01  # 工作流基础
   make run-02  # Activity 模式
   make run-03  # 错误处理
   make run-04  # 信号与查询
   ```

## 练习说明

每个练习目录包含一个完整的示例程序。建议学习顺序：

1. **阅读代码**：先理解每个练习的代码结构
2. **运行程序**：执行程序观察输出
3. **修改代码**：尝试修改参数和逻辑
4. **查看 UI**：在 Temporal Web UI 中查看工作流执行情况
5. **实验错误**：故意引入错误观察错误处理行为

## 常见问题

### Q: 连接不上 Temporal 服务器？

确保 Temporal 服务器正在运行：
```bash
# 检查 Docker 容器状态
docker ps | grep temporal

# 或检查本地服务器
temporal server start-dev
```

### Q: 找不到 Task Queue？

确保 Worker 和 Starter 使用相同的 Task Queue 名称。

### Q: Workflow 执行卡住？

检查：
1. Worker 是否正在运行
2. Task Queue 名称是否匹配
3. Temporal 服务器是否正常

## 参考资料

- [Temporal Go SDK 官方文档](https://docs.temporal.io/dev-guide/go)
- [Temporal Go SDK API 参考](https://pkg.go.dev/go.temporal.io/sdk)
- [Temporal 官方示例](https://github.com/temporalio/samples-go)
# 06-engineering

将 Go 语言知识转化为工程能力。涵盖包管理、标准库实践与命令行工具开发。

## 学习目标

- 掌握 Go Modules 依赖管理全流程
- 熟练使用 time 包进行时间处理
- 能够使用 Cobra 构建命令行应用

## 章节

| 编号 | 章节 | 文件 |
|------|------|------|
| 1.13 | 包管理 | [01-包管理](./01-包管理.md) |
| 1.14 | 时间处理 | [02-时间处理](./02-时间处理.md) |
| 1.15 | Cobra 命令行工具 | [04-Cobra命令行工具](./04-Cobra命令行工具.md) |

## 学习建议

- **包管理是工程化的基石** — 花时间理解 go.mod/go.sum 的每个字段
- 时间处理看似简单，时区和格式化是最容易踩坑的地方
- Cobra 是 Kubernetes、etcd、Hugo 等项目的命令行框架标准

## 检查清单

- [ ] 掌握 go mod init/tidy/download/verify/graph
- [ ] 理解 go.mod（require/replace/exclude/retract）和 go.sum 的作用
- [ ] 掌握语义化版本与依赖升级/降级
- [ ] 理解 GOPRIVATE 与私有仓库配置
- [ ] 熟练使用 time.Time 的创建、格式化（2006-01-02 15:04:05）、时区转换
- [ ] 掌握 Timer/Ticker 的使用与资源释放
- [ ] 能够使用 Cobra 构建命令与子命令
- [ ] 掌握 flag 参数解析与验证

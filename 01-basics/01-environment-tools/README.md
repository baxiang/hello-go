# 01-environment-tools

本章节介绍Go语言的开发环境搭建和常用工具链，帮助开发者快速入门。

## 学习目标

完成本章节学习后，你将能够：

- 独立搭建Go开发环境
- 掌握版本管理工具使用
- 理解Go Modules依赖管理
- 配置开发工具和IDE
- 使用常用开发工具

## 章节内容

### 01-环境搭建与工具链

**核心内容**：
- Go安装与版本管理（goenv、asdf、mise）
- GOPATH与Go Modules
- GOPROXY配置（国内镜像）
- 开发工具（VS Code + Go插件、GoLand）
- go命令详解
- 代码格式化与检查
- 调试工具（Delve）

**学习时间**：2-3天

**实践要求**：
- [ ] 完成Go环境安装
- [ ] 配置VS Code或GoLand
- [ ] 创建第一个Go项目
- [ ] 运行并调试代码

### 02-常用工具

**核心内容**：
- 代码生成工具
- 文档生成工具
- 代码质量工具
- 性能分析工具
- 测试覆盖率工具

**学习时间**：1-2天

**实践要求**：
- [ ] 使用go generate
- [ ] 生成代码文档
- [ ] 运行代码检查
- [ ] 分析性能瓶颈

## 前置知识

- 基本的计算机操作能力
- 命令行基础使用
- 至少一门编程语言基础（推荐但非必需）

## 学习建议

1. **边学边练**：每学习一个知识点，立即动手实践
2. **环境优先**：确保开发环境配置正确，这是后续学习的基础
3. **工具熟练**：掌握常用工具能大幅提升开发效率
4. **遇到问题**：善用官方文档和社区资源

## 检查清单

完成本章节后，验证你的掌握程度：

- [ ] Go环境正确安装并配置
- [ ] 能够创建和管理Go项目
- [ ] 熟练使用go run、go build、go test
- [ ] 配置并使用代码格式化工具
- [ ] 能够调试Go程序
- [ ] 理解Go Modules工作原理

## 常见问题

**Q: 应该选择VS Code还是GoLand？**
A: VS Code免费轻量，适合入门；GoLand功能强大，适合专业开发。建议先从VS Code开始。

**Q: 国内如何加速依赖下载？**
A: 配置GOPROXY=https://goproxy.cn,direct

**Q: Go Modules和GOPATH有什么区别？**
A: Go Modules是现代化的依赖管理方式，不需要在GOPATH下开发，推荐使用。

## 延伸阅读

- [Go官方文档](https://golang.org/doc/)
- [Go安装指南](https://golang.org/doc/install)
- [Go Modules教程](https://blog.golang.org/using-go-modules)
- [VS Code Go插件](https://marketplace.visualstudio.com/items?itemName=golang.Go)
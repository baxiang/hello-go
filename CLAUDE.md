# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Repository Overview

This is a **Go language learning documentation repository** (go-learning-roadmap) containing comprehensive Chinese technical guides covering Go basics through cloud-native development.

**Structure**: 10 parts, 50+ markdown files organized by topic:

```
part01-basics/     # Go 基础：环境、语法、流程控制、数组、Map、函数、指针、时间、字符串
part02-core/       # 核心特性：结构体、接口、错误处理、包管理、泛型
part03-concurrency/# 并发编程：Goroutine、Channel、同步原语、Context、并发模式
part04-stdlib/     # 标准库：常用库、网络、数据库、测试、日志、文件 IO、正则、序列化
part05-web/        # Web 开发：框架、中间件、认证、API 设计、微服务
part06-engineering/# 工程实践：项目结构、代码规范、依赖管理、构建部署、监控
part07-advanced/   # 高级主题：内存管理、反射、汇编、设计模式、性能优化
part08-projects/   # 实战项目：入门、进阶、高级项目、开源贡献
part09-cloudnative/# 云原生：Docker、K8s、Helm、GitOps、ServiceMesh、Serverless
part10-performance/# 性能调优：方法论、CPU/内存/网络/数据库调优
```

## Common Commands

This repository contains **documentation only** (markdown files). No Go source code to compile.

**Useful commands for working with this repo:**

```bash
# View the learning roadmap
cat go-learning-roadmap.md

# List all topic files
find . -name "*.md" | sort

# Search for specific topics
grep -r "Goroutine" --include="*.md" .
grep -r "JSON" --include="*.md" .

# Git workflow
git status
git add .
git commit -m "docs: enhance <topic> content"
git push
```

## Content Guidelines

When creating or editing markdown files:

1. **File naming**: Use `NN-标题.md` format (e.g., `01-环境搭建与工具链.md`)
2. **Code examples**: All code must be complete, runnable Go code with package and imports
3. **Structure**: Use consistent heading hierarchy (##, ###, ####)
4. **Comments**: Chinese comments in code examples
5. **Best practices**: Include checklists and comparison tables where applicable

## Architecture Notes

- Each part directory contains independent topic files
- Files are designed to be read sequentially within each part
- Cross-references between parts use relative markdown links
- Code examples should follow Effective Go guidelines

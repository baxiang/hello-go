# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Repository Overview

This is a **Go language learning documentation repository** (go-learning-roadmap) containing comprehensive Chinese technical guides covering Go basics through cloud-native development.

**Structure**: 6 parts, 50+ markdown files organized by topic:

```
01-语言基础/    # Go 基础 + 核心特性：环境、语法、结构体、接口、错误处理、泛型
02-进阶编程/    # 并发编程 + 标准库：Goroutine、Channel、并发模式、网络、IO、测试
03-Web与工程/   # Web开发 + 工程实践：框架、中间件、项目结构、CI/CD、监控
04-高级话题/    # 高级主题 + 性能调优：内存管理、反射、pprof、CPU/内存调优
05-实战项目/    # 项目实战：gRPC、Kafka、Kratos 微服务、etcd、NATS、Redis 等
06-云原生/      # 云原生：Docker、K8s、Helm、GitOps、ServiceMesh、Serverless
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

- Each  directory contains independent topic files
- Files are designed to be read sequentially within each 
- Cross-references between s use relative markdown links
- Code examples should follow Effective Go guidelines

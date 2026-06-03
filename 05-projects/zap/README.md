# Zap 学习笔记

本目录包含完整的 Zap 结构化日志学习资料，包括理论文档、实战项目和代码示例。

## 目录结构

```
zap/
├── docs/                    # 理论文档
│   ├── README.md           # 系列导航和学习路径
│   ├── 01-Zap概述.md
│   ├── 02-基本使用.md
│   └── ...
├── projects/               # 实战项目
│   ├── 01-getting-started-helloworld/
│   ├── 02-advanced-webapp日志/
│   └── ...
├── hands-on/               # 动手练习代码
└── example/                # 完整示例应用
```

## 学习路径

### 入门路径
docs/01-Zap概述 → docs/02-基本使用 → docs/03-配置选项

### 进阶路径
docs/04-高级特性 → docs/05-性能优化 → docs/06-最佳实践

### 实战路径
完成理论学习后 → projects/01-getting-started-helloworld → projects/02-advanced-webapp日志 → ...

## 本地开发环境

### 安装 Zap

```bash
go get go.uber.org/zap@v1.27.0
```

### 使用示例

```go
package main

import (
    "go.uber.org/zap"
)

func main() {
    logger, _ := zap.NewProduction()
    defer logger.Sync()
    
    logger.Info("hello zap",
        zap.String("key", "value"),
        zap.Int("count", 42),
    )
}
```

## 快速开始

1. 阅读 [docs/README.md](./docs/README.md) 了解完整学习路径
2. 从 [docs/01-Zap概述.md](./docs/01-Zap概述.md) 开始学习
3. 使用 hands-on/ 目录中的代码进行练习

## 版本信息

| 组件 | 版本 |
|------|------|
| Zap | 1.27.0 |
| Go | 1.21+ |
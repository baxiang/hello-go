# Zap 概述

Zap 是 Uber 开源的高性能结构化日志库。

## 核心特性

- **高性能**：零分配、极快的日志记录
- **结构化**：JSON 格式，易于解析
- **灵活配置**：多种日志级别和输出格式
- **类型安全**：强类型字段

## 快速开始

### 安装

```bash
go get go.uber.org/zap@v1.27.0
```

### 使用

```go
package main

import "go.uber.org/zap"

func main() {
    // 生产环境
    logger, _ := zap.NewProduction()
    defer logger.Sync()
    
    logger.Info("hello zap",
        zap.String("key", "value"),
        zap.Int("count", 42),
    )
    
    // 开发环境
    devLogger, _ := zap.NewDevelopment()
    devLogger.Debug("debug message")
    
    // Sugar（性能略低，语法更简洁）
    sugar := logger.Sugar()
    sugar.Infow("sugar logger", "url", "http://example.com", "attempt", 3)
}
```

### 输出示例

```json
{
  "level": "info",
  "ts": "2024-01-01T12:00:00.000Z",
  "caller": "main.go:10",
  "msg": "hello zap",
  "key": "value",
  "count": 42
}
```

## 性能对比

| 日志库 | 禁用日志 | 输出日志 |
|--------|---------|---------|
| Zap | 0.5ns | 1200ns |
| Logrus | 1000ns | 11000ns |
| Log15 | 900ns | 5000ns |

## 版本信息

| 组件 | 版本 |
|------|------|
| Zap | 1.27.0 |

在下一章中，我们将学习基本使用。

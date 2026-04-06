# Wire 概述

Wire 是 Google 开源的 Go 依赖注入代码生成工具。

## 核心概念

### Provider
创建依赖的函数。

```go
func NewDatabase() *Database {
    return &Database{}
}
```

### Injector
组装依赖的函数，由 Wire 生成。

```go
// wire.go
//+build wireinject

func InitializeApp() *App {
    wire.Build(NewDatabase, NewRepository, NewService, NewApp)
    return nil
}
```

## 快速开始

### 安装

```bash
go install github.com/google/wire/cmd/wire@v0.6.0
```

### 定义 Provider

```go
// user.go
type User struct {}

func NewUser() *User {
    return &User{}
}
```

### 定义 Injector

```go
// wire.go
//+build wireinject

package main

import "github.com/google/wire"

func InitializeUser() *User {
    wire.Build(NewUser)
    return nil
}
```

### 生成代码

```bash
wire
```

生成的 `wire_gen.go`:

```go
func InitializeUser() *User {
    user := NewUser()
    return user
}
```

## 版本信息

| 组件 | 版本 |
|------|------|
| Wire | 0.6.0 |

在下一章中，我们将学习基本概念。

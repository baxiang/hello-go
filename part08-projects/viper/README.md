# Viper 学习笔记

本目录包含完整的 Viper 配置管理学习资料，包括理论文档、实战项目和代码示例。

## 目录结构

```
viper/
├── docs/                    # 理论文档
│   ├── README.md           # 系列导航和学习路径
│   ├── 01-Viper概述.md
│   ├── 02-配置读取.md
│   └── ...
├── projects/               # 实战项目
│   ├── 01-入门-HelloWorld/
│   ├── 02-进阶-多环境配置/
│   └── ...
├── hands-on/               # 动手练习代码
└── example/                # 完整示例应用
```

## 学习路径

### 入门路径
docs/01-Viper概述 → docs/02-配置读取 → docs/03-配置写入

### 进阶路径
docs/04-多环境配置 → docs/05-配置热更新 → docs/06-最佳实践

### 实战路径
完成理论学习后 → projects/01-入门-HelloWorld → projects/02-进阶-多环境配置 → ...

## 本地开发环境

### 安装 Viper

```bash
go get github.com/spf13/viper@v1.18.2
```

### 使用示例

```go
package main

import (
    "fmt"
    "github.com/spf13/viper"
)

func main() {
    viper.SetConfigName("config")
    viper.SetConfigType("yaml")
    viper.AddConfigPath(".")
    
    if err := viper.ReadInConfig(); err != nil {
        panic(err)
    }
    
    fmt.Println(viper.GetString("app.name"))
}
```

## 快速开始

1. 阅读 [docs/README.md](./docs/README.md) 了解完整学习路径
2. 从 [docs/01-Viper概述.md](./docs/01-Viper概述.md) 开始学习
3. 使用 hands-on/ 目录中的代码进行练习

## 版本信息

| 组件 | 版本 |
|------|------|
| Viper | 1.18.2 |
| Go | 1.21+ |
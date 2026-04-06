# Viper 概述

Viper 是 Go 应用程序配置解决方案，支持多种配置格式。

## 核心特性

- **多格式**：JSON、YAML、TOML、HCL、ENV
- **多源**：配置文件、环境变量、命令行参数
- **热更新**：监听配置文件变更
- **默认值**：设置配置默认值

## 快速开始

### 安装

```bash
go get github.com/spf13/viper@v1.18.2
```

### 使用

```go
package main

import (
    "fmt"
    "github.com/spf13/viper"
)

func main() {
    // 设置默认值
    viper.SetDefault("app.name", "myapp")
    
    // 读取配置文件
    viper.SetConfigName("config")
    viper.SetConfigType("yaml")
    viper.AddConfigPath(".")
    viper.ReadInConfig()
    
    // 读取环境变量
    viper.AutomaticEnv()
    
    // 获取配置
    fmt.Println(viper.GetString("app.name"))
}
```

### config.yaml

```yaml
app:
  name: myapp
  port: 8080
database:
  host: localhost
  port: 3306
```

## 版本信息

| 组件 | 版本 |
|------|------|
| Viper | 1.18.2 |

在下一章中，我们将学习配置读取。

# Cobra 核心概念与基础

> Cobra 是一个强大的 CLI 框架，用于构建现代命令行应用程序。本文将详细介绍 Cobra 的核心概念和使用方法。

## 1.1 Cobra 简介

### 什么是 Cobra？

Cobra 是 Go 语言中最流行的命令行界面（CLI）应用程序框架之一，由 spf13 开发并维护。它被广泛应用于各种知名项目中，包括：

- **Kubernetes** - 云原生容器编排平台
- **Docker** - 容器化平台
- **Hugo** - 静态网站生成器
- **Istio** - 服务网格
- **GitHub CLI** - GitHub 命令行工具

```
┌─────────────────────────────────────────────────────────────────┐
│                    Cobra 生态                                   │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│   ┌─────────────┐  ┌─────────────┐  ┌─────────────┐          │
│   │ Kubernetes  │  │   Docker   │  │   Hugo     │          │
│   │   (k8s)    │  │            │  │            │          │
│   └─────────────┘  └─────────────┘  └─────────────┘          │
│                                                                 │
│   ┌─────────────┐  ┌─────────────┐  ┌─────────────┐          │
│   │   Istio     │  │ GitHub CLI  │  │   Terraform  │          │
│   │            │  │    (gh)    │  │            │          │
│   └─────────────┘  └─────────────┘  └─────────────┘          │
│                                                                 │
│   ... 以及 thousands more!                                      │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

### 核心特性

| 特性 | 说明 |
|------|------|
| **Easy CLI** | 简单易用的 API |
| **Flags** | 支持 POSIX 兼容的 flags |
| **Subcommands** | 支持嵌套命令 |
| **Templates** | 支持命令模板 |
| **Auto Generate** | 自动生成代码和文档 |
| **Help** | 自动生成帮助信息 |
| **Custom Help** | 支持自定义帮助模板 |

### 安装 Cobra

```bash
# 安装 Cobra CLI 工具
go install github.com/spf13/cobra-cli@latest

# 验证安装
cobra-cli --version
```

## 1.2 核心概念

### 1.2.1 命令结构

Cobra 的命令结构如下：

```
┌─────────────────────────────────────────────────────────────────┐
│                    命令结构层次                                  │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  Root Command (根命令)                                          │
│       │                                                         │
│       ├── Command 1 (子命令)                                    │
│       │       │                                                 │
│       │       ├── SubCommand 1.1 (孙命令)                      │
│       │       │                                                 │
│       │       └── SubCommand 1.2                               │
│       │                                                         │
│       ├── Command 2 (子命令)                                    │
│       │                                                         │
│       └── Command 3 (子命令)                                    │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

### 1.2.2 核心组件

Cobra 主要由三个核心组件构成：

```
┌─────────────────────────────────────────────────────────────────┐
│                    Cobra 三大组件                                │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐  │
│  │    Command     │  │     Arg        │  │    Flag       │  │
│  │   (命令)       │  │    (参数)       │  │    (标志)      │  │
│  │                │  │                │  │                │  │
│  │  Execute()    │  │  位置参数      │  │  -f, --force   │  │
│  │  Run()        │  │  必须参数      │  │  -n, --name    │  │
│  │  RunE()       │  │  可选参数      │  │  -v, --verbose │  │
│  │                │  │                │  │                │  │
│  └─────────────────┘  └─────────────────┘  └─────────────────┘  │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

#### Command（命令）

```go
// Command 是 Cobra 的核心结构
type Command struct {
    // 命令名称
    Use string
    
    // 命令简短的描述
    Short string
    
    // 命令的详细描述
    Long string
    
    // 命令的处理函数
    Run func(cmd *Command, args []string)
    
    // 带错误的处理函数
    RunE func(cmd *Command, args []string) error
    
    // 子命令
    Commands []*Command
    
    // 标志
    Flags *flag.FlagSet
    
    // ... 更多字段
}
```

#### Arg（参数）

参数是命令的位置参数：

```bash
# 示例：myapp user add <username> <email>
# <username> 和 <email> 就是位置参数
```

#### Flag（标志）

标志是命令的可选修饰符：

```bash
# 示例：myapp user add --name john --email john@example.com
# --name 和 --email 就是标志
```

### 1.2.3 工作流程

```
┌─────────────────────────────────────────────────────────────────┐
│                    命令执行流程                                  │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  用户输入                                                      │
│       │                                                         │
│       ▼                                                         │
│  ┌─────────────────────────────────────────────────────────┐  │
│  │ 1. 解析命令行参数                                        │  │
│  │    - 解析 flags (如 --verbose)                          │  │
│  │    - 解析 args (如 username)                            │  │
│  └─────────────────────────────────────────────────────────┘  │
│       │                                                         │
│       ▼                                                         │
│  ┌─────────────────────────────────────────────────────────┐  │
│  │ 2. 验证参数                                             │  │
│  │    - 检查必需的参数                                      │  │
│  │    - 验证参数格式                                        │  │
│  └─────────────────────────────────────────────────────────┘  │
│       │                                                         │
│       ▼                                                         │
│  ┌─────────────────────────────────────────────────────────┐  │
│  │ 3. 执行 PreRun 钩子 (可选)                              │  │
│  └─────────────────────────────────────────────────────────┘  │
│       │                                                         │
│       ▼                                                         │
│  ┌─────────────────────────────────────────────────────────┐  │
│  │ 4. 执行 Run 或 RunE 函数                               │  │
│  └─────────────────────────────────────────────────────────┘  │
│       │                                                         │
│       ▼                                                         │
│  ┌─────────────────────────────────────────────────────────┐  │
│  │ 5. 执行 PostRun 钩子 (可选)                              │  │
│  └─────────────────────────────────────────────────────────┘  │
│       │                                                         │
│       ▼                                                         │
│  返回结果                                                      │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

## 1.3 快速开始

### 1.3.1 使用 Cobra CLI 创建项目

```bash
# 1. 初始化 Go 模块
mkdir myapp && cd myapp
go mod init myapp

# 2. 使用 Cobra CLI 初始化项目
cobra-cli init --author "Your Name" --license MIT

# 3. 查看生成的项目结构
tree
```

生成的项目结构：

```
myapp/
├── cmd/
│   └── root.go          # 根命令
├── main.go              # 入口文件
├── go.mod
└── go.sum
```

### 1.3.2 手动创建命令

```go
// main.go
package main

import (
    "fmt"
    "github.com/spf13/cobra"
)

func main() {
    // 创建根命令
    var rootCmd = &cobra.Command{
        Use:   "myapp",
        Short: "MyApp 是一个示例命令行工具",
        Long: `MyApp 是一个功能强大的命令行工具，
              用于演示 Cobra 框架的使用方法。`,
        Run: func(cmd *cobra.Command, args []string) {
            fmt.Println("执行根命令")
        },
    }

    // 执行命令
    if err := rootCmd.Execute(); err != nil {
        fmt.Fprintf(os.Stderr, "执行错误: %v\n", err)
        os.Exit(1)
    }
}
```

### 1.3.3 添加子命令

```go
package main

import (
    "fmt"
    "github.com/spf13/cobra"
)

func main() {
    // 创建根命令
    rootCmd := &cobra.Command{
        Use:   "myapp",
        Short: "MyApp CLI",
    }

    // 创建子命令: greet
    greetCmd := &cobra.Command{
        Use:   "greet [name]",
        Short: "问候某人",
        Args:  cobra.ExactArgs(1),  // 精确要求 1 个参数
        Run: func(cmd *cobra.Command, args []string) {
            fmt.Printf("你好, %s!\n", args[0])
        },
    }

    // 创建子命令: version
    versionCmd := &cobra.Command{
        Use:   "version",
        Short: "显示版本信息",
        Run: func(cmd *cobra.Command, args []string) {
            fmt.Println("MyApp v1.0.0")
        },
    }

    // 添加子命令
    rootCmd.AddCommand(greetCmd)
    rootCmd.AddCommand(versionCmd)

    // 执行
    rootCmd.Execute()
}
```

运行示例：

```bash
# 运行根命令
$ go run main.go
# (无输出，因为根命令没有定义 Run)

# 运行 greet 子命令
$ go run main.go greet 张三
你好, 张三!

# 运行 version 子命令
$ go run main.go version
MyApp v1.0.0
```

## 1.4 命令详解

### 1.4.1 命令属性

```go
cmd := &cobra.Command{
    // Use: 命令名称（必需）
    // 格式: command [subcommand] [flags] [args]
    Use: "app [command]",
    
    // Aliases: 命令别名
    Aliases: []string{"a", "app"},
    
    // Short: 命令简短描述（用于帮助信息）
    Short: "这是一个简短的描述",
    
    // Long: 命令详细描述
    Long: `这是一个非常详细的描述，
           可以包含多行文本。`,
    
    // Example: 使用示例
    Example: `  app greet 张三
  app greet -l john`,
    
    // Version: 版本号
    Version: "1.0.0",
    
    // Args: 参数验证器
    Args: cobra.ExactArgs(1),
    
    // SuggestFor: 建议替代命令
    SuggestFor: []string{"old-command"},
    
    // Deprecated: 标记为废弃
    Deprecated: "请使用 new-command 代替",
    
    // Hidden: 隐藏命令（不在帮助中显示）
    Hidden: false,
}
```

### 1.4.2 命令执行方式

Cobra 支持多种命令执行方式：

```go
// 方式 1: 直接定义 Run
cmd := &cobra.Command{
    Run: func(cmd *cobra.Command, args []string) {
        fmt.Println("执行命令")
    },
}

// 方式 2: 使用 RunE 返回错误
cmd := &cobra.Command{
    RunE: func(cmd *cobra.Command, args []string) error {
        if len(args) == 0 {
            return fmt.Errorf("需要至少一个参数")
        }
        fmt.Println("执行命令")
        return nil
    },
}

// 方式 3: 使用 RunE 和 PreRun
cmd := &cobra.Command{
    // 前置处理
    PreRun: func(cmd *cobra.Command, args []string) {
        fmt.Println("PreRun: 准备执行")
    },
    
    // 主处理
    RunE: func(cmd *cobra.Command, args []string) error {
        fmt.Println("RunE: 执行中")
        return nil
    },
    
    // 后置处理
    PostRun: func(cmd *cobra.Command, args []string) {
        fmt.Println("PostRun: 执行完成")
    },
}
```

### 1.4.3 命令生命周期

```
┌─────────────────────────────────────────────────────────────────┐
│                    命令执行生命周期                              │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  1. ValidateArgs()     ← 验证参数                               │
│         │                                                       │
│         ▼                                                       │
│  2. ValidateRequiredFlags()  ← 验证必需标志                    │
│         │                                                       │
│         ▼                                                       │
│  3. ValidateFlagGroups()  ← 验证标志组                         │
│         │                                                       │
│         ▼                                                       │
│  4. PreRun()            ← 前置钩子（可选）                      │
│         │                                                       │
│         ▼                                                       │
│  5. Run() / RunE()     ← 主处理函数                             │
│         │                                                       │
│         ▼                                                       │
│  6. PostRun()          ← 后置钩子（可选）                       │
│         │                                                       │
│         ▼                                                       │
│  7. Execute()          ← 返回结果                               │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

## 1.5 参数（Args）

### 1.5.1 内置参数验证器

Cobra 提供了多种内置的参数验证器：

```go
// 无参数验证
cmd.Args = nil

// 精确参数数量
cmd.Args = cobra.ExactArgs(1)      // 恰好 1 个参数
cmd.Args = cobra.MinimumNArgs(1)    // 至少 1 个参数
cmd.Args = cobra.MaximumNArgs(3)    // 最多 3 个参数
cmd.Args = cobra.RangeArgs(1, 3)    // 1-3 个参数

// 参数范围
cmd.Args = cobra.OnlyValidArgs      // 必须是有效参数

// 无参数（无参数时不会报错）
cmd.Args = cobra.NoArgs

// 任意参数
cmd.Args = cobra.ArbitraryArgs
```

### 1.5.2 自定义参数验证

```go
// 自定义参数验证函数
func validateName(cmd *cobra.Command, args []string) error {
    if len(args) < 1 {
        return fmt.Errorf("需要至少一个参数")
    }
    
    name := args[0]
    if len(name) < 2 {
        return fmt.Errorf("名字长度必须至少为 2 个字符")
    }
    
    return nil
}

// 使用自定义验证
cmd.Args = validateName
```

## 1.6 标志（Flags）

### 1.6.1 标志类型

```go
// 字符串标志
cmd.Flags().String("name", "", "姓名")
cmd.Flags().StringVar(&name, "name", "", "姓名")

// 整数标志
cmd.Flags().Int("age", 0, "年龄")
cmd.Flags().IntVar(&age, "age", 0, "年龄")

// 布尔标志
cmd.Flags().Bool("verbose", false, "详细输出")
cmd.Flags().BoolP("verbose", "v", false, "详细输出")

// 浮点数标志
cmd.Flags().Float64("rate", 0.0, "速率")

// 字符串数组标志
cmd.Flags().StringArray("tags", []string{}, "标签")

// 持久标志（对子命令也生效）
cmd.Flags().String("config", "", "配置文件")
cmd.MarkFlagRequired("config")

// 本地标志（仅对当前命令生效）
cmd.Flags().String("local", "", "本地标志")
cmd.Flags().MarkPersistentRequired("config")
```

### 1.6.2 标志示例

```go
package main

import (
    "fmt"
    "github.com/spf13/cobra"
)

var (
    name    string
    age     int
    verbose bool
)

func main() {
    rootCmd := &cobra.Command{
        Use:   "user",
        Short: "用户管理",
    }

    // 添加用户命令
    addCmd := &cobra.Command{
        Use:   "add",
        Short: "添加用户",
        Run: func(cmd *cobra.Command, args []string) {
            fmt.Printf("添加用户: name=%s, age=%d, verbose=%v\n", 
                name, age, verbose)
        },
    }

    // 注册标志
    addCmd.Flags().StringVarP(&name, "name", "n", "", "用户姓名")
    addCmd.Flags().IntVarP(&age, "age", "a", 0, "用户年龄")
    addCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "详细输出")

    // 标记必需
    addCmd.MarkFlagRequired("name")

    rootCmd.AddCommand(addCmd)
    rootCmd.Execute()
}
```

运行示例：

```bash
# 基本使用
$ go run main.go add -n 张三 -a 25
添加用户: name=张三, age=25, verbose=false

# 使用短标志
$ go run main.go add -n 李四 -a 30 -v
添加用户: name=李四, age=30, verbose=true

# 缺少必需参数
$ go run main.go add -a 20
Error: required flag(s) not set: --name

# 查看帮助
$ go run main.go add --help
Usage:
  user add [flags]

Flags:
  -a, --age int        用户年龄
  -h, --help           help for add
  -n, --name string     用户姓名 (required)
  -v, --verbose        详细输出
```

### 1.6.3 标志分组

```go
// 创建标志组
cmd.Flags().String("host", "localhost", "服务器地址")
cmd.Flags().Int("port", 8080, "服务器端口")

// 标记必需
cmd.MarkFlagRequired("host")
cmd.MarkFlagRequired("port")

// 标志分组（用于帮助信息）
cmd.Flags().SetAnnotation("host", "group", "Server")
cmd.Flags().SetAnnotation("port", "group", "Server")
```

## 1.7 帮助信息

### 1.7.1 自动生成的帮助

Cobra 自动生成完整的帮助信息：

```bash
$ myapp --help
MyApp CLI

MyApp 是一个功能强大的命令行工具

Usage:
  myapp [command]

Available Commands:
  add       添加用户
  delete    删除用户
  help      Help about any command
  list      列出用户
  version   显示版本信息

Flags:
  -h, --help   help for myapp

Use "myapp [command] --help" for more information about a command.
```

### 1.7.2 自定义帮助模板

```go
// 自定义帮助模板
helpTemplate := `Usage:{{if .Runnable}}
  {{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}
  {{.CommandPath}} [command]{{end}}

{{if .HasAvailableSubCommands}}Available Commands:{{range .Commands}}{{if .IsAvailableCommand}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableFlags}}

Flags:
{{.Flags.FlagUsages | indent 4}}{{end}}{{if .HasAvailableSubCommands}}

Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}
`

cmd.SetHelpTemplate(helpTemplate)
```

## 1.8 本章小结

```
┌─────────────────────────────────────────────────────────────┐
│                      本章总结                                │
├─────────────────────────────────────────────────────────────────┤
│                                                             │
│  ✓ Cobra 是 Go 语言最流行的 CLI 框架                         │
│                                                             │
│  ✓ 核心组件：Command、Arg、Flag                              │
│                                                             │
│  ✓ 命令执行流程：解析 → 验证 → 执行 → 返回                   │
│                                                             │
│  ✓ 支持多种参数验证器                                        │
│                                                             │
│  ✓ 支持多种标志类型                                         │
│                                                             │
│  ✓ 自动生成帮助信息                                         │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

---

## 下章预告

下一章我们将学习 **Cobra 进阶功能详解**，包括：
- 子命令嵌套
- 持久标志
- 命令组
- 配置管理
- 交互式输入

敬请期待！
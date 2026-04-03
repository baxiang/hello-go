# 05-命令行工具

本章节介绍如何使用Cobra框架构建专业的命令行工具。

## 学习目标

完成本章节学习后，你将能够：

- 使用Cobra创建CLI应用
- 实现命令和子命令
- 添加命令行参数和选项
- 生成帮助文档
- 构建专业的命令行工具

## 章节内容

### 01-Cobra命令行工具

**核心内容**：
- Cobra框架介绍
- 命令和子命令
- 参数和选项
- 帮助文档生成
- 实战案例

**学习时间**：2-3天

**实践要求**：
- [ ] 安装Cobra工具
- [ ] 创建基础CLI应用
- [ ] 添加命令和子命令
- [ ] 实现参数解析

## 前置知识

- 完成01-环境与工具章节
- 完成04-函数与指针章节
- 理解Go Modules依赖管理

## 核心概念

### Cobra简介

Cobra是Go语言最流行的命令行应用框架，被众多知名项目使用：
- Docker
- Kubernetes
- Hugo
- GitHub CLI

### 基础使用

```bash
# 安装Cobra
go get -u github.com/spf13/cobra/cobra

# 创建新项目
cobra init --pkg-name myapp

# 添加命令
cobra add version
```

### 命令结构

```go
package cmd

import (
    "github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
    Use:   "myapp",
    Short: "A brief description of your application",
    Long: `A longer description that spans multiple lines and
likely contains examples and usage of using your application.`,
    Run: func(cmd *cobra.Command, args []string) {
        // 主逻辑
    },
}

func Execute() {
    if err := rootCmd.Execute(); err != nil {
        fmt.Println(err)
        os.Exit(1)
    }
}
```

### 子命令

```go
var versionCmd = &cobra.Command{
    Use:   "version",
    Short: "Print the version number",
    Run: func(cmd *cobra.Command, args []string) {
        fmt.Println("MyApp v1.0.0")
    },
}

func init() {
    rootCmd.AddCommand(versionCmd)
}
```

### 参数和选项

```go
var (
    name     string
    verbose  bool
    port     int
)

func init() {
    rootCmd.Flags().StringVarP(&name, "name", "n", "", "Your name")
    rootCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output")
    rootCmd.Flags().IntVarP(&port, "port", "p", 8080, "Port number")
}

// 使用
// myapp --name John --verbose --port 9090
```

## 学习重点

### 1. 命令组织

```go
// 多级子命令
git add
git commit
git remote add

// 实现
var remoteCmd = &cobra.Command{
    Use:   "remote",
    Short: "Manage remote repositories",
}

var remoteAddCmd = &cobra.Command{
    Use:   "add",
    Short: "Add a remote repository",
    Run: func(cmd *cobra.Command, args []string) {
        // 添加远程仓库逻辑
    },
}

func init() {
    rootCmd.AddCommand(remoteCmd)
    remoteCmd.AddCommand(remoteAddCmd)
}
```

### 2. 参数验证

```go
var createCmd = &cobra.Command{
    Use:   "create [name]",
    Short: "Create a new resource",
    Args:  cobra.ExactArgs(1),  // 必须有且只有1个参数
    Run: func(cmd *cobra.Command, args []string) {
        name := args[0]
        fmt.Printf("Creating %s...\n", name)
    },
}

// Args验证器
// NoArgs - 不接受参数
// ArbitraryArgs - 接受任意参数
// MinimumNArgs(int) - 最少N个参数
// MaximumNArgs(int) - 最多N个参数
// ExactArgs(int) - 精确N个参数
```

### 3. PreRun和PostRun

```go
var cmd = &cobra.Command{
    Use:   "example",
    Short: "Example command",
    PreRun: func(cmd *cobra.Command, args []string) {
        fmt.Println("PreRun: 准备工作")
    },
    Run: func(cmd *cobra.Command, args []string) {
        fmt.Println("Run: 主逻辑")
    },
    PostRun: func(cmd *cobra.Command, args []string) {
        fmt.Println("PostRun: 清理工作")
    },
}
// 执行顺序: PreRun -> Run -> PostRun
```

## 检查清单

完成本章节后，验证你的掌握程度：

- [ ] 安装并配置Cobra
- [ ] 创建基础CLI应用
- [ ] 添加命令和子命令
- [ ] 实现参数和选项解析
- [ ] 生成帮助文档
- [ ] 完成一个实用的CLI工具

## 最佳实践

### 1. 项目结构

```
myapp/
├── cmd/
│   ├── root.go
│   ├── version.go
│   └── create.go
├── main.go
├── go.mod
└── go.sum
```

### 2. 配置文件集成

```go
import "github.com/spf13/viper"

func init() {
    cobra.OnInitialize(initConfig)
    rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file")
}

func initConfig() {
    if cfgFile != "" {
        viper.SetConfigFile(cfgFile)
    } else {
        viper.AddConfigPath("$HOME/.myapp")
        viper.SetConfigName("config")
    }
    viper.AutomaticEnv()
    viper.ReadInConfig()
}
```

### 3. 错误处理

```go
var cmd = &cobra.Command{
    Use:   "create",
    Short: "Create a resource",
    RunE: func(cmd *cobra.Command, args []string) error {
        // 返回error而不是直接退出
        if err := create(); err != nil {
            return err
        }
        return nil
    },
}
```

## 实战案例

### 构建一个待办事项CLI

```go
package cmd

import (
    "encoding/json"
    "fmt"
    "os"
    "github.com/spf13/cobra"
)

type Todo struct {
    ID    int    `json:"id"`
    Title string `json:"title"`
    Done  bool   `json:"done"`
}

var todos []Todo
var dataFile = "todos.json"

var rootCmd = &cobra.Command{
    Use:   "todo",
    Short: "A todo list CLI",
}

var addCmd = &cobra.Command{
    Use:   "add [title]",
    Short: "Add a new todo",
    Args:  cobra.ExactArgs(1),
    Run: func(cmd *cobra.Command, args []string) {
        loadTodos()
        todo := Todo{
            ID:    len(todos) + 1,
            Title: args[0],
            Done:  false,
        }
        todos = append(todos, todo)
        saveTodos()
        fmt.Printf("Added: %s\n", args[0])
    },
}

var listCmd = &cobra.Command{
    Use:   "list",
    Short: "List all todos",
    Run: func(cmd *cobra.Command, args []string) {
        loadTodos()
        for _, todo := range todos {
            status := " "
            if todo.Done {
                status = "✓"
            }
            fmt.Printf("[%s] %d. %s\n", status, todo.ID, todo.Title)
        }
    },
}

func init() {
    rootCmd.AddCommand(addCmd, listCmd)
}

func loadTodos() {
    data, err := os.ReadFile(dataFile)
    if err != nil {
        return
    }
    json.Unmarshal(data, &todos)
}

func saveTodos() {
    data, _ := json.MarshalIndent(todos, "", "  ")
    os.WriteFile(dataFile, data, 0644)
}

func Execute() {
    rootCmd.Execute()
}
```

## 延伸阅读

- [Cobra官方文档](https://github.com/spf13/cobra)
- [Cobra生成器](https://github.com/spf13/cobra/blob/master/cobra/README.md)
- [Go命令行应用最佳实践](https://github.com/urfave/cli)
- [12 Factor CLI Apps](https://medium.com/@jdxcode/12-factor-cli-apps-dd3c227a0e46)
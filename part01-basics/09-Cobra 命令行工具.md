# Cobra 命令行工具

## 简介

Cobra 是 Go 的命令行应用框架，特点：
- 子命令系统 (如 git clone, git commit)
- 自动帮助信息生成
- 标志 (flags) 解析
- Bash/Zsh/Fish 自动补全
- 被 hugo, kubectl, docker 等项目使用

## 安装

```bash
go get github.com/spf13/cobra
```

## 快速开始

### 基础示例

```go
package main

import (
    "fmt"
    "github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
    Use:   "myapp",
    Short: "我的应用",
    Long: `myapp 是一个示例命令行应用。
它可以执行各种有用的任务。`,
    Run: func(cmd *cobra.Command, args []string) {
        fmt.Println("Hello, Cobra!")
    },
}

func main() {
    if err := rootCmd.Execute(); err != nil {
        fmt.Println(err)
    }
}
```

### 运行

```bash
go run main.go
# Hello, Cobra!

go run main.go --help
# myapp 是一个示例命令行应用。
# 它可以执行各种有用的任务。
# Usage:
#   myapp [flags]
```

## 命令详解

### 创建子命令

```go
// cmd/user.go
var userCmd = &cobra.Command{
    Use:   "user",
    Short: "用户管理",
    Long:  "用于管理用户的子命令",
}

// cmd/user_create.go
var userCreateCmd = &cobra.Command{
    Use:     "create [username]",
    Short:   "创建用户",
    Long:    "创建一个新的用户账号",
    Args:    cobra.ExactArgs(1),  // 必须有一个参数
    Aliases: []string{"add", "new"},
    Run: func(cmd *cobra.Command, args []string) {
        username := args[0]
        fmt.Println("创建用户:", username)
    },
}

// 在 rootCmd 初始化时添加
func init() {
    rootCmd.AddCommand(userCmd)
    userCmd.AddCommand(userCreateCmd)
}
```

### 使用

```bash
./myapp user create alice
./myapp user add alice      # 使用别名
./myapp user --help
```

## 标志 (Flags)

### 定义标志

```go
var (
    name     string
    age      int
    debug    bool
    tags     []string
    metadata map[string]string
)

func init() {
    // 本地标志 (仅当前命令)
    userCreateCmd.Flags().StringVarP(&name, "name", "n", "", "用户名称")
    userCreateCmd.Flags().IntVarP(&age, "age", "a", 0, "用户年龄")
    userCreateCmd.Flags().BoolVarP(&debug, "debug", "d", false, "调试模式")
    userCreateCmd.Flags().StringSliceVar(&tags, "tags", []string{}, "标签")
    userCreateCmd.Flags().StringToStringVar(&metadata, "meta", map[string]string{}, "元数据")

    // 必需标志
    userCreateCmd.MarkFlagRequired("name")

    // 全局标志
    rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "配置文件路径")
    rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "详细输出")
}
```

### 读取标志

```go
Run: func(cmd *cobra.Command, args []string) {
    // 方式 1: 使用变量
    fmt.Println("Name:", name)
    fmt.Println("Age:", age)

    // 方式 2: 从命令获取
    name2, _ := cmd.Flags().GetString("name")
    age2, _ := cmd.Flags().GetInt("age")
    debug2, _ := cmd.Flags().GetBool("debug")
    tags2, _ := cmd.Flags().GetStringSlice("tags")

    // 获取父命令的全局标志
    verbose, _ := cmd.Root().PersistentFlags().GetBool("verbose")

    fmt.Printf("创建用户：%s, 年龄：%d\n", name, age)
}
```

### 标志类型

```go
// 所有支持的标志类型
cmd.Flags().String("name", "", "名称")
cmd.Flags().StringP("name", "n", "", "名称 (带缩写)")
cmd.Flags().Bool("debug", false, "调试")
cmd.Flags().Int("port", 8080, "端口")
cmd.Flags().Int64("size", 0, "大小")
cmd.Flags().Float64("ratio", 0.0, "比例")
cmd.Flags().Duration("timeout", time.Second, "超时")
cmd.Flags().StringSlice("tags", []string{}, "标签")
cmd.Flags().IntSlice("ports", []int{}, "端口列表")
cmd.Flags().StringToString("meta", map[string]string{}, "元数据")
cmd.Flags().StringArray("command", []string{}, "命令 (保留顺序)")

// 绑定到环境变量
cmd.Flags().String("api-key", "", "API 密钥")
viper.BindPFlag("api-key", cmd.Flags().Lookup("api-key"))
```

## 参数验证

```go
import "github.com/spf13/cobra"

// 参数验证器
var cmd = &cobra.Command{
    Use: "cmd",

    // 不接收参数
    Args: cobra.NoArgs,

    // 恰好 1 个参数
    Args: cobra.ExactArgs(1),

    // 最少 1 个参数
    Args: cobra.MinimumNArgs(1),

    // 最多 2 个参数
    Args: cobra.MaximumNArgs(2),

    // 1 到 2 个参数
    Args: cobra.RangeArgs(1, 2),

    // 任意参数 (默认)
    Args: cobra.ArbitraryArgs,

    // 只接受已定义的 flag
    Args: cobra.OnlyValidArgs,

    // 自定义验证
    Args: func(cmd *cobra.Command, args []string) error {
        if len(args) < 2 {
            return fmt.Errorf("至少需要 2 个参数")
        }
        if args[0] != "valid" {
            return fmt.Errorf("第一个参数必须是 'valid'")
        }
        return nil
    },

    Run: func(cmd *cobra.Command, args []string) {
        // ...
    },
}
```

## 完整示例：Todo CLI

### 项目结构

```
todo/
├── cmd/
│   └── root.go
├── main.go
├── todo/
│   └── store.go
└── go.mod
```

### 主程序

```go
// main.go
package main

import "github.com/example/todo/cmd"

func main() {
    cmd.Execute()
}
```

### 数据存储

```go
// todo/store.go
package todo

import (
    "encoding/json"
    "os"
    "path/filepath"
)

type Todo struct {
    ID     int    `json:"id"`
    Title  string `json:"title"`
    Done   bool   `json:"done"`
}

type Store struct {
    file string
    Todos []Todo `json:"todos"`
}

func NewStore() (*Store, error) {
    home, _ := os.UserHomeDir()
    file := filepath.Join(home, ".todo.json")

    store := &Store{file: file, Todos: []Todo{}}

    // 读取现有数据
    if data, err := os.ReadFile(file); err == nil {
        json.Unmarshal(data, &store.Todos)
    }

    return store, nil
}

func (s *Store) Save() error {
    data, err := json.MarshalIndent(s.Todos, "", "  ")
    if err != nil {
        return err
    }
    return os.WriteFile(s.file, data, 0644)
}

func (s *Store) Add(title string) *Todo {
    todo := Todo{
        ID:    len(s.Todos) + 1,
        Title: title,
        Done:  false,
    }
    s.Todos = append(s.Todos, &todo)
    s.Save()
    return &todo
}

func (s *Store) List() []Todo {
    return s.Todos
}

func (s *Store) Complete(id int) error {
    for i, todo := range s.Todos {
        if todo.ID == id {
            s.Todos[i].Done = true
            return s.Save()
        }
    }
    return fmt.Errorf("todo %d not found", id)
}

func (s *Store) Delete(id int) error {
    for i, todo := range s.Todos {
        if todo.ID == id {
            s.Todos = append(s.Todos[:i], s.Todos[i+1:]...)
            return s.Save()
        }
    }
    return fmt.Errorf("todo %d not found", id)
}
```

### 根命令

```go
// cmd/root.go
package cmd

import (
    "fmt"
    "os"
    "github.com/spf13/cobra"
    "github.com/example/todo/todo"
)

var (
    store *todo.Store
)

var rootCmd = &cobra.Command{
    Use:   "todo",
    Short: "Todo 命令行工具",
    Long:  "一个简单易用的待办事项管理工具",
    PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
        var err error
        store, err = todo.NewStore()
        return err
    },
}

func Execute() {
    if err := rootCmd.Execute(); err != nil {
        os.Exit(1)
    }
}

func init() {
    rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "配置文件")
}
```

### 添加命令

```go
// cmd/add.go
var addCmd = &cobra.Command{
    Use:     "add [title]",
    Short:   "添加待办事项",
    Aliases: []string{"a"},
    Args:    cobra.ExactArgs(1),
    RunE: func(cmd *cobra.Command, args []string) error {
        title := args[0]
        todo := store.Add(title)
        fmt.Printf("已添加待办事项 #%d: %s\n", todo.ID, todo.Title)
        return nil
    },
}

func init() {
    rootCmd.AddCommand(addCmd)
}
```

### 列表命令

```go
// cmd/list.go
var (
    showDone bool
    limit    int
)

var listCmd = &cobra.Command{
    Use:     "list",
    Short:   "列出待办事项",
    Aliases: []string{"ls"},
    RunE: func(cmd *cobra.Command, args []string) error {
        todos := store.List()

        for _, t := range todos {
            if !showDone && t.Done {
                continue
            }

            status := "[ ]"
            if t.Done {
                status = "[x]"
            }
            fmt.Printf("%s #%d %s\n", status, t.ID, t.Title)
        }

        return nil
    },
}

func init() {
    rootCmd.AddCommand(listCmd)

    listCmd.Flags().BoolVar(&showDone, "all", false, "显示已完成的")
    listCmd.Flags().IntVarP(&limit, "limit", "l", 10, "显示数量限制")
}
```

### 完成命令

```go
// cmd/complete.go
var completeCmd = &cobra.Command{
    Use:     "complete [id]",
    Short:   "完成待办事项",
    Aliases: []string{"done"},
    Args:    cobra.ExactArgs(1),
    RunE: func(cmd *cobra.Command, args []string) error {
        id, _ := strconv.Atoi(args[0])
        if err := store.Complete(id); err != nil {
            return err
        }
        fmt.Printf("已完成待办事项 #%d\n", id)
        return nil
    },
}

func init() {
    rootCmd.AddCommand(completeCmd)
}
```

### 删除命令

```go
// cmd/delete.go
var deleteCmd = &cobra.Command{
    Use:     "delete [id]",
    Short:   "删除待办事项",
    Aliases: []string{"rm"},
    Args:    cobra.ExactArgs(1),
    RunE: func(cmd *cobra.Command, args []string) error {
        id, _ := strconv.Atoi(args[0])
        if err := store.Delete(id); err != nil {
            return err
        }
        fmt.Printf("已删除待办事项 #%d\n", id)
        return nil
    },
}

func init() {
    rootCmd.AddCommand(deleteCmd)
}
```

### 使用

```bash
# 添加
./todo add "学习 Go"
./todo a "学习 Cobra"

# 列表
./todo list
./todo ls --all
./todo ls -l 5

# 完成
./todo complete 1
./todo done 1

# 删除
./todo delete 1
./todo rm 1

# 帮助
./todo --help
./todo add --help
```

## 自动生成补全脚本

```go
// cmd/completion.go
var completionCmd = &cobra.Command{
    Use:   "completion [bash|zsh|fish|powershell]",
    Short: "生成补全脚本",
    Long: `生成 shell 补全脚本。

示例:
  # Bash
  source <(todo completion bash)

  # Zsh
  source <(todo completion zsh)

  # Fish
  todo completion fish | source
`,
    ValidArgs: []string{"bash", "zsh", "fish", "powershell"},
    Args:      cobra.ExactValidArgs(1),
    Run: func(cmd *cobra.Command, args []string) {
        switch args[0] {
        case "bash":
            cmd.Root().GenBashCompletion(os.Stdout)
        case "zsh":
            cmd.Root().GenZshCompletion(os.Stdout)
        case "fish":
            cmd.Root().GenFishCompletion(os.Stdout, true)
        case "powershell":
            cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
        }
    },
}

func init() {
    rootCmd.AddCommand(completionCmd)
}
```

## 配置和颜色

```go
import (
    "github.com/fatih/color"
)

// 定义颜色
var (
    red    = color.New(color.FgRed).SprintFunc()
    green  = color.New(color.FgGreen).SprintFunc()
    yellow = color.New(color.FgYellow).SprintFunc()
)

// 使用
fmt.Println(green("成功"), red("失败"), yellow("警告"))

// 带格式的 Printf
green.Printf("任务已完成\n")
```

## Cobra 检查清单

```
[ ] 使用子命令组织功能
[ ] 提供清晰的 Short 和 Long 描述
[ ] 为命令设置 Aliases
[ ] 使用 Args 验证参数
[ ] 使用 RunE 返回错误
[ ] 添加 completion 命令
[ ] 使用 PersistentFlags 全局标志
[ ] 设置必需标志
[ ] 提供 --help 帮助信息
[ ] 错误信息清晰有用
```

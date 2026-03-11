# Cobra 进阶功能详解

> 本章将深入介绍 Cobra 的进阶功能，包括子命令嵌套、持久标志、命令组、配置管理和交互式输入等。

## 2.1 子命令嵌套

### 2.1.1 多层嵌套结构

Cobra 支持无限层级的命令嵌套：

```
┌─────────────────────────────────────────────────────────────────┐
│                    子命令嵌套示例                                │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  app (根命令)                                                  │
│       │                                                         │
│       ├── user (子命令)                                        │
│       │       │                                                 │
│       │       ├── create (孙命令)                              │
│       │       │                                                 │
│       │       ├── delete (孙命令)                              │
│       │       │                                                 │
│       │       └── list (孙命令)                                │
│       │                                                         │
│       ├── config (子命令)                                      │
│       │       │                                                 │
│       │       ├── get (孙命令)                                 │
│       │       │                                                 │
│       │       └── set (孙命令)                                 │
│       │                                                         │
│       └── server (子命令)                                      │
│               │                                                 │
│               ├── start (孙命令)                               │
│               │                                                 │
│               └── stop (孙命令)                                 │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

### 2.1.2 代码实现

```go
package main

import (
    "fmt"
    "github.com/spf13/cobra"
)

func main() {
    // 根命令
    var rootCmd = &cobra.Command{Use: "app"}

    // user 子命令
    var userCmd = &cobra.Command{
        Use:   "user",
        Short: "用户管理",
    }

    // user create 孙命令
    var createCmd = &cobra.Command{
        Use:   "create",
        Short: "创建用户",
        Run: func(cmd *cobra.Command, args []string) {
            fmt.Println("创建用户")
        },
    }

    // user delete 孙命令
    var deleteCmd = &cobra.Command{
        Use:   "delete",
        Short: "删除用户",
        Run: func(cmd *cobra.Command, args []string) {
            fmt.Println("删除用户")
        },
    }

    // user list 孙命令
    var listCmd = &cobra.Command{
        Use:   "list",
        Short: "列出用户",
        Run: func(cmd *cobra.Command, args []string) {
            fmt.Println("列出用户")
        },
    }

    // config 子命令
    var configCmd = &cobra.Command{
        Use:   "config",
        Short: "配置管理",
    }

    // config get 孙命令
    var configGetCmd = &cobra.Command{
        Use:   "get",
        Short: "获取配置",
        Run: func(cmd *cobra.Command, args []string) {
            fmt.Println("获取配置")
        },
    }

    // config set 孙命令
    var configSetCmd = &cobra.Command{
        Use:   "set",
        Short: "设置配置",
        Run: func(cmd *cobra.Command, args []string) {
            fmt.Println("设置配置")
        },
    }

    // 组装命令树
    userCmd.AddCommand(createCmd, deleteCmd, listCmd)
    configCmd.AddCommand(configGetCmd, configSetCmd)
    rootCmd.AddCommand(userCmd, configCmd)

    rootCmd.Execute()
}
```

运行示例：

```bash
$ go run main.go user create
创建用户

$ go run main.go config get
获取配置

$ go run main.go --help
Usage:
  app [command]

Available Commands:
  config   配置管理
  user    用户管理

Use "app [command] --help" for more information about a command.
```

## 2.2 持久标志与局部标志

### 2.2.1 概念区分

```
┌─────────────────────────────────────────────────────────────────┐
│                    标志作用域                                  │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  持久标志 (Persistent Flags)                                    │
│  ├── 对当前命令及其所有子命令生效                               │
│  ├── 通常用于全局配置，如 --verbose、--config                  │
│  └── 通过 cmd.Flags().MarkPersistentFlag() 设置               │
│                                                                 │
│  局部标志 (Local Flags)                                        │
│  ├── 仅对当前命令生效                                         │
│  ├── 用于特定命令的特定功能                                    │
│  └── 通过 cmd.Flags().String() 等直接设置                      │
│                                                                 │
│  继承标志 (Inherited Flags)                                    │
│  ├── 子命令自动继承父命令的持久标志                            │
│  └── 子命令可以直接使用父命令的标志                            │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

### 2.2.2 代码示例

```go
package main

import (
    "fmt"
    "github.com/spf13/cobra"
)

var (
    verbose   bool
    config    string
    name      string
    age       int
)

func main() {
    // 根命令
    rootCmd := &cobra.Command{Use: "app"}

    // user 子命令
    userCmd := &cobra.Command{
        Use:   "user",
        Short: "用户管理",
    }

    // user create 孙命令
    createCmd := &cobra.Command{
        Use:   "create",
        Short: "创建用户",
        Run: func(cmd *cobra.Command, args []string) {
            fmt.Printf("创建用户: name=%s, age=%d\n", name, age)
            fmt.Printf("全局配置: verbose=%v, config=%s\n", verbose, config)
        },
    }

    // ===== 持久标志 ======
    // 对 rootCmd 设置持久标志，对所有子命令生效
    rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "详细输出")
    rootCmd.PersistentFlags().StringVar(&config, "config", "", "配置文件路径")

    // ===== 局部标志 ======
    // 仅对 createCmd 生效
    createCmd.Flags().StringVarP(&name, "name", "n", "", "用户姓名")
    createCmd.Flags().IntVarP(&age, "age", "a", 0, "用户年龄")
    createCmd.MarkFlagRequired("name")

    userCmd.AddCommand(createCmd)
    rootCmd.AddCommand(userCmd)
    rootCmd.Execute()
}
```

运行示例：

```bash
# 根命令的持久标志可以在子命令中使用
$ go run main.go user create -n 张三 -a 25 --verbose
创建用户: name=张三, age=25
全局配置: verbose=true, config=

# 从根命令继承的持久标志
$ go run main.go user create -n 李四 --config /etc/app.yaml
创建用户: name=李四, age=0
全局配置: verbose=false, config=/etc/app.yaml

# 局部标志只能在特定命令中使用
$ go run main.go user --help
Usage:
  app user [command]

Available Commands:
  create   创建用户

Flags:
  -h, --help   help for user

# create 命令的局部标志
$ go run main.go user create --help
Usage:
  app user create [flags]

Flags:
  -a, --age int        用户年龄
  -h, --help           help for create
  -n, --name string     用户姓名 (required)
```

## 2.3 命令组

### 2.3.1 命令分组

```go
package main

import (
    "fmt"
    "github.com/spf13/cobra"
)

func main() {
    rootCmd := &cobra.Command{Use: "app"}

    // 创建用户管理命令组
    userCmd := &cobra.Command{
        Use:     "user",
        Short:   "用户管理",
        aliases: []string{"u"},
    }

    // 创建管理命令组
    adminCmd := &cobra.Command{
        Use:     "admin",
        Short:   "系统管理",
        aliases: []string{"a"},
    }

    // 添加命令到组
    userCmd.AddCommand(
        &cobra.Command{Use: "create", Short: "创建用户", Run: func(cmd *cobra.Command, args []string) { fmt.Println("创建用户") }},
        &cobra.Command{Use: "delete", Short: "删除用户", Run: func(cmd *cobra.Command, args []string) { fmt.Println("删除用户") }},
    )

    adminCmd.AddCommand(
        &cobra.Command{Use: "backup", Short: "备份系统", Run: func(cmd *cobra.Command, args []string) { fmt.Println("备份系统") }},
        &cobra.Command{Use: "restore", Short: "恢复系统", Run: func(cmd *cobra.Command, args []string) { fmt.Println("恢复系统") }},
    )

    // 设置命令组
    rootCmd.AddCommand(userCmd, adminCmd)

    // 设置命令组描述
    userCmd.GroupID = "user"
    adminCmd.GroupID = "admin"

    rootCmd.Execute()
}
```

### 2.3.2 帮助信息中的命令组

```
$ app --help
Usage:
  app [command]

Commands:
  User Management:
    user         用户管理

  System Administration:
    admin        系统管理

Flags:
  -h, --help   help for app
```

## 2.4 配置管理

### 2.4.1 使用 Viper 集成

Cobra 可以与 Viper 无缝集成，实现配置管理：

```go
package main

import (
    "fmt"
    "github.com/spf13/cobra"
    "github.com/spf13/viper"
)

var (
    cfgFile string
    name   string
)

func main() {
    rootCmd := &cobra.Command{
        Use:   "app",
        Short: "示例应用",
    }

    // 添加配置文件标志
    rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "配置文件路径")

    // 初始化 Viper
    initConfig()

    // server 命令
    serverCmd := &cobra.Command{
        Use:   "server",
        Short: "启动服务器",
        Run: func(cmd *cobra.Command, args []string) {
            fmt.Printf("启动服务器: %s\n", viper.GetString("server.host"))
            fmt.Printf("端口: %d\n", viper.GetInt("server.port"))
        },
    }

    rootCmd.AddCommand(serverCmd)
    rootCmd.Execute()
}

func initConfig() {
    if cfgFile != "" {
        viper.SetConfigFile(cfgFile)
    } else {
        viper.SetConfigName(".app")
        viper.AddConfigPath(".")
        viper.AddConfigPath("$HOME/.app")
        viper.AddConfigPath("/etc/app/")
    }

    viper.AutomaticEnv()

    if err := viper.ReadInConfig(); err == nil {
        fmt.Println("使用配置文件:", viper.ConfigFileUsed())
    }
}
```

### 2.4.2 配置文件示例

```yaml
# config.yaml
server:
  host: localhost
  port: 8080
  timeout: 30

database:
  host: localhost
  port: 3306
  name: myapp

log:
  level: info
  file: /var/log/app.log
```

## 2.5 交互式输入

### 2.5.1 使用 Survey 库

Cobra 可以与 survey 库集成，实现交互式输入：

```go
package main

import (
    "fmt"
    "github.com/AlecAivazis/survey/v2"
    "github.com/spf13/cobra"
)

func main() {
    rootCmd := &cobra.Command{
        Use:   "app",
        Short: "示例应用",
    }

    // 交互式创建用户
    createCmd := &cobra.Command{
        Use:   "create-interactive",
        Short: "交互式创建用户",
        Run: func(cmd *cobra.Command, args []string) {
            // 定义问题
            qs := []*survey.Question{
                {
                    Name: "name",
                    Prompt: &survey.Input{
                        Message: "请输入用户名:",
                    },
                    Validate: survey.Required,
                },
                {
                    Name: "email",
                    Prompt: &survey.Input{
                        Message: "请输入邮箱:",
                    },
                    Validate: survey.Required,
                },
                {
                    Name: "age",
                    Prompt: &survey.Input{
                        Message: "请输入年龄:",
                        Default: "18",
                    },
                },
            }

            // 收集答案
            answers := struct {
                Name  string
                Email string
                Age   string
            }{}

            if err := survey.Ask(qs, &answers); err != nil {
                fmt.Println("错误:", err)
                return
            }

            fmt.Printf("创建用户: name=%s, email=%s, age=%s\n", 
                answers.Name, answers.Email, answers.Age)
        },
    }

    rootCmd.AddCommand(createCmd)
    rootCmd.Execute()
}
```

### 2.5.2 选择确认示例

```go
package main

import (
    "fmt"
    "github.com/AlecAivazis/survey/v2"
    "github.com/spf13/cobra"
)

func main() {
    rootCmd := &cobra.Command{
        Use:   "app",
        Short: "示例应用",
    }

    // 确认操作
    confirmCmd := &cobra.Command{
        Use:   "confirm",
        Short: "确认操作",
        Run: func(cmd *cobra.Command, args []string) {
            var confirm bool
            prompt := &survey.Confirm{
                Message: "确定要删除这个文件吗?",
                Default: false,
            }
            survey.AskOne(prompt, &confirm)

            if confirm {
                fmt.Println("文件已删除")
            } else {
                fmt.Println("操作已取消")
            }
        },
    }

    // 选择操作
    selectCmd := &cobra.Command{
        Use:   "select",
        Short: "选择操作",
        Run: func(cmd *cobra.Command, args []string) {
            var choice string
            prompt := &survey.Select{
                Message: "请选择操作:",
                Options: []string{"创建", "更新", "删除", "查看"},
                Default: "查看",
            }
            survey.AskOne(prompt, &choice)

            fmt.Printf("选择: %s\n", choice)
        },
    }

    rootCmd.AddCommand(confirmCmd, selectCmd)
    rootCmd.Execute()
}
```

## 2.6 错误处理

### 2.6.1 命令级错误处理

```go
package main

import (
    "errors"
    "fmt"
    "github.com/spf13/cobra"
)

func main() {
    rootCmd := &cobra.Command{Use: "app"}

    // 使用 RunE 返回错误
    cmd := &cobra.Command{
        Use:   "divide",
        Short: "除法运算",
        Args:  cobra.ExactArgs(2),
        RunE: func(cmd *cobra.Command, args []string) error {
            var a, b float64
            fmt.Sscanf(args[0], "%f", &a)
            fmt.Sscanf(args[1], "%f", &b)

            if b == 0 {
                return errors.New("除数不能为零")
            }

            fmt.Printf("结果: %.2f\n", a/b)
            return nil
        },
    }

    rootCmd.AddCommand(cmd)
    rootCmd.Execute()
}
```

### 2.6.2 全局错误处理

```go
package main

import (
    "fmt"
    "github.com/spf13/cobra"
)

func main() {
    rootCmd := &cobra.Command{
        Use:   "app",
        Short: "示例应用",
    }

    // 设置错误处理函数
    rootCmd.SetFlagErrorFunc(func(cmd *cobra.Command, err error) error {
        fmt.Fprintf(cmd.OutOrStderr(), "错误: %v\n", err)
        return err
    })

    // 设置帮助函数
    rootCmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
        fmt.Fprintf(cmd.OutOrStdout(), "帮助信息\n")
    })

    rootCmd.Execute()
}
```

## 2.7 钩子函数

### 2.7.1 命令钩子

```go
package main

import (
    "fmt"
    "github.com/spf13/cobra"
)

func main() {
    var verbose bool

    rootCmd := &cobra.Command{Use: "app"}

    // 父命令
    parentCmd := &cobra.Command{
        Use:   "parent",
        Short: "父命令",
        PreRun: func(cmd *cobra.Command, args []string) {
            fmt.Println("Parent PreRun")
        },
        Run: func(cmd *cobra.Command, args []string) {
            fmt.Println("Parent Run")
        },
        PostRun: func(cmd *cobra.Command, args []string) {
            fmt.Println("Parent PostRun")
        },
    }

    // 子命令
    childCmd := &cobra.Command{
        Use:   "child",
        Short: "子命令",
        PreRun: func(cmd *cobra.Command, args []string) {
            fmt.Println("Child PreRun")
        },
        Run: func(cmd *cobra.Command, args []string) {
            fmt.Println("Child Run")
        },
        PostRun: func(cmd *cobra.Command, args []string) {
            fmt.Println("Child PostRun")
        },
    }

    parentCmd.AddCommand(childCmd)
    rootCmd.AddCommand(parentCmd)
    rootCmd.Execute()
}
```

执行顺序：

```
$ go run main.go parent child
Parent PreRun
Child PreRun
Child Run
Child PostRun
Parent PostRun
```

## 2.8 本章小结

```
┌─────────────────────────────────────────────────────────────┐
│                      本章总结                                │
├─────────────────────────────────────────────────────────────────┤
│                                                             │
│  ✓ 子命令嵌套 - 支持无限层级命令结构                         │
│                                                             │
│  ✓ 持久标志 - 全局配置标志                                   │
│                                                             │
│  ✓ 命令组 - 帮助信息中的命令分组                             │
│                                                             │
│  ✓ 配置管理 - 与 Viper 集成                                 │
│                                                             │
│  ✓ 交互式输入 - 与 Survey 集成                               │
│                                                             │
│  ✓ 错误处理 - RunE 和错误函数                              │
│                                                             │
│  ✓ 钩子函数 - PreRun/Run/PostRun                          │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

---

## 下章预告

下一章我们将学习 **Cobra 完整项目示例**，包括：
- 项目结构设计
- 完整 CRUD 功能
- 配置文件支持
- 测试编写

敬请期待！
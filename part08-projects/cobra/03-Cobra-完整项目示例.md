# Cobra 完整项目示例

> 本章将创建一个完整的命令行工具项目，包含用户管理的 CRUD 功能，展示 Cobra 的最佳实践。

## 3.1 项目概述

### 3.1.1 项目功能

我们将创建一个名为 `usermgr` 的用户管理命令行工具：

```
┌─────────────────────────────────────────────────────────────────┐
│                    usermgr 功能列表                             │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  用户管理                                                      │
│  ├── user create     创建用户                                 │
│  ├── user list       列出用户                                 │
│  ├── user get        获取用户详情                              │
│  ├── user update     更新用户                                 │
│  └── user delete     删除用户                                 │
│                                                                 │
│  配置管理                                                      │
│  ├── config init     初始化配置                               │
│  ├── config show     显示配置                                 │
│  └── config edit    编辑配置                                 │
│                                                                 │
│  工具                                                          │
│  ├── version        显示版本                                  │
│  └── completion     生成自动补全脚本                          │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

### 3.1.2 项目结构

```
usermgr/
├── cmd/
│   ├── root.go           # 根命令
│   ├── user/
│   │   ├── create.go     # 创建用户
│   │   ├── list.go       # 列出用户
│   │   ├── get.go        # 获取用户
│   │   ├── update.go     # 更新用户
│   │   └── delete.go     # 删除用户
│   ├── config/
│   │   ├── init.go       # 初始化配置
│   │   ├── show.go       # 显示配置
│   │   └── edit.go       # 编辑配置
│   └── completion.go     # 自动补全
├── internal/
│   ├── config/
│   │   └── config.go     # 配置管理
│   ├── model/
│   │   └── user.go       # 用户模型
│   └── store/
│       └── store.go      # 数据存储
├── main.go               # 入口文件
├── go.mod
└── config.yaml           # 配置文件
```

## 3.2 入口文件

```go
// main.go
package main

import (
    "os"

    "usermgr/cmd"
)

func main() {
    // 执行命令
    if err := cmd.Execute(); err != nil {
        os.Exit(1)
    }
}
```

## 3.3 根命令

```go
// cmd/root.go
package cmd

import (
    "fmt"
    "os"

    "github.com/spf13/cobra"
    "github.com/spf13/viper"

    "usermgr/cmd/config"
    "usermgr/cmd/user"
)

var (
    // 版本信息
    version = "1.0.0"
    commit  = "dev"
    date    = "now"
)

// Execute 执行所有命令
func Execute() {
    rootCmd.Execute()
}

// rootCmd 是根命令
var rootCmd = &cobra.Command{
    Use:   "usermgr",
    Short: "usermgr 是一个用户管理命令行工具",
    Long: `usermgr 是一个功能强大的用户管理命令行工具。

支持的功能:
  - 用户 CRUD 操作
  - 配置管理
  - 自动补全

更多信息请访问: https://github.com/example/usermgr`,
    Version: version,
    PersistentPreRun: func(cmd *cobra.Command, args []string) {
        // 初始化配置
        initConfig()
    },
}

func init() {
    // 添加子命令
    rootCmd.AddCommand(user.UserCmd)
    rootCmd.AddCommand(config.ConfigCmd)
    rootCmd.AddCommand(completionCmd)

    // 全局标志
    rootCmd.PersistentFlags().StringP("config", "c", "", "配置文件路径")
    rootCmd.PersistentFlags().BoolP("verbose", "v", false, "详细输出")
}

func initConfig() {
    // 从标志获取配置文件路径
    configFile, _ := rootCmd.Flags().GetString("config")
    if configFile != "" {
        viper.SetConfigFile(configFile)
    } else {
        // 默认配置
        viper.SetConfigName("config")
        viper.AddConfigPath(".")
        viper.AddConfigPath("$HOME/.usermgr")
        viper.AddConfigPath("/etc/usermgr")
    }

    // 环境变量前缀
    viper.SetEnvPrefix("USERMGR")
    viper.AutomaticEnv()

    // 读取配置
    if err := viper.ReadInConfig(); err != nil {
        // 配置文件不存在时忽略错误
        if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
            fmt.Fprintf(os.Stderr, "读取配置文件错误: %v\n", err)
        }
    }
}
```

## 3.4 用户模型

```go
// internal/model/user.go
package model

import "time"

// User 用户模型
type User struct {
    ID        int       `json:"id"`
    Username  string    `json:"username"`
    Email     string    `json:"email"`
    Phone     string    `json:"phone"`
    Status    int       `json:"status"` // 1: 正常, 0: 禁用
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}

// UserStore 用户存储接口
type UserStore interface {
    Create(user *User) error
    Get(id int) (*User, error)
    GetByUsername(username string) (*User, error)
    List() ([]*User, error)
    Update(user *User) error
    Delete(id int) error
}
```

## 3.5 数据存储

```go
// internal/store/store.go
package store

import (
    "errors"
    "sync"

	"usermgr/internal/model"
)

// MemoryStore 内存存储
type MemoryStore struct {
    mu    sync.RWMutex
    users map[int]*model.User
    nextID int
}

// NewMemoryStore 创建内存存储
func NewMemoryStore() *MemoryStore {
    return &MemoryStore{
        users:  make(map[int]*model.User),
        nextID: 1,
    }
}

// Create 创建用户
func (s *MemoryStore) Create(user *model.User) error {
    s.mu.Lock()
    defer s.mu.Unlock()

    // 检查用户名是否存在
    for _, u := range s.users {
        if u.Username == user.Username {
            return errors.New("用户名已存在")
        }
    }

    user.ID = s.nextID
    user.CreatedAt = time.Now()
    user.UpdatedAt = time.Now()
    s.users[user.ID] = user
    s.nextID++

    return nil
}

// Get 获取用户
func (s *MemoryStore) Get(id int) (*model.User, error) {
    s.mu.RLock()
    defer s.mu.RUnlock()

    user, ok := s.users[id]
    if !ok {
        return nil, errors.New("用户不存在")
    }

    return user, nil
}

// GetByUsername 根据用户名获取
func (s *MemoryStore) GetByUsername(username string) (*model.User, error) {
    s.mu.RLock()
    defer s.mu.RUnlock()

    for _, u := range s.users {
        if u.Username == username {
            return u, nil
        }
    }

    return nil, errors.New("用户不存在")
}

// List 列出所有用户
func (s *MemoryStore) List() ([]*model.User, error) {
    s.mu.RLock()
    defer s.mu.RUnlock()

    users := make([]*model.User, 0, len(s.users))
    for _, u := range s.users {
        users = append(users, u)
    }

    return users, nil
}

// Update 更新用户
func (s *MemoryStore) Update(user *model.User) error {
    s.mu.Lock()
    defer s.mu.Unlock()

    if _, ok := s.users[user.ID]; !ok {
        return errors.New("用户不存在")
    }

    user.UpdatedAt = time.Now()
    s.users[user.ID] = user

    return nil
}

// Delete 删除用户
func (s *MemoryStore) Delete(id int) error {
    s.mu.Lock()
    defer s.mu.Unlock()

    if _, ok := s.users[id]; !ok {
        return errors.New("用户不存在")
    }

    delete(s.users, id)

    return nil
}

// 全局存储实例
var globalStore = NewMemoryStore()

// GetStore 获取全局存储
func GetStore() model.UserStore {
    return globalStore
}
```

## 3.6 用户命令

### 3.6.1 用户子命令

```go
// cmd/user/user.go
package user

import (
    "github.com/spf13/cobra"
)

// UserCmd 用户命令
var UserCmd = &cobra.Command{
    Use:   "user",
    Short: "用户管理",
    Long:  "用户管理命令，包括创建、查询、更新、删除等操作",
}

func init() {
    UserCmd.AddCommand(createCmd)
    UserCmd.AddCommand(listCmd)
    UserCmd.AddCommand(getCmd)
    UserCmd.AddCommand(updateCmd)
    UserCmd.AddCommand(deleteCmd)
}
```

### 3.6.2 创建用户

```go
// cmd/user/create.go
package user

import (
    "fmt"

    "github.com/spf13/cobra"
    "github.com/spf13/viper"

    "usermgr/internal/model"
    "usermgr/internal/store"
)

var (
    username  string
    email     string
    phone     string
)

var createCmd = &cobra.Command{
    Use:   "create",
    Short: "创建用户",
    Long:  "创建一个新用户",
    Example: `  usermgr user create -n john -e john@example.com -p 13800138000
  usermgr user create --name jane --email jane@example.com`,
    RunE: func(cmd *cobra.Command, args []string) error {
        // 验证参数
        if username == "" {
            return fmt.Errorf("用户名不能为空")
        }
        if email == "" {
            return fmt.Errorf("邮箱不能为空")
        }

        // 创建用户
        user := &model.User{
            Username: username,
            Email:    email,
            Phone:    phone,
            Status:   1,
        }

        // 保存到存储
        s := store.GetStore()
        if err := s.Create(user); err != nil {
            return err
        }

        // 输出结果
        fmt.Printf("用户创建成功!\n")
        fmt.Printf("ID: %d, 用户名: %s, 邮箱: %s\n", user.ID, user.Username, user.Email)

        // 如果是详细模式，输出更多信息
        if verbose, _ := cmd.Flags().GetBool("verbose"); verbose {
            fmt.Printf("创建时间: %s\n", user.CreatedAt.Format("2006-01-02 15:04:05"))
        }

        return nil
    },
}

func init() {
    createCmd.Flags().StringVarP(&username, "name", "n", "", "用户名 (必需)")
    createCmd.Flags().StringVarP(&email, "email", "e", "", "邮箱 (必需)")
    createCmd.Flags().StringVarP(&phone, "phone", "p", "", "手机号")

    createCmd.MarkFlagRequired("name")
    createCmd.MarkFlagRequired("email")
}
```

### 3.6.3 列出用户

```go
// cmd/user/list.go
package user

import (
    "fmt"
    "os"
    "text/tabwriter"

    "github.com/spf13/cobra"

    "usermgr/internal/store"
)

var listCmd = &cobra.Command{
    Use:   "list",
    Short: "列出用户",
    Long:  "列出所有用户",
    Example: `  usermgr user list
  usermgr user list --format json`,
    RunE: func(cmd *cobra.Command, args []string) error {
        s := store.GetStore()
        users, err := s.List()
        if err != nil {
            return err
        }

        // 检查是否有用户
        if len(users) == 0 {
            fmt.Println("暂无用户")
            return nil
        }

        // 输出格式
        format, _ := cmd.Flags().GetString("format")
        if format == "json" {
            // JSON 格式输出
            fmt.Println("TODO: JSON 输出")
        } else {
            // 表格格式输出
            w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
            fmt.Fprintln(w, "ID\t用户名\t邮箱\t手机号\t状态")
            fmt.Fprintln(w, "---\t------\t----\t----\t----")
            for _, u := range users {
                status := "正常"
                if u.Status == 0 {
                    status = "禁用"
                }
                fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%s\n", 
                    u.ID, u.Username, u.Email, u.Phone, status)
            }
            w.Flush()
        }

        fmt.Printf("\n共 %d 个用户\n", len(users))

        return nil
    },
}

func init() {
    listCmd.Flags().String("format", "", "输出格式 (table/json)")
}
```

### 3.6.4 获取用户

```go
// cmd/user/get.go
package user

import (
    "fmt"

    "github.com/spf13/cobra"

    "usermgr/internal/store"
)

var userID int

var getCmd = &cobra.Command{
    Use:   "get [id]",
    Short: "获取用户详情",
    Long:  "根据 ID 获取用户详情",
    Args:  cobra.ExactArgs(1),
    Example: `  usermgr user get 1
  usermgr user get --id 1`,
    RunE: func(cmd *cobra.Command, args []string) error {
        // 获取用户 ID
        var id int
        fmt.Sscanf(args[0], "%d", &id)

        // 查询用户
        s := store.GetStore()
        user, err := s.Get(id)
        if err != nil {
            return err
        }

        // 输出用户信息
        fmt.Println("用户信息:")
        fmt.Println("----------")
        fmt.Printf("ID:       %d\n", user.ID)
        fmt.Printf("用户名:   %s\n", user.Username)
        fmt.Printf("邮箱:     %s\n", user.Email)
        fmt.Printf("手机号:   %s\n", user.Phone)
        fmt.Printf("状态:     %s\n", map[int]string{1: "正常", 0: "禁用"}[user.Status])
        fmt.Printf("创建时间: %s\n", user.CreatedAt.Format("2006-01-02 15:04:05"))
        fmt.Printf("更新时间: %s\n", user.UpdatedAt.Format("2006-01-02 15:04:05"))

        return nil
    },
}
```

### 3.6.5 更新用户

```go
// cmd/user/update.go
package user

import (
    "fmt"

    "github.com/spf13/cobra"

    "usermgr/internal/store"
)

var (
    updateID    int
    newEmail    string
    newPhone    string
    newStatus   int
)

var updateCmd = &cobra.Command{
    Use:   "update",
    Short: "更新用户",
    Long:  "更新用户信息",
    Example: `  usermgr user update 1 -e newemail@example.com
  usermgr user update --id 1 --phone 13900139000 --status 0`,
    RunE: func(cmd *cobra.Command, args []string) error {
        // 获取用户 ID
        var id int
        if len(args) > 0 {
            fmt.Sscanf(args[0], "%d", &id)
        } else {
            id = updateID
        }

        if id == 0 {
            return fmt.Errorf("用户 ID 不能为空")
        }

        // 获取用户
        s := store.GetStore()
        user, err := s.Get(id)
        if err != nil {
            return err
        }

        // 更新字段
        if newEmail != "" {
            user.Email = newEmail
        }
        if newPhone != "" {
            user.Phone = newPhone
        }
        if newStatus != 0 {
            user.Status = newStatus
        }

        // 保存更新
        if err := s.Update(user); err != nil {
            return err
        }

        fmt.Printf("用户 %d 更新成功!\n", id)

        return nil
    },
}

func init() {
    updateCmd.Flags().IntVarP(&updateID, "id", "i", 0, "用户 ID")
    updateCmd.Flags().StringVarP(&newEmail, "email", "e", "", "新邮箱")
    updateCmd.Flags().StringVarP(&newPhone, "phone", "p", "", "新手机号")
    updateCmd.Flags().IntVarP(&newStatus, "status", "s", 0, "状态 (0: 禁用, 1: 正常)")
}
```

### 3.6.6 删除用户

```go
// cmd/user/delete.go
package user

import (
    "fmt"

    "github.com/spf13/cobra"

    "usermgr/internal/store"
)

var force bool

var deleteCmd = &cobra.Command{
    Use:   "delete [id]",
    Short: "删除用户",
    Long:  "删除指定用户",
    Example: `  usermgr user delete 1
  usermgr user delete 1 --force`,
    Args:  cobra.ExactArgs(1),
    RunE: func(cmd *cobra.Command, args []string) error {
        // 获取用户 ID
        var id int
        fmt.Sscanf(args[0], "%d", &id)

        // 获取用户
        s := store.GetStore()
        user, err := s.Get(id)
        if err != nil {
            return err
        }

        // 确认删除
        if !force {
            fmt.Printf("确定要删除用户 %s 吗? (y/N): ", user.Username)
            var confirm string
            fmt.Scanln(&confirm)
            if confirm != "y" && confirm != "Y" {
                fmt.Println("已取消")
                return nil
            }
        }

        // 删除用户
        if err := s.Delete(id); err != nil {
            return err
        }

        fmt.Printf("用户 %d 已删除\n", id)

        return nil
    },
}

func init() {
    deleteCmd.Flags().BoolVarP(&force, "force", "f", false, "强制删除")
}
```

## 3.7 配置命令

```go
// cmd/config/config.go
package config

import (
    "github.com/spf13/cobra"
)

// ConfigCmd 配置命令
var ConfigCmd = &cobra.Command{
    Use:   "config",
    Short: "配置管理",
    Long:  "管理应用程序配置",
}

func init() {
    ConfigCmd.AddCommand(initCmd)
    ConfigCmd.AddCommand(showCmd)
    ConfigCmd.AddCommand(editCmd)
}
```

## 3.8 自动补全

```go
// cmd/completion.go
package cmd

import (
    "os"

    "github.com/spf13/cobra"
)

var completionCmd = &cobra.Command{
    Use:   "completion [shell]",
    Short: "生成自动补全脚本",
    Long: `生成指定 shell 的自动补全脚本。

支持的 shell:
  - bash
  - zsh
  - fish
  - powershell

示例:
  # Bash
  source <(usermgr completion bash)

  # Zsh
  source <(usermgr completion zsh)

  # 持久化
  usermgr completion bash > /etc/bash_completion.d/usermgr`,
    Args:      cobra.ExactArgs(1),
    ValidArgs: []string{"bash", "zsh", "fish", "powershell"},
    RunE: func(cmd *cobra.Command, args []string) error {
        shell := args[0]

        var err error
        switch shell {
        case "bash":
            err = rootCmd.GenBashCompletion(os.Stdout)
        case "zsh":
            err = rootCmd.GenZshCompletion(os.Stdout)
        case "fish":
            err = rootCmd.GenFishCompletion(os.Stdout, true)
        case "powershell":
            err = rootCmd.GenPowerShellCompletionWithDesc(os.Stdout)
        default:
            err = fmt.Errorf("不支持的 shell: %s", shell)
        }

        return err
    },
}
```

## 3.9 运行项目

### 3.9.1 构建和运行

```bash
# 构建
go build -o usermgr

# 运行
./usermgr --help

# 创建用户
./usermgr user create -n john -e john@example.com

# 列出用户
./usermgr user list

# 获取用户
./usermgr user get 1

# 更新用户
./usermgr user update 1 -e newemail@example.com

# 删除用户
./usermgr user delete 1 --force
```

### 3.9.2 帮助信息

```
$ ./usermgr --help
usermgr 是一个用户管理命令行工具。

支持的功能:
  - 用户 CRUD 操作
  - 配置管理
  - 自动补全

更多信息请访问: https://github.com/example/usermgr

Usage:
  usermgr [command]

Available Commands:
  completion  生成自动补全脚本
  config      配置管理
  help        Help about any command
  user        用户管理
  version     显示版本

Flags:
  -c, --config string   配置文件路径
  -h, --help           help for usermgr
  -v, --verbose         详细输出
      --version         version for usermgr

Use "usermgr [command] --help" for more information about a command.
```

## 3.10 本章小结

```
┌─────────────────────────────────────────────────────────────┐
│                      本章总结                                │
├─────────────────────────────────────────────────────────────────┤
│                                                             │
│  ✓ 完整的项目结构                                           │
│                                                             │
│  ✓ 用户 CRUD 完整实现                                       │
│                                                             │
│  ✓ 配置管理                                                │
│                                                             │
│  ✓ 自动补全                                                │
│                                                             │
│  ✓ 最佳实践                                                │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

---

## 附录：完整命令列表

```bash
# 用户管理
usermgr user create -n <name> -e <email> [-p <phone>]
usermgr user list [--format table|json]
usermgr user get <id>
usermgr user update <id> [-e <email>] [-p <phone>] [-s <status>]
usermgr user delete <id> [--force]

# 配置管理
usermgr config init
usermgr config show
usermgr config edit

# 工具
usermgr version
usermgr completion bash|zsh|fish|powershell

# 全局选项
usermgr --config <path> -c <path>
usermgr --verbose -v
usermgr --help -h
```
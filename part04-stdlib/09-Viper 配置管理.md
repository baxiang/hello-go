# Viper 配置管理

## 简介

Viper 是 Go 的完整配置解决方案，支持：
- JSON/YAML/TOML/HCL/Properties 等多种格式
- 环境变量
- 命令行标志
- 远程配置系统 (etcd/Consul)
- 配置热重载

## 快速开始

### 安装

```bash
go get github.com/spf13/viper
```

### 基础用法

```go
package main

import (
    "fmt"
    "github.com/spf13/viper"
)

func main() {
    // 设置配置文件名 (不含扩展名)
    viper.SetConfigName("config")

    // 设置配置文件类型
    viper.SetConfigType("yaml")

    // 添加搜索路径
    viper.AddConfigPath(".")
    viper.AddConfigPath("./configs")
    viper.AddConfigPath("/etc/myapp")

    // 读取配置
    if err := viper.ReadInConfig(); err != nil {
        panic(fmt.Errorf("读取配置失败：%w", err))
    }

    // 读取配置值
    port := viper.GetInt("server.port")
    host := viper.GetString("server.host")
    debug := viper.GetBool("debug")

    fmt.Printf("Server: %s:%d\n", host, port)
    fmt.Printf("Debug: %v\n", debug)
}
```

## 配置文件示例

### YAML 配置

```yaml
# config.yaml
app:
  name: myapp
  version: 1.0.0
  env: development

server:
  host: localhost
  port: 8080
  read_timeout: 30s
  write_timeout: 30s

database:
  driver: mysql
  host: localhost
  port: 3306
  username: root
  password: secret
  database: myapp
  max_open_conns: 100
  max_idle_conns: 10

redis:
  host: localhost
  port: 6379
  password: ""
  db: 0
  pool_size: 10

log:
  level: info
  format: json
  output: stdout

# 嵌套配置
features:
  - name: feature1
    enabled: true
  - name: feature2
    enabled: false

# 多环境配置
environments:
  development:
    debug: true
  production:
    debug: false
```

### JSON 配置

```json
{
  "app": {
    "name": "myapp",
    "version": "1.0.0"
  },
  "server": {
    "host": "localhost",
    "port": 8080
  },
  "database": {
    "driver": "mysql",
    "host": "localhost",
    "port": 3306
  }
}
```

### TOML 配置

```toml
# config.toml
[app]
name = "myapp"
version = "1.0.0"

[server]
host = "localhost"
port = 8080

[database]
driver = "mysql"
host = "localhost"
port = 3306
```

## 配置结构体

### 定义配置结构

```go
package config

import (
    "time"
)

// Config 应用配置
type Config struct {
    App      AppConfig      `mapstructure:"app"`
    Server   ServerConfig   `mapstructure:"server"`
    Database DatabaseConfig `mapstructure:"database"`
    Redis    RedisConfig    `mapstructure:"redis"`
    Log      LogConfig      `mapstructure:"log"`
}

type AppConfig struct {
    Name    string `mapstructure:"name"`
    Version string `mapstructure:"version"`
    Env     string `mapstructure:"env"`
}

type ServerConfig struct {
    Host         string        `mapstructure:"host"`
    Port         int           `mapstructure:"port"`
    ReadTimeout  time.Duration `mapstructure:"read_timeout"`
    WriteTimeout time.Duration `mapstructure:"write_timeout"`
}

type DatabaseConfig struct {
    Driver        string `mapstructure:"driver"`
    Host          string `mapstructure:"host"`
    Port          int    `mapstructure:"port"`
    Username      string `mapstructure:"username"`
    Password      string `mapstructure:"password"`
    Database      string `mapstructure:"database"`
    MaxOpenConns  int    `mapstructure:"max_open_conns"`
    MaxIdleConns  int    `mapstructure:"max_idle_conns"`
    ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
}

type RedisConfig struct {
    Host     string `mapstructure:"host"`
    Port     int    `mapstructure:"port"`
    Password string `mapstructure:"password"`
    DB       int    `mapstructure:"db"`
    PoolSize int    `mapstructure:"pool_size"`
}

type LogConfig struct {
    Level  string `mapstructure:"level"`
    Format string `mapstructure:"format"`
    Output string `mapstructure:"output"`
}
```

### 读取到结构体

```go
package main

import (
    "fmt"
    "github.com/spf13/viper"
)

func LoadConfig(path string) (*Config, error) {
    viper.SetConfigFile(path)

    // 读取环境变量 (自动转换)
    // 例如：APP_NAME 会映射到 app.name
    viper.SetEnvPrefix("APP")
    viper.AutomaticEnv()

    if err := viper.ReadInConfig(); err != nil {
        return nil, err
    }

    var cfg Config
    if err := viper.Unmarshal(&cfg); err != nil {
        return nil, err
    }

    return &cfg, nil
}

func main() {
    cfg, err := LoadConfig("config.yaml")
    if err != nil {
        panic(err)
    }

    fmt.Printf("App: %s v%s\n", cfg.App.Name, cfg.App.Version)
    fmt.Printf("Server: %s:%d\n", cfg.Server.Host, cfg.Server.Port)
}
```

## 环境变量支持

```go
package main

import (
    "fmt"
    "strings"
    "github.com/spf13/viper"
)

func main() {
    viper.SetConfigName("config")
    viper.AddConfigPath(".")

    // 环境变量配置
    viper.SetEnvPrefix("MYAPP")  // 前缀
    viper.AutomaticEnv()          // 自动绑定

    // 手动绑定环境变量
    viper.BindEnv("server.port", "MYAPP_SERVER_PORT")
    viper.BindEnv("database.host", "MYAPP_DB_HOST")

    // 自定义键名转换 (将 server.port 转为 SERVER_PORT)
    viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

    // 设置默认值
    viper.SetDefault("server.port", 8080)
    viper.SetDefault("server.host", "localhost")
    viper.SetDefault("debug", false)

    viper.ReadInConfig()

    // 优先级：环境变量 > 配置文件 > 默认值
    fmt.Println("Port:", viper.GetInt("server.port"))
}
```

```bash
# 使用环境变量覆盖配置
export MYAPP_SERVER_PORT=9090
export MYAPP_DB_HOST=remote-db.example.com
./myapp
```

## 命令行标志

```go
package main

import (
    "github.com/spf13/viper"
    "github.com/spf13/pflag"
)

func main() {
    // 创建 flagset
    flags := pflag.NewFlagSet("myapp", pflag.ExitOnError)

    // 定义命令行标志
    flags.String("config", "config.yaml", "配置文件路径")
    flags.String("server.port", "", "服务器端口")
    flags.String("server.host", "", "服务器地址")
    flags.Bool("debug", false, "调试模式")

    // 绑定到 viper
    viper.BindPFlags(flags)

    // 绑定环境变量
    viper.BindEnv("server.port", "MYAPP_PORT")

    // 读取配置
    viper.SetConfigFile(viper.GetString("config"))
    viper.ReadInConfig()

    // 使用配置
    port := viper.GetInt("server.port")
    host := viper.GetString("server.host")
}
```

## 远程配置

### 从 etcd 读取

```go
package main

import (
    "github.com/spf13/viper"
)

func main() {
    // 添加 etcd 远程配置
    viper.AddRemoteProvider("etcd", "http://localhost:2379", "/config/myapp")
    viper.SetConfigType("yaml")

    if err := viper.ReadRemoteConfig(); err != nil {
        panic(err)
    }

    // 监听配置变化
    go func() {
        for {
            viper.WatchRemoteConfig()
            // 配置已变更，重新加载
            reloadConfig()
        }
    }()
}
```

### 从 Consul 读取

```go
viper.AddRemoteProvider("consul", "http://localhost:8500", "config/myapp")
viper.SetConfigType("json")

if err := viper.ReadRemoteConfig(); err != nil {
    panic(err)
}
```

## 配置热重载

```go
package main

import (
    "fmt"
    "log"
    "sync"
    "github.com/spf13/viper"
)

type ConfigManager struct {
    mu     sync.RWMutex
    config *Config
    hooks  []func(*Config)
}

func NewConfigManager(path string) (*ConfigManager, error) {
    cm := &ConfigManager{}

    viper.SetConfigFile(path)
    viper.WatchConfig()  // 监听文件变化

    // 配置变化回调
    viper.OnConfigChange(func(e viper.ConfigEvent) {
        log.Println("配置文件变更:", e.Key, "->", e.NewValue)
        cm.reload()
    })

    if err := viper.ReadInConfig(); err != nil {
        return nil, err
    }

    if err := cm.reload(); err != nil {
        return nil, err
    }

    return cm, nil
}

func (cm *ConfigManager) reload() error {
    cm.mu.Lock()
    defer cm.mu.Unlock()

    var cfg Config
    if err := viper.Unmarshal(&cfg); err != nil {
        return err
    }
    cm.config = &cfg

    // 通知回调
    for _, hook := range cm.hooks {
        hook(cm.config)
    }

    return nil
}

func (cm *ConfigManager) Get() *Config {
    cm.mu.RLock()
    defer cm.mu.RUnlock()
    return cm.config
}

func (cm *ConfigManager) RegisterHook(hook func(*Config)) {
    cm.hooks = append(cm.hooks, hook)
}

// 使用示例
func main() {
    cm, err := NewConfigManager("config.yaml")
    if err != nil {
        panic(err)
    }

    // 注册配置变更回调
    cm.RegisterHook(func(cfg *Config) {
        log.Println("日志级别变更为:", cfg.Log.Level)
        // 重新配置 logger...
    })

    // 获取当前配置
    cfg := cm.Get()
    fmt.Println("Port:", cfg.Server.Port)
}
```

## 多环境配置

```go
package main

import (
    "fmt"
    "github.com/spf13/viper"
)

func LoadConfigByEnv(env string) (*Config, error) {
    // 根据环境加载不同配置
    viper.SetConfigFile(fmt.Sprintf("config.%s.yaml", env))

    // 通用配置
    viper.SetConfigFile("config.base.yaml")
    viper.MergeInConfig()

    if err := viper.ReadInConfig(); err != nil {
        return nil, err
    }

    var cfg Config
    viper.Unmarshal(&cfg)

    return &cfg, nil
}

// 或者使用子树
func loadConfigWithSubtree() {
    viper.SetConfigFile("config.yaml")
    viper.ReadInConfig()

    // 获取当前环境
    env := viper.GetString("app.env")

    // 从 environments.development 或 environments.production 读取
    dbConfig := viper.Sub(fmt.Sprintf("environments.%s", env))

    fmt.Println("DB Host:", dbConfig.GetString("database.host"))
}
```

## 配置验证

```go
package config

import (
    "errors"
    "github.com/spf13/viper"
)

var (
    ErrMissingPort     = errors.New("server.port is required")
    ErrInvalidLogLevel = errors.New("invalid log level")
)

func Validate() error {
    // 必需配置检查
    required := []string{
        "server.host",
        "server.port",
        "database.driver",
        "database.host",
    }

    for _, key := range required {
        if !viper.IsSet(key) {
            return fmt.Errorf("%s is required", key)
        }
    }

    // 值验证
    port := viper.GetInt("server.port")
    if port <= 0 || port > 65535 {
        return ErrMissingPort
    }

    level := viper.GetString("log.level")
    validLevels := []string{"debug", "info", "warn", "error"}
    if !contains(validLevels, level) {
        return ErrInvalidLogLevel
    }

    return nil
}

func contains(slice []string, item string) bool {
    for _, s := range slice {
        if s == item {
            return true
        }
    }
    return false
}
```

## 完整项目实践

```go
// config/config.go
package config

import (
    "fmt"
    "sync"
    "github.com/spf13/viper"
)

var (
    once     sync.Once
    instance *ConfigManager
)

type ConfigManager struct {
    viper  *viper.Viper
    config *Config
    mu     sync.RWMutex
}

func Init(configPath string) error {
    var err error
    once.Do(func() {
        instance, err = NewConfigManager(configPath)
    })
    return err
}

func NewConfigManager(path string) (*ConfigManager, error) {
    v := viper.New()
    v.SetConfigFile(path)
    v.SetEnvPrefix("APP")
    v.AutomaticEnv()

    if err := v.ReadInConfig(); err != nil {
        return nil, err
    }

    cm := &ConfigManager{viper: v}
    if err := cm.reload(); err != nil {
        return nil, err
    }

    v.WatchConfig()
    v.OnConfigChange(func(e viper.ConfigEvent) {
        cm.reload()
    })

    return cm, nil
}

func (cm *ConfigManager) reload() error {
    cm.mu.Lock()
    defer cm.mu.Unlock()

    var cfg Config
    if err := cm.viper.Unmarshal(&cfg); err != nil {
        return err
    }
    cm.config = &cfg
    return nil
}

func Get() *Config {
    if instance == nil {
        return nil
    }
    instance.mu.RLock()
    defer instance.mu.RUnlock()
    return instance.config
}

// 便捷函数
func GetServerPort() int {
    return Get().Server.Port
}

func GetDSN() string {
    db := Get().Database
    return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
        db.Username, db.Password, db.Host, db.Port, db.Database)
}
```

## Viper 检查清单

```
[ ] 使用结构体定义配置
[ ] 设置合理的默认值
[ ] 支持环境变量覆盖
[ ] 配置文件与代码分离
[ ] 敏感信息使用环境变量
[ ] 实现配置验证
[ ] 支持配置热重载
[ ] 记录配置变更日志
[ ] 多环境配置分离
```

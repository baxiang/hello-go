package config

import (
	"os"

	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/config/file"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/spf13/viper"
)

// Config 应用配置
type Config struct {
	Server   ServerConfig   `mapstructure:"server" json:"server"`
	Database DatabaseConfig `mapstructure:"database" json:"database"`
	NATS     NATSConfig     `mapstructure:"nats" json:"nats"`
	Redis    RedisConfig    `mapstructure:"redis" json:"redis"`
	JWT      JWTConfig      `mapstructure:"jwt" json:"jwt"`
	Logger   LoggerConfig   `mapstructure:"logger" json:"logger"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	HTTP HTTPConfig `mapstructure:"http" json:"http"`
	GRPC GRPCConfig `mapstructure:"grpc" json:"grpc"`
}

// HTTPConfig HTTP 配置
type HTTPConfig struct {
	Network  string `mapstructure:"network" json:"network"`
	Addr    string `mapstructure:"addr" json:"addr"`
	Timeout int64  `mapstructure:"timeout" json:"timeout"`
}

// GRPCConfig gRPC 配置
type GRPCConfig struct {
	Network  string `mapstructure:"network" json:"network"`
	Addr    string `mapstructure:"addr" json:"addr"`
	Timeout int64  `mapstructure:"timeout" json:"timeout"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Driver   string `mapstructure:"driver" json:"driver"`
	Source   string `mapstructure:"source" json:"source"`
	MaxOpen  int    `mapstructure:"max_open" json:"max_open"`
	MaxIdle  int    `mapstructure:"max_idle" json:"max_idle"`
	MaxLife  int    `mapstructure:"max_life" json:"max_life"`
}

// NATSConfig NATS 配置
type NATSConfig struct {
	URL        string `mapstructure:"url" json:"url"`
	ClusterID  string `mapstructure:"cluster_id" json:"cluster_id"`
	ClientID   string `mapstructure:"client_id" json:"client_id"`
	StreamName string `mapstructure:"stream_name" json:"stream_name"`
}

// RedisConfig Redis 配置
type RedisConfig struct {
	Addr     string `mapstructure:"addr" json:"addr"`
	Password string `mapstructure:"password" json:"password"`
	DB       int    `mapstructure:"db" json:"db"`
}

// JWTConfig JWT 配置
type JWTConfig struct {
	Secret     string `mapstructure:"secret" json:"secret"`
	ExpireHour int    `mapstructure:"expire_hour" json:"expire_hour"`
}

// LoggerConfig 日志配置
type LoggerConfig struct {
	Level      string `mapstructure:"level" json:"level"`
	Format     string `mapstructure:"format" json:"format"`
	OutputPath string `mapstructure:"output_path" json:"output_path"`
}

// Load 加载配置
func Load(path string) (*Config, error) {
	v := viper.New()
	v.SetConfigFile(path)
	v.SetConfigType("yaml")

	// 设置默认值
	setDefaults(v)

	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// LoadConfig 使用 Kratos 配置加载
func LoadConfig(path string) (*Config, error) {
	c := config.New(
		config.WithSource(
			file.NewSource(path),
		),
	)
	defer c.Close()

	if err := c.Load(); err != nil {
		return nil, err
	}

	var cfg Config
	if err := c.Scan(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("server.http.network", "tcp")
	v.SetDefault("server.http.addr", ":8080")
	v.SetDefault("server.http.timeout", 60)

	v.SetDefault("server.grpc.network", "tcp")
	v.SetDefault("server.grpc.addr", ":50051")
	v.SetDefault("server.grpc.timeout", 60)

	v.SetDefault("database.driver", "mysql")
	v.SetDefault("database.max_open", 25)
	v.SetDefault("database.max_idle", 5)
	v.SetDefault("database.max_life", 300)

	v.SetDefault("nats.url", "nats://localhost:4222")
	v.SetDefault("nats.stream_name", "ECOMMERCE")

	v.SetDefault("redis.addr", "localhost:6379")
	v.SetDefault("redis.db", 0)

	v.SetDefault("jwt.secret", "kratos-secret-key")
	v.SetDefault("jwt.expire_hour", 24)

	v.SetDefault("logger.level", "info")
	v.SetDefault("logger.format", "json")
	v.SetDefault("logger.output_path", "logs/app.log")
}

// GetEnv 获取环境变量
func GetEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// MustLoad 强制加载配置
func MustLoad(path string) *Config {
	cfg, err := Load(path)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}
	return cfg
}
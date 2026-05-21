// Package config 提供配置加载功能
package config

import (
	"fmt"

	"github.com/spf13/viper"
)

// ServerConfig 服务器配置
type ServerConfig struct {
	HTTP HTTPConfig `mapstructure:"http"`
	GRPC GRPCConfig `mapstructure:"grpc"`
}

// HTTPConfig HTTP 服务配置
type HTTPConfig struct {
	Network string `mapstructure:"network"`
	Addr    string `mapstructure:"addr"`
	Timeout int64  `mapstructure:"timeout"`
}

// GRPCConfig gRPC 服务配置
type GRPCConfig struct {
	Network string `mapstructure:"network"`
	Addr    string `mapstructure:"addr"`
	Timeout int64  `mapstructure:"timeout"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Driver  string `mapstructure:"driver"`
	Source  string `mapstructure:"source"`
	MaxOpen int    `mapstructure:"max_open"`
	MaxIdle int    `mapstructure:"max_idle"`
	MaxLife int    `mapstructure:"max_life"`
}

// NATSConfig NATS 配置
type NATSConfig struct {
	URL        string `mapstructure:"url"`
	ClientID   string `mapstructure:"client_id"`
	StreamName string `mapstructure:"stream_name"`
}

// RedisConfig Redis 配置
type RedisConfig struct {
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

// JWTConfig JWT 配置
type JWTConfig struct {
	Secret     string `mapstructure:"secret"`
	ExpireHour int    `mapstructure:"expire_hour"`
}

// LoggerConfig 日志配置
type LoggerConfig struct {
	Level      string `mapstructure:"level"`
	Format     string `mapstructure:"format"`
	OutputPath string `mapstructure:"output_path"`
}

// ServiceEndpoint 服务端点配置
type ServiceEndpoint struct {
	Addr string `mapstructure:"addr"`
}

// ServicesConfig 远程服务依赖配置（用于 order/payment/gateway）
type ServicesConfig struct {
	User    ServiceEndpoint `mapstructure:"user"`
	Product ServiceEndpoint `mapstructure:"product"`
	Order   ServiceEndpoint `mapstructure:"order"`
	Payment ServiceEndpoint `mapstructure:"payment"`
}

// AppConfig 通用应用配置（每个服务可嵌入扩展）
type AppConfig struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	NATS     NATSConfig     `mapstructure:"nats"`
	Redis    RedisConfig    `mapstructure:"redis"`
	JWT      JWTConfig      `mapstructure:"jwt"`
	Logger   LoggerConfig   `mapstructure:"logger"`
	Services ServicesConfig `mapstructure:"services"`
}

// Load 从指定路径加载配置到目标结构体
func Load(path string, cfg interface{}) error {
	v := viper.New()
	v.SetConfigFile(path)
	v.SetConfigType("yaml")

	if err := v.ReadInConfig(); err != nil {
		return fmt.Errorf("读取配置文件失败: %w", err)
	}

	if err := v.Unmarshal(cfg); err != nil {
		return fmt.Errorf("解析配置失败: %w", err)
	}
	return nil
}

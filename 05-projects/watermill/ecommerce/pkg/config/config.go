// Package config 提供 YAML 配置加载
package config

import (
	"github.com/spf13/viper"
)

// Config 应用总配置
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Kafka    KafkaConfig    `mapstructure:"kafka"`
	Database DatabaseConfig `mapstructure:"database"`
}

// ServerConfig 服务配置
type ServerConfig struct {
	HTTPPort int `mapstructure:"http_port"`
}

// KafkaConfig Kafka 配置
type KafkaConfig struct {
	Brokers       []string `mapstructure:"brokers"`
	ConsumerGroup string   `mapstructure:"consumer_group"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	DSN string `mapstructure:"dsn"`
}

// Load 从 YAML 文件加载配置
func Load(path string, cfg *Config) error {
	v := viper.New()
	v.SetConfigFile(path)
	v.SetConfigType("yaml")
	if err := v.ReadInConfig(); err != nil {
		return err
	}
	return v.Unmarshal(cfg)
}

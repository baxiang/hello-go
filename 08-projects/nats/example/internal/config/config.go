// Package config 配置定义
package config

import "time"

// NATS 服务器地址
const (
	DefaultNATSURL = "nats://localhost:4222"
)

// DeviceConfig 设备配置
type DeviceConfig struct {
	ID               string        // 设备 ID
	AccountID        string        // 账号 ID
	NodeType         string        // 节点类型
	NATSURL          string        // NATS 服务器地址
	HeartbeatInterval time.Duration // 心跳间隔
}

// HubConfig Hub 配置
type HubConfig struct {
	NATSURL          string        // NATS 服务器地址
	ExecuteTimeout   time.Duration // 执行超时
	HeartbeatTimeout time.Duration // 心跳超时
}

// ClientConfig 客户端配置
type ClientConfig struct {
	HubURL string // Hub 服务地址
}

// DefaultDeviceConfig 默认设备配置
func DefaultDeviceConfig() *DeviceConfig {
	return &DeviceConfig{
		NATSURL:          DefaultNATSURL,
		HeartbeatInterval: 30 * time.Second,
	}
}

// DefaultHubConfig 默认 Hub 配置
func DefaultHubConfig() *HubConfig {
	return &HubConfig{
		NATSURL:          DefaultNATSURL,
		ExecuteTimeout:   60 * time.Second,
		HeartbeatTimeout: 90 * time.Second,
	}
}

// DefaultClientConfig 默认客户端配置
func DefaultClientConfig() *ClientConfig {
	return &ClientConfig{
		HubURL: "http://localhost:8080",
	}
}
// Package natsclient 提供 NATS JetStream 客户端封装
package natsclient

import (
	"context"
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"go.uber.org/zap"

	"services/pkg/config"
)

// Client NATS 客户端
type Client struct {
	nc  *nats.Conn
	js  jetstream.JetStream
	log *zap.Logger
}

// New 创建 NATS 客户端
func New(cfg *config.NATSConfig, log *zap.Logger) (*Client, error) {
	nc, err := nats.Connect(cfg.URL,
		nats.Name(cfg.ClientID),
		nats.MaxReconnects(5),
		nats.ReconnectWait(time.Second),
		nats.PingInterval(20*time.Second),
		nats.Timeout(10*time.Second),
	)
	if err != nil {
		return nil, fmt.Errorf("连接 NATS 失败: %w", err)
	}

	js, err := jetstream.New(nc)
	if err != nil {
		nc.Close()
		return nil, fmt.Errorf("创建 JetStream 失败: %w", err)
	}

	log.Info("NATS 连接成功", zap.String("url", cfg.URL))
	return &Client{nc: nc, js: js, log: log}, nil
}

// Publish 发布消息到 JetStream
func (c *Client) Publish(ctx context.Context, subject string, data []byte) error {
	_, err := c.js.Publish(ctx, subject, data)
	if err != nil {
		return fmt.Errorf("发布消息失败: %w", err)
	}
	return nil
}

// Subscribe 订阅消息（核心 NATS）
func (c *Client) Subscribe(subject string, handler func(*nats.Msg)) (*nats.Subscription, error) {
	return c.nc.Subscribe(subject, handler)
}

// EnsureStream 确保 Stream 存在
func (c *Client) EnsureStream(ctx context.Context, name string, subjects []string) error {
	_, err := c.js.Stream(ctx, name)
	if err == nil {
		return nil
	}
	_, err = c.js.CreateStream(ctx, jetstream.StreamConfig{
		Name:     name,
		Subjects: subjects,
		Storage:  jetstream.FileStorage,
		MaxBytes: 1024 * 1024 * 1024,
		MaxAge:   24 * time.Hour * 7,
	})
	if err != nil {
		return fmt.Errorf("创建 Stream 失败: %w", err)
	}
	c.log.Info("Stream 创建成功", zap.String("name", name))
	return nil
}

// Close 关闭连接
func (c *Client) Close() {
	if c.nc != nil {
		c.nc.Close()
	}
}

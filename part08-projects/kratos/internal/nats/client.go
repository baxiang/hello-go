package nats

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"

	"go.uber.org/zap"
)

// Client NATS 客户端封装
type Client struct {
	nc    *nats.Conn
	js    jetstream.JetStream
	mu    sync.RWMutex
	log   *zap.Logger
}

// NewClient 创建 NATS 客户端
func NewClient(url string, logger *zap.Logger) (*Client, error) {
	nc, err := nats.Connect(url,
		nats.Name("kratos-ecommerce"),
		nats.MaxReconnects(5),
		nats.ReconnectWait(time.Second),
		nats.PingInterval(20*time.Second),
		nats.DialTimeout(10*time.Second),
	)
	if err != nil {
		return nil, fmt.Errorf("连接 NATS 失败: %w", err)
	}

	js, err := jetstream.New(nc)
	if err != nil {
		nc.Close()
		return nil, fmt.Errorf("创建 JetStream 失败: %w", err)
	}

	return &Client{
		nc:  nc,
		js:  js,
		log: logger,
	}, nil
}

// GetConn 获取连接
func (c *Client) GetConn() *nats.Conn {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.nc
}

// GetJS 获取 JetStream
func (c *Client) GetJS() jetstream.JetStream {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.js
}

// Close 关闭连接
func (c *Client) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.nc != nil {
		c.nc.Close()
	}
}

// IsConnected 检查连接状态
func (c *Client) IsConnected() bool {
	return c.nc != nil && c.nc.IsConnected()
}

// CreateStream 创建 Stream
func (c *Client) CreateStream(ctx context.Context, name string, subjects []string) error {
	_, err := c.js.Stream(ctx, name)
	if err == nil {
		c.log.Info("Stream 已存在", zap.String("name", name))
		return nil
	}

	_, err = c.js.CreateStream(ctx, jetstream.StreamConfig{
		Name:      name,
		Subjects:  subjects,
		Storage:   jetstream.FileStorage,
		Retention: jetstream.LimitsPolicy{
			MaxBytes: 1024 * 1024 * 1024, // 1GB
			MaxAge:   time.Hour * 24 * 7, // 7 天
		},
		Discard:  jetstream.DiscardOld,
		Replicas: 1,
	})
	if err != nil {
		return fmt.Errorf("创建 Stream 失败: %w", err)
	}

	c.log.Info("创建 Stream 成功", zap.String("name", name))
	return nil
}

// Publish 发布消息
func (c *Client) Publish(ctx context.Context, subject string, data []byte) error {
	_, err := c.js.Publish(ctx, subject, data)
	if err != nil {
		return fmt.Errorf("发布消息失败: %w", err)
	}
	return nil
}

// PublishWithOptions 带选项发布
func (c *Client) PublishWithOptions(ctx context.Context, subject string, data []byte, opts ...jetstream.PublishOption) error {
	_, err := c.js.Publish(ctx, subject, data, opts...)
	if err != nil {
		return fmt.Errorf("发布消息失败: %w", err)
	}
	return nil
}

// Subscribe 订阅消息
func (c *Client) Subscribe(subject string, handler func(msg *nats.Msg)) (*nats.Subscription, error) {
	return c.nc.Subscribe(subject, handler)
}

// QueueSubscribe 队列订阅
func (c *Client) QueueSubscribe(subject, queue string, handler func(msg *nats.Msg)) (*nats.Subscription, error) {
	return c.nc.QueueSubscribe(subject, queue, handler)
}

// CreateConsumer 创建消费者
func (c *Client) CreateConsumer(ctx context.Context, stream, name, subject string) (jetstream.Consumer, error) {
	return c.js.CreateConsumer(ctx, stream, jetstream.ConsumerConfig{
		Name:          name,
		Durable:       true,
		FilterSubject: subject,
		DeliverPolicy: jetstream.DeliverAll,
		AckPolicy:     jetstream.AckExplicit,
		AckWait:       30 * time.Second,
		MaxDeliver:    3,
		MaxAckPending: 100,
	})
}

// Consume 消费消息
func (c *Client) Consume(stream, consumer string, opts ...jetstream.ConsumeOption) (jetstream.ConsumeContext, error) {
	return c.js.Consume(stream, consumer, opts...)
}

// GetServerInfo 获取服务器信息
func (c *Client) GetServerInfo() *nats.ServerInfo {
	return c.nc.ServerInfo()
}
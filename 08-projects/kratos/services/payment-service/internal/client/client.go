// Package client 提供 payment-service 对其他服务的 gRPC 客户端
package client

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	orderV1 "services/api/order/v1"
)

// OrderClient 订单服务客户端
type OrderClient struct {
	orderV1.OrderServiceClient
	conn *grpc.ClientConn
}

// NewOrderClient 创建订单服务客户端
func NewOrderClient(ctx context.Context, addr string, log *zap.Logger) (*OrderClient, error) {
	conn, err := grpc.DialContext(ctx, addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, fmt.Errorf("连接 order-service(%s) 失败: %w", addr, err)
	}
	log.Info("已连接 order-service", zap.String("addr", addr))
	return &OrderClient{
		OrderServiceClient: orderV1.NewOrderServiceClient(conn),
		conn:               conn,
	}, nil
}

// Close 关闭连接
func (c *OrderClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// Package client 提供对外部服务的 gRPC 客户端封装
package client

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	productV1 "services/api/product/v1"
	userV1 "services/api/user/v1"
)

// UserClient 用户服务客户端
type UserClient struct {
	userV1.UserServiceClient
	conn *grpc.ClientConn
}

// NewUserClient 创建用户服务客户端
func NewUserClient(ctx context.Context, addr string, log *zap.Logger) (*UserClient, error) {
	conn, err := grpc.DialContext(ctx, addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, fmt.Errorf("连接 user-service(%s) 失败: %w", addr, err)
	}
	log.Info("已连接 user-service", zap.String("addr", addr))
	return &UserClient{
		UserServiceClient: userV1.NewUserServiceClient(conn),
		conn:              conn,
	}, nil
}

// Close 关闭连接
func (c *UserClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// ProductClient 商品服务客户端
type ProductClient struct {
	productV1.ProductServiceClient
	conn *grpc.ClientConn
}

// NewProductClient 创建商品服务客户端
func NewProductClient(ctx context.Context, addr string, log *zap.Logger) (*ProductClient, error) {
	conn, err := grpc.DialContext(ctx, addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, fmt.Errorf("连接 product-service(%s) 失败: %w", addr, err)
	}
	log.Info("已连接 product-service", zap.String("addr", addr))
	return &ProductClient{
		ProductServiceClient: productV1.NewProductServiceClient(conn),
		conn:                 conn,
	}, nil
}

// Close 关闭连接
func (c *ProductClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

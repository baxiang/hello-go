// Package client 提供 api-gateway 调用后端服务的 gRPC 客户端
package client

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	orderV1 "services/api/order/v1"
	paymentV1 "services/api/payment/v1"
	productV1 "services/api/product/v1"
	userV1 "services/api/user/v1"
)

// Clients 聚合所有后端 gRPC 客户端
type Clients struct {
	User    userV1.UserServiceClient
	Product productV1.ProductServiceClient
	Order   orderV1.OrderServiceClient
	Payment paymentV1.PaymentServiceClient

	conns []*grpc.ClientConn
}

// New 创建并连接所有后端服务客户端
func New(ctx context.Context, userAddr, productAddr, orderAddr, paymentAddr string, log *zap.Logger) (*Clients, error) {
	dial := func(name, addr string) (*grpc.ClientConn, error) {
		conn, err := grpc.DialContext(ctx, addr,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithBlock(),
		)
		if err != nil {
			return nil, fmt.Errorf("连接 %s(%s) 失败: %w", name, addr, err)
		}
		log.Info("已连接后端服务", zap.String("service", name), zap.String("addr", addr))
		return conn, nil
	}

	userConn, err := dial("user-service", userAddr)
	if err != nil {
		return nil, err
	}
	productConn, err := dial("product-service", productAddr)
	if err != nil {
		userConn.Close()
		return nil, err
	}
	orderConn, err := dial("order-service", orderAddr)
	if err != nil {
		userConn.Close()
		productConn.Close()
		return nil, err
	}
	paymentConn, err := dial("payment-service", paymentAddr)
	if err != nil {
		userConn.Close()
		productConn.Close()
		orderConn.Close()
		return nil, err
	}

	return &Clients{
		User:    userV1.NewUserServiceClient(userConn),
		Product: productV1.NewProductServiceClient(productConn),
		Order:   orderV1.NewOrderServiceClient(orderConn),
		Payment: paymentV1.NewPaymentServiceClient(paymentConn),
		conns:   []*grpc.ClientConn{userConn, productConn, orderConn, paymentConn},
	}, nil
}

// Close 关闭所有连接
func (c *Clients) Close() {
	for _, conn := range c.conns {
		_ = conn.Close()
	}
}

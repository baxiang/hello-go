// Package server 提供 gRPC 服务启动封装
package server

import (
	"context"
	"errors"
	"fmt"
	"net"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"

	"services/pkg/config"
)

// GRPCServer gRPC 服务包装
type GRPCServer struct {
	server *grpc.Server
	addr   string
	log    *zap.Logger
}

// NewGRPCServer 创建 gRPC 服务（外部需通过 RawServer() 注册服务）
func NewGRPCServer(cfg *config.GRPCConfig, log *zap.Logger, opts ...grpc.ServerOption) *GRPCServer {
	defaultOpts := []grpc.ServerOption{
		grpc.UnaryInterceptor(loggingUnaryInterceptor(log)),
	}
	defaultOpts = append(defaultOpts, opts...)

	s := grpc.NewServer(defaultOpts...)
	return &GRPCServer{
		server: s,
		addr:   cfg.Addr,
		log:    log,
	}
}

// RawServer 返回原始 *grpc.Server，外部用此注册服务
func (s *GRPCServer) RawServer() *grpc.Server {
	return s.server
}

// Start 启动 gRPC 服务（阻塞）
func (s *GRPCServer) Start() error {
	lis, err := net.Listen("tcp", s.addr)
	if err != nil {
		return fmt.Errorf("监听 %s 失败: %w", s.addr, err)
	}
	s.log.Info("gRPC 服务启动", zap.String("addr", s.addr))
	if err := s.server.Serve(lis); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
		return fmt.Errorf("gRPC 服务异常退出: %w", err)
	}
	return nil
}

// Stop 优雅关闭
func (s *GRPCServer) Stop(ctx context.Context) error {
	s.log.Info("gRPC 服务正在关闭...")
	done := make(chan struct{})
	go func() {
		s.server.GracefulStop()
		close(done)
	}()
	select {
	case <-done:
		return nil
	case <-ctx.Done():
		s.server.Stop()
		return ctx.Err()
	}
}

// loggingUnaryInterceptor 简单的日志拦截器
func loggingUnaryInterceptor(log *zap.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		start := time.Now()
		resp, err := handler(ctx, req)
		fields := []zap.Field{
			zap.String("method", info.FullMethod),
			zap.Duration("cost", time.Since(start)),
		}
		if err != nil {
			fields = append(fields, zap.Error(err))
			log.Warn("gRPC 调用失败", fields...)
		} else {
			log.Debug("gRPC 调用成功", fields...)
		}
		return resp, err
	}
}

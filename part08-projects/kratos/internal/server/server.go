package server

import (
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	"github.com/go-kratos/kratos/v2/middleware/validate"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/v2/transport/http"

	"kratos/internal/config"
)

// NewHTTPServer 创建 HTTP 服务器
func NewHTTPServer(cfg *config.Config, logger log.Logger) *http.Server {
	var opts = []http.ServerOption{
		http.Middleware(
			recovery.Recovery(),
			tracing.Server(),
			validate.Validator(),
		),
	}

	if cfg.Server.HTTP.Network != "" {
		opts = append(opts, http.Network(cfg.Server.HTTP.Network))
	}
	if cfg.Server.HTTP.Addr != "" {
		opts = append(opts, http.Address(cfg.Server.HTTP.Addr))
	}
	if cfg.Server.HTTP.Timeout > 0 {
		opts = append(opts, http.Timeout(cfg.Server.HTTP.Timeout))
	}

	return http.NewServer(opts...)
}

// NewGRPCServer 创建 gRPC 服务器
func NewGRPCServer(cfg *config.Config, logger log.Logger) *grpc.Server {
	var opts = []grpc.ServerOption{
		grpc.Middleware(
			recovery.Recovery(),
			tracing.Server(),
			validate.Validator(),
		),
	}

	if cfg.Server.GRPC.Network != "" {
		opts = append(opts, grpc.Network(cfg.Server.GRPC.Network))
	}
	if cfg.Server.GRPC.Addr != "" {
		opts = append(opts, grpc.Address(cfg.Server.GRPC.Addr))
	}
	if cfg.Server.GRPC.Timeout > 0 {
		opts = append(opts, grpc.Timeout(cfg.Server.GRPC.Timeout))
	}

	return grpc.NewServer(opts...)
}

// NewMiddlewares 创建中间件
func NewMiddlewares(logger log.Logger) []interface{} {
	return []interface{}{
		recovery.Recovery(),
		tracing.Server(),
		validate.Validator(),
		log.NewLogger(logger),
	}
}
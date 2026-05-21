package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"

	"services/pkg/config"
)

// HTTPServer HTTP 服务包装（适用于 api-gateway）
type HTTPServer struct {
	server *http.Server
	addr   string
	log    *zap.Logger
}

// NewHTTPServer 创建 HTTP 服务
func NewHTTPServer(cfg *config.HTTPConfig, handler http.Handler, log *zap.Logger) *HTTPServer {
	timeout := time.Duration(cfg.Timeout) * time.Second
	if timeout <= 0 {
		timeout = 60 * time.Second
	}
	return &HTTPServer{
		server: &http.Server{
			Addr:         cfg.Addr,
			Handler:      handler,
			ReadTimeout:  timeout,
			WriteTimeout: timeout,
		},
		addr: cfg.Addr,
		log:  log,
	}
}

// Start 启动 HTTP 服务（阻塞）
func (s *HTTPServer) Start() error {
	s.log.Info("HTTP 服务启动", zap.String("addr", s.addr))
	if err := s.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("HTTP 服务异常退出: %w", err)
	}
	return nil
}

// Stop 优雅关闭
func (s *HTTPServer) Stop(ctx context.Context) error {
	s.log.Info("HTTP 服务正在关闭...")
	return s.server.Shutdown(ctx)
}

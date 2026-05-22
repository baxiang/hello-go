package server

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"
)

// Runner 可启动 + 可停止的服务接口
type Runner interface {
	Start() error
	Stop(ctx context.Context) error
}

// RunAndWait 启动所有服务并等待退出信号
func RunAndWait(log *zap.Logger, runners ...Runner) {
	errCh := make(chan error, len(runners))
	for _, r := range runners {
		go func(r Runner) {
			if err := r.Start(); err != nil {
				errCh <- err
			}
		}(r)
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-sigCh:
		log.Info("收到退出信号", zap.String("signal", sig.String()))
	case err := <-errCh:
		log.Error("服务启动失败", zap.Error(err))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	for _, r := range runners {
		if err := r.Stop(ctx); err != nil {
			log.Warn("服务关闭异常", zap.Error(err))
		}
	}
	log.Info("所有服务已关闭")
}

// NewLogger 创建生产级 zap logger
func NewLogger(serviceName string) *zap.Logger {
	cfg := zap.NewProductionConfig()
	cfg.OutputPaths = []string{"stdout"}
	logger, _ := cfg.Build()
	return logger.With(zap.String("service", serviceName))
}

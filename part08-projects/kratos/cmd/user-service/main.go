package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/config/file"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/v2/transport/http"
	"go.uber.org/zap"

	"kratos/internal/biz"
	"kratos/internal/config"
	"kratos/internal/data"
	"kratos/internal/repo"
	"kratos/internal/server"
	"kratos/internal/service"
	userV1 "kratos/api/user/v1"
)

var (
	// Version 版本号
	Version = "v1.0.0"
	// BuildTime 构建时间
	BuildTime = "unknown"
)

func main() {
	// 初始化日志
	logger := initLogger()
	defer logger.Sync()

	log := log.NewHelper(log.With(logger, "module", "main"))

	// 加载配置
	cfg := loadConfig()

	// 创建数据层
	dataData, cleanup, err := newData(cfg, logger)
	if err != nil {
		log.Fatal("创建数据层失败", zap.Error(err))
	}
	defer cleanup()

	// 创建业务层
	userUC := biz.NewUserUseCase(repo.NewUserRepo(dataData, logger), logger)

	// 创建 HTTP 服务器
	httpServer := server.NewHTTPServer(cfg, logger)
	userService := service.NewUserService(userUC)
	userV1.RegisterUserServiceHTTPServer(httpServer, userService)

	// 创建 gRPC 服务器
	grpcServer := server.NewGRPCServer(cfg, logger)
	userV1.RegisterUserServiceServer(grpcServer, userService)

	// 创建应用
	app := kratos.New(
		kratos.Name("user-service"),
		kratos.Version(Version),
		kratos.Metadata(map[string]string{"build_time": BuildTime}),
		kratos.Logger(logger),
		kratos.Server(
			httpServer,
			grpcServer,
		),
	)

	// 启动应用
	if err := app.Run(); err != nil {
		log.Fatal("启动服务失败", zap.Error(err))
	}
}

// initLogger 初始化日志
func initLogger() *zap.Logger {
	zapConfig := zap.NewProductionConfig()
	zapConfig.OutputPaths = []string{"stdout"}
	zapConfig.EncoderConfig.TimeKey = "time"
	zapConfig.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	zapConfig.EncoderConfig.EncodeDuration = zapcore.SecondsDurationEncoder

	logger, _ := zapConfig.Build()
	return logger
}

// loadConfig 加载配置
func loadConfig() *config.Config {
	path := os.Getenv("CONFIG_PATH")
	if path == "" {
		path = "configs/user-service.yaml"
	}

	c := config.New(
		config.WithSource(
			file.NewSource(path),
		),
	)
	defer c.Close()

	if err := c.Load(); err != nil {
		panic(err)
	}

	var cfg config.Config
	if err := c.Scan(&cfg); err != nil {
		panic(err)
	}

	return &cfg
}

// newData 创建数据层
func newData(cfg *config.Config, logger *zap.Logger) (*data.Data, func(), error) {
	db, err := data.InitDB(cfg, logger)
	if err != nil {
		return nil, nil, err
	}

	natsClient, err := data.InitNATS(cfg, logger)
	if err != nil {
		return nil, nil, err
	}

	redisClient, err := data.InitRedis(cfg, logger)
	if err != nil {
		return nil, nil, err
	}

	return data.NewData(db, natsClient, redisClient, logger, cfg)
}

// 确保 zapcore 已导入
var _ = zapcore.ISO8601TimeEncoder
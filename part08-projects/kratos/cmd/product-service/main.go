package main

import (
	"os"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/config/file"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/v2/transport/http"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"kratos/internal/biz"
	"kratos/internal/config"
	"kratos/internal/data"
	"kratos/internal/repo"
	"kratos/internal/server"
	"kratos/internal/service"
	productV1 "kratos/api/product/v1"
)

func main() {
	logger := initLogger()
	defer logger.Sync()

	log := log.NewHelper(log.With(logger, "module", "main"))

	cfg := loadConfig()

	dataData, cleanup, err := newData(cfg, logger)
	if err != nil {
		log.Fatal("创建数据层失败", zap.Error(err))
	}
	defer cleanup()

	productUC := biz.NewProductUseCase(repo.NewProductRepo(dataData, logger), dataData.NATS, logger)

	httpServer := server.NewHTTPServer(cfg, logger)
	productService := service.NewProductService(productUC)
	productV1.RegisterProductServiceHTTPServer(httpServer, productService)

	grpcServer := server.NewGRPCServer(cfg, logger)
	productV1.RegisterProductServiceServer(grpcServer, productService)

	app := kratos.New(
		kratos.Name("product-service"),
		kratos.Version("v1.0.0"),
		kratos.Logger(logger),
		kratos.Server(
			httpServer,
			grpcServer,
		),
	)

	if err := app.Run(); err != nil {
		log.Fatal("启动服务失败", zap.Error(err))
	}
}

func initLogger() *zap.Logger {
	zapConfig := zap.NewProductionConfig()
	zapConfig.OutputPaths = []string{"stdout"}
	logger, _ := zapConfig.Build()
	return logger
}

func loadConfig() *config.Config {
	path := os.Getenv("CONFIG_PATH")
	if path == "" {
		path = "configs/product-service.yaml"
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
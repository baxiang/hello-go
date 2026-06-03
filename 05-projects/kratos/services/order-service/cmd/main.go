// order-service 入口
package main

import (
	"context"
	"flag"
	"log"
	"time"

	v1 "services/api/order/v1"
	"services/order-service/internal/biz"
	"services/order-service/internal/client"
	"services/order-service/internal/data"
	"services/order-service/internal/service"
	"services/pkg/config"
	"services/pkg/database"
	"services/pkg/server"
)

var configPath = flag.String("config", "configs/config.yaml", "配置文件路径")

func main() {
	flag.Parse()

	logger := server.NewLogger("order-service")
	defer func() { _ = logger.Sync() }()

	var cfg config.AppConfig
	if err := config.Load(*configPath, &cfg); err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// DB
	db, err := database.New(&cfg.Database, logger)
	if err != nil {
		log.Fatalf("初始化数据库失败: %v", err)
	}
	defer db.Close()

	if err := data.AutoMigrate(db); err != nil {
		log.Fatalf("迁移数据库失败: %v", err)
	}

	// 连接 product-service
	dialCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	productClient, err := client.NewProductClient(dialCtx, cfg.Services.Product.Addr, logger)
	if err != nil {
		log.Fatalf("连接 product-service 失败: %v", err)
	}
	defer productClient.Close()

	// 组装依赖
	orderRepo := data.NewOrderRepo(db, logger)
	orderUC := biz.NewOrderUseCase(orderRepo, productClient, logger)
	orderSvc := service.NewOrderService(orderUC)

	grpcSrv := server.NewGRPCServer(&cfg.Server.GRPC, logger)
	v1.RegisterOrderServiceServer(grpcSrv.RawServer(), orderSvc)

	server.RunAndWait(logger, grpcSrv)
}

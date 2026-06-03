// payment-service 入口
package main

import (
	"context"
	"flag"
	"log"
	"time"

	v1 "services/api/payment/v1"
	"services/payment-service/internal/biz"
	"services/payment-service/internal/client"
	"services/payment-service/internal/data"
	"services/payment-service/internal/service"
	"services/pkg/config"
	"services/pkg/database"
	"services/pkg/server"
)

var configPath = flag.String("config", "configs/config.yaml", "配置文件路径")

func main() {
	flag.Parse()

	logger := server.NewLogger("payment-service")
	defer func() { _ = logger.Sync() }()

	var cfg config.AppConfig
	if err := config.Load(*configPath, &cfg); err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	db, err := database.New(&cfg.Database, logger)
	if err != nil {
		log.Fatalf("初始化数据库失败: %v", err)
	}
	defer db.Close()

	if err := data.AutoMigrate(db); err != nil {
		log.Fatalf("迁移数据库失败: %v", err)
	}

	// 连接 order-service
	dialCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	orderClient, err := client.NewOrderClient(dialCtx, cfg.Services.Order.Addr, logger)
	if err != nil {
		log.Fatalf("连接 order-service 失败: %v", err)
	}
	defer orderClient.Close()

	paymentRepo := data.NewPaymentRepo(db, logger)
	paymentUC := biz.NewPaymentUseCase(paymentRepo, orderClient, logger)
	paymentSvc := service.NewPaymentService(paymentUC)

	grpcSrv := server.NewGRPCServer(&cfg.Server.GRPC, logger)
	v1.RegisterPaymentServiceServer(grpcSrv.RawServer(), paymentSvc)

	server.RunAndWait(logger, grpcSrv)
}

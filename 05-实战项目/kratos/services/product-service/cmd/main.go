// product-service 入口
package main

import (
	"flag"
	"log"

	v1 "services/api/product/v1"
	"services/pkg/config"
	"services/pkg/database"
	"services/pkg/server"
	"services/product-service/internal/biz"
	"services/product-service/internal/data"
	"services/product-service/internal/service"
)

var configPath = flag.String("config", "configs/config.yaml", "配置文件路径")

func main() {
	flag.Parse()

	logger := server.NewLogger("product-service")
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

	productRepo := data.NewProductRepo(db, logger)
	productUC := biz.NewProductUseCase(productRepo, logger)
	productSvc := service.NewProductService(productUC)

	grpcSrv := server.NewGRPCServer(&cfg.Server.GRPC, logger)
	v1.RegisterProductServiceServer(grpcSrv.RawServer(), productSvc)

	server.RunAndWait(logger, grpcSrv)
}

// user-service 入口
package main

import (
	"flag"
	"log"

	v1 "services/api/user/v1"
	"services/pkg/config"
	"services/pkg/database"
	"services/pkg/server"
	"services/pkg/token"
	"services/user-service/internal/biz"
	"services/user-service/internal/data"
	"services/user-service/internal/service"
)

var configPath = flag.String("config", "configs/config.yaml", "配置文件路径")

func main() {
	flag.Parse()

	logger := server.NewLogger("user-service")
	defer func() { _ = logger.Sync() }()

	// 加载配置
	var cfg config.AppConfig
	if err := config.Load(*configPath, &cfg); err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 初始化数据库
	db, err := database.New(&cfg.Database, logger)
	if err != nil {
		log.Fatalf("初始化数据库失败: %v", err)
	}
	defer db.Close()

	// 自动迁移表
	if err := data.AutoMigrate(db); err != nil {
		log.Fatalf("迁移数据库失败: %v", err)
	}

	// 组装依赖
	tokenMgr := token.NewManager(cfg.JWT.Secret, cfg.JWT.ExpireHour)
	userRepo := data.NewUserRepo(db, logger)
	userUC := biz.NewUserUseCase(userRepo, tokenMgr, logger)
	userSvc := service.NewUserService(userUC)

	// 启动 gRPC 服务
	grpcSrv := server.NewGRPCServer(&cfg.Server.GRPC, logger)
	v1.RegisterUserServiceServer(grpcSrv.RawServer(), userSvc)

	server.RunAndWait(logger, grpcSrv)
}

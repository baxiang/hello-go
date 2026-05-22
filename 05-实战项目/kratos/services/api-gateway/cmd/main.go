// api-gateway 入口
package main

import (
	"context"
	"flag"
	"log"
	"time"

	"github.com/gorilla/mux"

	"services/api-gateway/internal/client"
	"services/api-gateway/internal/handler"
	"services/api-gateway/internal/middleware"
	"services/pkg/config"
	"services/pkg/server"
)

var configPath = flag.String("config", "configs/config.yaml", "配置文件路径")

func main() {
	flag.Parse()

	logger := server.NewLogger("api-gateway")
	defer func() { _ = logger.Sync() }()

	var cfg config.AppConfig
	if err := config.Load(*configPath, &cfg); err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 连接所有后端服务
	dialCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	clients, err := client.New(dialCtx,
		cfg.Services.User.Addr,
		cfg.Services.Product.Addr,
		cfg.Services.Order.Addr,
		cfg.Services.Payment.Addr,
		logger,
	)
	if err != nil {
		log.Fatalf("连接后端服务失败: %v", err)
	}
	defer clients.Close()

	// 路由
	router := mux.NewRouter()
	h := handler.New(clients, logger)
	h.Register(router)

	// 中间件链
	whitelist := []string{
		"/health",
		"/api/v1/auth/login",
		"/api/v1/users", // 注册不需要登录
	}
	authMW := middleware.Auth(clients.User, whitelist, logger)
	logMW := middleware.Logger(logger)

	finalHandler := logMW(authMW(router))

	// 启动 HTTP 服务
	httpSrv := server.NewHTTPServer(&cfg.Server.HTTP, finalHandler, logger)

	server.RunAndWait(logger, httpSrv)
}

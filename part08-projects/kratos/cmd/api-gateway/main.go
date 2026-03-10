package main

import (
	"os"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/config/file"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	"github.com/go-kratos/kratos/v2/transport/http"
	"github.com/gorilla/mux"
	"go.uber.org/zap"

	"kratos/internal/config"
)

func main() {
	logger := initLogger()
	defer logger.Sync()

	log := log.NewHelper(log.With(logger, "module", "main"))

	cfg := loadConfig()

	// 创建 HTTP 服务器
	opts := []http.ServerOption{
		http.Middleware(
			recovery.Recovery(),
			tracing.Server(),
		),
		http.Address(cfg.Server.HTTP.Addr),
	}

	httpServer := http.NewServer(opts...)

	// 设置路由
	router := mux.NewRouter()

	// 用户服务代理
	router.HandleFunc("/api/v1/users", handleUserRequest).Methods("GET", "POST")
	router.HandleFunc("/api/v1/users/{id}", handleUserRequest).Methods("GET", "PUT", "DELETE")

	// 商品服务代理
	router.HandleFunc("/api/v1/products", handleProductRequest).Methods("GET", "POST")
	router.HandleFunc("/api/v1/products/{id}", handleProductRequest).Methods("GET", "PUT", "DELETE")
	router.HandleFunc("/api/v1/products/deduct", handleProductRequest).Methods("POST")

	// 订单服务代理
	router.HandleFunc("/api/v1/orders", handleOrderRequest).Methods("GET", "POST")
	router.HandleFunc("/api/v1/orders/{id}", handleOrderRequest).Methods("GET")
	router.HandleFunc("/api/v1/orders/{id}/cancel", handleOrderRequest).Methods("POST")
	router.HandleFunc("/api/v1/orders/{id}/pay", handleOrderRequest).Methods("POST")

	// 支付服务代理
	router.HandleFunc("/api/v1/payments", handlePaymentRequest).Methods("GET", "POST")
	router.HandleFunc("/api/v1/payments/callback", handlePaymentRequest).Methods("POST")

	// 认证
	router.HandleFunc("/api/v1/auth/login", handleAuthRequest).Methods("POST")
	router.HandleFunc("/api/v1/auth/logout", handleAuthRequest).Methods("POST")

	// 健康检查
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"status":"ok"}`))
	})

	httpServer.HandlePrefix("/", router)

	// 创建应用
	app := kratos.New(
		kratos.Name("api-gateway"),
		kratos.Version("v1.0.0"),
		kratos.Logger(logger),
		kratos.Server(
			httpServer,
		),
	)

	log.Info("API Gateway 启动", zap.String("addr", cfg.Server.HTTP.Addr))

	if err := app.Run(); err != nil {
		log.Fatal("启动服务失败", zap.Error(err))
	}
}

// 处理用户请求
func handleUserRequest(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(`{"message":"user service proxy"}`))
}

// 处理商品请求
func handleProductRequest(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(`{"message":"product service proxy"}`))
}

// 处理订单请求
func handleOrderRequest(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(`{"message":"order service proxy"}`))
}

// 处理支付请求
func handlePaymentRequest(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(`{"message":"payment service proxy"}`))
}

// 处理认证请求
func handleAuthRequest(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(`{"message":"auth service proxy"}`))
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
		path = "configs/api-gateway.yaml"
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
package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"ecommerce/internal/order/biz"
	"ecommerce/internal/order/data"
	"ecommerce/internal/order/service"
	"ecommerce/pkg/config"
	"ecommerce/pkg/database"
	"ecommerce/pkg/events"
	"ecommerce/pkg/kafka"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/message/router/middleware"
	"go.uber.org/zap"
)

var configPath = flag.String("config", "configs/config.yaml", "配置文件路径")

func main() {
	flag.Parse()

	log, _ := zap.NewProduction()
	defer log.Sync()

	// 加载配置
	var cfg config.Config
	if err := config.Load(*configPath, &cfg); err != nil {
		log.Fatal("加载配置失败", zap.Error(err))
	}

	// 初始化数据库
	db, err := database.New(cfg.Database.DSN, log)
	if err != nil {
		log.Fatal("数据库连接失败", zap.Error(err))
	}
	defer db.Close()
	data.AutoMigrate(db)

	// 初始化 Kafka
	watermillLogger := watermill.NewStdLogger(false, false)
	pub, err := kafka.NewPublisher(cfg.Kafka.Brokers, watermillLogger)
	if err != nil {
		log.Fatal("Kafka Publisher 创建失败", zap.Error(err))
	}
	sub, err := kafka.NewSubscriber(cfg.Kafka.Brokers, "order-service", watermillLogger)
	if err != nil {
		log.Fatal("Kafka Subscriber 创建失败", zap.Error(err))
	}

	// 依赖组装
	repo := data.NewOrderRepo(db, log)
	uc := biz.NewOrderUseCase(repo, log)
	eventHandler := biz.NewOrderEventHandler(uc, log)

	// Router — 消费外部事件
	router, err := message.NewRouter(message.RouterConfig{}, watermillLogger)
	if err != nil {
		log.Fatal("创建 Router 失败", zap.Error(err))
	}
	router.AddMiddleware(middleware.Recoverer)
	router.AddMiddleware(middleware.Retry{MaxRetries: 3, InitialInterval: time.Second, Logger: watermillLogger}.Middleware)

	router.AddNoPublisherHandler("payment_completed", events.TopicPaymentCompleted, sub, eventHandler.HandlePaymentCompleted)
	router.AddNoPublisherHandler("payment_failed", events.TopicPaymentFailed, sub, eventHandler.HandlePaymentFailed)
	router.AddNoPublisherHandler("inventory_insufficient", events.TopicInventoryInsufficient, sub, eventHandler.HandleInventoryInsufficient)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() { router.Run(ctx) }()
	<-router.Running()

	// HTTP Server
	svc := service.NewOrderService(pub, uc, log)
	http.HandleFunc("/orders", svc.HandleCreateOrder)
	addr := fmt.Sprintf(":%d", cfg.Server.HTTPPort)
	log.Info("订单服务启动", zap.String("addr", addr))

	go func() { http.ListenAndServe(addr, nil) }()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
	log.Info("订单服务关闭")
}

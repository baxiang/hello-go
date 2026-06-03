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

	"ecommerce/internal/inventory/biz"
	"ecommerce/internal/inventory/data"
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

	var cfg config.Config
	if err := config.Load(*configPath, &cfg); err != nil {
		log.Fatal("加载配置失败", zap.Error(err))
	}

	db, err := database.New(cfg.Database.DSN, log)
	if err != nil {
		log.Fatal("数据库连接失败", zap.Error(err))
	}
	defer db.Close()
	data.AutoMigrate(db)

	watermillLogger := watermill.NewStdLogger(false, false)
	pub, err := kafka.NewPublisher(cfg.Kafka.Brokers, watermillLogger)
	if err != nil {
		log.Fatal("Kafka Publisher 创建失败", zap.Error(err))
	}
	sub, err := kafka.NewSubscriber(cfg.Kafka.Brokers, "inventory-service", watermillLogger)
	if err != nil {
		log.Fatal("Kafka Subscriber 创建失败", zap.Error(err))
	}

	repo := data.NewInventoryRepo(db, log)
	uc := biz.NewInventoryUseCase(repo, log)
	eventHandler := biz.NewInventoryEventHandler(uc, pub, log)

	router, err := message.NewRouter(message.RouterConfig{}, watermillLogger)
	if err != nil {
		log.Fatal("创建 Router 失败", zap.Error(err))
	}
	router.AddMiddleware(middleware.Recoverer)
	router.AddMiddleware(middleware.Retry{MaxRetries: 3, InitialInterval: time.Second, Logger: watermillLogger}.Middleware)

	router.AddNoPublisherHandler("order_created", events.TopicOrderCreated, sub, eventHandler.HandleOrderCreated)
	router.AddNoPublisherHandler("inventory_release", events.TopicInventoryRelease, sub, eventHandler.HandleInventoryRelease)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() { router.Run(ctx) }()
	<-router.Running()

	addr := fmt.Sprintf(":%d", cfg.Server.HTTPPort)
	log.Info("库存服务启动", zap.String("addr", addr))
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})

	go func() { http.ListenAndServe(addr, nil) }()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
	log.Info("库存服务关闭")
}

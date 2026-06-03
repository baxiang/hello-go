package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"ecommerce/internal/notification/biz"
	"ecommerce/pkg/config"
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

	watermillLogger := watermill.NewStdLogger(false, false)
	sub, err := kafka.NewSubscriber(cfg.Kafka.Brokers, "notification-service", watermillLogger)
	if err != nil {
		log.Fatal("Kafka Subscriber 创建失败", zap.Error(err))
	}

	handler := biz.NewNotificationHandler(log)

	router, err := message.NewRouter(message.RouterConfig{}, watermillLogger)
	if err != nil {
		log.Fatal("创建 Router 失败", zap.Error(err))
	}
	router.AddMiddleware(middleware.Recoverer)

	router.AddNoPublisherHandler("order_created_notify", events.TopicOrderCreated, sub, handler.HandleOrderCreated)
	router.AddNoPublisherHandler("payment_completed_notify", events.TopicPaymentCompleted, sub, handler.HandleOrderConfirmed)
	router.AddNoPublisherHandler("order_cancelled_notify", events.TopicOrderCancelled, sub, handler.HandleOrderCancelled)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() { router.Run(ctx) }()
	<-router.Running()

	addr := fmt.Sprintf(":%d", cfg.Server.HTTPPort)
	log.Info("通知服务启动", zap.String("addr", addr))
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})

	go func() { http.ListenAndServe(addr, nil) }()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
	log.Info("通知服务关闭")
}

// Package main 演示 Watermill 集成 Prometheus 指标采集
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/pubsub/gochannel"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	messagesProcessed = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "demo_messages_processed_total",
		Help: "Total number of messages processed",
	})
)

func init() {
	prometheus.MustRegister(messagesProcessed)
}

func main() {
	logger := watermill.NewStdLogger(false, false)
	pubSub := gochannel.NewGoChannel(gochannel.Config{}, logger)

	router, err := message.NewRouter(message.RouterConfig{}, logger)
	if err != nil {
		log.Fatal(err)
	}

	router.AddNoPublisherHandler(
		"demo-handler",
		"demo.topic",
		pubSub,
		func(msg *message.Message) error {
			messagesProcessed.Inc()
			fmt.Printf("处理消息: %s\n", string(msg.Payload))
			time.Sleep(100 * time.Millisecond)
			return nil
		},
	)

	// 启动 Prometheus HTTP 端点
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		fmt.Println("指标端点: http://localhost:2112/metrics")
		http.ListenAndServe(":2112", nil)
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		if err := router.Run(ctx); err != nil {
			log.Fatal(err)
		}
	}()
	<-router.Running()

	// 持续发布消息，产生指标
	go func() {
		for {
			msg := message.NewMessage(watermill.NewUUID(), []byte("ping"))
			pubSub.Publish("demo.topic", msg)
			time.Sleep(500 * time.Millisecond)
		}
	}()

	fmt.Println("持续发布消息中，访问 http://localhost:2112/metrics 查看指标")
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
	fmt.Println("\n05-metrics 演示完成")
}

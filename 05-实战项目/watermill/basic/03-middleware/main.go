// Package main 演示 Watermill 中间件：重试、超时、恢复
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/message/router/middleware"
	"github.com/ThreeDotsLabs/watermill/pubsub/gochannel"
)

var failCount int

func main() {
	logger := watermill.NewStdLogger(false, false)
	pubSub := gochannel.NewGoChannel(gochannel.Config{}, logger)

	router, err := message.NewRouter(message.RouterConfig{}, logger)
	if err != nil {
		log.Fatal(err)
	}

	// 重试中间件：最多重试 3 次，初始间隔 500ms
	router.AddMiddleware(middleware.Retry{
		MaxRetries:      3,
		InitialInterval: 500 * time.Millisecond,
		Logger:          logger,
	}.Middleware)

	// 超时中间件：30 秒超时
	router.AddMiddleware(middleware.Timeout(30 * time.Second))

	// 恢复中间件：捕获 panic
	router.AddMiddleware(middleware.Recoverer)

	// Handler：前 2 次会失败，模拟重试
	router.AddNoPublisherHandler(
		"retry-handler",
		"tasks",
		pubSub,
		func(msg *message.Message) error {
			failCount++
			if failCount <= 2 {
				fmt.Printf("[retry-handler] 第 %d 次处理失败，将被重试\n", failCount)
				return fmt.Errorf("临时错误，第 %d 次", failCount)
			}
			fmt.Printf("[retry-handler] 第 %d 次处理成功: %s\n", failCount, string(msg.Payload))
			return nil
		},
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		if err := router.Run(ctx); err != nil {
			log.Fatal(err)
		}
	}()
	<-router.Running()

	msg := message.NewMessage(watermill.NewUUID(), []byte("重要任务"))
	pubSub.Publish("tasks", msg)
	fmt.Println("发布: 重要任务（将重试 2 次后成功）")

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
	fmt.Println("\n03-middleware 演示完成")
}

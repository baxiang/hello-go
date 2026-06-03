// Package main 演示 Watermill Router 的消息路由与多 Handler
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/message/router/middleware"
	"github.com/ThreeDotsLabs/watermill/pubsub/gochannel"
)

func main() {
	logger := watermill.NewStdLogger(false, false)
	pubSub := gochannel.NewGoChannel(gochannel.Config{}, logger)

	router, err := message.NewRouter(message.RouterConfig{}, logger)
	if err != nil {
		log.Fatal(err)
	}
	router.AddMiddleware(middleware.Recoverer)

	// Handler 1: 处理 orders 事件，转换为偶数/奇数分流
	router.AddHandler(
		"order-handler",
		"order.created",
		pubSub,
		"order.even",
		pubSub,
		func(msg *message.Message) ([]*message.Message, error) {
			payload := string(msg.Payload)
			fmt.Printf("[order-handler] 收到: %s\n", payload)

			var orderID int
			fmt.Sscanf(payload, "订单 #%d", &orderID)

			if orderID%2 == 0 {
				return []*message.Message{
					message.NewMessage(watermill.NewUUID(), []byte(fmt.Sprintf("偶数订单 #%d", orderID))),
				}, nil
			}
			return []*message.Message{
				message.NewMessage(watermill.NewUUID(), []byte(fmt.Sprintf("奇数订单 #%d", orderID))),
			}, nil
		},
	)

	// Handler 2: 消费偶数订单（无输出 topic）
	router.AddNoPublisherHandler(
		"even-handler",
		"order.even",
		pubSub,
		func(msg *message.Message) error {
			fmt.Printf("  [even-handler] 处理: %s\n", string(msg.Payload))
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

	// 发布 5 条订单
	for i := 1; i <= 5; i++ {
		msg := message.NewMessage(watermill.NewUUID(), []byte(fmt.Sprintf("订单 #%d", i)))
		if err := pubSub.Publish("order.created", msg); err != nil {
			log.Fatal(err)
		}
		fmt.Printf("发布: 订单 #%d\n", i)
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
	fmt.Println("\n02-router 演示完成")
}

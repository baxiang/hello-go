// Package main 演示 Watermill 最基础的发布-订阅模式
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/pubsub/gochannel"
)

func main() {
	logger := watermill.NewStdLogger(false, false)
	pubSub := gochannel.NewGoChannel(gochannel.Config{}, logger)

	ctx := context.Background()

	// 订阅 topic "hello"
	messages, err := pubSub.Subscribe(ctx, "hello")
	if err != nil {
		log.Fatal(err)
	}

	// 在 goroutine 中消费消息
	go func() {
		for msg := range messages {
			fmt.Printf("收到消息: %s (UUID: %s)\n", string(msg.Payload), msg.UUID)
			msg.Ack()
		}
	}()

	// 发布 3 条消息
	for i := 1; i <= 3; i++ {
		payload := []byte(fmt.Sprintf("消息 #%d", i))
		msg := message.NewMessage(watermill.NewUUID(), payload)
		if err := pubSub.Publish("hello", msg); err != nil {
			log.Fatal(err)
		}
		fmt.Printf("发布消息: %s\n", payload)
	}

	// 等待消费完成
	time.Sleep(time.Second)
	fmt.Println("01-pubsub 演示完成")
}

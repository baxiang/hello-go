package main

import (
	"context"
	"log"
	"sync/atomic"
	"time"

	"github.com/IBM/sarama"
)

type ConsumerGroupHandler struct {
	messageCount int64
}

func (h *ConsumerGroupHandler) Setup(session sarama.ConsumerGroupSession) error {
	log.Printf("消费者组启动，分配到分区: %v", session.Claims())
	return nil
}

func (h *ConsumerGroupHandler) Cleanup(session sarama.ConsumerGroupSession) error {
	log.Printf("消费者组关闭，共消费 %d 条消息", atomic.LoadInt64(&h.messageCount))
	return nil
}

func (h *ConsumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		atomic.AddInt64(&h.messageCount, 1)
		session.MarkMessage(msg, "")
	}
	return nil
}

func main() {
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true
	config.Consumer.Offsets.Initial = sarama.OffsetOldest
	config.Consumer.Offsets.AutoCommit.Enable = true
	config.Consumer.Offsets.AutoCommit.Interval = 5 * time.Second

	group, err := sarama.NewConsumerGroup([]string{"localhost:9092"}, "advanced-group", config)
	if err != nil {
		log.Fatalf("创建消费者组失败: %v", err)
	}
	defer group.Close()

	// 处理错误
	go func() {
		for err := range group.Errors() {
			log.Printf("错误: %v", err)
		}
	}()

	handler := &ConsumerGroupHandler{}
	ctx := context.Background()

	log.Println("消费者启动，等待消息...")
	for {
		err := group.Consume(ctx, []string{"advanced-topic"}, handler)
		if err != nil {
			log.Printf("消费失败: %v", err)
		}
		if ctx.Err() != nil {
			break
		}
	}
}
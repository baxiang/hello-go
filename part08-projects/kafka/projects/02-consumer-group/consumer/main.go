package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/IBM/sarama"
)

// ConsumerGroupHandler 实现 sarama.ConsumerGroupHandler 接口
type ConsumerGroupHandler struct {
	consumerID string
}

func (h *ConsumerGroupHandler) Setup(session sarama.ConsumerGroupSession) error {
	log.Printf("[%s] 消费者组重平衡完成，分配到分区: %v", h.consumerID, session.Claims())
	return nil
}

func (h *ConsumerGroupHandler) Cleanup(session sarama.ConsumerGroupSession) error {
	log.Printf("[%s] 消费者组清理", h.consumerID)
	return nil
}

func (h *ConsumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		log.Printf("[%s] 消费消息: partition=%d, offset=%d, key=%s, value=%s",
			h.consumerID, msg.Partition, msg.Offset, string(msg.Key), string(msg.Value))

		// 模拟消息处理
		time.Sleep(100 * time.Millisecond)

		// 手动提交 offset
		session.MarkMessage(msg, "")
	}
	return nil
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("用法: go run consumer/main.go <消费者ID>")
		fmt.Println("示例: go run consumer/main.go consumer-1")
		os.Exit(1)
	}

	consumerID := os.Args[1]
	groupID := "hello-group"
	topic := "group-topic"

	// 消费者配置
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true
	config.Consumer.Offsets.Initial = sarama.OffsetOldest
	config.Consumer.Offsets.AutoCommit.Enable = false // 手动提交
	config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin

	// 创建消费者组
	group, err := sarama.NewConsumerGroup([]string{"localhost:9092"}, groupID, config)
	if err != nil {
		log.Fatalf("[%s] 创建消费者组失败: %v", consumerID, err)
	}
	defer func() {
		if err := group.Close(); err != nil {
			log.Printf("[%s] 关闭消费者组失败: %v", consumerID, err)
		}
	}()

	// 处理错误
	go func() {
		for err := range group.Errors() {
			log.Printf("[%s] 消费者组错误: %v", consumerID, err)
		}
	}()

	// 消费消息
	handler := &ConsumerGroupHandler{consumerID: consumerID}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 监听退出信号
	go func() {
		sigterm := make(chan os.Signal, 1)
		signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)
		<-sigterm
		log.Printf("[%s] 收到退出信号，正在关闭...", consumerID)
		cancel()
	}()

	for {
		err := group.Consume(ctx, []string{topic}, handler)
		if err != nil {
			log.Printf("[%s] 消费失败: %v", consumerID, err)
		}
		if ctx.Err() != nil {
			break
		}
	}

	log.Printf("[%s] 消费者已退出", consumerID)
}
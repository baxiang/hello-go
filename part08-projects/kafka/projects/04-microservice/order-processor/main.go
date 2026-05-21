package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/IBM/sarama"
)

// OrderEvent 订单事件
type OrderEvent struct {
	OrderID   string  `json:"order_id"`
	UserID    string  `json:"user_id"`
	Amount    float64 `json:"amount"`
	Status    string  `json:"status"`
	Timestamp string  `json:"timestamp"`
}

type OrderProcessor struct{}

func (p *OrderProcessor) Setup(session sarama.ConsumerGroupSession) error {
	log.Println("订单处理器启动")
	return nil
}

func (p *OrderProcessor) Cleanup(session sarama.ConsumerGroupSession) error {
	log.Println("订单处理器关闭")
	return nil
}

func (p *OrderProcessor) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		var order OrderEvent
		if err := json.Unmarshal(msg.Value, &order); err != nil {
			log.Printf("解析订单失败: %v", err)
			continue
		}

		// 模拟订单处理
		switch order.Status {
		case "created":
			log.Printf("[订单处理] Order %s: 创建成功，金额 %.2f", order.OrderID, order.Amount)
		case "paid":
			log.Printf("[订单处理] Order %s: 支付完成", order.OrderID)
		case "shipped":
			log.Printf("[订单处理] Order %s: 已发货", order.OrderID)
		case "delivered":
			log.Printf("[订单处理] Order %s: 已送达", order.OrderID)
		case "cancelled":
			log.Printf("[订单处理] Order %s: 订单已取消", order.OrderID)
		}

		session.MarkMessage(msg, "")
	}
	return nil
}

func main() {
	config := sarama.NewConfig()
	config.Consumer.Offsets.Initial = sarama.OffsetOldest
	config.Consumer.Offsets.AutoCommit.Enable = true

	group, err := sarama.NewConsumerGroup([]string{"localhost:9092"}, "order-processor-group", config)
	if err != nil {
		log.Fatalf("创建消费者组失败: %v", err)
	}
	defer group.Close()

	go func() {
		for err := range group.Errors() {
			log.Printf("错误: %v", err)
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		sigterm := make(chan os.Signal, 1)
		signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)
		<-sigterm
		cancel()
	}()

	processor := &OrderProcessor{}
	for {
		err := group.Consume(ctx, []string{"orders"}, processor)
		if err != nil {
			log.Printf("消费失败: %v", err)
		}
		if ctx.Err() != nil {
			break
		}
	}

	log.Println("订单处理器已退出")
}
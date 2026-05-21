package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/IBM/sarama"
)

// OrderEvent 订单事件
type OrderEvent struct {
	OrderID     string    `json:"order_id"`
	UserID      string    `json:"user_id"`
	ProductID   string    `json:"product_id"`
	Amount      float64   `json:"amount"`
	Status      string    `json:"status"`
	Timestamp   time.Time `json:"timestamp"`
}

func main() {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.RequiredAcks = sarama.WaitForLocal
	config.Producer.Retry.Max = 3
	config.Producer.Compression = sarama.CompressionSnappy

	producer, err := sarama.NewSyncProducer([]string{"localhost:9092"}, config)
	if err != nil {
		log.Fatalf("创建生产者失败: %v", err)
	}
	defer producer.Close()

	// 发送订单事件到不同 topic
	topics := []string{"orders", "order-logs"}
	log.Printf("开始模拟订单系统，发送事件到 topics: %v", topics)

	statuses := []string{"created", "paid", "shipped", "delivered", "cancelled"}

	for i := 0; i < 100; i++ {
		event := OrderEvent{
			OrderID:   fmt.Sprintf("order-%d", i),
			UserID:    fmt.Sprintf("user-%d", rand.Intn(100)),
			ProductID: fmt.Sprintf("product-%d", rand.Intn(50)),
			Amount:    float64(rand.Intn(10000)) / 100,
			Status:    statuses[rand.Intn(len(statuses))],
			Timestamp: time.Now(),
		}

		data, _ := json.Marshal(event)

		// 发送订单到 orders topic
		msg := &sarama.ProducerMessage{
			Topic: "orders",
			Key:   sarama.StringEncoder(event.OrderID),
			Value: sarama.ByteEncoder(data),
		}
		_, _, err := producer.SendMessage(msg)
		if err != nil {
			log.Printf("发送订单失败: %v", err)
			continue
		}

		// 发送日志到 order-logs topic
		logMsg := &sarama.ProducerMessage{
			Topic: "order-logs",
			Key:   sarama.StringEncoder(event.OrderID),
			Value: sarama.StringEncoder(fmt.Sprintf("[%s] Order %s status changed to %s",
				event.Timestamp.Format(time.RFC3339), event.OrderID, event.Status)),
		}
		_, _, err = producer.SendMessage(logMsg)
		if err != nil {
			log.Printf("发送日志失败: %v", err)
			continue
		}

		log.Printf("订单 %s 已创建，状态: %s，金额: %.2f", event.OrderID, event.Status, event.Amount)
		time.Sleep(200 * time.Millisecond)
	}

	log.Println("订单模拟完成")
}
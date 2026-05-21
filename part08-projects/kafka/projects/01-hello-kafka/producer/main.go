package main

import (
	"fmt"
	"log"

	"github.com/IBM/sarama"
)

func main() {
	// 生产者配置
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.RequiredAcks = sarama.WaitForLocal       // 等待本地确认
	config.Producer.Retry.Max = 3                            // 失败重试3次

	// 创建生产者
	producer, err := sarama.NewSyncProducer([]string{"localhost:9092"}, config)
	if err != nil {
		log.Fatalf("创建生产者失败: %v", err)
	}
	defer func() {
		if err := producer.Close(); err != nil {
			log.Printf("关闭生产者失败: %v", err)
		}
	}()

	// 发送消息
	topic := "hello-topic"
	for i := 0; i < 10; i++ {
		msg := &sarama.ProducerMessage{
			Topic: topic,
			Key:   sarama.StringEncoder(fmt.Sprintf("key-%d", i%3)),
			Value: sarama.StringEncoder(fmt.Sprintf("Hello Kafka %d", i)),
		}

		partition, offset, err := producer.SendMessage(msg)
		if err != nil {
			log.Printf("发送消息失败: %v", err)
			continue
		}

		log.Printf("消息已发送: topic=%s, partition=%d, offset=%d", topic, partition, offset)
	}

	log.Println("生产者发送完成")
}
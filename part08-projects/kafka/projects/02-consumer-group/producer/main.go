package main

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/IBM/sarama"
)

func main() {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.Partitioner = sarama.NewHashPartitioner

	producer, err := sarama.NewSyncProducer([]string{"localhost:9092"}, config)
	if err != nil {
		log.Fatalf("创建生产者失败: %v", err)
	}
	defer producer.Close()

	topic := "group-topic"
	log.Printf("开始向主题 %s 发送消息（按 Ctrl+C 停止）...", topic)

	for {
		key := fmt.Sprintf("user-%d", rand.Intn(5))
		value := fmt.Sprintf("消息-%d-%s", time.Now().Unix(), key)

		msg := &sarama.ProducerMessage{
			Topic: topic,
			Key:   sarama.StringEncoder(key),
			Value: sarama.StringEncoder(value),
		}

		partition, offset, err := producer.SendMessage(msg)
		if err != nil {
			log.Printf("发送失败: %v", err)
			continue
		}

		log.Printf("已发送: partition=%d, offset=%d, key=%s", partition, offset, key)
		time.Sleep(500 * time.Millisecond)
	}
}
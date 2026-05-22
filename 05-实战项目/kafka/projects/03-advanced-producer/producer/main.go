package main

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/IBM/sarama"
)

func main() {
	config := sarama.NewConfig()

	// 可靠性配置
	config.Producer.RequiredAcks = sarama.WaitForAll          // 等待所有 ISR 副本确认
	config.Producer.Retry.Max = 5                              // 失败重试 5 次
	config.Producer.Retry.Backoff = 100 * time.Millisecond     // 重试间隔
	config.Producer.Idempotent = true                          // 开启幂等性

	// 批量发送配置
	config.Producer.Flush.Bytes = 1024 * 1024                  // 批次大小 1MB
	config.Producer.Flush.Messages = 1000                      // 每批最多 1000 条
	config.Producer.Flush.Frequency = 100 * time.Millisecond   // 最长等待 100ms

	// 压缩配置
	config.Producer.Compression = sarama.CompressionSnappy     // Snappy 压缩

	// 性能配置
	config.Producer.MaxMessageBytes = 1024 * 1024              // 最大消息 1MB
	config.Net.MaxOpenRequests = 5                             // 最大并发请求数

	// 事务配置（可选，需配合 Transaction.ID）
	// config.Producer.Transaction.ID = "my-transaction-id"

	producer, err := sarama.NewSyncProducer([]string{"localhost:9092"}, config)
	if err != nil {
		log.Fatalf("创建生产者失败: %v", err)
	}
	defer producer.Close()

	topic := "advanced-topic"
	messageCount := 10000

	log.Println("开始批量发送消息...")
	start := time.Now()

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for j := 0; j < messageCount/10; j++ {
				msg := &sarama.ProducerMessage{
					Topic: topic,
					Key:   sarama.StringEncoder(fmt.Sprintf("key-%d", workerID)),
					Value: sarama.StringEncoder(fmt.Sprintf("消息-%d-%d", workerID, j)),
				}
				_, _, err := producer.SendMessage(msg)
				if err != nil {
					log.Printf("发送失败: %v", err)
				}
			}
		}(i)
	}

	wg.Wait()
	elapsed := time.Since(start)
	throughput := float64(messageCount) / elapsed.Seconds()

	log.Printf("发送完成: %d 条消息, 耗时: %v, 吞吐量: %.2f msg/s", messageCount, elapsed, throughput)
}
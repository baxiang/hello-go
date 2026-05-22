package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/IBM/sarama"
)

func main() {
	// 消费者配置
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true
	config.Consumer.Offsets.Initial = sarama.OffsetOldest // 从最早消息开始消费

	// 创建消费者
	consumer, err := sarama.NewConsumer([]string{"localhost:9092"}, config)
	if err != nil {
		log.Fatalf("创建消费者失败: %v", err)
	}
	defer func() {
		if err := consumer.Close(); err != nil {
			log.Printf("关闭消费者失败: %v", err)
		}
	}()

	// 获取主题分区列表
	partitions, err := consumer.Partitions("hello-topic")
	if err != nil {
		log.Fatalf("获取分区失败: %v", err)
	}

	// 为每个分区创建消费实例
	var partitionConsumers []sarama.PartitionConsumer
	for _, partition := range partitions {
		pc, err := consumer.ConsumePartition("hello-topic", partition, sarama.OffsetOldest)
		if err != nil {
			log.Printf("创建分区消费者失败: %v", err)
			continue
		}
		partitionConsumers = append(partitionConsumers, pc)

		// 异步消费消息
		go func(p int32, pc sarama.PartitionConsumer) {
			for msg := range pc.Messages() {
				log.Printf("收到消息: partition=%d, offset=%d, key=%s, value=%s",
					msg.Partition, msg.Offset, string(msg.Key), string(msg.Value))
			}
		}(partition, pc)
	}

	// 等待退出信号
	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)
	<-sigterm

	log.Println("消费者正在关闭...")
	for _, pc := range partitionConsumers {
		if err := pc.Close(); err != nil {
			log.Printf("关闭分区消费者失败: %v", err)
		}
	}
	log.Println("消费者已关闭")
}
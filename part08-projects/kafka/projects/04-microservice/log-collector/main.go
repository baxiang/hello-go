package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/IBM/sarama"
)

type LogCollector struct {
	logCount int
}

func (c *LogCollector) Setup(session sarama.ConsumerGroupSession) error {
	log.Println("日志收集器启动")
	return nil
}

func (c *LogCollector) Cleanup(session sarama.ConsumerGroupSession) error {
	log.Printf("日志收集器关闭，共收集 %d 条日志", c.logCount)
	return nil
}

func (c *LogCollector) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		c.logCount++
		value := string(msg.Value)

		// 简单的日志分类
		var level string
		switch {
		case strings.Contains(value, "cancelled"):
			level = "WARN"
		case strings.Contains(value, "paid"):
			level = "INFO"
		default:
			level = "DEBUG"
		}

		fmt.Printf("[%s] %s\n", level, value)
		session.MarkMessage(msg, "")
	}
	return nil
}

func main() {
	config := sarama.NewConfig()
	config.Consumer.Offsets.Initial = sarama.OffsetOldest
	config.Consumer.Offsets.AutoCommit.Enable = true

	group, err := sarama.NewConsumerGroup([]string{"localhost:9092"}, "log-collector-group", config)
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

	collector := &LogCollector{}
	for {
		err := group.Consume(ctx, []string{"order-logs"}, collector)
		if err != nil {
			log.Printf("消费失败: %v", err)
		}
		if ctx.Err() != nil {
			break
		}
	}

	log.Println("日志收集器已退出")
}
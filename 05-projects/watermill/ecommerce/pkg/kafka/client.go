// Package kafka 提供 Kafka Publisher 和 Subscriber 管理
package kafka

import (
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-kafka/v3/pkg/kafka"
	"github.com/ThreeDotsLabs/watermill/message"
)

// NewPublisher 创建 Kafka Publisher
func NewPublisher(brokers []string, logger watermill.LoggerAdapter) (message.Publisher, error) {
	return kafka.NewPublisher(
		kafka.PublisherConfig{
			Brokers:   brokers,
			Marshaler: kafka.DefaultMarshaler{},
		},
		logger,
	)
}

// NewSubscriber 创建 Kafka Subscriber
func NewSubscriber(brokers []string, consumerGroup string, logger watermill.LoggerAdapter) (message.Subscriber, error) {
	return kafka.NewSubscriber(
		kafka.SubscriberConfig{
			Brokers:       brokers,
			ConsumerGroup: consumerGroup,
			Unmarshaler:   kafka.DefaultMarshaler{},
		},
		logger,
	)
}

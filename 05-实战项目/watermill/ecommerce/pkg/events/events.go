// Package events 提供事件序列化与 Topic 常量
package events

import (
	"fmt"

	"google.golang.org/protobuf/proto"
)

// Kafka 主题常量
const (
	TopicOrderCreated          = "order.events"
	TopicInventoryReserved     = "inventory.events"
	TopicInventoryInsufficient = "inventory.events"
	TopicPaymentRequested      = "payment.events"
	TopicPaymentCompleted      = "payment.events"
	TopicPaymentFailed         = "payment.events"
	TopicOrderConfirmed        = "order.events"
	TopicOrderCancelled        = "order.events"
	TopicInventoryRelease      = "inventory.events"
	TopicNotificationSent      = "notification.events"
)

// ProtoMarshaler protobuf 序列化器
type ProtoMarshaler struct{}

func (m ProtoMarshaler) Marshal(v interface{}) ([]byte, error) {
	return proto.Marshal(v.(proto.Message))
}

func (m ProtoMarshaler) Unmarshal(data []byte, v interface{}) error {
	msg, ok := v.(proto.Message)
	if !ok {
		return fmt.Errorf("expected proto.Message, got %T", v)
	}
	return proto.Unmarshal(data, msg)
}

// MustMarshal 序列化 protobuf 消息，panic on error
func MustMarshal(m proto.Message) []byte {
	data, err := proto.Marshal(m)
	if err != nil {
		panic(err)
	}
	return data
}

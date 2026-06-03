// Package events 提供事件序列化与 Topic 常量
package events

import (
	"fmt"

	"google.golang.org/protobuf/proto"
)

// Kafka 主题常量 — 每个事件类型使用独立 topic，避免不同类型消息被同一 handler 误消费
const (
	TopicOrderCreated          = "order.created"
	TopicInventoryReserved     = "inventory.reserved"
	TopicInventoryInsufficient = "inventory.insufficient"
	TopicPaymentRequested      = "payment.requested"
	TopicPaymentCompleted      = "payment.completed"
	TopicPaymentFailed         = "payment.failed"
	TopicOrderConfirmed        = "order.confirmed"
	TopicOrderCancelled        = "order.cancelled"
	TopicInventoryRelease      = "inventory.release"
	TopicNotificationSent      = "notification.sent"
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

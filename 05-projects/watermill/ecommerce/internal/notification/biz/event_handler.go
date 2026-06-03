package biz

import (
	"fmt"

	eventsv1 "ecommerce/api/proto/events/v1"

	"github.com/ThreeDotsLabs/watermill/message"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

type NotificationHandler struct {
	log *zap.Logger
}

func NewNotificationHandler(log *zap.Logger) *NotificationHandler {
	return &NotificationHandler{log: log}
}

func (h *NotificationHandler) HandleOrderCreated(msg *message.Message) error {
	var event eventsv1.OrderCreated
	if err := proto.Unmarshal(msg.Payload, &event); err != nil {
		return err
	}
	h.log.Info(fmt.Sprintf("通知: 用户 %s 的订单 %s 已创建", event.UserId, event.OrderId))
	return nil
}

func (h *NotificationHandler) HandleOrderConfirmed(msg *message.Message) error {
	var event eventsv1.PaymentCompleted
	if err := proto.Unmarshal(msg.Payload, &event); err != nil {
		return err
	}
	h.log.Info(fmt.Sprintf("通知: 订单 %s 支付成功，payment_id: %s", event.OrderId, event.PaymentId))
	return nil
}

func (h *NotificationHandler) HandleOrderCancelled(msg *message.Message) error {
	var event eventsv1.OrderCancelled
	if err := proto.Unmarshal(msg.Payload, &event); err != nil {
		return err
	}
	h.log.Info(fmt.Sprintf("通知: 订单 %s 已取消，原因: %s", event.OrderId, event.Reason))
	return nil
}

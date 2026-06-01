package biz

import (
	"context"

	eventsv1 "ecommerce/api/proto/events/v1"

	"github.com/ThreeDotsLabs/watermill/message"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

type OrderEventHandler struct {
	uc  *OrderUseCase
	log *zap.Logger
}

func NewOrderEventHandler(uc *OrderUseCase, log *zap.Logger) *OrderEventHandler {
	return &OrderEventHandler{uc: uc, log: log}
}

// HandlePaymentCompleted 处理支付完成事件 → 确认订单
func (h *OrderEventHandler) HandlePaymentCompleted(msg *message.Message) error {
	var event eventsv1.PaymentCompleted
	if err := proto.Unmarshal(msg.Payload, &event); err != nil {
		return err
	}
	h.log.Info("收到支付完成事件", zap.String("order_id", event.OrderId))
	return h.uc.Confirm(context.Background(), event.OrderId)
}

// HandlePaymentFailed 处理支付失败事件 → 取消订单
func (h *OrderEventHandler) HandlePaymentFailed(msg *message.Message) error {
	var event eventsv1.PaymentFailed
	if err := proto.Unmarshal(msg.Payload, &event); err != nil {
		return err
	}
	h.log.Info("收到支付失败事件，取消订单", zap.String("order_id", event.OrderId))
	return h.uc.Cancel(context.Background(), event.OrderId)
}

// HandleInventoryInsufficient 处理库存不足事件 → 取消订单
func (h *OrderEventHandler) HandleInventoryInsufficient(msg *message.Message) error {
	var event eventsv1.InventoryInsufficient
	if err := proto.Unmarshal(msg.Payload, &event); err != nil {
		return err
	}
	h.log.Info("收到库存不足事件，取消订单", zap.String("order_id", event.OrderId))
	return h.uc.Cancel(context.Background(), event.OrderId)
}

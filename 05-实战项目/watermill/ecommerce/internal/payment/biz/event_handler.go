package biz

import (
	"context"

	eventsv1 "ecommerce/api/proto/events/v1"

	"github.com/ThreeDotsLabs/watermill/message"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

type PaymentEventHandler struct {
	uc  *PaymentUseCase
	log *zap.Logger
}

func NewPaymentEventHandler(uc *PaymentUseCase, log *zap.Logger) *PaymentEventHandler {
	return &PaymentEventHandler{uc: uc, log: log}
}

// HandleInventoryReserved 库存预留成功后 → 发起支付
func (h *PaymentEventHandler) HandleInventoryReserved(msg *message.Message) error {
	var event eventsv1.InventoryReserved
	if err := proto.Unmarshal(msg.Payload, &event); err != nil {
		return err
	}
	h.log.Info("库存已预留，发起支付", zap.String("order_id", event.OrderId))

	paymentReq := &eventsv1.PaymentRequested{
		OrderId: event.OrderId,
		Amount:  100.0,
	}
	h.uc.ProcessPayment(context.Background(), paymentReq)
	return nil
}

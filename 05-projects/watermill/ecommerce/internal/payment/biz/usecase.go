package biz

import (
	"context"
	"math/rand"

	eventsv1 "ecommerce/api/proto/events/v1"
	"ecommerce/pkg/events"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"go.uber.org/zap"
)

type PaymentUseCase struct {
	pub message.Publisher
	log *zap.Logger
}

func NewPaymentUseCase(pub message.Publisher, log *zap.Logger) *PaymentUseCase {
	return &PaymentUseCase{pub: pub, log: log}
}

// ProcessPayment 模拟支付处理（90% 成功率）
func (uc *PaymentUseCase) ProcessPayment(ctx context.Context, event *eventsv1.PaymentRequested) {
	paymentID := watermill.NewUUID()

	if rand.Intn(100) < 90 {
		completed := &eventsv1.PaymentCompleted{
			OrderId:   event.OrderId,
			PaymentId: paymentID,
		}
		msg := message.NewMessage(watermill.NewUUID(), events.MustMarshal(completed))
		uc.pub.Publish(events.TopicPaymentCompleted, msg)
		uc.log.Info("支付成功", zap.String("order_id", event.OrderId), zap.String("payment_id", paymentID))
	} else {
		failed := &eventsv1.PaymentFailed{
			OrderId: event.OrderId,
			Reason:  "余额不足",
		}
		msg := message.NewMessage(watermill.NewUUID(), events.MustMarshal(failed))
		uc.pub.Publish(events.TopicPaymentFailed, msg)
		uc.log.Warn("支付失败", zap.String("order_id", event.OrderId))
	}
}

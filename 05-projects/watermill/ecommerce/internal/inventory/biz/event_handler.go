package biz

import (
	"context"

	eventsv1 "ecommerce/api/proto/events/v1"
	"ecommerce/pkg/events"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

type InventoryEventHandler struct {
	uc  *InventoryUseCase
	pub message.Publisher
	log *zap.Logger
}

func NewInventoryEventHandler(uc *InventoryUseCase, pub message.Publisher, log *zap.Logger) *InventoryEventHandler {
	return &InventoryEventHandler{uc: uc, pub: pub, log: log}
}

func (h *InventoryEventHandler) HandleOrderCreated(msg *message.Message) error {
	var event eventsv1.OrderCreated
	if err := proto.Unmarshal(msg.Payload, &event); err != nil {
		return err
	}
	h.log.Info("收到 OrderCreated 事件", zap.String("order_id", event.OrderId))

	for _, item := range event.Items {
		if err := h.uc.Deduct(context.Background(), item.ProductId, item.Quantity); err != nil {
			h.log.Warn("库存不足", zap.String("product_id", item.ProductId), zap.Error(err))
			for _, prev := range event.Items {
				if prev.ProductId == item.ProductId {
					break
				}
				h.uc.Restore(context.Background(), prev.ProductId, prev.Quantity)
			}
			insufficient := &eventsv1.InventoryInsufficient{
				OrderId:   event.OrderId,
				ProductId: item.ProductId,
				Required:  item.Quantity,
			}
			h.pub.Publish(events.TopicInventoryInsufficient,
				message.NewMessage(watermill.NewUUID(), events.MustMarshal(insufficient)))
			return nil
		}
	}

	reserved := &eventsv1.InventoryReserved{OrderId: event.OrderId}
	h.pub.Publish(events.TopicInventoryReserved,
		message.NewMessage(watermill.NewUUID(), events.MustMarshal(reserved)))

	h.log.Info("库存预留成功", zap.String("order_id", event.OrderId))
	return nil
}

func (h *InventoryEventHandler) HandleInventoryRelease(msg *message.Message) error {
	var event eventsv1.InventoryRelease
	if err := proto.Unmarshal(msg.Payload, &event); err != nil {
		return err
	}
	h.log.Info("收到库存释放事件", zap.String("order_id", event.OrderId))
	for _, item := range event.Items {
		h.uc.Restore(context.Background(), item.ProductId, item.Quantity)
	}
	return nil
}

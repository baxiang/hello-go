package service

import (
	"encoding/json"
	"net/http"

	eventsv1 "ecommerce/api/proto/events/v1"
	"ecommerce/internal/order/biz"
	"ecommerce/pkg/events"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"go.uber.org/zap"
)

type OrderService struct {
	pub message.Publisher
	uc  *biz.OrderUseCase
	log *zap.Logger
}

func NewOrderService(pub message.Publisher, uc *biz.OrderUseCase, log *zap.Logger) *OrderService {
	return &OrderService{pub: pub, uc: uc, log: log}
}

type CreateOrderReq struct {
	OrderID string          `json:"order_id"`
	UserID  string          `json:"user_id"`
	Items   []biz.OrderItem `json:"items"`
	Total   float64         `json:"total"`
}

// HandleCreateOrder POST /orders — 创建订单并发布 OrderCreated 事件
func (s *OrderService) HandleCreateOrder(w http.ResponseWriter, r *http.Request) {
	var req CreateOrderReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "无效的请求体", http.StatusBadRequest)
		return
	}

	order := &biz.Order{
		OrderID: req.OrderID,
		UserID:  req.UserID,
		Items:   req.Items,
		Total:   req.Total,
	}
	if err := s.uc.Create(r.Context(), order); err != nil {
		s.log.Error("创建订单失败", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 构造并发布 OrderCreated 事件
	items := make([]*eventsv1.OrderItem, len(req.Items))
	for i, item := range req.Items {
		items[i] = &eventsv1.OrderItem{
			ProductId: item.ProductID,
			Quantity:  item.Quantity,
			Price:     item.Price,
		}
	}

	event := &eventsv1.OrderCreated{
		OrderId: req.OrderID,
		UserId:  req.UserID,
		Items:   items,
		Total:   req.Total,
	}
	payload := events.MustMarshal(event)
	msg := message.NewMessage(watermill.NewUUID(), payload)
	s.pub.Publish(events.TopicOrderCreated, msg)

	s.log.Info("已发布 OrderCreated 事件", zap.String("order_id", req.OrderID))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"order_id": req.OrderID, "status": "pending"})
}

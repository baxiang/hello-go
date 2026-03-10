package service

import (
	"context"

	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	v1 "kratos/api/order/v1"
	"kratos/internal/biz"
)

// OrderService 订单服务
type OrderService struct {
	v1.UnimplementedOrderServiceServer

	uc *biz.OrderUseCase
}

// NewOrderService 创建订单服务
func NewOrderService(uc *biz.OrderUseCase) *OrderService {
	return &OrderService{
		uc: uc,
	}
}

// CreateOrder 创建订单
func (s *OrderService) CreateOrder(ctx context.Context, req *v1.CreateOrderRequest) (*v1.Order, error) {
	items := make([]*biz.OrderItem, len(req.Items))
	for i, item := range req.Items {
		items[i] = &biz.OrderItem{
			ProductID:   item.ProductId,
			ProductName: item.ProductName,
			Price:       item.Price,
			Quantity:    item.Quantity,
		}
	}

	order, err := s.uc.Create(ctx, req.UserId, items, req.Remark)
	if err != nil {
		return nil, err
	}

	return s.toProto(order), nil
}

// GetOrder 获取订单
func (s *OrderService) GetOrder(ctx context.Context, req *v1.GetOrderRequest) (*v1.Order, error) {
	order, err := s.uc.Get(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	return s.toProto(order), nil
}

// ListOrder 订单列表
func (s *OrderService) ListOrder(ctx context.Context, req *v1.ListOrderRequest) (*v1.ListOrderReply, error) {
	orders, total, err := s.uc.List(ctx, int(req.Page), int(req.PageSize), req.UserId, req.Status)
	if err != nil {
		return nil, err
	}

	protoOrders := make([]*v1.Order, len(orders))
	for i, order := range orders {
		protoOrders[i] = s.toProto(order)
	}

	return &v1.ListOrderReply{
		Orders:   protoOrders,
		Total:    int32(total),
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

// CancelOrder 取消订单
func (s *OrderService) CancelOrder(ctx context.Context, req *v1.CancelOrderRequest) (*v1.Order, error) {
	order, err := s.uc.Cancel(ctx, req.Id, req.Reason)
	if err != nil {
		return nil, err
	}

	return s.toProto(order), nil
}

// PayOrder 支付订单
func (s *OrderService) PayOrder(ctx context.Context, req *v1.PayOrderRequest) (*v1.Order, error) {
	order, err := s.uc.Pay(ctx, req.Id, req.PaymentMethod)
	if err != nil {
		return nil, err
	}

	return s.toProto(order), nil
}

// toProto 转换为 Protobuf 消息
func (s *OrderService) toProto(order *biz.Order) *v1.Order {
	items := make([]*v1.OrderItem, len(order.Items))
	for i, item := range order.Items {
		items[i] = &v1.OrderItem{
			Id:          item.ID,
			OrderId:     item.OrderID,
			ProductId:   item.ProductID,
			ProductName: item.ProductName,
			Price:       item.Price,
			Quantity:    item.Quantity,
			Subtotal:    item.Subtotal,
		}
	}

	return &v1.Order{
		Id:          order.ID,
		OrderNo:     order.OrderNo,
		UserId:      order.UserID,
		TotalAmount: order.TotalAmount,
		Status:      string(order.Status),
		PaymentId:   order.PaymentID,
		Remark:      order.Remark,
		Items:       items,
		CreatedAt:   timestamppb.New(order.CreatedAt),
		UpdatedAt:   timestamppb.New(order.UpdatedAt),
	}
}
// Package service 提供订单 gRPC 服务实现
package service

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	v1 "services/api/order/v1"
	"services/order-service/internal/biz"
)

// OrderService 订单服务
type OrderService struct {
	v1.UnimplementedOrderServiceServer
	uc *biz.OrderUseCase
}

// NewOrderService 创建订单服务
func NewOrderService(uc *biz.OrderUseCase) *OrderService {
	return &OrderService{uc: uc}
}

// CreateOrder 创建订单
func (s *OrderService) CreateOrder(ctx context.Context, req *v1.CreateOrderRequest) (*v1.Order, error) {
	items := make([]*biz.OrderItem, len(req.GetItems()))
	for i, item := range req.GetItems() {
		items[i] = &biz.OrderItem{
			ProductID:   item.GetProductId(),
			ProductName: item.GetProductName(),
			Price:       item.GetPrice(),
			Quantity:    item.GetQuantity(),
		}
	}
	order, err := s.uc.Create(ctx, req.GetUserId(), items, req.GetRemark())
	if err != nil {
		return nil, toGRPCError(err)
	}
	return toProto(order), nil
}

// GetOrder 获取订单
func (s *OrderService) GetOrder(ctx context.Context, req *v1.GetOrderRequest) (*v1.Order, error) {
	order, err := s.uc.Get(ctx, req.GetId())
	if err != nil {
		return nil, toGRPCError(err)
	}
	return toProto(order), nil
}

// ListOrder 订单列表
func (s *OrderService) ListOrder(ctx context.Context, req *v1.ListOrderRequest) (*v1.ListOrderReply, error) {
	orders, total, err := s.uc.List(ctx, int(req.GetPage()), int(req.GetPageSize()), req.GetUserId(), req.GetStatus())
	if err != nil {
		return nil, toGRPCError(err)
	}
	out := make([]*v1.Order, len(orders))
	for i, o := range orders {
		out[i] = toProto(o)
	}
	return &v1.ListOrderReply{
		Orders:   out,
		Total:    int32(total),
		Page:     req.GetPage(),
		PageSize: req.GetPageSize(),
	}, nil
}

// CancelOrder 取消订单
func (s *OrderService) CancelOrder(ctx context.Context, req *v1.CancelOrderRequest) (*v1.Order, error) {
	order, err := s.uc.Cancel(ctx, req.GetId(), req.GetReason())
	if err != nil {
		return nil, toGRPCError(err)
	}
	return toProto(order), nil
}

// PayOrder 支付订单
func (s *OrderService) PayOrder(ctx context.Context, req *v1.PayOrderRequest) (*v1.Order, error) {
	order, err := s.uc.Pay(ctx, req.GetId(), req.GetPaymentMethod())
	if err != nil {
		return nil, toGRPCError(err)
	}
	return toProto(order), nil
}

// UpdatePaymentID 更新订单的支付 ID
func (s *OrderService) UpdatePaymentID(ctx context.Context, req *v1.UpdatePaymentIDRequest) (*v1.Order, error) {
	order, err := s.uc.UpdatePaymentID(ctx, req.GetOrderId(), req.GetPaymentId())
	if err != nil {
		return nil, toGRPCError(err)
	}
	return toProto(order), nil
}

func toProto(o *biz.Order) *v1.Order {
	items := make([]*v1.OrderItem, len(o.Items))
	for i, item := range o.Items {
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
		Id:          o.ID,
		OrderNo:     o.OrderNo,
		UserId:      o.UserID,
		TotalAmount: o.TotalAmount,
		Status:      string(o.Status),
		PaymentId:   o.PaymentID,
		Remark:      o.Remark,
		Items:       items,
		CreatedAt:   timestamppb.New(o.CreatedAt),
		UpdatedAt:   timestamppb.New(o.UpdatedAt),
	}
}

func toGRPCError(err error) error {
	switch {
	case errors.Is(err, biz.ErrOrderNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, biz.ErrOrderStatus):
		return status.Error(codes.FailedPrecondition, err.Error())
	default:
		return status.Error(codes.Internal, err.Error())
	}
}

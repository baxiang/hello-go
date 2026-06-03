// Package service 提供支付 gRPC 服务实现
package service

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	v1 "services/api/payment/v1"
	"services/payment-service/internal/biz"
)

// PaymentService 支付服务
type PaymentService struct {
	v1.UnimplementedPaymentServiceServer
	uc *biz.PaymentUseCase
}

// NewPaymentService 创建支付服务
func NewPaymentService(uc *biz.PaymentUseCase) *PaymentService {
	return &PaymentService{uc: uc}
}

// CreatePayment 创建支付
func (s *PaymentService) CreatePayment(ctx context.Context, req *v1.CreatePaymentRequest) (*v1.Payment, error) {
	p, err := s.uc.Create(ctx, req.GetOrderId(), req.GetOrderNo(), req.GetAmount(), req.GetMethod())
	if err != nil {
		return nil, toGRPCError(err)
	}
	return toProto(p), nil
}

// GetPayment 获取支付
func (s *PaymentService) GetPayment(ctx context.Context, req *v1.GetPaymentRequest) (*v1.Payment, error) {
	p, err := s.uc.Get(ctx, req.GetId())
	if err != nil {
		return nil, toGRPCError(err)
	}
	return toProto(p), nil
}

// PaymentCallback 支付回调
func (s *PaymentService) PaymentCallback(ctx context.Context, req *v1.PaymentCallbackRequest) (*v1.Payment, error) {
	if err := s.uc.ProcessCallback(ctx, req.GetPaymentNo(), req.GetTransactionId(), req.GetStatus()); err != nil {
		return nil, toGRPCError(err)
	}
	p, err := s.uc.GetByPaymentNo(ctx, req.GetPaymentNo())
	if err != nil {
		return nil, toGRPCError(err)
	}
	return toProto(p), nil
}

func toProto(p *biz.Payment) *v1.Payment {
	out := &v1.Payment{
		Id:            p.ID,
		PaymentNo:     p.PaymentNo,
		OrderId:       p.OrderID,
		OrderNo:       p.OrderNo,
		Amount:        p.Amount,
		Method:        p.Method,
		Status:        string(p.Status),
		TransactionId: p.TransactionID,
		CreatedAt:     timestamppb.New(p.CreatedAt),
		UpdatedAt:     timestamppb.New(p.UpdatedAt),
	}
	if p.PaidAt != nil {
		out.PaidAt = timestamppb.New(*p.PaidAt)
	}
	return out
}

func toGRPCError(err error) error {
	switch {
	case errors.Is(err, biz.ErrPaymentNotFound):
		return status.Error(codes.NotFound, err.Error())
	default:
		return status.Error(codes.Internal, err.Error())
	}
}

package service

import (
	"context"

	"google.golang.org/protobuf/types/known/timestamppb"

	v1 "kratos/api/payment/v1"
	"kratos/internal/biz"
)

// PaymentService 支付服务
type PaymentService struct {
	v1.UnimplementedPaymentServiceServer

	uc *biz.PaymentUseCase
}

// NewPaymentService 创建支付服务
func NewPaymentService(uc *biz.PaymentUseCase) *PaymentService {
	return &PaymentService{
		uc: uc,
	}
}

// CreatePayment 创建支付
func (s *PaymentService) CreatePayment(ctx context.Context, req *v1.CreatePaymentRequest) (*v1.Payment, error) {
	payment, err := s.uc.Create(ctx, req.OrderId, req.OrderNo, req.Amount, req.Method)
	if err != nil {
		return nil, err
	}

	return s.toProto(payment), nil
}

// GetPayment 获取支付
func (s *PaymentService) GetPayment(ctx context.Context, req *v1.GetPaymentRequest) (*v1.Payment, error) {
	payment, err := s.uc.Get(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	return s.toProto(payment), nil
}

// PaymentCallback 支付回调
func (s *PaymentService) PaymentCallback(ctx context.Context, req *v1.PaymentCallbackRequest) (*v1.Payment, error) {
	err := s.uc.ProcessCallback(ctx, req.PaymentNo, req.TransactionId, req.Status)
	if err != nil {
		return nil, err
	}

	payment, err := s.uc.GetByPaymentNo(ctx, req.PaymentNo)
	if err != nil {
		return nil, err
	}

	return s.toProto(payment), nil
}

// toProto 转换为 Protobuf 消息
func (s *PaymentService) toProto(payment *biz.Payment) *v1.Payment {
	proto := &v1.Payment{
		Id:            payment.ID,
		PaymentNo:     payment.PaymentNo,
		OrderId:       payment.OrderID,
		OrderNo:       payment.OrderNo,
		Amount:        payment.Amount,
		Method:        payment.Method,
		Status:        string(payment.Status),
		TransactionId: payment.TransactionID,
		CreatedAt:     timestamppb.New(payment.CreatedAt),
		UpdatedAt:     timestamppb.New(payment.UpdatedAt),
	}

	if payment.PaidAt != nil {
		proto.PaidAt = timestamppb.New(*payment.PaidAt)
	}

	return proto
}
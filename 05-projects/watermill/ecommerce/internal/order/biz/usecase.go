package biz

import (
	"context"
	"errors"
	"fmt"

	"go.uber.org/zap"
)

var (
	ErrOrderNotFound = errors.New("订单不存在")
)

type Order struct {
	OrderID string
	UserID  string
	Total   float64
	Status  string
	Items   []OrderItem
}

type OrderItem struct {
	ProductID string
	Quantity  int32
	Price     float64
}

type OrderRepo interface {
	Create(ctx context.Context, o *Order) error
	UpdateStatus(ctx context.Context, orderID string, status string) error
	FindByOrderID(ctx context.Context, orderID string) (*Order, error)
}

type OrderUseCase struct {
	repo OrderRepo
	log  *zap.Logger
}

func NewOrderUseCase(repo OrderRepo, log *zap.Logger) *OrderUseCase {
	return &OrderUseCase{repo: repo, log: log}
}

func (uc *OrderUseCase) Create(ctx context.Context, req *Order) error {
	req.Status = "pending"
	if req.OrderID == "" {
		return fmt.Errorf("订单ID不能为空")
	}
	if err := uc.repo.Create(ctx, req); err != nil {
		uc.log.Error("创建订单失败", zap.Error(err))
		return err
	}
	uc.log.Info("订单创建成功", zap.String("order_id", req.OrderID))
	return nil
}

func (uc *OrderUseCase) Confirm(ctx context.Context, orderID string) error {
	uc.log.Info("确认订单", zap.String("order_id", orderID))
	return uc.repo.UpdateStatus(ctx, orderID, "confirmed")
}

func (uc *OrderUseCase) Cancel(ctx context.Context, orderID string) error {
	uc.log.Info("取消订单", zap.String("order_id", orderID))
	return uc.repo.UpdateStatus(ctx, orderID, "cancelled")
}

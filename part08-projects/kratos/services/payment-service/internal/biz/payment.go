// Package biz 提供支付业务逻辑
package biz

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	orderV1 "services/api/order/v1"
	"services/payment-service/internal/client"
)

// 业务错误
var (
	ErrPaymentNotFound = errors.New("支付不存在")
)

// PaymentStatus 支付状态
type PaymentStatus string

const (
	PaymentStatusPending PaymentStatus = "pending"
	PaymentStatusSuccess PaymentStatus = "success"
	PaymentStatusFailed  PaymentStatus = "failed"
)

// Payment 支付实体
type Payment struct {
	ID            int64         `gorm:"primaryKey;autoIncrement" json:"id"`
	PaymentNo     string        `gorm:"type:varchar(64);uniqueIndex;not null" json:"payment_no"`
	OrderID       int64         `gorm:"index;not null" json:"order_id"`
	OrderNo       string        `gorm:"type:varchar(64)" json:"order_no"`
	Amount        float64       `gorm:"type:decimal(10,2);not null" json:"amount"`
	Method        string        `gorm:"type:varchar(32)" json:"method"`
	Status        PaymentStatus `gorm:"type:varchar(32);default:'pending'" json:"status"`
	TransactionID string        `gorm:"type:varchar(128)" json:"transaction_id"`
	PaidAt        *time.Time    `json:"paid_at"`
	CreatedAt     time.Time     `json:"created_at"`
	UpdatedAt     time.Time     `json:"updated_at"`
}

// TableName 自定义表名
func (Payment) TableName() string {
	return "payments"
}

// PaymentRepo 支付仓库接口
type PaymentRepo interface {
	Create(ctx context.Context, p *Payment) error
	FindByID(ctx context.Context, id int64) (*Payment, error)
	FindByPaymentNo(ctx context.Context, paymentNo string) (*Payment, error)
	Update(ctx context.Context, p *Payment) error
}

// PaymentUseCase 支付用例
type PaymentUseCase struct {
	repo        PaymentRepo
	orderClient *client.OrderClient
	log         *zap.Logger
}

// NewPaymentUseCase 创建支付用例
func NewPaymentUseCase(repo PaymentRepo, orderClient *client.OrderClient, log *zap.Logger) *PaymentUseCase {
	return &PaymentUseCase{repo: repo, orderClient: orderClient, log: log}
}

// Create 创建支付
func (uc *PaymentUseCase) Create(ctx context.Context, orderID int64, orderNo string, amount float64, method string) (*Payment, error) {
	p := &Payment{
		PaymentNo: generatePaymentNo(),
		OrderID:   orderID,
		OrderNo:   orderNo,
		Amount:    amount,
		Method:    method,
		Status:    PaymentStatusPending,
	}
	if err := uc.repo.Create(ctx, p); err != nil {
		return nil, err
	}

	// 模拟异步支付回调
	go func() {
		time.Sleep(2 * time.Second)
		_ = uc.ProcessCallback(context.Background(), p.PaymentNo, "TXN"+p.PaymentNo, string(PaymentStatusSuccess))
	}()

	uc.log.Info("创建支付", zap.Int64("id", p.ID), zap.String("payment_no", p.PaymentNo))
	return p, nil
}

// Get 获取支付
func (uc *PaymentUseCase) Get(ctx context.Context, id int64) (*Payment, error) {
	p, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPaymentNotFound
		}
		return nil, err
	}
	return p, nil
}

// GetByPaymentNo 通过支付号获取
func (uc *PaymentUseCase) GetByPaymentNo(ctx context.Context, paymentNo string) (*Payment, error) {
	p, err := uc.repo.FindByPaymentNo(ctx, paymentNo)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPaymentNotFound
		}
		return nil, err
	}
	return p, nil
}

// ProcessCallback 处理第三方回调
func (uc *PaymentUseCase) ProcessCallback(ctx context.Context, paymentNo, transactionID, statusStr string) error {
	p, err := uc.repo.FindByPaymentNo(ctx, paymentNo)
	if err != nil {
		return err
	}

	p.TransactionID = transactionID
	p.Status = PaymentStatus(statusStr)
	if p.Status == PaymentStatusSuccess {
		now := time.Now()
		p.PaidAt = &now
	}
	if err := uc.repo.Update(ctx, p); err != nil {
		return err
	}

	// 支付成功 → 调用 order-service 更新订单
	if p.Status == PaymentStatusSuccess {
		if _, err := uc.orderClient.UpdatePaymentID(ctx, &orderV1.UpdatePaymentIDRequest{
			OrderId:   p.OrderID,
			PaymentId: p.ID,
		}); err != nil {
			uc.log.Warn("更新订单支付ID失败", zap.Error(err))
		}
		if _, err := uc.orderClient.PayOrder(ctx, &orderV1.PayOrderRequest{
			Id:            p.OrderID,
			PaymentMethod: p.Method,
		}); err != nil {
			uc.log.Warn("更新订单状态失败", zap.Error(err))
		}
	}

	uc.log.Info("支付回调处理", zap.String("payment_no", paymentNo), zap.String("status", statusStr))
	return nil
}

func generatePaymentNo() string {
	return fmt.Sprintf("PAY%s%04d", time.Now().Format("20060102150405"), time.Now().UnixNano()%10000)
}

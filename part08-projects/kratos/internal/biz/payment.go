package biz

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"

	"kratos/internal/repo"
)

var (
	ErrPaymentNotFound = errors.New("支付不存在")
	ErrPaymentStatus   = errors.New("支付状态错误")
)

// PaymentStatus 支付状态
type PaymentStatus string

const (
	PaymentStatusPending  PaymentStatus = "pending"
	PaymentStatusSuccess  PaymentStatus = "success"
	PaymentStatusFailed   PaymentStatus = "failed"
	PaymentStatusRefunded PaymentStatus = "refunded"
)

// Payment 支付业务实体
type Payment struct {
	ID            int64          `json:"id"`
	PaymentNo     string         `json:"payment_no"`
	OrderID       int64          `json:"order_id"`
	OrderNo       string         `json:"order_no"`
	Amount        float64        `json:"amount"`
	Method        string         `json:"method"`
	Status        PaymentStatus  `json:"status"`
	TransactionID string         `json:"transaction_id"`
	PaidAt        *time.Time     `json:"paid_at"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
}

// PaymentUseCase 支付用例
type PaymentUseCase struct {
	paymentRepo  repo.PaymentRepo
	orderUC      *OrderUseCase
	natsClient   interface {
		Publish(ctx context.Context, subject string, data []byte) error
	}
	log *log.Helper
}

// NewPaymentUseCase 创建支付用例
func NewPaymentUseCase(paymentRepo repo.PaymentRepo, orderUC *OrderUseCase, natsClient interface {
	Publish(ctx context.Context, subject string, data []byte) error
}, logger log.Logger) *PaymentUseCase {
	return &PaymentUseCase{
		paymentRepo: paymentRepo,
		orderUC:     orderUC,
		natsClient:  natsClient,
		log:         log.NewHelper(logger),
	}
}

// Create 创建支付
func (uc *PaymentUseCase) Create(ctx context.Context, orderID int64, orderNo string, amount float64, method string) (*Payment, error) {
	// 生成支付单号
	paymentNo := generatePaymentNo()

	payment := &Payment{
		PaymentNo: paymentNo,
		OrderID:   orderID,
		OrderNo:   orderNo,
		Amount:    amount,
		Method:    method,
		Status:    PaymentStatusPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// 创建支付记录
	if err := uc.paymentRepo.Create(ctx, payment); err != nil {
		return nil, err
	}

	// 模拟支付处理 (实际应调用第三方支付接口)
	go func() {
		time.Sleep(2 * time.Second)
		// 这里应该调用第三方支付接口
		// 模拟支付成功
		uc.ProcessCallback(ctx, paymentNo, "TXN"+paymentNo, string(PaymentStatusSuccess))
	}()

	uc.log.Info("创建支付成功", log.Any("payment", payment))
	return payment, nil
}

// Get 获取支付
func (uc *PaymentUseCase) Get(ctx context.Context, id int64) (*Payment, error) {
	payment, err := uc.paymentRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPaymentNotFound
		}
		return nil, err
	}
	return payment, nil
}

// GetByPaymentNo 根据支付单号获取
func (uc *PaymentUseCase) GetByPaymentNo(ctx context.Context, paymentNo string) (*Payment, error) {
	payment, err := uc.paymentRepo.FindByPaymentNo(ctx, paymentNo)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPaymentNotFound
		}
		return nil, err
	}
	return payment, nil
}

// ProcessCallback 处理支付回调
func (uc *PaymentUseCase) ProcessCallback(ctx context.Context, paymentNo, transactionID, status string) error {
	payment, err := uc.paymentRepo.FindByPaymentNo(ctx, paymentNo)
	if err != nil {
		return err
	}

	// 更新支付状态
	payment.TransactionID = transactionID
	payment.Status = PaymentStatus(status)
	if status == string(PaymentStatusSuccess) {
		now := time.Now()
		payment.PaidAt = &now
	}

	if err := uc.paymentRepo.Update(ctx, payment); err != nil {
		return err
	}

	// 如果支付成功，更新订单状态
	if status == string(PaymentStatusSuccess) {
		uc.orderUC.UpdatePaymentID(ctx, payment.OrderID, payment.ID)
		uc.orderUC.Pay(ctx, payment.OrderID, payment.Method)
	}

	// 发布支付回调事件
	uc.publishEvent(ctx, "payment.callback", payment)

	uc.log.Info("处理支付回调成功",
		log.Any("payment_no", paymentNo),
		log.Any("status", status))

	return nil
}

func (uc *PaymentUseCase) publishEvent(ctx context.Context, eventType string, data interface{}) {
	uc.log.Info("发布事件", log.Any("type", eventType), log.Any("data", data))
}

func generatePaymentNo() string {
	return fmt.Sprintf("PAY%s%d", time.Now().Format("20060102150405"), time.Now().UnixNano()%10000)
}
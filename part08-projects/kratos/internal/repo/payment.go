package repo

import (
	"context"

	"gorm.io/gorm"

	"kratos/internal/biz"
)

// PaymentRepo 支付仓库接口
type PaymentRepo interface {
	Create(ctx context.Context, payment *biz.Payment) error
	FindByID(ctx context.Context, id int64) (*biz.Payment, error)
	FindByPaymentNo(ctx context.Context, paymentNo string) (*biz.Payment, error)
	FindByOrderID(ctx context.Context, orderID int64) (*biz.Payment, error)
	Update(ctx context.Context, payment *biz.Payment) error
	Delete(ctx context.Context, id int64) error
}

// paymentRepo 支付仓库实现
type paymentRepo struct {
	data *data.Data
	log  interface{}
}

// NewPaymentRepo 创建支付仓库
func NewPaymentRepo(data *data.Data, logger interface{}) PaymentRepo {
	return &paymentRepo{
		data: data,
		log:  logger,
	}
}

// Create 创建支付
func (r *paymentRepo) Create(ctx context.Context, payment *biz.Payment) error {
	return r.data.DB.WithContext(ctx).Create(payment).Error
}

// FindByID 根据 ID 查找
func (r *paymentRepo) FindByID(ctx context.Context, id int64) (*biz.Payment, error) {
	var payment biz.Payment
	err := r.data.DB.WithContext(ctx).First(&payment, id).Error
	return &payment, err
}

// FindByPaymentNo 根据支付单号查找
func (r *paymentRepo) FindByPaymentNo(ctx context.Context, paymentNo string) (*biz.Payment, error) {
	var payment biz.Payment
	err := r.data.DB.WithContext(ctx).Where("payment_no = ?", paymentNo).First(&payment).Error
	return &payment, err
}

// FindByOrderID 根据订单ID查找
func (r *paymentRepo) FindByOrderID(ctx context.Context, orderID int64) (*biz.Payment, error) {
	var payment biz.Payment
	err := r.data.DB.WithContext(ctx).Where("order_id = ?", orderID).First(&payment).Error
	return &payment, err
}

// Update 更新支付
func (r *paymentRepo) Update(ctx context.Context, payment *biz.Payment) error {
	return r.data.DB.WithContext(ctx).Save(payment).Error
}

// Delete 删除支付
func (r *paymentRepo) Delete(ctx context.Context, id int64) error {
	return r.data.DB.WithContext(ctx).Delete(&biz.Payment{}, id).Error
}

// 确保实现接口
var _ PaymentRepo = (*paymentRepo)(nil)
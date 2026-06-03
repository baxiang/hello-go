// Package data 提供支付数据访问
package data

import (
	"context"

	"go.uber.org/zap"

	"services/payment-service/internal/biz"
	"services/pkg/database"
)

type paymentRepo struct {
	db  *database.DB
	log *zap.Logger
}

// NewPaymentRepo 创建支付仓库
func NewPaymentRepo(db *database.DB, log *zap.Logger) biz.PaymentRepo {
	return &paymentRepo{db: db, log: log}
}

// AutoMigrate 自动迁移支付表
func AutoMigrate(db *database.DB) error {
	return db.AutoMigrate(&biz.Payment{})
}

func (r *paymentRepo) Create(ctx context.Context, p *biz.Payment) error {
	return r.db.WithContext(ctx).Create(p).Error
}

func (r *paymentRepo) FindByID(ctx context.Context, id int64) (*biz.Payment, error) {
	var p biz.Payment
	err := r.db.WithContext(ctx).First(&p, id).Error
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *paymentRepo) FindByPaymentNo(ctx context.Context, paymentNo string) (*biz.Payment, error) {
	var p biz.Payment
	err := r.db.WithContext(ctx).Where("payment_no = ?", paymentNo).First(&p).Error
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *paymentRepo) Update(ctx context.Context, p *biz.Payment) error {
	return r.db.WithContext(ctx).Save(p).Error
}

var _ biz.PaymentRepo = (*paymentRepo)(nil)

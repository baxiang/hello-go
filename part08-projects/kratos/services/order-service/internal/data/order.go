// Package data 提供订单数据访问
package data

import (
	"context"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"services/order-service/internal/biz"
	"services/pkg/database"
)

type orderRepo struct {
	db  *database.DB
	log *zap.Logger
}

// NewOrderRepo 创建订单仓库
func NewOrderRepo(db *database.DB, log *zap.Logger) biz.OrderRepo {
	return &orderRepo{db: db, log: log}
}

// AutoMigrate 自动迁移订单相关表
func AutoMigrate(db *database.DB) error {
	return db.AutoMigrate(&biz.Order{}, &biz.OrderItem{})
}

func (r *orderRepo) Create(ctx context.Context, order *biz.Order) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(order).Error; err != nil {
			return err
		}
		for _, item := range order.Items {
			item.OrderID = order.ID
		}
		if len(order.Items) > 0 {
			if err := tx.Create(order.Items).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *orderRepo) FindByID(ctx context.Context, id int64) (*biz.Order, error) {
	var order biz.Order
	err := r.db.WithContext(ctx).First(&order, id).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *orderRepo) Update(ctx context.Context, order *biz.Order) error {
	return r.db.WithContext(ctx).Save(order).Error
}

func (r *orderRepo) UpdateStatus(ctx context.Context, id int64, status string) error {
	return r.db.WithContext(ctx).Model(&biz.Order{}).Where("id = ?", id).Update("status", status).Error
}

func (r *orderRepo) UpdatePaymentID(ctx context.Context, orderID, paymentID int64) error {
	return r.db.WithContext(ctx).Model(&biz.Order{}).Where("id = ?", orderID).Update("payment_id", paymentID).Error
}

func (r *orderRepo) List(ctx context.Context, page, pageSize int, userID int64, status string) ([]*biz.Order, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	var orders []*biz.Order
	var total int64

	query := r.db.WithContext(ctx).Model(&biz.Order{})
	if userID > 0 {
		query = query.Where("user_id = ?", userID)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&orders).Error; err != nil {
		return nil, 0, err
	}
	return orders, total, nil
}

func (r *orderRepo) FindItemsByOrderID(ctx context.Context, orderID int64) ([]*biz.OrderItem, error) {
	var items []*biz.OrderItem
	err := r.db.WithContext(ctx).Where("order_id = ?", orderID).Find(&items).Error
	return items, err
}

var _ biz.OrderRepo = (*orderRepo)(nil)

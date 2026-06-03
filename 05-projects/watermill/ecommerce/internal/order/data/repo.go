package data

import (
	"context"

	"ecommerce/internal/order/biz"
	"ecommerce/pkg/database"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

var _ biz.OrderRepo = (*orderRepo)(nil)

type orderRepo struct {
	db  *database.DB
	log *zap.Logger
}

func NewOrderRepo(db *database.DB, log *zap.Logger) biz.OrderRepo {
	return &orderRepo{db: db, log: log}
}

func (r *orderRepo) Create(ctx context.Context, o *biz.Order) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		order := &Order{OrderID: o.OrderID, UserID: o.UserID, Total: o.Total, Status: o.Status}
		if err := tx.Create(order).Error; err != nil {
			return err
		}
		for _, item := range o.Items {
			oi := &OrderItem{
				OrderID:   o.OrderID,
				ProductID: item.ProductID,
				Quantity:  item.Quantity,
				Price:     item.Price,
			}
			if err := tx.Create(oi).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *orderRepo) UpdateStatus(ctx context.Context, orderID string, status string) error {
	return r.db.WithContext(ctx).Model(&Order{}).Where("order_id = ?", orderID).Update("status", status).Error
}

func (r *orderRepo) FindByOrderID(ctx context.Context, orderID string) (*biz.Order, error) {
	var order Order
	if err := r.db.WithContext(ctx).Where("order_id = ?", orderID).First(&order).Error; err != nil {
		return nil, err
	}
	return &biz.Order{
		OrderID: order.OrderID, UserID: order.UserID,
		Total: order.Total, Status: order.Status,
	}, nil
}

func AutoMigrate(db *database.DB) error {
	return db.AutoMigrate(&Order{}, &OrderItem{})
}

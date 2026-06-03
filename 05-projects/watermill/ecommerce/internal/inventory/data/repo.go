package data

import (
	"context"
	"errors"

	"ecommerce/internal/inventory/biz"
	"ecommerce/pkg/database"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

var _ biz.InventoryRepo = (*inventoryRepo)(nil)

var ErrInsufficientStock = errors.New("库存不足")

type inventoryRepo struct {
	db  *database.DB
	log *zap.Logger
}

func NewInventoryRepo(db *database.DB, log *zap.Logger) biz.InventoryRepo {
	return &inventoryRepo{db: db, log: log}
}

func (r *inventoryRepo) Deduct(ctx context.Context, productID string, quantity int32) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var inv Inventory
		if err := tx.Where("product_id = ?", productID).First(&inv).Error; err != nil {
			return err
		}
		if inv.Stock < quantity {
			return ErrInsufficientStock
		}
		result := tx.Model(&inv).Where("version = ?", inv.Version).
			Updates(map[string]interface{}{"stock": inv.Stock - quantity, "version": inv.Version + 1})
		if result.RowsAffected == 0 {
			return errors.New("并发冲突，请重试")
		}
		return nil
	})
}

func (r *inventoryRepo) Restore(ctx context.Context, productID string, quantity int32) error {
	return r.db.WithContext(ctx).Model(&Inventory{}).
		Where("product_id = ?", productID).
		Updates(map[string]interface{}{
			"stock":   gorm.Expr("stock + ?", quantity),
			"version": gorm.Expr("version + 1"),
		}).Error
}

func (r *inventoryRepo) FindByProductID(ctx context.Context, productID string) (*biz.Inventory, error) {
	var inv Inventory
	if err := r.db.WithContext(ctx).Where("product_id = ?", productID).First(&inv).Error; err != nil {
		return nil, err
	}
	return &biz.Inventory{ProductID: inv.ProductID, Stock: inv.Stock}, nil
}

func AutoMigrate(db *database.DB) error {
	return db.AutoMigrate(&Inventory{})
}

// Package data 提供商品数据访问
package data

import (
	"context"
	"errors"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"services/pkg/database"
	"services/product-service/internal/biz"
)

type productRepo struct {
	db  *database.DB
	log *zap.Logger
}

// NewProductRepo 创建商品仓库
func NewProductRepo(db *database.DB, log *zap.Logger) biz.ProductRepo {
	return &productRepo{db: db, log: log}
}

// AutoMigrate 自动迁移商品表
func AutoMigrate(db *database.DB) error {
	return db.AutoMigrate(&biz.Product{})
}

func (r *productRepo) Create(ctx context.Context, p *biz.Product) error {
	return r.db.WithContext(ctx).Create(p).Error
}

func (r *productRepo) FindByID(ctx context.Context, id int64) (*biz.Product, error) {
	var p biz.Product
	err := r.db.WithContext(ctx).First(&p, id).Error
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *productRepo) Update(ctx context.Context, p *biz.Product) error {
	return r.db.WithContext(ctx).Save(p).Error
}

func (r *productRepo) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&biz.Product{}, id).Error
}

func (r *productRepo) List(ctx context.Context, page, pageSize int, category, keyword string) ([]*biz.Product, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	var products []*biz.Product
	var total int64

	query := r.db.WithContext(ctx).Model(&biz.Product{})
	if category != "" {
		query = query.Where("category = ?", category)
	}
	if keyword != "" {
		like := "%" + keyword + "%"
		query = query.Where("name LIKE ? OR description LIKE ?", like, like)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Offset((page - 1) * pageSize).Limit(pageSize).Find(&products).Error; err != nil {
		return nil, 0, err
	}
	return products, total, nil
}

// DeductStock 通过事务扣减库存
func (r *productRepo) DeductStock(ctx context.Context, productID int64, quantity int32) (*biz.Product, error) {
	var p biz.Product
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 锁行查询
		if err := tx.Where("id = ?", productID).First(&p).Error; err != nil {
			return err
		}
		if p.Stock < quantity {
			return biz.ErrStockNotEnough
		}
		p.Stock -= quantity
		return tx.Save(&p).Error
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, biz.ErrProductNotFound
		}
		return nil, err
	}
	return &p, nil
}

// RestoreStock 恢复库存
func (r *productRepo) RestoreStock(ctx context.Context, productID int64, quantity int32) (*biz.Product, error) {
	var p biz.Product
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("id = ?", productID).First(&p).Error; err != nil {
			return err
		}
		p.Stock += quantity
		return tx.Save(&p).Error
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, biz.ErrProductNotFound
		}
		return nil, err
	}
	return &p, nil
}

var _ biz.ProductRepo = (*productRepo)(nil)

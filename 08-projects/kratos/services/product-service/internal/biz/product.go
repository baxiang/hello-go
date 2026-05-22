// Package biz 提供商品业务逻辑
package biz

import (
	"context"
	"errors"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// 业务错误
var (
	ErrProductNotFound = errors.New("商品不存在")
	ErrStockNotEnough  = errors.New("库存不足")
)

// Product 商品实体
type Product struct {
	ID          int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	Name        string    `gorm:"type:varchar(128);index;not null" json:"name"`
	Description string    `gorm:"type:text" json:"description"`
	Category    string    `gorm:"type:varchar(64);index" json:"category"`
	Price       float64   `gorm:"type:decimal(10,2);not null" json:"price"`
	Stock       int32     `gorm:"default:0;not null" json:"stock"`
	ImageURL    string    `gorm:"type:varchar(256)" json:"image_url"`
	Status      int32     `gorm:"default:1" json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// TableName 自定义表名
func (Product) TableName() string {
	return "products"
}

// ProductRepo 商品仓库接口
type ProductRepo interface {
	Create(ctx context.Context, p *Product) error
	FindByID(ctx context.Context, id int64) (*Product, error)
	Update(ctx context.Context, p *Product) error
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context, page, pageSize int, category, keyword string) ([]*Product, int64, error)
	DeductStock(ctx context.Context, productID int64, quantity int32) (*Product, error)
	RestoreStock(ctx context.Context, productID int64, quantity int32) (*Product, error)
}

// ProductUseCase 商品用例
type ProductUseCase struct {
	repo ProductRepo
	log  *zap.Logger
}

// NewProductUseCase 创建商品用例
func NewProductUseCase(repo ProductRepo, log *zap.Logger) *ProductUseCase {
	return &ProductUseCase{repo: repo, log: log}
}

// Create 创建商品
func (uc *ProductUseCase) Create(ctx context.Context, p *Product) (*Product, error) {
	if p.Status == 0 {
		p.Status = 1
	}
	if err := uc.repo.Create(ctx, p); err != nil {
		return nil, err
	}
	uc.log.Info("创建商品成功", zap.Int64("id", p.ID), zap.String("name", p.Name))
	return p, nil
}

// Get 获取商品
func (uc *ProductUseCase) Get(ctx context.Context, id int64) (*Product, error) {
	p, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProductNotFound
		}
		return nil, err
	}
	return p, nil
}

// Update 更新商品
func (uc *ProductUseCase) Update(ctx context.Context, p *Product) (*Product, error) {
	existing, err := uc.repo.FindByID(ctx, p.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProductNotFound
		}
		return nil, err
	}
	existing.Name = p.Name
	existing.Description = p.Description
	existing.Category = p.Category
	existing.Price = p.Price
	existing.Stock = p.Stock
	existing.ImageURL = p.ImageURL
	if p.Status != 0 {
		existing.Status = p.Status
	}
	if err := uc.repo.Update(ctx, existing); err != nil {
		return nil, err
	}
	return existing, nil
}

// Delete 删除商品
func (uc *ProductUseCase) Delete(ctx context.Context, id int64) error {
	if _, err := uc.repo.FindByID(ctx, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrProductNotFound
		}
		return err
	}
	return uc.repo.Delete(ctx, id)
}

// List 商品列表
func (uc *ProductUseCase) List(ctx context.Context, page, pageSize int, category, keyword string) ([]*Product, int64, error) {
	return uc.repo.List(ctx, page, pageSize, category, keyword)
}

// DeductStock 扣减库存
func (uc *ProductUseCase) DeductStock(ctx context.Context, productID int64, quantity int32) (*Product, error) {
	p, err := uc.repo.DeductStock(ctx, productID, quantity)
	if err != nil {
		return nil, err
	}
	uc.log.Info("扣减库存",
		zap.Int64("product_id", productID),
		zap.Int32("quantity", quantity),
		zap.Int32("remaining", p.Stock),
	)
	return p, nil
}

// RestoreStock 恢复库存
func (uc *ProductUseCase) RestoreStock(ctx context.Context, productID int64, quantity int32) (*Product, error) {
	return uc.repo.RestoreStock(ctx, productID, quantity)
}

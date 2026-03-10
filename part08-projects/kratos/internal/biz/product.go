package biz

import (
	"context"
	"errors"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"

	"kratos/internal/repo"
)

var (
	ErrProductNotFound = errors.New("商品不存在")
	ErrStockNotEnough  = errors.New("库存不足")
)

// Product 商品业务实体
type Product struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Category    string    `json:"category"`
	Price       float64   `json:"price"`
	Stock       int32     `json:"stock"`
	ImageURL    string    `json:"image_url"`
	Status      int32     `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ProductUseCase 商品用例
type ProductUseCase struct {
	productRepo repo.ProductRepo
	natsClient  interface {
		Publish(ctx context.Context, subject string, data []byte) error
	}
	log *log.Helper
}

// NewProductUseCase 创建商品用例
func NewProductUseCase(productRepo repo.ProductRepo, natsClient interface {
	Publish(ctx context.Context, subject string, data []byte) error
}, logger log.Logger) *ProductUseCase {
	return &ProductUseCase{
		productRepo: productRepo,
		natsClient:  natsClient,
		log:         log.NewHelper(logger),
	}
}

// Create 创建商品
func (uc *ProductUseCase) Create(ctx context.Context, p *Product) (*Product, error) {
	if err := uc.productRepo.Create(ctx, p); err != nil {
		return nil, err
	}

	// 发布商品创建事件
	uc.publishEvent(ctx, "product.created", p)

	uc.log.Info("创建商品成功", log.Any("product", p))
	return p, nil
}

// Get 获取商品
func (uc *ProductUseCase) Get(ctx context.Context, id int64) (*Product, error) {
	product, err := uc.productRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProductNotFound
		}
		return nil, err
	}
	return product, nil
}

// Update 更新商品
func (uc *ProductUseCase) Update(ctx context.Context, p *Product) (*Product, error) {
	existing, err := uc.productRepo.FindByID(ctx, p.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProductNotFound
		}
		return nil, err
	}

	// 更新字段
	existing.Name = p.Name
	existing.Description = p.Description
	existing.Category = p.Category
	existing.Price = p.Price
	existing.Stock = p.Stock
	existing.ImageURL = p.ImageURL
	existing.Status = p.Status

	if err := uc.productRepo.Update(ctx, existing); err != nil {
		return nil, err
	}

	// 发布商品更新事件
	uc.publishEvent(ctx, "product.updated", existing)

	uc.log.Info("更新商品成功", log.Any("product", existing))
	return existing, nil
}

// Delete 删除商品
func (uc *ProductUseCase) Delete(ctx context.Context, id int64) error {
	existing, err := uc.productRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrProductNotFound
		}
		return err
	}

	if err := uc.productRepo.Delete(ctx, existing.ID); err != nil {
		return err
	}

	// 发布商品删除事件
	uc.publishEvent(ctx, "product.deleted", map[string]int64{"id": id})

	uc.log.Info("删除商品成功", log.Any("id", id))
	return nil
}

// List 商品列表
func (uc *ProductUseCase) List(ctx context.Context, page, pageSize int, category, keyword string) ([]*Product, int64, error) {
	return uc.productRepo.List(ctx, page, pageSize, category, keyword)
}

// DeductStock 扣减库存
func (uc *ProductUseCase) DeductStock(ctx context.Context, productID int64, quantity int32) (*Product, error) {
	product, err := uc.productRepo.FindByID(ctx, productID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProductNotFound
		}
		return nil, err
	}

	// 检查库存
	if product.Stock < quantity {
		return nil, ErrStockNotEnough
	}

	// 扣减库存
	product.Stock -= quantity
	if err := uc.productRepo.Update(ctx, product); err != nil {
		return nil, err
	}

	// 发布库存扣减事件
	uc.publishEvent(ctx, "stock.deducted", map[string]interface{}{
		"product_id": productID,
		"quantity":   quantity,
		"remaining":  product.Stock,
	})

	uc.log.Info("扣减库存成功",
		log.Any("product_id", productID),
		log.Any("quantity", quantity),
		log.Any("remaining", product.Stock))

	return product, nil
}

// RestoreStock 恢复库存
func (uc *ProductUseCase) RestoreStock(ctx context.Context, productID int64, quantity int32) (*Product, error) {
	product, err := uc.productRepo.FindByID(ctx, productID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProductNotFound
		}
		return nil, err
	}

	// 恢复库存
	product.Stock += quantity
	if err := uc.productRepo.Update(ctx, product); err != nil {
		return nil, err
	}

	uc.log.Info("恢复库存成功",
		log.Any("product_id", productID),
		log.Any("quantity", quantity),
		log.Any("remaining", product.Stock))

	return product, nil
}

func (uc *ProductUseCase) publishEvent(ctx context.Context, eventType string, data interface{}) {
	// 简化的事件发布
	uc.log.Info("发布事件", log.Any("type", eventType), log.Any("data", data))
}
package repo

import (
	"context"

	"gorm.io/gorm"

	"kratos/internal/biz"
)

// ProductRepo 商品仓库接口
type ProductRepo interface {
	Create(ctx context.Context, product *biz.Product) error
	FindByID(ctx context.Context, id int64) (*biz.Product, error)
	FindByName(ctx context.Context, name string) (*biz.Product, error)
	Update(ctx context.Context, product *biz.Product) error
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context, page, pageSize int, category, keyword string) ([]*biz.Product, int64, error)
}

// productRepo 商品仓库实现
type productRepo struct {
	data *data.Data
	log  interface{}
}

// NewProductRepo 创建商品仓库
func NewProductRepo(data *data.Data, logger interface{}) ProductRepo {
	return &productRepo{
		data: data,
		log:  logger,
	}
}

// Create 创建商品
func (r *productRepo) Create(ctx context.Context, product *biz.Product) error {
	return r.data.DB.WithContext(ctx).Create(product).Error
}

// FindByID 根据 ID 查找
func (r *productRepo) FindByID(ctx context.Context, id int64) (*biz.Product, error) {
	var product biz.Product
	err := r.data.DB.WithContext(ctx).First(&product, id).Error
	return &product, err
}

// FindByName 根据名称查找
func (r *productRepo) FindByName(ctx context.Context, name string) (*biz.Product, error) {
	var product biz.Product
	err := r.data.DB.WithContext(ctx).Where("name = ?", name).First(&product).Error
	return &product, err
}

// Update 更新商品
func (r *productRepo) Update(ctx context.Context, product *biz.Product) error {
	return r.data.DB.WithContext(ctx).Save(product).Error
}

// Delete 删除商品
func (r *productRepo) Delete(ctx context.Context, id int64) error {
	return r.data.DB.WithContext(ctx).Delete(&biz.Product{}, id).Error
}

// List 商品列表
func (r *productRepo) List(ctx context.Context, page, pageSize int, category, keyword string) ([]*biz.Product, int64, error) {
	var products []*biz.Product
	var total int64

	query := r.data.DB.WithContext(ctx).Model(&biz.Product{})

	// 分类筛选
	if category != "" {
		query = query.Where("category = ?", category)
	}

	// 关键词搜索
	if keyword != "" {
		query = query.Where("name LIKE ? OR description LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	// 统计总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Find(&products).Error; err != nil {
		return nil, 0, err
	}

	return products, total, nil
}

// 确保实现接口
var _ ProductRepo = (*productRepo)(nil)
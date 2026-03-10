package repo

import (
	"context"

	"gorm.io/gorm"

	"kratos/internal/biz"
)

// OrderRepo 订单仓库接口
type OrderRepo interface {
	Create(ctx context.Context, order *biz.Order) error
	FindByID(ctx context.Context, id int64) (*biz.Order, error)
	FindByOrderNo(ctx context.Context, orderNo string) (*biz.Order, error)
	Update(ctx context.Context, order *biz.Order) error
	UpdateStatus(ctx context.Context, id int64, status string) error
	UpdatePaymentID(ctx context.Context, orderID, paymentID int64) error
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context, page, pageSize int, userID int64, status string) ([]*biz.Order, int64, error)
	FindItemsByOrderID(ctx context.Context, orderID int64) ([]*biz.OrderItem, error)
	CreateItems(ctx context.Context, items []*biz.OrderItem) error
}

// orderRepo 订单仓库实现
type orderRepo struct {
	data *data.Data
	log  interface{}
}

// NewOrderRepo 创建订单仓库
func NewOrderRepo(data *data.Data, logger interface{}) OrderRepo {
	return &orderRepo{
		data: data,
		log:  logger,
	}
}

// Create 创建订单
func (r *orderRepo) Create(ctx context.Context, order *biz.Order) error {
	return r.data.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 创建订单
		if err := tx.Create(order).Error; err != nil {
			return err
		}

		// 创建订单商品
		for _, item := range order.Items {
			item.OrderID = order.ID
			if err := tx.Create(item).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

// FindByID 根据 ID 查找
func (r *orderRepo) FindByID(ctx context.Context, id int64) (*biz.Order, error) {
	var order biz.Order
	err := r.data.DB.WithContext(ctx).First(&order, id).Error
	return &order, err
}

// FindByOrderNo 根据订单号查找
func (r *orderRepo) FindByOrderNo(ctx context.Context, orderNo string) (*biz.Order, error) {
	var order biz.Order
	err := r.data.DB.WithContext(ctx).Where("order_no = ?", orderNo).First(&order).Error
	return &order, err
}

// Update 更新订单
func (r *orderRepo) Update(ctx context.Context, order *biz.Order) error {
	return r.data.DB.WithContext(ctx).Save(order).Error
}

// UpdateStatus 更新订单状态
func (r *orderRepo) UpdateStatus(ctx context.Context, id int64, status string) error {
	return r.data.DB.WithContext(ctx).Model(&biz.Order{}).Where("id = ?", id).Update("status", status).Error
}

// UpdatePaymentID 更新支付ID
func (r *orderRepo) UpdatePaymentID(ctx context.Context, orderID, paymentID int64) error {
	return r.data.DB.WithContext(ctx).Model(&biz.Order{}).Where("id = ?", orderID).Update("payment_id", paymentID).Error
}

// Delete 删除订单
func (r *orderRepo) Delete(ctx context.Context, id int64) error {
	return r.data.DB.WithContext(ctx).Delete(&biz.Order{}, id).Error
}

// List 订单列表
func (r *orderRepo) List(ctx context.Context, page, pageSize int, userID int64, status string) ([]*biz.Order, int64, error) {
	var orders []*biz.Order
	var total int64

	query := r.data.DB.WithContext(ctx).Model(&biz.Order{})

	// 用户筛选
	if userID > 0 {
		query = query.Where("user_id = ?", userID)
	}

	// 状态筛选
	if status != "" {
		query = query.Where("status = ?", status)
	}

	// 统计总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&orders).Error; err != nil {
		return nil, 0, err
	}

	return orders, total, nil
}

// FindItemsByOrderID 根据订单ID查找订单商品
func (r *orderRepo) FindItemsByOrderID(ctx context.Context, orderID int64) ([]*biz.OrderItem, error) {
	var items []*biz.OrderItem
	err := r.data.DB.WithContext(ctx).Where("order_id = ?", orderID).Find(&items).Error
	return items, err
}

// CreateItems 创建订单商品
func (r *orderRepo) CreateItems(ctx context.Context, items []*biz.OrderItem) error {
	return r.data.DB.WithContext(ctx).Create(items).Error
}

// 确保实现接口
var _ OrderRepo = (*orderRepo)(nil)
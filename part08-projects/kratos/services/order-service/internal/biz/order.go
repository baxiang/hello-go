// Package biz 提供订单业务逻辑
package biz

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	productV1 "services/api/product/v1"
	"services/order-service/internal/client"
)

// 业务错误
var (
	ErrOrderNotFound = errors.New("订单不存在")
	ErrOrderStatus   = errors.New("订单状态错误")
)

// OrderStatus 订单状态
type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "pending"
	OrderStatusPaid      OrderStatus = "paid"
	OrderStatusShipped   OrderStatus = "shipped"
	OrderStatusCompleted OrderStatus = "completed"
	OrderStatusCancelled OrderStatus = "cancelled"
	OrderStatusFailed    OrderStatus = "failed"
)

// Order 订单实体
type Order struct {
	ID          int64        `gorm:"primaryKey;autoIncrement" json:"id"`
	OrderNo     string       `gorm:"type:varchar(64);uniqueIndex;not null" json:"order_no"`
	UserID      int64        `gorm:"index;not null" json:"user_id"`
	TotalAmount float64      `gorm:"type:decimal(10,2);not null" json:"total_amount"`
	Status      OrderStatus  `gorm:"type:varchar(32);default:'pending'" json:"status"`
	PaymentID   int64        `json:"payment_id"`
	Remark      string       `gorm:"type:varchar(256)" json:"remark"`
	Items       []*OrderItem `gorm:"foreignKey:OrderID" json:"items"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

// TableName 自定义表名
func (Order) TableName() string {
	return "orders"
}

// OrderItem 订单商品
type OrderItem struct {
	ID          int64   `gorm:"primaryKey;autoIncrement" json:"id"`
	OrderID     int64   `gorm:"index;not null" json:"order_id"`
	ProductID   int64   `gorm:"not null" json:"product_id"`
	ProductName string  `gorm:"type:varchar(128)" json:"product_name"`
	Price       float64 `gorm:"type:decimal(10,2)" json:"price"`
	Quantity    int32   `gorm:"not null" json:"quantity"`
	Subtotal    float64 `gorm:"type:decimal(10,2)" json:"subtotal"`
}

// TableName 自定义表名
func (OrderItem) TableName() string {
	return "order_items"
}

// OrderRepo 订单仓库接口
type OrderRepo interface {
	Create(ctx context.Context, order *Order) error
	FindByID(ctx context.Context, id int64) (*Order, error)
	Update(ctx context.Context, order *Order) error
	UpdateStatus(ctx context.Context, id int64, status string) error
	UpdatePaymentID(ctx context.Context, orderID, paymentID int64) error
	List(ctx context.Context, page, pageSize int, userID int64, status string) ([]*Order, int64, error)
	FindItemsByOrderID(ctx context.Context, orderID int64) ([]*OrderItem, error)
}

// OrderUseCase 订单用例
type OrderUseCase struct {
	repo          OrderRepo
	productClient *client.ProductClient
	log           *zap.Logger
}

// NewOrderUseCase 创建订单用例
func NewOrderUseCase(repo OrderRepo, productClient *client.ProductClient, log *zap.Logger) *OrderUseCase {
	return &OrderUseCase{
		repo:          repo,
		productClient: productClient,
		log:           log,
	}
}

// Create 创建订单（含库存扣减，跨服务调用 product-service）
func (uc *OrderUseCase) Create(ctx context.Context, userID int64, items []*OrderItem, remark string) (*Order, error) {
	// 计算总价
	var total float64
	for _, item := range items {
		item.Subtotal = item.Price * float64(item.Quantity)
		total += item.Subtotal
	}

	order := &Order{
		OrderNo:     generateOrderNo(),
		UserID:      userID,
		TotalAmount: total,
		Status:      OrderStatusPending,
		Remark:      remark,
		Items:       items,
	}

	// 1. 创建订单（DB 写入）
	if err := uc.repo.Create(ctx, order); err != nil {
		return nil, err
	}

	// 2. 跨服务调用 product-service 扣减库存
	deductedItems := make([]*OrderItem, 0, len(items))
	for _, item := range items {
		resp, err := uc.productClient.DeductStock(ctx, &productV1.DeductStockRequest{
			ProductId: item.ProductID,
			Quantity:  item.Quantity,
		})
		if err != nil || !resp.GetSuccess() {
			uc.log.Warn("库存扣减失败，回滚",
				zap.Int64("product_id", item.ProductID),
				zap.Error(err),
			)
			// 回滚已扣减的
			for _, done := range deductedItems {
				_, _ = uc.productClient.RestoreStock(ctx, &productV1.RestoreStockRequest{
					ProductId: done.ProductID,
					Quantity:  done.Quantity,
				})
			}
			_ = uc.repo.UpdateStatus(ctx, order.ID, string(OrderStatusFailed))
			if err != nil {
				return nil, fmt.Errorf("扣减库存失败: %w", err)
			}
			return nil, errors.New("库存不足: " + resp.GetMessage())
		}
		deductedItems = append(deductedItems, item)
	}

	uc.log.Info("订单创建成功", zap.Int64("id", order.ID), zap.String("order_no", order.OrderNo))
	return order, nil
}

// Get 获取订单
func (uc *OrderUseCase) Get(ctx context.Context, id int64) (*Order, error) {
	order, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOrderNotFound
		}
		return nil, err
	}
	items, err := uc.repo.FindItemsByOrderID(ctx, id)
	if err != nil {
		return nil, err
	}
	order.Items = items
	return order, nil
}

// List 订单列表
func (uc *OrderUseCase) List(ctx context.Context, page, pageSize int, userID int64, status string) ([]*Order, int64, error) {
	orders, total, err := uc.repo.List(ctx, page, pageSize, userID, status)
	if err != nil {
		return nil, 0, err
	}
	for _, o := range orders {
		items, _ := uc.repo.FindItemsByOrderID(ctx, o.ID)
		o.Items = items
	}
	return orders, total, nil
}

// Cancel 取消订单（含库存恢复）
func (uc *OrderUseCase) Cancel(ctx context.Context, id int64, reason string) (*Order, error) {
	order, err := uc.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if order.Status != OrderStatusPending && order.Status != OrderStatusPaid {
		return nil, ErrOrderStatus
	}

	// 恢复库存
	for _, item := range order.Items {
		_, _ = uc.productClient.RestoreStock(ctx, &productV1.RestoreStockRequest{
			ProductId: item.ProductID,
			Quantity:  item.Quantity,
		})
	}

	order.Status = OrderStatusCancelled
	order.Remark = reason
	if err := uc.repo.Update(ctx, order); err != nil {
		return nil, err
	}
	return order, nil
}

// Pay 标记订单已支付
func (uc *OrderUseCase) Pay(ctx context.Context, id int64, paymentMethod string) (*Order, error) {
	order, err := uc.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if order.Status != OrderStatusPending {
		return nil, ErrOrderStatus
	}
	order.Status = OrderStatusPaid
	if err := uc.repo.Update(ctx, order); err != nil {
		return nil, err
	}
	uc.log.Info("订单已支付", zap.Int64("id", id), zap.String("method", paymentMethod))
	return order, nil
}

// UpdatePaymentID 更新支付 ID
func (uc *OrderUseCase) UpdatePaymentID(ctx context.Context, orderID, paymentID int64) (*Order, error) {
	if err := uc.repo.UpdatePaymentID(ctx, orderID, paymentID); err != nil {
		return nil, err
	}
	return uc.Get(ctx, orderID)
}

// generateOrderNo 生成订单号
func generateOrderNo() string {
	return fmt.Sprintf("ORD%s%04d", time.Now().Format("20060102150405"), time.Now().UnixNano()%10000)
}

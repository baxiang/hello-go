package biz

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"

	"kratos/internal/data"
	"kratos/internal/repo"
)

var (
	ErrOrderNotFound = errors.New("订单不存在")
	ErrOrderStatus   = errors.New("订单状态错误")
)

// OrderStatus 订单状态
type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "pending"
	OrderStatusPaid      OrderStatus = "paid"
	OrderStatusProcessing OrderStatus = "processing"
	OrderStatusShipped   OrderStatus = "shipped"
	OrderStatusCompleted OrderStatus = "completed"
	OrderStatusCancelled OrderStatus = "cancelled"
	OrderStatusFailed    OrderStatus = "failed"
)

// Order 订单业务实体
type Order struct {
	ID          int64       `json:"id"`
	OrderNo     string      `json:"order_no"`
	UserID      int64       `json:"user_id"`
	TotalAmount float64     `json:"total_amount"`
	Status      OrderStatus `json:"status"`
	PaymentID   int64       `json:"payment_id"`
	Remark      string      `json:"remark"`
	Items       []*OrderItem `json:"items"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
}

// OrderItem 订单商品
type OrderItem struct {
	ID          int64   `json:"id"`
	OrderID     int64   `json:"order_id"`
	ProductID   int64   `json:"product_id"`
	ProductName string  `json:"product_name"`
	Price       float64 `json:"price"`
	Quantity    int32   `json:"quantity"`
	Subtotal    float64 `json:"subtotal"`
}

// OrderUseCase 订单用例
type OrderUseCase struct {
	orderRepo  repo.OrderRepo
	productUC  *ProductUseCase
	natsClient interface {
		Publish(ctx context.Context, subject string, data []byte) error
	}
	log *log.Helper
}

// NewOrderUseCase 创建订单用例
func NewOrderUseCase(orderRepo repo.OrderRepo, productUC *ProductUseCase, natsClient interface {
	Publish(ctx context.Context, subject string, data []byte) error
}, logger log.Logger) *OrderUseCase {
	return &OrderUseCase{
		orderRepo:  orderRepo,
		productUC:  productUC,
		natsClient: natsClient,
		log:        log.NewHelper(logger),
	}
}

// Create 创建订单
func (uc *OrderUseCase) Create(ctx context.Context, userID int64, items []*OrderItem, remark string) (*Order, error) {
	// 生成订单号
	orderNo := generateOrderNo()

	// 计算总价
	var totalAmount float64
	for _, item := range items {
		item.Subtotal = item.Price * float64(item.Quantity)
		totalAmount += item.Subtotal
	}

	order := &Order{
		OrderNo:     orderNo,
		UserID:      userID,
		TotalAmount: totalAmount,
		Status:      OrderStatusPending,
		Remark:     remark,
		Items:       items,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// 创建订单
	if err := uc.orderRepo.Create(ctx, order); err != nil {
		return nil, err
	}

	// 扣减库存
	for _, item := range items {
		_, err := uc.productUC.DeductStock(ctx, item.ProductID, item.Quantity)
		if err != nil {
			// 库存扣减失败，取消订单
			uc.orderRepo.UpdateStatus(ctx, order.ID, string(OrderStatusFailed))
			return nil, err
		}
	}

	// 发布订单创建事件
	uc.publishEvent(ctx, "order.created", order)

	uc.log.Info("创建订单成功", log.Any("order", order))
	return order, nil
}

// Get 获取订单
func (uc *OrderUseCase) Get(ctx context.Context, id int64) (*Order, error) {
	order, err := uc.orderRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOrderNotFound
		}
		return nil, err
	}

	// 获取订单商品
	items, err := uc.orderRepo.FindItemsByOrderID(ctx, id)
	if err != nil {
		return nil, err
	}
	order.Items = items

	return order, nil
}

// List 订单列表
func (uc *OrderUseCase) List(ctx context.Context, page, pageSize int, userID int64, status string) ([]*Order, int64, error) {
	orders, total, err := uc.orderRepo.List(ctx, page, pageSize, userID, status)
	if err != nil {
		return nil, 0, err
	}

	// 填充订单商品
	for _, order := range orders {
		items, _ := uc.orderRepo.FindItemsByOrderID(ctx, order.ID)
		order.Items = items
	}

	return orders, total, nil
}

// Cancel 取消订单
func (uc *OrderUseCase) Cancel(ctx context.Context, id int64, reason string) (*Order, error) {
	order, err := uc.orderRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOrderNotFound
		}
		return nil, err
	}

	// 检查订单状态
	if order.Status != OrderStatusPending && order.Status != OrderStatusPaid {
		return nil, ErrOrderStatus
	}

	// 恢复库存
	for _, item := range order.Items {
		uc.productUC.RestoreStock(ctx, item.ProductID, item.Quantity)
	}

	// 更新订单状态
	order.Status = OrderStatusCancelled
	order.Remark = reason
	if err := uc.orderRepo.Update(ctx, order); err != nil {
		return nil, err
	}

	// 发布订单取消事件
	uc.publishEvent(ctx, "order.cancelled", order)

	uc.log.Info("取消订单成功", log.Any("order_id", id))
	return order, nil
}

// Pay 支付订单
func (uc *OrderUseCase) Pay(ctx context.Context, id int64, paymentMethod string) (*Order, error) {
	order, err := uc.orderRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOrderNotFound
		}
		return nil, err
	}

	// 检查订单状态
	if order.Status != OrderStatusPending {
		return nil, ErrOrderStatus
	}

	// 更新订单状态
	order.Status = OrderStatusPaid
	if err := uc.orderRepo.Update(ctx, order); err != nil {
		return nil, err
	}

	// 发布订单支付事件
	uc.publishEvent(ctx, "order.paid", order)

	uc.log.Info("订单支付成功", log.Any("order_id", id))
	return order, nil
}

// UpdatePaymentID 更新支付ID
func (uc *OrderUseCase) UpdatePaymentID(ctx context.Context, orderID, paymentID int64) error {
	return uc.orderRepo.UpdatePaymentID(ctx, orderID, paymentID)
}

func (uc *OrderUseCase) publishEvent(ctx context.Context, eventType string, data interface{}) {
	// 简化的事件发布
	uc.log.Info("发布事件", log.Any("type", eventType), log.Any("data", data))
}

func generateOrderNo() string {
	return fmt.Sprintf("ORD%s%d", time.Now().Format("20060102150405"), time.Now().UnixNano()%10000)
}
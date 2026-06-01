package data

import "time"

// Order 订单 GORM 模型
type Order struct {
	ID        int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	OrderID   string    `gorm:"type:varchar(36);uniqueIndex;not null" json:"order_id"`
	UserID    string    `gorm:"type:varchar(36);not null" json:"user_id"`
	Total     float64   `gorm:"type:decimal(10,2);not null" json:"total"`
	Status    string    `gorm:"type:varchar(20);not null;default:pending" json:"status"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (Order) TableName() string { return "orders" }

// OrderItem 订单项 GORM 模型
type OrderItem struct {
	ID        int64   `gorm:"primaryKey;autoIncrement" json:"id"`
	OrderID   string  `gorm:"type:varchar(36);not null;index" json:"order_id"`
	ProductID string  `gorm:"type:varchar(36);not null" json:"product_id"`
	Quantity  int32   `gorm:"not null" json:"quantity"`
	Price     float64 `gorm:"type:decimal(10,2);not null" json:"price"`
}

func (OrderItem) TableName() string { return "order_items" }

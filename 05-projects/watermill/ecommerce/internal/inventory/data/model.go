package data

type Inventory struct {
	ID        int64  `gorm:"primaryKey;autoIncrement"`
	ProductID string `gorm:"type:varchar(36);uniqueIndex;not null"`
	Stock     int32  `gorm:"not null"`
	Version   int32  `gorm:"not null;default:1"`
}

func (Inventory) TableName() string { return "inventory" }

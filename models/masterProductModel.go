package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type MasterProduct struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;index" json:"id"`
	Name        string    `gorm:"type:varchar(100);not null" json:"name"`
	Description string    `gorm:"type:text;not null" json:"description"`
	Price       float64   `gorm:"type:decimal(10,2);not null" json:"price"`
	Image       string    `gorm:"type:text;not null" json:"image"`
	Category    string    `gorm:"type:varchar(100);not null" json:"category"`

	IsActive   bool `gorm:"default:false" json:"is_active"`
	IsBeverage bool `gorm:"default:false" json:"is_beverage"`

	BasketOrders []BasketOrder `gorm:"foreignKey:ProductID" json:"basket_orders"`

	//audit fields
	CreatedAt time.Time      `json:"created_at"`
	CreatedBy string         `gorm:"size:100;not null" json:"created_by"`
	UpdatedAt time.Time      `json:"updated_at"`
	UpdatedBy string         `gorm:"size:100;not null" json:"updated_by"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at"`
	DeletedBy *string        `gorm:"size:100" json:"deleted_by"`
}

// Hook sebelum membuat data
func (mp *MasterProduct) BeforeCreate(tx *gorm.DB) (err error) {
	mp.ID = uuid.New()
	return
}

package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// BasketOrder represents a user's order item in the basket.
type BasketOrder struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	UserIDOrder uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id_order"`
	ProductID   uuid.UUID `gorm:"type:uuid;not null;index" json:"product_id"`
	IceStatus   string    `gorm:"type:varchar(50);not null" json:"ice_status"`
	Size        string    `gorm:"type:varchar(50);not null" json:"size"`
	Quantity    int64     `gorm:"not null" json:"quantity"`        // int64 for large quantities
	Note        *string   `gorm:"type:text" json:"note"`           // Optional customer note
	Price       float64   `gorm:"not null" json:"price"`           // PostgreSQL handles as double precision
	TotalPrice  float64   `gorm:"not null" json:"total_price"`     // PostgreSQL handles as double precision
	IsOrdered   bool      `gorm:"default:false" json:"is_ordered"` // Whether this item has been checked out

	// Optional: Relasi balik (belongs to)
	MasterProduct MasterProduct `gorm:"foreignKey:ProductID" json:"master_product"`
	User          User          `gorm:"foreignKey:UserIDOrder" json:"-"`

	// Audit fields
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	CreatedBy string         `gorm:"type:varchar(100);not null" json:"created_by"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	UpdatedBy string         `gorm:"type:varchar(100);not null" json:"updated_by"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at"`
	DeletedBy *string        `gorm:"type:varchar(100)" json:"deleted_by"`
}

// BeforeCreate generates a new UUID for the BasketOrder before creation.
func (b *BasketOrder) BeforeCreate(tx *gorm.DB) error {
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}
	return nil
}

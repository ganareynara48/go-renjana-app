package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ProductCategory struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Name        string    `gorm:"size:100" json:"name"` // contoh: "Rumah", "Kantor"
	Description *string   `gorm:"type:text" json:"description"`

	//audit fields
	CreatedAt time.Time      `json:"created_at"`
	CreatedBy string         `gorm:"size:100;not null" json:"created_by"`
	UpdatedAt time.Time      `json:"updated_at"`
	UpdatedBy string         `gorm:"size:100;not null" json:"updated_by"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at"`
	DeletedBy *string        `gorm:"size:100" json:"deleted_by"`
}

// Hook sebelum membuat data
func (pc *ProductCategory) BeforeCreate(tx *gorm.DB) (err error) {
	pc.ID = uuid.New()
	return
}

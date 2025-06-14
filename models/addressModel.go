package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Address struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	UserID      uuid.UUID `gorm:"type:uuid;not null" json:"user_id"`
	Label       string    `gorm:"size:100" json:"label"` // contoh: "Rumah", "Kantor"
	ReceiveName string    `gorm:"size:100" json:"receive_name"`
	PhoneNumber string    `gorm:"size:20" json:"phone_number"`
	FullAddress string    `gorm:"type:text" json:"full_address"`
	PostCode    string    `gorm:"size:10" json:"post_code"`

	ProvinceName string `gorm:"size:100" json:"province_name"`
	CityName     string `gorm:"size:100" json:"city_name"`
	DistrictName string `gorm:"size:100" json:"district_name"`
	VillageName  string `gorm:"size:100" json:"village_name"`

	IsDefault bool `gorm:"default:false" json:"is_default"`

	// Optional: Relasi balik (belongs to)
	User User `gorm:"foreignKey:UserID" json:"-"`

	//audit fields
	CreatedAt time.Time      `json:"created_at"`
	CreatedBy string         `gorm:"size:100;not null" json:"created_by"`
	UpdatedAt time.Time      `json:"updated_at"`
	UpdatedBy string         `gorm:"size:100;not null" json:"updated_by"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at"`
	DeletedBy *string        `gorm:"size:100" json:"deleted_by"`
}

// Hook sebelum membuat data
func (a *Address) BeforeCreate(tx *gorm.DB) (err error) {
	a.ID = uuid.New()
	return
}

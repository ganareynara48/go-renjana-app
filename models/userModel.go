package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID             uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"` // UUID sebagai primary key
	Fullname       string    `gorm:"size:100;not null" json:"fulname"`
	Username       string    `gorm:"size:100;uniqueIndex;not null" json:"username"`
	Email          string    `gorm:"size:100;uniqueIndex;not null" json:"email"`
	Password       string    `gorm:"not null" json:"-"`
	ProfilePicture *string   `gorm:"type:text" json:"profile_picture"`

	Token *string `gorm:"type:text"`

	// Relations
	Addresses []Address `gorm:"foreignKey:UserID" json:"addresses"`

	// Audit fields
	CreatedAt time.Time      `json:"created_at"`
	CreatedBy string         `gorm:"size:100;not null" json:"created_by"`
	UpdatedAt time.Time      `json:"updated_at"`
	UpdatedBy string         `gorm:"size:100;not null" json:"updated_by"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at"`
	DeletedBy *string        `gorm:"size:100" json:"deleted_by"`
}

// Hook sebelum membuat data
func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	u.ID = uuid.New()
	return
}

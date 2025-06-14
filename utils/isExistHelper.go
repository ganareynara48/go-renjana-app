package utils

import (
	"errors"

	"renjana-app/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Cek apakah username sudah digunakan
func IsUsernameTaken(tx *gorm.DB, username string) (bool, error) {
	var user models.User
	err := tx.Where("username = ?", username).First(&user).Error
	if err == nil {
		return true, nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	}
	return false, err
}

// Cek apakah email sudah digunakan
func IsEmailTaken(tx *gorm.DB, email string) (bool, error) {
	var user models.User
	err := tx.Where("email = ?", email).First(&user).Error
	if err == nil {
		return true, nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	}
	return false, err
}

// Cek apakah label sebelumnya sudah digunakan
func IsLabelExist(tx *gorm.DB, label string, userID uuid.UUID) (bool, error) {
	var address models.Address
	err := tx.Where(map[string]interface{}{
		"label":   label,
		"user_id": userID,
	}).First(&address).Error

	if err == nil {
		return true, nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	}
	return false, err
}

func IsUserExist(tx *gorm.DB, userID string) (bool, error) {
	var user models.User
	err := tx.Where(map[string]interface{}{
		"id": userID,
	}).First(&user).Error

	if err == nil {
		return true, nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	}
	return false, err
}

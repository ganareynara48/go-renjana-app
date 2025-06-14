package controllers

import (
	"errors"
	"fmt"
	"net/http"
	"renjana-app/database"
	"renjana-app/models"
	"renjana-app/utils"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func InsertAddress(c *gin.Context) {
	tokenStr := c.GetHeader("Authorization")
	_, username, err := utils.VerifyJWT(tokenStr)

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	user_id := utils.Request(c, "user_id")
	label := utils.Request(c, "label")
	receive_name := utils.Request(c, "receive_name")
	phone_number := utils.Request(c, "phone_number")
	full_address := utils.Request(c, "full_address")
	post_code := utils.Request(c, "post_code")
	province_name := utils.Request(c, "province_name")
	city_name := utils.Request(c, "city_name")
	district_name := utils.Request(c, "district_name")
	village_name := utils.Request(c, "village_name")
	is_default := utils.Request(c, "is_default")

	if user_id == nil || user_id == "" {
		c.JSON(400, gin.H{"error": "User not found!"})
		return
	}

	if label == nil || label == "" {
		c.JSON(400, gin.H{"error": "Label can't empty!"})
		return
	}

	if receive_name == nil || receive_name == "" {
		c.JSON(400, gin.H{"error": "Receive name can't empty!"})
		return
	}

	if phone_number == nil || phone_number == "" {
		c.JSON(400, gin.H{"error": "Phone number can't empty!"})
		return
	}

	if full_address == nil || full_address == "" {
		c.JSON(400, gin.H{"error": "Full address can't empty!"})
		return
	}

	if post_code == nil || post_code == "" {
		c.JSON(400, gin.H{"error": "Post code can't empty!"})
		return
	}

	if province_name == nil || province_name == "" {
		c.JSON(400, gin.H{"error": "Province name can't empty!"})
		return
	}

	if city_name == nil || city_name == "" {
		c.JSON(400, gin.H{"error": "City name can't empty!"})
		return
	}

	if district_name == nil || district_name == "" {
		c.JSON(400, gin.H{"error": "District name can't empty!"})
		return
	}

	if village_name == nil || village_name == "" {
		c.JSON(400, gin.H{"error": "Village name can't empty!"})
		return
	}

	if is_default == nil || is_default == "" {
		is_default = false
	}

	userIDStr, ok := user_id.(string)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id must be a string"})
		return
	}

	user_uuid, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid UUID format for user_id"})
		return
	}

	address := models.Address{
		UserID:       user_uuid,
		Label:        strings.ToLower(label.(string)),
		ReceiveName:  strings.ToLower(receive_name.(string)),
		PhoneNumber:  phone_number.(string),
		FullAddress:  strings.ToLower(full_address.(string)),
		PostCode:     strings.ToLower(post_code.(string)),
		ProvinceName: strings.ToLower(province_name.(string)),
		CityName:     strings.ToLower(city_name.(string)),
		DistrictName: strings.ToLower(district_name.(string)),
		VillageName:  strings.ToLower(village_name.(string)),
		IsDefault:    is_default.(bool),
		CreatedBy:    username,
		UpdatedBy:    username,
	}

	tx := database.DB.Begin()
	if tx.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Transaction DB failed!"})
		return
	}

	// Cek user
	if taken, err := utils.IsUserExist(tx, address.UserID.String()); err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Crash when checking user!"})
		return
	} else if !taken {
		tx.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{"error": "User not found!"})
		return
	}

	// Cek label
	if taken, err := utils.IsLabelExist(tx, address.Label, address.UserID); err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Crash when checking label!"})
		return
	} else if taken {
		tx.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{"error": "Label already available!"})
		return
	}

	if err := tx.Create(&address).Error; err != nil {
		tx.Rollback()
		c.JSON(500, gin.H{"error": "Address can't registered!"})
		return
	}

	// Jika address baru adalah default, maka ubah address lama milik user itu menjadi non-default
	if address.IsDefault {
		err = tx.Model(&models.Address{}).
			Where("user_id = ? AND id != ?", address.UserID, address.ID).
			Updates(map[string]any{
				"is_default": false,
			}).Error

		if err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to unset other default addresses"})
			return
		}
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		c.JSON(500, gin.H{"error": "Crash when registering user!"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Address registered!",
		"data": gin.H{
			"user_id":       address.UserID,
			"label":         address.Label,
			"receive_name":  address.ReceiveName,
			"phone_number":  address.PhoneNumber,
			"full_address":  address.FullAddress,
			"post_code":     address.PostCode,
			"province_name": address.ProvinceName,
			"city_name":     address.CityName,
			"district_name": address.DistrictName,
			"village_name":  address.VillageName,
			"is_default":    address.IsDefault,
		},
	})
}

func GetAllAddressByUserID(c *gin.Context) {
	user_id := utils.Request(c, "user_id")

	var address []models.Address

	tx := database.DB.Begin()
	if tx.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
		return
	}

	if err := tx.Where("user_id = ?", user_id).Find(&address).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Success catch all users!",
		"data":    address,
	})
}

func GetActiveAddress(c *gin.Context) {
	user_id := utils.Request(c, "user_id")

	var address models.Address

	tx := database.DB.Begin()
	if tx.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
		return
	}

	if err := tx.Where("user_id = ? AND is_default is true", user_id).Find(&address).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Success catch all users!",
		"data":    address,
	})
}

func UpdateAddress(c *gin.Context) {
	tokenStr := c.GetHeader("Authorization")
	_, username, err := utils.VerifyJWT(tokenStr)

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	address_id := utils.Request(c, "address_id")
	user_id := utils.Request(c, "user_id")
	label := utils.Request(c, "label")
	receive_name := utils.Request(c, "receive_name")
	phone_number := utils.Request(c, "phone_number")
	full_address := utils.Request(c, "full_address")
	post_code := utils.Request(c, "post_code")
	province_name := utils.Request(c, "province_name")
	city_name := utils.Request(c, "city_name")
	district_name := utils.Request(c, "district_name")
	village_name := utils.Request(c, "village_name")
	is_default := utils.Request(c, "is_default")

	if user_id == nil || user_id == "" {
		c.JSON(400, gin.H{"error": "User not found!"})
		return
	}

	if address_id == nil || address_id == "" {
		c.JSON(400, gin.H{"error": "Address data not found!"})
		return
	}

	tx := database.DB.Begin()
	if tx.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Transaction DB failed!"})
		return
	}

	var address models.Address

	res := tx.Where(map[string]any{
		"id":         address_id,
		"user_id":    user_id,
		"deleted_at": nil,
		"deleted_by": nil,
	}).First(&address)

	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Address data not found!"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": res.Error.Error()})
		}
		tx.Rollback()
		return
	}

	new_label := utils.GetStringOrDefault(label, address.Label)
	new_receive_name := utils.GetStringOrDefault(receive_name, address.ReceiveName)
	new_phone_number := utils.GetStringOrDefault(phone_number, address.PhoneNumber)
	new_full_address := utils.GetStringOrDefault(full_address, address.FullAddress)
	new_post_code := utils.GetStringOrDefault(post_code, address.PostCode)
	new_province_name := utils.GetStringOrDefault(province_name, address.ProvinceName)
	new_city_name := utils.GetStringOrDefault(city_name, address.CityName)
	new_district_name := utils.GetStringOrDefault(district_name, address.DistrictName)
	new_village_name := utils.GetStringOrDefault(village_name, address.VillageName)

	new_is_default := false
	if is_default == nil || is_default == "" {
		new_is_default = false
	} else {
		new_is_default = is_default.(bool)
	}

	updateAddress := map[string]any{
		"id":            address_id,
		"user_id":       user_id,
		"label":         new_label,
		"receive_name":  new_receive_name,
		"phone_number":  new_phone_number,
		"full_address":  new_full_address,
		"post_code":     new_post_code,
		"province_name": new_province_name,
		"city_name":     new_city_name,
		"district_name": new_district_name,
		"village_name":  new_village_name,
		"is_default":    new_is_default,
		"updated_by":    username,
	}

	if err := tx.Model(&address).Updates(updateAddress).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update address"})
		return
	}

	// Jika address baru adalah default, maka ubah address lama milik user itu menjadi non-default
	if address.IsDefault {
		err = tx.Model(&models.Address{}).
			Where("user_id = ? AND id != ?", address.UserID, address.ID).
			Updates(map[string]any{
				"is_default": false,
			}).Error

		if err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to unset other default addresses"})
			return
		}
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("Address with receive name %s updated!", new_receive_name),
		"data":    updateAddress,
	})
}

func DeleteAddress(c *gin.Context) {
	tokenStr := c.GetHeader("Authorization")
	_, username, err := utils.VerifyJWT(tokenStr)

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	address_id := utils.Request(c, "address_id")
	user_id := utils.Request(c, "user_id")

	if user_id == nil || user_id == "" {
		c.JSON(400, gin.H{"error": "User not found!"})
		return
	}

	if address_id == nil || address_id == "" {
		c.JSON(400, gin.H{"error": "Address data not found!"})
		return
	}

	tx := database.DB.Begin()
	if tx.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Transaction DB failed!"})
		return
	}

	var address models.Address

	// Ambil data address yang belum dihapus
	res := tx.Where("id = ? AND user_id = ? AND deleted_at IS NULL AND deleted_by IS NULL",
		address_id, user_id).First(&address)

	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Address data not found!"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": res.Error.Error()})
		}
		tx.Rollback()
		return
	}

	// Soft delete
	if err := tx.Model(&address).Updates(map[string]any{
		"is_default": false,
		"deleted_at": time.Now(),
		"deleted_by": username,
	}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete address"})
		return
	}

	// Jika address yang dihapus adalah default, maka cari penggantinya
	if address.IsDefault {
		var replacement models.Address
		err := tx.Where("user_id = ? AND id != ? AND deleted_at IS NULL AND deleted_by IS NULL",
			address.UserID, address.ID).First(&replacement).Error

		if err == nil {
			// Set satu address lain sebagai default
			if err := tx.Model(&replacement).Update("is_default", true).Error; err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to set replacement default address"})
				return
			}
		}
		// Kalau tidak ada address lain, tidak perlu error
	}

	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Address deleted!"})
}

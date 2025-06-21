package controllers

import (
	"errors"
	"net/http"
	"renjana-app/database"
	"renjana-app/models"
	"renjana-app/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func InsertBasket(c *gin.Context) {
	tokenStr := c.GetHeader("Authorization")
	userID, username, err := utils.VerifyJWT(tokenStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	productID := utils.Request(c, "product_id")
	iceStatus := utils.Request(c, "ice_status")
	size := utils.Request(c, "size")
	quantityRaw := utils.Request(c, "quantity")
	note := utils.Request(c, "note")

	if productID == "" || iceStatus == "" || size == "" || quantityRaw == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "All fields are required"})
		return
	}

	quantityStr, ok := quantityRaw.(string)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid quantity type"})
		return
	}

	qtyInt, err := utils.SafeAtoi(quantityStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Quantity must be a valid number"})
		return
	}

	tx := database.DB.Begin()
	if tx.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
		return
	}

	// Ambil data produk
	var masterProduct models.MasterProduct
	if err := tx.Where("id = ?", productID).First(&masterProduct).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}

	// Cek apakah item basket sudah ada
	var basketDetail models.BasketOrder
	err = tx.Where("user_id_order = ? AND product_id = ? AND ice_status = ? AND size = ? AND is_ordered = false",
		userID, productID, iceStatus, size).
		First(&basketDetail).Error

	// Handle note (default "-")
	var noteStr string
	if note == nil || note == "" {
		noteStr = "-"
	} else {
		noteStr, _ = note.(string)
	}

	var basketID uuid.UUID

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Buat data baru
			newDetail := models.BasketOrder{
				UserIDOrder: userID,
				ProductID:   masterProduct.ID,
				IceStatus:   iceStatus.(string),
				Size:        size.(string),
				Quantity:    int64(qtyInt),
				Note:        &noteStr,
				Price:       masterProduct.Price,
				TotalPrice:  masterProduct.Price,
				IsOrdered:   false,
				CreatedBy:   username,
				UpdatedBy:   username,
			}

			if err := tx.Create(&newDetail).Error; err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add new item to basket"})
				return
			}

			basketID = newDetail.ID
		} else {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error checking basket item"})
			return
		}
	} else {
		// Sudah ada, update quantity
		basketDetail.Quantity += int64(qtyInt)
		basketDetail.UpdatedBy = username
		basketDetail.TotalPrice = basketDetail.Price * float64(basketDetail.Quantity)

		if err := tx.Save(&basketDetail).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update basket item"})
			return
		}

		basketID = basketDetail.ID
	}

	// Commit transaksi
	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
		return
	}

	var reFetchBasket models.BasketOrder
	if err := database.DB.Preload("MasterProduct").
		First(&reFetchBasket, "id = ?", basketID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load basket with product info"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Basket inserted successfully",
		"data":    reFetchBasket,
	})
}

func AddQty(c *gin.Context) {
	tokenStr := c.GetHeader("Authorization")
	userID, username, err := utils.VerifyJWT(tokenStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	basketID := utils.Request(c, "basket_id")
	if basketID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "basket_id is required"})
		return
	}

	tx := database.DB.Begin()

	// Ambil data basket termasuk relasi MasterProduct
	var basketOrder models.BasketOrder
	if err := tx.Preload("MasterProduct").
		Where("id = ? AND user_id_order = ?", basketID, userID).
		First(&basketOrder).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusNotFound, gin.H{"error": "Basket item not found"})
		return
	}

	// Update qty dan total
	basketOrder.Quantity += 1
	basketOrder.Price = basketOrder.MasterProduct.Price
	basketOrder.TotalPrice = float64(basketOrder.Quantity) * basketOrder.Price
	basketOrder.UpdatedBy = username

	if err := tx.Save(&basketOrder).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update quantity"})
		return
	}

	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
		return
	}

	// Ambil ulang dengan Preload jika perlu
	if err := database.DB.Preload("MasterProduct").
		First(&basketOrder, "id = ?", basketID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load updated basket"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Quantity added successfully",
		"data":    basketOrder,
	})
}

func RemoveQty(c *gin.Context) {
	tokenStr := c.GetHeader("Authorization")
	userID, username, err := utils.VerifyJWT(tokenStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	basketID := utils.Request(c, "basket_id")
	if basketID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "basket_id is required"})
		return
	}

	tx := database.DB.Begin()

	// Ambil data basket dengan relasi MasterProduct
	var basketOrder models.BasketOrder
	if err := tx.Preload("MasterProduct").
		Where("id = ? AND user_id_order = ?", basketID, userID).
		First(&basketOrder).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusNotFound, gin.H{"error": "Basket item not found"})
		return
	}

	// Kurangi qty
	basketOrder.Quantity -= 1
	basketOrder.Price = basketOrder.MasterProduct.Price
	basketOrder.TotalPrice = float64(basketOrder.Quantity) * basketOrder.Price
	basketOrder.UpdatedBy = username

	// Jika quantity <= 0, hapus
	if basketOrder.Quantity <= 0 {
		basketOrder.DeletedBy = &username

		if err := tx.Delete(&basketOrder).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete basket item"})
			return
		}
	}

	// Jika masih > 0, update normal
	if err := tx.Save(&basketOrder).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update quantity"})
		return
	}

	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Quantity decreased successfully",
		"data":    basketOrder,
	})
}

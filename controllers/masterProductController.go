package controllers

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"renjana-app/database"
	"renjana-app/models"
	"renjana-app/utils"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func InsertMasterProduct(c *gin.Context) {
	tokenStr := c.GetHeader("Authorization")
	_, username, err := utils.VerifyJWT(tokenStr)

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	name := utils.Request(c, "name")
	description := utils.Request(c, "description")
	price := utils.Request(c, "price")
	image := utils.Request(c, "image")
	category := utils.Request(c, "category")
	is_active := utils.Request(c, "is_active")
	is_beverage := utils.Request(c, "is_beverage")

	if name == nil || name == "" {
		c.JSON(400, gin.H{"error": "Name of product can't empty!"})
		return
	}

	if description == nil || description == "" {
		c.JSON(400, gin.H{"error": "Description of product can't empty!"})
		return
	}

	if price == nil || price == "" {
		c.JSON(400, gin.H{"error": "Price of product can't empty!"})
		return
	}

	if is_active == nil || is_active == "" {
		c.JSON(400, gin.H{"error": "Activated status of product can't empty!"})
		return
	}

	if is_beverage == nil || is_beverage == "" {
		c.JSON(400, gin.H{"error": "Beverage status of product can't empty!"})
		return
	}

	if image == nil || image == "" {
		c.JSON(400, gin.H{"error": "Image of product can't empty!"})
		return
	}

	if category == nil || category == "" {
		c.JSON(400, gin.H{"error": "Category of product can't empty!"})
		return
	}

	// image product
	base64Str, ok := image.(string)
	if !ok || strings.TrimSpace(base64Str) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid base64 string for profile picture"})
		return
	}

	imageData, imageExt, err := utils.DecodeBase64Image(base64Str)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	allowedImageExt := []string{".jpg", ".png", ".gif", ".webp"}
	if !slices.Contains(allowedImageExt, imageExt) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Image extension only allowed: jpg, png, gif, webp"})
		return
	}

	filename := fmt.Sprintf("product_%s%s", strings.ReplaceAll(name.(string), " ", "_"), imageExt)
	savePath := filepath.Join("storage", "upload", "product", filename)

	if err := os.MkdirAll(filepath.Dir(savePath), os.ModePerm); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create directory"})
		return
	}

	if err := os.WriteFile(savePath, imageData, 0644); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save image"})
		return
	}

	mp := models.MasterProduct{
		Name:        strings.ToLower(name.(string)),
		Description: strings.ToLower(description.(string)),
		Price:       price.(float64),
		Category:    category.(string),
		Image:       savePath,
		IsActive:    is_active.(bool),
		IsBeverage:  is_beverage.(bool),
		CreatedBy:   username,
		UpdatedBy:   username,
	}

	tx := database.DB.Begin()
	if tx.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Transaction DB failed!"})
		return
	}

	//check category
	if taken, err := utils.IsProductCategoryNameTaken(tx, mp.Category); err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Crash when checking name product category!"})
		return
	} else if !taken {
		tx.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{"error": "Product category isn't exist!"})
		return
	}

	// Cek product
	if taken, err := utils.IsProductExist(tx, mp.Name); err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Crash when checking name product!"})
		return
	} else if taken {
		tx.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{"error": "Product already exist!"})
		return
	}

	if err := tx.Create(&mp).Error; err != nil {
		tx.Rollback()
		c.JSON(500, gin.H{"error": "Product can't registered!"})
		return
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		c.JSON(500, gin.H{"error": "Crash when registering user!"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Product ready to serve!",
		"data":    mp,
	})
}

func GetAllProduct(c *gin.Context) {
	var masterProduct []models.MasterProduct

	tx := database.DB.Begin()
	if tx.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
		return
	}

	if err := tx.Find(&masterProduct).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Success catch all master products!",
		"data":    masterProduct,
	})
}

func GetProductByID(c *gin.Context) {
	product_id := utils.Request(c, "product_id")
	var masterProduct models.MasterProduct

	tx := database.DB.Begin()

	if tx.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
		return
	}

	if err := tx.Where("id = ?", product_id).First(&masterProduct).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Success catch master product!",
		"data":    masterProduct,
	})
}

func UpdateProduct(c *gin.Context) {
	tokenStr := c.GetHeader("Authorization")
	_, username, err := utils.VerifyJWT(tokenStr)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Token too old!"})
		return
	}

	// Extract input
	productID := utils.Request(c, "product_id")
	name := utils.Request(c, "name")
	description := utils.Request(c, "description")
	price := utils.Request(c, "price")
	image := utils.Request(c, "image")
	isActive := utils.Request(c, "is_active")
	isBeverage := utils.Request(c, "is_beverage")
	category := utils.Request(c, "category")

	if productID == nil || productID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Product ID is required!"})
		return
	}

	tx := database.DB.Begin()
	if tx.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction!"})
		return
	}

	var product models.MasterProduct
	if err := tx.Where("id = ?", productID).First(&product).Error; err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Product not found!"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch product data!"})
		}
		return
	}

	// Use old data if field is not updated
	newName := product.Name
	if name != nil && name != "" {
		newName = strings.ToLower(name.(string))
	}

	if product.Name != newName {
		if taken, err := utils.IsProductExist(tx, newName); err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Crash when checking name product!"})
			return
		} else if taken {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{"error": "Product name same like another product!"})
			return
		}
	}

	newCategory := product.Category
	if category != nil && category != "" {
		newCategory = strings.ToLower(category.(string))
	}

	if product.Category != newCategory {
		//check category
		if taken, err := utils.IsProductCategoryNameTaken(tx, product.Category); err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Crash when checking product category!"})
			return
		} else if !taken {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{"error": "Product category isn't exist!"})
			return
		}
	}

	newDesc := product.Description
	if description != nil && description != "" {
		newDesc = strings.ToLower(description.(string))
	}

	newPrice := product.Price
	if price != nil {
		switch v := price.(type) {
		case float64:
			newPrice = v
		case string:
			if parsed, err := strconv.ParseFloat(v, 64); err == nil {
				newPrice = parsed
			}
		}
	}

	newImage := product.Image
	if image != nil {
		base64Str, ok := image.(string)
		if !ok || strings.TrimSpace(base64Str) == "" {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid base64 string for image"})
			return
		}

		// Delete old image if exists
		if product.Image != "" {
			if err := os.Remove(product.Image); err != nil && !os.IsNotExist(err) {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete old image"})
				return
			}
		}

		// Decode & Save new image
		imageData, imageExt, err := utils.DecodeBase64Image(base64Str)
		if err != nil {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		allowedExt := []string{".jpg", ".png", ".gif", ".webp"}
		if !slices.Contains(allowedExt, imageExt) {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{"error": "Image must be jpg, png, gif, or webp"})
			return
		}

		filename := fmt.Sprintf("product_%s%s", strings.ReplaceAll(newName, " ", "_"), imageExt)
		savePath := filepath.Join("storage", "upload", "product", filename)

		if err := os.MkdirAll(filepath.Dir(savePath), os.ModePerm); err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create image directory"})
			return
		}

		if err := os.WriteFile(savePath, imageData, 0644); err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save image"})
			return
		}

		newImage = savePath
	}

	newIsActive := product.IsActive
	if isActive != nil {
		if val, ok := isActive.(bool); ok {
			newIsActive = val
		}
	}

	newIsBeverage := product.IsBeverage
	if isBeverage != nil {
		if val, ok := isBeverage.(bool); ok {
			newIsBeverage = val
		}
	}

	// Update product
	updateData := map[string]any{
		"name":        newName,
		"description": newDesc,
		"price":       newPrice,
		"image":       newImage,
		"is_active":   newIsActive,
		"is_beverage": newIsBeverage,
		"updated_by":  username,
	}

	if err := tx.Model(&product).Updates(updateData).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update product"})
		return
	}

	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Transaction commit failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Product updated successfully!",
		"data":    updateData,
	})
}

func DeleteProduct(c *gin.Context) {
	tokenStr := c.GetHeader("Authorization")
	_, username, err := utils.VerifyJWT(tokenStr)

	if err != nil {
		c.JSON(500, gin.H{"error": "Token is too old!"})
		return
	}

	product_id := utils.Request(c, "product_id")

	if product_id == nil {
		c.JSON(500, gin.H{"error": "Product id not found!"})
		return
	}

	tx := database.DB.Begin()
	if tx.Error != nil {
		c.JSON(500, gin.H{"error": "Failed to start transaction!"})
		return
	}

	var masterProduct models.MasterProduct

	if err := tx.Where("id = ?", product_id).First(&masterProduct).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found!"})
		return
	}

	deleteProduct := tx.Model(&masterProduct).Where("id = ?", product_id).Updates(map[string]any{
		"is_active":  false,
		"deleted_at": time.Now(),
		"deleted_by": username,
	})

	if deleteProduct.Error != nil {
		tx.Rollback()
		c.JSON(500, gin.H{"error": "Product failed to delete!"})
		return
	}

	if masterProduct.Image != "" {
		if err := os.Remove(masterProduct.Image); err != nil && !os.IsNotExist(err) {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete old image"})
			return
		}
	}

	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Transaction commit failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Product deleted successfully!",
	})
}

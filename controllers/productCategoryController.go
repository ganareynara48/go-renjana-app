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
	"gorm.io/gorm"
)

func InsertCategory(c *gin.Context) {
	tokenStr := c.GetHeader("Authorization")
	_, username, err := utils.VerifyJWT(tokenStr)

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	name := utils.Request(c, "name")
	description := utils.Request(c, "description")

	if name == nil || name == "" {
		c.JSON(400, gin.H{"error": "Label can't empty!"})
		return
	}

	var descPtr *string
	if desc, ok := description.(string); ok && strings.TrimSpace(desc) != "" {
		d := strings.ToLower(desc)
		descPtr = &d
	} else {
		d := "-"
		descPtr = &d
	}

	ProductCategory := models.ProductCategory{
		Name:        strings.ToLower(name.(string)),
		Description: descPtr,
		CreatedBy:   username,
		UpdatedBy:   username,
	}

	tx := database.DB.Begin()
	if tx.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Transaction DB failed!"})
		return
	}

	// Cek nama kategori
	if taken, err := utils.IsProductCategoryNameTaken(tx, strings.ToLower(name.(string))); err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Crash when checking user!"})
		return
	} else if taken {
		tx.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{"error": "Product name already exists!"})
		return
	}

	if err := tx.Create(&ProductCategory).Error; err != nil {
		tx.Rollback()
		c.JSON(500, gin.H{"error": "Address can't registered!"})
		return
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		c.JSON(500, gin.H{"error": "Crash when registering user!"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Category registered!",
		"data":    ProductCategory,
	})
}

func GetAllCategory(c *gin.Context) {
	var ProductCategory []models.ProductCategory

	tx := database.DB.Begin()
	if tx.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
		return
	}

	if err := tx.Find(&ProductCategory).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusNotFound, gin.H{"error": "Product category not found"})
		return
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Success catch all product category!",
		"data":    ProductCategory,
	})
}

func GetCategoryByID(c *gin.Context) {
	product_category_id := utils.Request(c, "product_category_id")

	var ProductCategory []models.ProductCategory

	tx := database.DB.Begin()
	if tx.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
		return
	}

	if err := tx.Where("id = ?", product_category_id).Find(&ProductCategory).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusNotFound, gin.H{"error": "Product category not found"})
		return
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Success catch product category!",
		"data":    ProductCategory,
	})
}

func UpdateCategory(c *gin.Context) {
	tokenStr := c.GetHeader("Authorization")
	_, username, err := utils.VerifyJWT(tokenStr)

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	productCategoryID := utils.Request(c, "product_category_id")
	name := utils.Request(c, "name")
	description := utils.Request(c, "description")

	if productCategoryID == nil || productCategoryID == "" {
		c.JSON(400, gin.H{"error": "Product category not found!"})
		return
	}

	if name == nil || name == "" {
		c.JSON(400, gin.H{"error": "Name product category data not found!"})
		return
	}

	if description == nil || description == "" {
		c.JSON(400, gin.H{"error": "Description product category data not found!"})
		return
	}

	tx := database.DB.Begin()
	if tx.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Transaction DB failed!"})
		return
	}

	var productCategory models.ProductCategory
	res := tx.Where(map[string]any{
		"id":         productCategoryID,
		"deleted_by": nil,
	}).First(&productCategory)

	if res.Error != nil {
		tx.Rollback()
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Product category data not found!"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": res.Error.Error()})
		}
		return
	}

	newName := strings.ToLower(name.(string))

	// Cek duplikat jika nama berubah
	if newName != productCategory.Name {
		taken, err := utils.IsProductCategoryNameTaken(tx, newName)
		if err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Crash when checking product name!"})
			return
		}
		if taken {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{"error": "Product name already exists!"})
			return
		}
	}

	newDescription := strings.ToLower(description.(string))

	updateData := map[string]any{
		"name":        newName,
		"description": newDescription,
		"updated_by":  username,
	}

	if err := tx.Model(&productCategory).Updates(updateData).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update product category"})
		return
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("Product category with category name %s updated!", newName),
		"data":    productCategory,
	})
}

func DeleteCategory(c *gin.Context) {
	tokenStr := c.GetHeader("Authorization")
	_, username, err := utils.VerifyJWT(tokenStr)

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	product_category_id := utils.Request(c, "product_category_id")

	if product_category_id == nil || product_category_id == "" {
		c.JSON(400, gin.H{"error": "Product category not found!"})
		return
	}

	tx := database.DB.Begin()
	if tx.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Transaction DB failed!"})
		return
	}

	var ProductCategory models.ProductCategory
	// Ambil data address yang belum dihapus
	res := tx.Where("id = ? AND deleted_at IS NULL AND deleted_by IS NULL", product_category_id).First(&ProductCategory)

	var masterProduct models.MasterProduct
	err = tx.Where("category = ?", ProductCategory.Name).Where("IsActive", true).First(&masterProduct).Error

	if err == nil {
		// Ditemukan product yang menggunakan kategori
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("Category '%s' is currently used by product '%s'", ProductCategory.Name, masterProduct.Name),
		})
		tx.Rollback()
		return
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		// Jika error selain 'data tidak ditemukan', berarti query gagal
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		tx.Rollback()
		return
	}

	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Product category data not found!"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": res.Error.Error()})
		}
		tx.Rollback()
		return
	}

	// Soft delete
	if err := tx.Model(&ProductCategory).Updates(map[string]any{
		"deleted_at": time.Now(),
		"deleted_by": username,
	}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete product category"})
		return
	}

	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Product category deleted!"})
}

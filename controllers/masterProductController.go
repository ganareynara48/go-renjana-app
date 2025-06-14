package controllers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"renjana-app/database"
	"renjana-app/models"
	"renjana-app/utils"
	"slices"
	"strings"

	"github.com/gin-gonic/gin"
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
	is_active := utils.Request(c, "is_active")

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

	if image == nil || image == "" {
		c.JSON(400, gin.H{"error": "Image of product can't empty!"})
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

	filename := fmt.Sprintf("product_%s%s", name, imageExt)
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
		Image:       savePath,
		IsActive:    is_active.(bool),
		CreatedBy:   username,
		UpdatedBy:   username,
	}

	tx := database.DB.Begin()
	if tx.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Transaction DB failed!"})
		return
	}

	// Cek user
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

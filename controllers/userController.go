package controllers

import (
	"errors"
	"fmt"
	"net/http"
	"net/mail"
	"os"
	"path/filepath"
	"renjana-app/database"
	"renjana-app/models"
	"renjana-app/utils"
	"slices"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterUser(c *gin.Context) {
	fullname := utils.Request(c, "fullname")
	username := utils.Request(c, "username")
	email := utils.Request(c, "email")
	password := utils.Request(c, "password")

	if fullname == nil || fullname == "" {
		c.JSON(400, gin.H{"error": "Fullname can't empty!"})
		return
	}

	if username == nil || username == "" {
		c.JSON(400, gin.H{"error": "Username can't empty!"})
		return
	}

	if email == nil || email == "" {
		c.JSON(400, gin.H{"error": "Email can't empty!"})
		return
	}

	// Validasi format email
	if _, err := mail.ParseAddress(email.(string)); err != nil {
		c.JSON(400, gin.H{"error": "Email isn't valid!"})
		return
	}

	if password == nil || password == "" {
		c.JSON(400, gin.H{"error": "Password can't empty!"})
		return
	}

	// Buat struct user (sesuaikan dengan model kamu)
	user := models.User{
		Fullname:  strings.ToLower(fullname.(string)),
		Username:  strings.ToLower(username.(string)),
		Email:     email.(string),
		Password:  password.(string),
		CreatedBy: strings.ToLower(username.(string)),
		UpdatedBy: strings.ToLower(username.(string)),
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(user.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to encrypt password!"})
		return
	}

	//replace with hashed password
	user.Password = hashedPassword

	// Mulai transaksi
	tx := database.DB.Begin()
	if tx.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Transaction DB failed!"})
		return
	}

	// Cek username
	if taken, err := utils.IsUsernameTaken(tx, user.Username); err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Crash when checking username!"})
		return
	} else if taken {
		tx.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username already used!"})
		return
	}

	// Cek email
	if taken, err := utils.IsEmailTaken(tx, user.Email); err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Crash when checking email!"})
		return
	} else if taken {
		tx.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email already used!"})
		return
	}

	// Simpan user
	if err := tx.Create(&user).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User can't registered!"})
		return
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Crash when registering user!"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "User registered!",
		"data": gin.H{
			"fullname": user.Fullname,
			"username": user.Username,
			"email":    user.Email,
		},
	})
}

func LoginUser(c *gin.Context) {
	username := utils.Request(c, "username")
	password := utils.Request(c, "password")

	if username == nil || username == "" {
		c.JSON(400, gin.H{"error": "Username can't be empty!"})
		return
	}

	if password == nil || password == "" {
		c.JSON(400, gin.H{"error": "Password can't be empty!"})
		return
	}

	usernameStr := strings.ToLower(username.(string))
	passwordStr := password.(string)

	tx := database.DB.Begin()
	if tx.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Transaction DB failed!"})
		return
	}

	var user models.User
	if err := tx.Where("username = ?", usernameStr).First(&user).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Username might be wrong!"})
		return
	}

	if !utils.CheckPasswordHash(passwordStr, user.Password) {
		tx.Rollback()
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username/password combination!"})
		return
	}

	token, err := utils.GenerateJWT(user.ID, user.Username)
	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	err = tx.Model(&models.User{}).
		Where("username = ?", usernameStr).
		Updates(map[string]interface{}{
			"updated_by": usernameStr,
			"token":      token,
		}).Error

	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save token"})
		return
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Transaction commit failed!"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}

func InfoUser(c *gin.Context) {
	token := c.GetHeader("Authorization")

	userID, username, err := utils.VerifyJWT(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	tx := database.DB.Begin()
	if tx.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
		return
	}

	var user models.User
	if err := tx.First(&user, userID).Error; err != nil {
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
		"message":  "Success catch user!",
		"username": username, // tambahan jika ingin return username dari token
		"data":     user,
	})
}

func GetUserByID(c *gin.Context) {
	userID := utils.Request(c, "user_id")

	var user models.User

	tx := database.DB.Begin()
	if tx.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
		return
	}

	if err := tx.Where("id = ?", userID).First(&user).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	if err := tx.Preload("Addresses", "is_default = ?", true).First(&user, user.ID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load updated user with default address"})
		return
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Success catch all users!",
		"data":    user,
	})
}

func GetAllUser(c *gin.Context) {
	var user []models.User

	tx := database.DB.Begin()
	if tx.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
		return
	}

	if err := tx.Find(&user).Error; err != nil {
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
		"data":    user,
	})
}

func UpdateUser(c *gin.Context) {
	tokenStr := c.GetHeader("Authorization")
	_, username, err := utils.VerifyJWT(tokenStr)

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	reqID := utils.Request(c, "id_user")
	reqFullname := utils.Request(c, "fullname")
	reqUsername := utils.Request(c, "username")
	reqEmail := utils.Request(c, "email")
	reqIsUpdatePassword := utils.Request(c, "is_update_password")
	reqNewPassword := utils.Request(c, "new_password")
	reqRetypePassword := utils.Request(c, "retype_password")
	reqProfilePicture := utils.Request(c, "profile_picture")

	tx := database.DB.Begin()
	if tx.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Transaction DB failed!"})
		return
	}

	var user models.User
	res := tx.Where(map[string]interface{}{
		"id":         reqID,
		"deleted_at": nil,
		"deleted_by": nil,
	}).First(&user)

	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "User data not found!"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": res.Error.Error()})
		}
		tx.Rollback()
		return
	}

	newFullname := utils.GetStringOrDefault(reqFullname, user.Fullname)
	newUsername := utils.GetStringOrDefault(reqUsername, user.Username)

	// Validasi email
	var newEmail string
	if reqEmail != nil {
		emailStr, ok := reqEmail.(string)
		if !ok || emailStr == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid email input"})
			tx.Rollback()
			return
		}
		if _, err := mail.ParseAddress(emailStr); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Email is not valid!"})
			tx.Rollback()
			return
		}
		newEmail = emailStr
	} else {
		newEmail = user.Email
	}

	// Password update
	newPassword := user.Password
	if isUpdate, ok := reqIsUpdatePassword.(bool); ok && isUpdate {
		newPass, ok1 := reqNewPassword.(string)
		retypePass, ok2 := reqRetypePassword.(string)

		if !ok1 || !ok2 || newPass != retypePass {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Passwords do not match!"})
			tx.Rollback()
			return
		}

		hashedPassword, err := utils.HashPassword(newPass)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to encrypt password!"})
			tx.Rollback()
			return
		}
		newPassword = hashedPassword
	}

	// Profile picture
	newProfilePicture := ""
	if reqProfilePicture != nil {
		base64Str, ok := reqProfilePicture.(string)
		if !ok || strings.TrimSpace(base64Str) == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid base64 string for profile picture"})
			tx.Rollback()
			return
		}

		if user.ProfilePicture != nil && *user.ProfilePicture != "" {
			oldPath := *user.ProfilePicture
			if err := os.Remove(oldPath); err != nil && !os.IsNotExist(err) {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete old profile picture"})
				tx.Rollback()
				return
			}
		}

		imageData, imageExt, err := utils.DecodeBase64Image(base64Str)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			tx.Rollback()
			return
		}

		allowedImageExt := []string{".jpg", ".png", ".gif", ".webp"}
		if !slices.Contains(allowedImageExt, imageExt) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Image extension only allowed: jpg, png, gif, webp"})
			tx.Rollback()
			return
		}

		filename := fmt.Sprintf("user_%d%s", time.Now().UnixNano(), imageExt)
		savePath := filepath.Join("storage", "upload", "profile", newUsername, filename)

		if err := os.MkdirAll(filepath.Dir(savePath), os.ModePerm); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create directory"})
			tx.Rollback()
			return
		}

		if err := os.WriteFile(savePath, imageData, 0644); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save image"})
			tx.Rollback()
			return
		}

		newProfilePicture = savePath
	}

	// Uniqueness
	if user.Username != newUsername {
		if taken, err := utils.IsUsernameTaken(tx, newUsername); err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error checking username"})
			return
		} else if taken {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{"error": "Username already used"})
			return
		}
	}

	if user.Email != newEmail {
		if taken, err := utils.IsEmailTaken(tx, newEmail); err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error checking email"})
			return
		} else if taken {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{"error": "Email already used"})
			return
		}
	}

	// Final update
	updates := map[string]any{
		"fullname":   strings.ToLower(newFullname),
		"username":   strings.ToLower(newUsername),
		"email":      newEmail,
		"password":   newPassword,
		"updated_at": time.Now(),
		"updated_by": username,
		"token":      user.Token,
	}

	if newProfilePicture != "" {
		updates["profile_picture"] = newProfilePicture
	}

	if err := tx.Model(&user).Updates(updates).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}

	if err := tx.Preload("Addresses", "is_default = ?", true).First(&user, user.ID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load updated user with default address"})
		return
	}

	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User updated!",
		"data":    user,
	})
}

func DeleteUser(c *gin.Context) {
	tokenStr := c.GetHeader("Authorization")
	_, username, err := utils.VerifyJWT(tokenStr)

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	reqID := utils.Request(c, "id_user")

	tx := database.DB.Begin()
	if tx.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Transaction DB failed!"})
		return
	}

	var user models.User
	res := tx.Where(map[string]interface{}{
		"id":         reqID,
		"deleted_at": nil,
		"deleted_by": nil,
	}).First(&user)

	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "User data not found!"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": res.Error.Error()})
		}
		tx.Rollback()
		return
	}

	// Final update
	updates := map[string]any{
		"deleted_at": time.Now(),
		"deleted_by": username,
	}

	if err := tx.Model(&user).Updates(updates).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
		return
	}

	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User deleted!",
	})
}

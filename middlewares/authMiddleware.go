package middlewares

import (
	"net/http"
	"renjana-app/utils"
	"strings"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token == "" || !strings.HasPrefix(token, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: missing or malformed token"})
			c.Abort()
			return
		}

		// Verifikasi dan ekstrak userID + username
		userID, username, err := utils.VerifyJWT(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			c.Abort()
			return
		}

		// Simpan ke context Gin
		c.Set("user_id", userID)
		c.Set("username", username)

		c.Next()
	}
}

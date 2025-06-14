package utils

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var jwtSecret = []byte(os.Getenv("JWT_SECRET"))

func GenerateJWT(userID uuid.UUID, username string) (string, error) {
	claims := jwt.MapClaims{
		"id":       userID.String(),
		"username": username,
		"iat":      time.Now().Unix(),                     // waktu token dibuat
		"exp":      time.Now().Add(24 * time.Hour).Unix(), // token berlaku 24 jam
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

func VerifyJWT(tokenString string) (uuid.UUID, string, error) {
	tokenString = strings.TrimPrefix(tokenString, "Bearer ")

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return jwtSecret, nil
	})
	if err != nil || !token.Valid {
		return uuid.Nil, "", errors.New("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return uuid.Nil, "", errors.New("invalid claims")
	}

	// Ambil ID user
	idStr, ok := claims["id"].(string)
	if !ok {
		return uuid.Nil, "", errors.New("user_id missing or not a string")
	}

	username, ok := claims["username"].(string)
	if !ok {
		return uuid.Nil, "", errors.New("username missing or not a string")
	}

	// Cek iat
	if iatFloat, ok := claims["iat"].(float64); ok {
		issuedAt := time.Unix(int64(iatFloat), 0)
		if time.Since(issuedAt) > 12*time.Hour {
			return uuid.Nil, "", errors.New("token too old, please re-login")
		}
	} else {
		return uuid.Nil, "", errors.New("iat claim missing or invalid")
	}

	userID, err := uuid.Parse(idStr)
	if err != nil {
		return uuid.Nil, "", errors.New("invalid UUID format")
	}

	return userID, username, nil
}

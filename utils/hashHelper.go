package utils

import (
	"golang.org/x/crypto/bcrypt"
)

// HashPassword melakukan hashing pada password sebelum disimpan ke database
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CheckPasswordHash memeriksa apakah password yang diberikan cocok dengan hash yang tersimpan di database
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

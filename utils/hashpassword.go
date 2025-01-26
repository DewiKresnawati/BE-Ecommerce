package utils

import (
	"golang.org/x/crypto/bcrypt"
)

// HashPassword meng-hash password sebelum disimpan ke database
func HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

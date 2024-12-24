package utils

import (
	"log"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/joho/godotenv"
)

// Load .env file untuk mengambil variabel environment
func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
}

// Mengambil JWT Secret Key dari .env
var jwtSecret = []byte(os.Getenv("JWT_SECRET_KEY"))

// GenerateJWT membuat dan menandatangani JWT token
func GenerateJWT(userID string) (string, error) {
	// Membuat klaim (claims) JWT
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Hour * 24).Unix(), // Token expired setelah 24 jam
	}

	// Membuat token JWT dengan signing method HS256
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Menandatangani token dengan secret key
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ValidateJWT memverifikasi JWT token dan mengembalikan klaim jika valid
func ValidateJWT(tokenString string) (jwt.MapClaims, error) {
	// Mem-parsing token dan memverifikasi dengan secret key
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Memastikan menggunakan signing method yang benar (HS256)
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.NewValidationError("invalid signing method", jwt.ValidationErrorSignatureInvalid)
		}
		return jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	// Mengembalikan klaim jika token valid
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	} else {
		return nil, jwt.NewValidationError("invalid token", jwt.ValidationErrorClaimsInvalid)
	}
}

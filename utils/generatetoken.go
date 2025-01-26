package utils

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
)

// GenerateRandomToken menghasilkan kode OTP atau token unik
func GenerateRandomToken(length int) string {
	token := make([]byte, length)
	_, err := rand.Read(token)
	if err != nil {
		fmt.Println("Failed to generate random token:", err)
		return ""
	}

	return base64.RawURLEncoding.EncodeToString(token)[:length]
}

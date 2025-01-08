package config

import (
	"github.com/gofiber/fiber/v2/middleware/cors"
)

// SetupCORS mengatur middleware CORS dengan konfigurasi khusus
func SetupCORS() cors.Config {
	return cors.Config{
		AllowOrigins: "http://127.0.0.1:5503",             // Mengizinkan semua origin
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS", // Metode yang diizinkan
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowCredentials: true, // Mengizinkan credentials seperti cookies
	}
}

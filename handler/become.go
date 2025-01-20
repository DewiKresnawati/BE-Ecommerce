package handler

import (
	"be_ecommerce/config"
	"be_ecommerce/utils"
	"context"
	"encoding/base64"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Endpoint untuk menjadi seller
func BecomeSeller(c *fiber.Ctx) error {
	// Struktur payload untuk menerima data dari request
	type RequestPayload struct {
		StoreName   string `json:"store_name"`
		FullAddress string `json:"full_address"`
		NIK         string `json:"nik"`
		PhotoBase64 string `json:"photo_base64"`
	}

	// Parsing body request ke dalam struktur payload
	var payload RequestPayload
	if err := c.BodyParser(&payload); err != nil {
		log.Println("Error parsing request body:", err)
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid request body",
		})
	}

	// Validasi input
	if payload.StoreName == "" || payload.FullAddress == "" || payload.NIK == "" || payload.PhotoBase64 == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"message": "All fields are required (store_name, full_address, nik, photo_base64)",
		})
	}

	// Ambil token dari header Authorization
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"message": "Missing authorization token",
		})
	}

	// Hapus prefix "Bearer " dari header
	token := strings.TrimPrefix(authHeader, "Bearer ")

	// Verifikasi dan parsing token
	claims, err := utils.ParseToken(token)
	if err != nil {
		log.Println("Invalid token:", err)
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"message": "Invalid token",
		})
	}

	// Ambil user_id dari token
	userID, ok := claims["user_id"].(string)
	if !ok || userID == "" {
		log.Println("User ID not found in token claims")
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"message": "Invalid token: user_id missing",
		})
	}

	// Konversi user_id ke ObjectID MongoDB
	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		log.Println("Invalid user ID format:", userID)
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid user ID format",
		})
	}

	// Decode foto dari Base64 dan simpan ke file
	photoData, err := base64.StdEncoding.DecodeString(payload.PhotoBase64)
	if err != nil {
		log.Println("Error decoding base64 photo:", err)
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid photo format",
		})
	}

	// Tentukan direktori upload
	uploadDir := "uploads/seller_photos"
	if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
		if err := os.MkdirAll(uploadDir, 0755); err != nil {
			log.Println("Error creating upload directory:", err)
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"message": "Failed to prepare upload directory",
			})
		}
	}

	// Tentukan nama file
	photoPath := filepath.Join(uploadDir, userID+".jpg")
	if err := ioutil.WriteFile(photoPath, photoData, 0644); err != nil {
		log.Println("Error saving photo file:", err)
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to save photo",
		})
	}

	// Filter untuk mencari user berdasarkan ID dan role "customer"
	filter := bson.M{
		"_id":   objectID,
		"roles": "customer",
	}

	// Update data user menjadi seller
	update := bson.M{
		"$set": bson.M{
			"store_info": bson.M{
				"store_name":   payload.StoreName,
				"full_address": payload.FullAddress,
				"nik":          payload.NIK,
				"photo_path":   photoPath, // Simpan path foto
			},
			"store_status": "pending", // Status awal aplikasi menjadi seller
		},
		"$addToSet": bson.M{
			"roles": "seller", // Tambahkan role seller jika belum ada
		},
	}

	// Gunakan koleksi dari fungsi getUserCollection
	collection := config.MongoClient.Database("ecommerce").Collection("users")

	// Update dokumen di MongoDB
	result, err := collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		log.Println("Error updating user to become seller:", err)
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to update user",
		})
	}

	// Jika tidak ada dokumen yang diubah, berarti user tidak ditemukan atau sudah menjadi seller
	if result.ModifiedCount == 0 {
		log.Println("No user updated. Either user not found or already a seller:", userID)
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"message": "User not found or already a seller",
		})
	}

	// Berhasil
	log.Println("User successfully updated to become seller:", userID)
	return c.JSON(fiber.Map{
		"message": "User successfully became a seller",
		"photo_path": photoPath, // Path foto untuk referensi
	})
}

package handler

import (
	"be_ecommerce/config"
	"be_ecommerce/model"
	"be_ecommerce/utils"
	"context"
	"strings"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ApplyAsSeller allows a customer to apply as a seller
func ApplyAsSeller(c *fiber.Ctx) error {
	// Ambil token dari header Authorization
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Missing Authorization header",
		})
	}

	// Hapus prefix "Bearer "
	token := strings.TrimPrefix(authHeader, "Bearer ")
	if token == authHeader {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Invalid token format",
		})
	}

	// Validasi token JWT
	claims, err := utils.ValidateJWT(token)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Invalid or expired token",
		})
	}

	// Ambil user_id dari klaim token
	userID, ok := claims["user_id"].(string)
	if !ok || userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Invalid token claims",
		})
	}

	// Convert userID ke ObjectID
	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid user ID format",
		})
	}

	// Parse data tambahan dari body request
	var storeInfo model.StoreInfo
	if err := c.BodyParser(&storeInfo); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid request body",
		})
	}

	// Validasi data tambahan
	if storeInfo.StoreName == "" || storeInfo.FullAddress == "" || storeInfo.NIK == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Store name, full address, and NIK are required",
		})
	}

	// Ambil user dari database
	collection := config.MongoClient.Database("ecommerce").Collection("users")
	var user model.User
	err = collection.FindOne(context.Background(), bson.M{"_id": objectID}).Decode(&user)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "User not found",
		})
	}

	// Periksa apakah user sudah memiliki permohonan pending
	if user.StoreStatus != nil && *user.StoreStatus == "pending" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Application already pending",
		})
	}

	// Perbarui status aplikasi toko dan informasi tambahan
	pendingStatus := "pending"
	_, err = collection.UpdateOne(context.Background(), bson.M{"_id": user.ID}, bson.M{
		"$set": bson.M{
			"store_status": pendingStatus,
			"store_info":   storeInfo,
		},
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to apply as seller",
		})
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Application submitted, waiting for admin approval",
	})
}

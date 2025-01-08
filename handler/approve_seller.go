package handler

import (
	"be_ecommerce/config"
	"be_ecommerce/model"
	"context"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ApproveSeller allows an admin to approve or reject a seller application
func ApproveSeller(c *fiber.Ctx) error {
	var request struct {
		UserID string `json:"user_id"` // ID pengguna
		Status string `json:"status"` // "approved" atau "rejected"
	}

	// Parse request body
	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid request body",
		})
	}

	// Validasi nilai status
	if request.Status != "approved" && request.Status != "rejected" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid status value",
		})
	}

	// Konversi UserID ke ObjectID
	objectID, err := primitive.ObjectIDFromHex(request.UserID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid user ID format",
		})
	}

	// Temukan user di database
	collection := config.MongoClient.Database("ecommerce").Collection("users")
	var user model.User
	err = collection.FindOne(context.Background(), bson.M{"_id": objectID}).Decode(&user)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"status":  "error",
			"message": "User not found",
		})
	}

	// Update role dan store_status
	update := bson.M{"store_status": request.Status}

	// Jika disetujui, tambahkan role "seller" jika belum ada
	if request.Status == "approved" {
		if !contains(user.Roles, "seller") {
			user.Roles = append(user.Roles, "seller")
		}
	} else if request.Status == "rejected" {
		// Jika ditolak, hapus role "seller" jika ada
		user.Roles = removeRole(user.Roles, "seller")
	}
	update["roles"] = user.Roles

	// Perbarui data user di database
	_, err = collection.UpdateOne(context.Background(), bson.M{"_id": user.ID}, bson.M{"$set": update})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to update application status",
		})
	}

	// Tentukan pesan berdasarkan status
	message := "Application rejected"
	if request.Status == "approved" {
		message = "Application approved, user is now a seller"
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": message,
	})
}

// Utility function to check if a role exists in roles slice
func contains(roles []string, role string) bool {
	for _, r := range roles {
		if r == role {
			return true
		}
	}
	return false
}

// Utility function to remove a role from roles slice
func removeRole(roles []string, role string) []string {
	var updatedRoles []string
	for _, r := range roles {
		if r != role {
			updatedRoles = append(updatedRoles, r)
		}
	}
	return updatedRoles
}

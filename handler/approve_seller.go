package handler

import (
	"be_ecommerce/config"
	"be_ecommerce/model"
	"context"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// ApproveSeller allows an admin to approve or reject a seller application
func ApproveSeller(c *fiber.Ctx) error {
	var request struct {
		UserID string `json:"user_id"` // ID pengguna
		Status string `json:"status"`  // "approved" atau "rejected"
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

type RejectRequest struct {
	UserID string `json:"user_id"`
	Status string `json:"status"`
}

func RejectSeller(c *fiber.Ctx) error {
	var req RejectRequest

	// Parse request body
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid request body",
			"error":   err.Error(),
		})
	}

	// Validate the request
	if req.UserID == "" || req.Status == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"message": "user_id and status are required",
		})
	}

	if req.Status != "rejected" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid status value. Only 'rejected' is allowed.",
		})
	}

	// Convert UserID to ObjectID
	objectID, err := primitive.ObjectIDFromHex(req.UserID)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid user ID format",
			"error":   err.Error(),
		})
	}

	// Connect to the database
	collection := config.MongoClient.Database("ecommerce").Collection("users")

	// Find the user in the database
	var user model.User
	err = collection.FindOne(context.Background(), bson.M{"_id": objectID}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return c.Status(http.StatusNotFound).JSON(fiber.Map{
				"message": "User not found",
			})
		}
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to fetch user",
			"error":   err.Error(),
		})
	}

	// Update store_status and remove "seller" role
	user.Roles = removeRole(user.Roles, "seller")
	update := bson.M{
		"store_status": req.Status,
		"roles":        user.Roles,
	}

	// Update the user in the database
	_, err = collection.UpdateOne(context.Background(), bson.M{"_id": objectID}, bson.M{"$set": update})
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to update user status",
			"error":   err.Error(),
		})
	}

	// Respond with success
	return c.JSON(fiber.Map{
		"message": "Application rejected, user status updated",
		"status":  "success",
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

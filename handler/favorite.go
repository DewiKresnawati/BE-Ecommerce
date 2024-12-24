package handler

import (
	"be_ecommerce/config"
	"be_ecommerce/model"
	"context"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// AddToFavorites menambahkan produk ke daftar favorit pengguna
func AddToFavorites(c *fiber.Ctx) error {
	var request struct {
		ProductID string `json:"product_id"`
	}
	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid request body"})
	}

	// Ambil user_id dari context
	userID := c.Locals("user_id").(string)

	// Ambil koleksi favorit
	collection := config.MongoClient.Database("ecommerce").Collection("favorites")

	// Periksa apakah favorit sudah ada
	var favorite model.Favorite
	err := collection.FindOne(context.Background(), bson.M{"user_id": userID}).Decode(&favorite)
	if err == mongo.ErrNoDocuments {
		// Jika tidak ada daftar favorit, buat baru
		favorite = model.Favorite{
			UserID:     userID,
			ProductIDs: []string{request.ProductID},
		}
		_, err := collection.InsertOne(context.Background(), favorite)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to add product to favorites"})
		}
	} else {
		// Jika favorit ada, tambahkan produk
		for _, id := range favorite.ProductIDs {
			if id == request.ProductID {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Product already in favorites"})
			}
		}
		favorite.ProductIDs = append(favorite.ProductIDs, request.ProductID)

		// Perbarui favorit
		_, err := collection.UpdateOne(context.Background(), bson.M{"user_id": userID}, bson.M{"$set": bson.M{"product_ids": favorite.ProductIDs}})
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to update favorites"})
		}
	}

	return c.JSON(fiber.Map{"message": "Product added to favorites successfully"})
}

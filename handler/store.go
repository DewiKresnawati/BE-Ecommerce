package handler

import (
	"be_ecommerce/config"
	"be_ecommerce/model"
	"context"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// GetStoreDetails returns store information and its products
func GetStoreDetails(c *fiber.Ctx) error {
	storeID := c.Params("id")

	// Konversi storeID ke ObjectID
	objectID, err := primitive.ObjectIDFromHex(storeID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid store ID format",
		})
	}

	// Ambil data toko dari database
	userCollection := config.MongoClient.Database("ecommerce").Collection("users")
	var store model.User
	err = userCollection.FindOne(context.Background(), bson.M{"_id": objectID, "roles": "seller"}).Decode(&store)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "Store not found",
		})
	}

	// Periksa apakah toko aktif
	if store.StoreStatus == nil || *store.StoreStatus != "approved" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Store is not active",
		})
	}

	// Ambil produk yang terkait dengan toko ini
	productCollection := config.MongoClient.Database("ecommerce").Collection("products")
	cursor, err := productCollection.Find(context.Background(), bson.M{"seller_id": objectID})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to fetch products",
		})
	}
	defer cursor.Close(context.Background())

	var products []model.Product
	if err := cursor.All(context.Background(), &products); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to parse products",
		})
	}

	// Kembalikan detail toko dan produk
	return c.JSON(fiber.Map{
		"store": fiber.Map{
			"id":           store.ID,
			"store_name":   store.StoreInfo.StoreName,
			"full_address": store.StoreInfo.FullAddress,
			"email":        store.Email,
			"status":       *store.StoreStatus,
		},
		"products": products,
	})
}


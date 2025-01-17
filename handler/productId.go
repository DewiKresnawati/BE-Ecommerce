package handler

import (
	"be_ecommerce/config"
	"be_ecommerce/model"
	"context"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// GetProductByID retrieves a product by its ID
func GetProductByID(c *fiber.Ctx) error {
	// Ambil ID dari URL parameter
	productID := c.Params("id")

	// Konversi ID ke ObjectID MongoDB
	objectID, err := primitive.ObjectIDFromHex(productID)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid product ID format",
		})
	}

	// Ambil koleksi produk
	productCollection := config.MongoClient.Database("ecommerce").Collection("products")

	// Cari produk berdasarkan ID
	var product model.Product
	err = productCollection.FindOne(context.Background(), bson.M{"_id": objectID}).Decode(&product)
	if err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"status":  "error",
			"message": "Product not found",
		})
	}

	// Kembalikan data produk
	return c.JSON(fiber.Map{
		"status": "success",
		"data":   product,
	})
}

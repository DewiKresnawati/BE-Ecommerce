package handler

import (
	"be_ecommerce/config"
	"be_ecommerce/model"
	"be_ecommerce/utils"
	"context"
	"strings"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
)

// CreateProduct handles creating a new product
func CreateProduct(c *fiber.Ctx) error {
	var product model.Product
	if err := c.BodyParser(&product); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Error parsing request body",
		})
	}

	collection := config.MongoClient.Database("ecommerce").Collection("products")
	result, err := collection.InsertOne(context.Background(), product)
	if err != nil {
		// Menangani error jika produk sudah ada (misalnya berdasarkan nama atau kategori)
		if strings.Contains(err.Error(), "duplicate key error") {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "Product already exists",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Error saving product to database",
		})
	}

	// Kembalikan response dengan ID produk yang baru saja disimpan
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message":    "Product created successfully",
		"product_id": result.InsertedID, // ID produk yang baru saja disimpan
	})
}

// GetAllProducts fetches all products
func GetAllProducts(c *fiber.Ctx) error {
	// Mendapatkan Authorization header
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"status":  "error",
			"message": "Authorization token is missing",
		})
	}

	// Memisahkan Bearer dan token
	authParts := strings.Split(authHeader, "Bearer ")
	if len(authParts) != 2 {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid authorization token format",
		})
	}
	tokenString := authParts[1]

	// Memverifikasi token
	claims, err := utils.ValidateJWT(tokenString)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid or expired token",
		})
	}

	// Mendapatkan user_id dari klaim
	userID, ok := claims["user_id"].(string)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid token claims",
		})
	}

	// Mengambil produk dari MongoDB
	collection := config.MongoClient.Database("ecommerce").Collection("products")
	cursor, err := collection.Find(c.Context(), bson.M{})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to fetch products",
		})
	}
	defer cursor.Close(c.Context())

	var products []model.Product
	if err := cursor.All(c.Context(), &products); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to parse products",
		})
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Products fetched successfully",
		"data":    products,
		"user_id": userID, // Mengembalikan user_id yang terkait dengan token
	})
}

func GetProductsUnderPrice(c *fiber.Ctx) error {
	priceLimit := 100000.0

	collection := config.MongoClient.Database("ecommerce").Collection("products")
	cursor, err := collection.Find(c.Context(), bson.M{"price": bson.M{"$lte": priceLimit}})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to fetch products under price limit",
		})
	}
	defer cursor.Close(c.Context())

	var products []model.Product
	if err := cursor.All(c.Context(), &products); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to parse products",
		})
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Products under Rp.100.000 fetched successfully",
		"data":    products,
	})
}
func GetBestSellers(c *fiber.Ctx) error {
	collection := config.MongoClient.Database("ecommerce").Collection("products")

	// Filter best sellers: Rating > 4.0 and Reviews > 1000
	filter := bson.M{
		"rating":  bson.M{"$gt": 4.0},
		"reviews": bson.M{"$gt": 1000},
	}

	cursor, err := collection.Find(context.Background(), filter)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to fetch best sellers",
		})
	}
	defer cursor.Close(context.Background())

	var products []model.Product
	if err := cursor.All(context.Background(), &products); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to parse best sellers",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Best sellers fetched successfully",
		"data":    products,
	})
}
package handler

import (
	"be_ecommerce/config"
	"be_ecommerce/model"
	"context"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// AddToFavorites menambahkan produk ke daftar favorit pengguna
func AddToFavorites(c *fiber.Ctx) error {
	var request struct {
		ProductID string `json:"product_id"`
		UserID    string `json:"user_id"`
	}

	// Parsing body request
	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid request body"})
	}

	// Validasi UserID dan ProductID
	if request.UserID == "" || request.ProductID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "User ID and Product ID are required"})
	}

	// Ambil koleksi favorit
	collection := config.MongoClient.Database("ecommerce").Collection("favorites")

	// Periksa apakah favorit sudah ada
	var favorite model.Favorite
	err := collection.FindOne(context.Background(), bson.M{"user_id": request.UserID}).Decode(&favorite)
	if err == mongo.ErrNoDocuments {
		// Jika tidak ada daftar favorit, buat baru
		favorite = model.Favorite{
			UserID:     request.UserID,
			ProductIDs: []string{request.ProductID},
		}
		_, err := collection.InsertOne(context.Background(), favorite)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to add product to favorites"})
		}
	} else if err == nil {
		// Jika favorit ada, tambahkan produk jika belum ada
		for _, id := range favorite.ProductIDs {
			if id == request.ProductID {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Product already in favorites"})
			}
		}
		favorite.ProductIDs = append(favorite.ProductIDs, request.ProductID)

		// Perbarui favorit
		_, err := collection.UpdateOne(context.Background(), bson.M{"user_id": request.UserID}, bson.M{
			"$set": bson.M{"product_ids": favorite.ProductIDs},
		})
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to update favorites"})
		}
	} else {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to fetch favorites"})
	}

	return c.JSON(fiber.Map{"message": "Product added to favorites successfully"})
}

func GetFavorites(c *fiber.Ctx) error {
	userID := c.Query("user_id")

	// Validasi UserID
	if userID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "user_id is required"})
	}

	// Ambil koleksi favorit
	collection := config.MongoClient.Database("ecommerce").Collection("favorites")

	// Cari favorit berdasarkan user_id
	var favorite model.Favorite
	err := collection.FindOne(context.Background(), bson.M{"user_id": userID}).Decode(&favorite)
	if err == mongo.ErrNoDocuments {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{"products": []string{}})
	} else if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to fetch favorites"})
	}

	// Konversi ProductIDs ke ObjectIDs
	var productObjectIDs []primitive.ObjectID
	for _, productID := range favorite.ProductIDs {
		objectID, err := primitive.ObjectIDFromHex(productID)
		if err != nil {
			continue // Abaikan ID produk yang tidak valid
		}
		productObjectIDs = append(productObjectIDs, objectID)
	}

	// Fetch detail produk berdasarkan ObjectIDs
	productCollection := config.MongoClient.Database("ecommerce").Collection("products")
	cursor, err := productCollection.Find(context.Background(), bson.M{"_id": bson.M{"$in": productObjectIDs}})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to fetch product details"})
	}
	defer cursor.Close(context.Background())

	var products []model.Product
	if err := cursor.All(context.Background(), &products); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to decode products"})
	}

	// Kembalikan produk favorit
	return c.JSON(fiber.Map{"products": products})
}


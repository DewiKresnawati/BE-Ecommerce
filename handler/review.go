package handler

import (
	"be_ecommerce/config"
	"be_ecommerce/model"
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// AddReview handles adding a new review
func AddReview(c *fiber.Ctx) error {
	var review model.Review
	if err := c.BodyParser(&review); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid request body",
		})
	}

	// Validasi ProductID
	if review.ProductID.IsZero() {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Product ID is required",
		})
	}

	// Validasi UserID
	if review.UserID.IsZero() {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "User ID is required",
		})
	}

	// Validasi nilai rating (harus antara 1.0 dan 5.0)
	if review.Rating < 1.0 || review.Rating > 5.0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Rating must be between 1.0 and 5.0",
		})
	}

	// Set waktu pembuatan
	review.ID = primitive.NewObjectID()
	review.CreatedAt = time.Now().Unix()

	// Simpan review ke database
	collection := config.MongoClient.Database("ecommerce").Collection("reviews")
	_, err := collection.InsertOne(context.Background(), review)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to save review",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message":   "Review added successfully",
		"review_id": review.ID,
	})
}

// GetReviews handles fetching all reviews for a product
func GetReviews(c *fiber.Ctx) error {
	productID := c.Params("product_id")

	// Konversi ProductID ke ObjectID
	objectID, err := primitive.ObjectIDFromHex(productID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid product ID format",
		})
	}

	// Ambil review dari database
	collection := config.MongoClient.Database("ecommerce").Collection("reviews")
	cursor, err := collection.Find(context.Background(), bson.M{"product_id": objectID})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to fetch reviews",
		})
	}
	defer cursor.Close(context.Background())

	var reviews []model.Review
	if err := cursor.All(context.Background(), &reviews); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to parse reviews",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Reviews fetched successfully",
		"data":    reviews,
	})
}

func GetProductRating(c *fiber.Ctx) error {
	productID := c.Params("product_id")

	// Konversi ProductID ke ObjectID
	objectID, err := primitive.ObjectIDFromHex(productID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid product ID format",
		})
	}

	// Agregasi untuk menghitung rata-rata rating dan jumlah ulasan
	collection := config.MongoClient.Database("ecommerce").Collection("reviews")
	cursor, err := collection.Aggregate(context.Background(), bson.A{
		bson.M{"$match": bson.M{"product_id": objectID}}, // Filter ulasan untuk produk tertentu
		bson.M{"$group": bson.M{
			"_id":        "$product_id",
			"avg_rating": bson.M{"$avg": "$rating"},
			"review_count": bson.M{"$sum": 1},
		}},
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to calculate product rating",
		})
	}
	defer cursor.Close(context.Background())

	var results []bson.M
	if err := cursor.All(context.Background(), &results); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to parse rating data",
		})
	}

	// Jika tidak ada ulasan
	if len(results) == 0 {
		return c.JSON(fiber.Map{
			"product_id":   productID,
			"avg_rating":   0,
			"review_count": 0,
		})
	}

	// Kembalikan rata-rata rating dan jumlah ulasan
	return c.JSON(fiber.Map{
		"product_id":   productID,
		"avg_rating":   results[0]["avg_rating"],
		"review_count": results[0]["review_count"],
	})
}

// UpdateReview handles updating an existing review
func UpdateReview(c *fiber.Ctx) error {
	reviewID := c.Params("review_id")

	// Konversi ReviewID ke ObjectID
	objectID, err := primitive.ObjectIDFromHex(reviewID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid review ID format",
		})
	}

	var updateData struct {
		Rating  float64 `json:"rating"`
		Comment string  `json:"comment"`
	}

	if err := c.BodyParser(&updateData); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid request body",
		})
	}

	// Validasi nilai rating (harus antara 1.0 dan 5.0)
	if updateData.Rating < 1.0 || updateData.Rating > 5.0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Rating must be between 1.0 and 5.0",
		})
	}

	// Update review di database
	collection := config.MongoClient.Database("ecommerce").Collection("reviews")
	_, err = collection.UpdateOne(context.Background(), bson.M{"_id": objectID}, bson.M{
		"$set": bson.M{
			"rating":  updateData.Rating,
			"comment": updateData.Comment,
		},
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to update review",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Review updated successfully",
	})
}

// DeleteReview handles deleting a review
func DeleteReview(c *fiber.Ctx) error {
	reviewID := c.Params("review_id")

	// Konversi ReviewID ke ObjectID
	objectID, err := primitive.ObjectIDFromHex(reviewID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid review ID format",
		})
	}

	// Hapus review dari database
	collection := config.MongoClient.Database("ecommerce").Collection("reviews")
	_, err = collection.DeleteOne(context.Background(), bson.M{"_id": objectID})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to delete review",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Review deleted successfully",
	})
}

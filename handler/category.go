package handler

import (
	"be_ecommerce/config"
	"be_ecommerce/model"
	"context"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// AddCategory handles adding a new category
func AddCategory(c *fiber.Ctx) error {
	var category model.Category
	if err := c.BodyParser(&category); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid request body",
		})
	}

	// Set ID untuk kategori
	category.ID = primitive.NewObjectID()

	// Simpan kategori ke database
	collection := config.MongoClient.Database("ecommerce").Collection("categories")
	_, err := collection.InsertOne(context.Background(), category)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to save category",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message":    "Category added successfully",
		"category_id": category.ID,
	})
}

// AddSubCategory handles adding a new sub-category to an existing category
func AddSubCategory(c *fiber.Ctx) error {
	var request struct {
		CategoryID primitive.ObjectID `json:"category_id"`
		Name       string             `json:"name"`
	}

	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid request body",
		})
	}

	// Periksa apakah kategori ada
	collection := config.MongoClient.Database("ecommerce").Collection("categories")
	var category model.Category
	err := collection.FindOne(context.Background(), bson.M{"_id": request.CategoryID}).Decode(&category)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "Category not found",
		})
	}

	// Tambahkan sub-kategori
	subCategory := model.SubCategory{
		ID:   primitive.NewObjectID(),
		Name: request.Name,
	}
	category.SubCategories = append(category.SubCategories, subCategory)

	// Perbarui kategori dengan sub-kategori baru
	_, err = collection.UpdateOne(context.Background(), bson.M{"_id": request.CategoryID}, bson.M{
		"$set": bson.M{"sub_categories": category.SubCategories},
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to add sub-category",
		})
	}

	return c.JSON(fiber.Map{
		"message":        "Sub-category added successfully",
		"sub_category_id": subCategory.ID,
	})
}

// GetCategories handles fetching all categories and their sub-categories
func GetCategories(c *fiber.Ctx) error {
	collection := config.MongoClient.Database("ecommerce").Collection("categories")

	cursor, err := collection.Find(context.Background(), bson.M{})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to fetch categories",
		})
	}
	defer cursor.Close(context.Background())

	var categories []model.Category
	if err := cursor.All(context.Background(), &categories); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to parse categories",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Categories fetched successfully",
		"data":    categories,
	})
}

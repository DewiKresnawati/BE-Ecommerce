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
			"error":   err.Error(),
		})
	}

	// Validasi duplikasi kategori berdasarkan nama
	collection := config.MongoClient.Database("ecommerce").Collection("categories")
	existingCategory := model.Category{}
	err := collection.FindOne(context.Background(), bson.M{"name": category.Name}).Decode(&existingCategory)
	if err == nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"message": "Category already exists",
		})
	}

	// Set ID untuk kategori
	category.ID = primitive.NewObjectID()

	// Simpan kategori ke database
	_, err = collection.InsertOne(context.Background(), category)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to save category",
			"error":   err.Error(),
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
			"error":   err.Error(),
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

	// Validasi jika sub-kategori dengan nama yang sama sudah ada
	for _, subCategory := range category.SubCategories {
		if subCategory.Name == request.Name {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"message": "Sub-category already exists in this category",
			})
		}
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
			"error":   err.Error(),
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
			"error":   err.Error(),
		})
	}
	defer cursor.Close(context.Background())

	var categories []model.Category
	if err := cursor.All(context.Background(), &categories); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to parse categories",
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Categories fetched successfully",
		"data":    categories,
	})
}
// UpdateCategory handles updating a category by its ID
func UpdateCategory(c *fiber.Ctx) error {
	id := c.Params("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid category ID",
		})
	}

	var payload struct {
		Name string `json:"name"`
	}

	if err := c.BodyParser(&payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid request body",
		})
	}

	collection := config.MongoClient.Database("ecommerce").Collection("categories")
	filter := bson.M{"_id": objectID}
	update := bson.M{"$set": bson.M{"name": payload.Name}}

	_, err = collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to update category",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Category updated successfully",
	})
}

// UpdateSubCategory handles updating a sub-category by its ID
func UpdateSubCategory(c *fiber.Ctx) error {
	id := c.Params("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid sub-category ID",
		})
	}

	var payload struct {
		CategoryID primitive.ObjectID `json:"category_id"`
		Name       string             `json:"name"`
	}

	if err := c.BodyParser(&payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid request body",
		})
	}

	collection := config.MongoClient.Database("ecommerce").Collection("categories")
	filter := bson.M{"_id": payload.CategoryID, "sub_categories._id": objectID}
	update := bson.M{"$set": bson.M{"sub_categories.$.name": payload.Name}}

	_, err = collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to update sub-category",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Sub-category updated successfully",
	})
}

// DeleteCategory handles deleting a category by its ID
func DeleteCategory(c *fiber.Ctx) error {
	id := c.Params("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid category ID",
		})
	}

	collection := config.MongoClient.Database("ecommerce").Collection("categories")
	_, err = collection.DeleteOne(context.Background(), bson.M{"_id": objectID})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to delete category",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Category deleted successfully",
	})
}

// DeleteSubCategory handles deleting a sub-category by its ID
func DeleteSubCategory(c *fiber.Ctx) error {
	id := c.Params("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid sub-category ID",
		})
	}

	var payload struct {
		CategoryID primitive.ObjectID `json:"category_id"`
	}

	if err := c.BodyParser(&payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid request body",
		})
	}

	collection := config.MongoClient.Database("ecommerce").Collection("categories")
	filter := bson.M{"_id": payload.CategoryID}
	update := bson.M{
		"$pull": bson.M{"sub_categories": bson.M{"_id": objectID}},
	}

	_, err = collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to delete sub-category",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Sub-category deleted successfully",
	})
}
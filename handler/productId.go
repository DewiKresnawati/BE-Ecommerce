package handler

import (
	"be_ecommerce/config"
	"be_ecommerce/model"
	"context"
	"fmt"
	"net/http"
	"strconv"

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
	categoryCollection := config.MongoClient.Database("ecommerce").Collection("categories")
	sellerCollection := config.MongoClient.Database("ecommerce").Collection("sellers")

	// Cari produk berdasarkan ID
	var product model.Product
	err = productCollection.FindOne(context.Background(), bson.M{"_id": objectID}).Decode(&product)
	if err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"status":  "error",
			"message": "Product not found",
		})
	}

	// Ambil kategori
	var category model.Category
	if err := categoryCollection.FindOne(context.Background(), bson.M{"_id": product.CategoryID}).Decode(&category); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to fetch category",
		})
	}

	// Cari sub-kategori
	var subCategoryName string
	for _, sub := range category.SubCategories {
		if sub.ID == product.SubCategoryID {
			subCategoryName = sub.Name
			break
		}
	}

	// Ambil data toko
	var store model.StoreInfo
	if err := sellerCollection.FindOne(context.Background(), bson.M{"_id": product.SellerID}).Decode(&store); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to fetch store",
		})
	}

	// Kembalikan data produk dengan detail tambahan
	return c.JSON(fiber.Map{
		"status": "success",
		"product": fiber.Map{
			"id":           product.ID.Hex(),
			"name":         product.Name,
			"price":        product.Price,
			"discount":     product.Discount,
			"image":        product.Image,
			"description":  product.Description,
			"category":     category.Name,
			"sub_category": subCategoryName,
			"seller_id": 	product.SellerID,
		},
		"store": fiber.Map{
			"store_name":   store.StoreName,
			"nik":          store.NIK,
			"photoselfie":  store.PhotoSelfie,
			"full_address": store.FullAddress,
		},
	})
}
func UpdateProductByID(c *fiber.Ctx) error {
	// Ambil ID dari URL parameter
	productID := c.Params("id")

	// Konversi ID ke ObjectID MongoDB
	objectID, err := primitive.ObjectIDFromHex(productID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid product ID format",
			"error":   err.Error(),
		})
	}

	// Parse multipart form
	form, err := c.MultipartForm()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid form data",
			"error":   err.Error(),
		})
	}

	// Extract fields and validate
	name := form.Value["name"]
	if len(name) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Product name is required",
		})
	}

	price, err := strconv.ParseInt(form.Value["price"][0], 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid price format",
			"error":   err.Error(),
		})
	}

	discount, err := strconv.ParseInt(form.Value["discount"][0], 10, 64)
	if err != nil {
		discount = 0 // Default discount jika tidak valid
	}

	categoryID, err := primitive.ObjectIDFromHex(form.Value["category_id"][0])
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid Category ID format",
			"error":   err.Error(),
		})
	}

	subCategoryID, err := primitive.ObjectIDFromHex(form.Value["sub_category_id"][0])
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid Sub-Category ID format",
			"error":   err.Error(),
		})
	}

	description := form.Value["description"]
	if len(description) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Description is required",
		})
	}

	// Handle file upload
	var imagePath string
	fileHeaders := form.File["image"]
	if len(fileHeaders) > 0 {
		file := fileHeaders[0]
		imagePath = fmt.Sprintf("./uploads/%s", file.Filename)

		// Save image file
		if err := c.SaveFile(file, imagePath); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": "Failed to save image",
				"error":   err.Error(),
			})
		}
	} else {
		// Jika tidak ada gambar baru, gunakan gambar lama
		productCollection := config.MongoClient.Database("ecommerce").Collection("products")
		var existingProduct model.Product
		err := productCollection.FindOne(context.Background(), bson.M{"_id": objectID}).Decode(&existingProduct)
		if err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"message": "Product not found",
				"error":   err.Error(),
			})
		}
		imagePath = existingProduct.Image
	}

	// Update data produk
	updateData := bson.M{
		"name":            name[0],
		"price":           price,
		"discount":        discount,
		"category_id":     categoryID,
		"sub_category_id": subCategoryID,
		"description":     description[0],
		"image":           imagePath,
	}

	// Update produk di database
	productCollection := config.MongoClient.Database("ecommerce").Collection("products")
	_, err = productCollection.UpdateOne(context.Background(), bson.M{"_id": objectID}, bson.M{"$set": updateData})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to update product",
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Product updated successfully",
		"status":  "success",
	})
}

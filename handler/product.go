package handler

import (
	"be_ecommerce/config"
	"be_ecommerce/model"
	"be_ecommerce/utils"
	"context"
	"log"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// CreateProduct handles creating a new product
func CreateProduct(c *fiber.Ctx) error {
	var product model.Product
	if err := c.BodyParser(&product); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid request body",
		})
	}

	// Validasi SellerID
	if product.SellerID.IsZero() {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Seller ID is required",
		})
	}

	// Periksa apakah seller valid dan tokonya aktif
	userCollection := config.MongoClient.Database("ecommerce").Collection("users")
	var seller model.User
	err := userCollection.FindOne(context.Background(), bson.M{"_id": product.SellerID}).Decode(&seller)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid Seller ID",
		})
	}

	if seller.StoreStatus == nil || *seller.StoreStatus != "approved" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Store is not active or approved",
		})
	}

	// Validasi CategoryID dan SubCategoryID
	categoryCollection := config.MongoClient.Database("ecommerce").Collection("categories")
	var category model.Category
	err = categoryCollection.FindOne(context.Background(), bson.M{"_id": product.CategoryID}).Decode(&category)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid Category ID",
		})
	}

	// Periksa apakah SubCategoryID valid dalam kategori ini
	subCategoryExists := false
	for _, subCat := range category.SubCategories {
		if subCat.ID == product.SubCategoryID {
			subCategoryExists = true
			break
		}
	}
	if !subCategoryExists {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid Sub-Category ID",
		})
	}

	// Simpan produk ke database
	productCollection := config.MongoClient.Database("ecommerce").Collection("products")
	product.ID = primitive.NewObjectID()
	result, err := productCollection.InsertOne(context.Background(), product)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Error saving product to database",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message":    "Product created successfully",
		"product_id": result.InsertedID,
	})
}

func GetProductDetail(c *fiber.Ctx) error {
	productID := c.Params("id")

	// Konversi productID ke ObjectID
	objectID, err := primitive.ObjectIDFromHex(productID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid product ID format",
		})
	}

	// Ambil produk dari database
	productCollection := config.MongoClient.Database("ecommerce").Collection("products")
	var product model.Product
	err = productCollection.FindOne(context.Background(), bson.M{"_id": objectID}).Decode(&product)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "Product not found",
		})
	}

	// Ambil detail toko berdasarkan SellerID
	userCollection := config.MongoClient.Database("ecommerce").Collection("users")
	var seller model.User
	err = userCollection.FindOne(context.Background(), bson.M{"_id": product.SellerID}).Decode(&seller)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "Seller not found",
		})
	}

	// Pastikan toko aktif (store_status == "approved")
	if seller.StoreStatus != nil && *seller.StoreStatus != "approved" {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "Store is not active",
		})
	}

	// Ambil detail kategori dan sub-kategori
	categoryCollection := config.MongoClient.Database("ecommerce").Collection("categories")
	var category model.Category
	err = categoryCollection.FindOne(context.Background(), bson.M{"_id": product.CategoryID}).Decode(&category)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "Category not found",
		})
	}

	// Cari sub-kategori dalam kategori
	var subCategoryName string
	for _, subCat := range category.SubCategories {
		if subCat.ID == product.SubCategoryID {
			subCategoryName = subCat.Name
			break
		}
	}
	if subCategoryName == "" {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "Sub-category not found",
		})
	}

	// Gabungkan data produk, kategori, dan toko
	response := fiber.Map{
		"product": fiber.Map{
			"id":           product.ID,
			"name":         product.Name,
			"price":        product.Price,
			"discount":     product.Discount,
			"category":     category.Name,
			"sub_category": subCategoryName,
			"description":  product.Description,
			"image":        product.Image,
			"rating":       product.Rating,
			"reviews":      product.Reviews,
		},
		"store": fiber.Map{
			"store_name":   seller.StoreInfo.StoreName,
			"full_address": seller.StoreInfo.FullAddress,
			"seller_email": seller.Email,
			"store_status": seller.StoreStatus,
			"seller_id":    seller.ID,
		},
	}

	return c.JSON(response)
}

// GetAllProducts fetches all products
func GetAllProducts(c *fiber.Ctx) error {
	// Ambil koleksi produk dari MongoDB
	collection := config.MongoClient.Database("ecommerce").Collection("products")

	// Ambil semua produk dari koleksi
	cursor, err := collection.Find(c.Context(), bson.M{})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to fetch products",
		})
	}
	defer cursor.Close(c.Context())

	// Decode produk ke slice
	var products []model.Product
	if err := cursor.All(c.Context(), &products); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to parse products",
		})
	}

	// Kembalikan daftar produk
	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Products fetched successfully",
		"data":    products,
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

func GetProductsByUserID(c *fiber.Ctx) error {
	// Ambil user_id dari parameter query atau autentikasi JWT
	userIDParam := c.Query("user_id")
	if userIDParam == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "User ID is required",
		})
	}

	// Konversi user_id ke ObjectID
	userID, err := primitive.ObjectIDFromHex(userIDParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid User ID format",
		})
	}

	// Ambil koleksi produk
	productCollection := config.MongoClient.Database("ecommerce").Collection("products")

	// Filter produk berdasarkan user_id
	filter := bson.M{"seller_id": userID}
	cursor, err := productCollection.Find(c.Context(), filter)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to fetch products",
			"error":   err.Error(),
		})
	}
	defer cursor.Close(c.Context())

	// Decode produk
	var products []model.Product
	if err := cursor.All(c.Context(), &products); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to parse products",
			"error":   err.Error(),
		})
	}

	// Kembalikan data produk
	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Products fetched successfully",
		"data":    products,
	})
}
func CreateProductForSeller(c *fiber.Ctx) error {
	// Ambil token dari header Authorization
	token := c.Get("Authorization")
	if token == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Unauthorized: Missing token",
		})
	}

	// Validasi token dan ambil klaim
	claims, err := utils.ValidateJWT(token[7:]) // Hapus prefix "Bearer "
	if err != nil {
		log.Println("Error validating token:", err)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Unauthorized: Invalid token",
		})
	}

	// Ambil user_id dari klaim
	userID, ok := claims["user_id"].(string)
	if !ok || userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Unauthorized: Invalid user ID",
		})
	}

	// Konversi user_id ke ObjectID
	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid User ID format",
		})
	}

	// Periksa apakah user adalah seller
	userCollection := config.MongoClient.Database("ecommerce").Collection("users")
	var seller model.User
	err = userCollection.FindOne(context.Background(), bson.M{"_id": objectID, "roles": "seller"}).Decode(&seller)
	if err != nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"message": "Forbidden: User is not a seller",
		})
	}

	// Periksa apakah toko aktif
	if seller.StoreStatus == nil || *seller.StoreStatus != "approved" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"message": "Forbidden: Store is not active or approved",
		})
	}

	// Parse request body
	var product model.Product
	if err := c.BodyParser(&product); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid request body",
		})
	}

	// Tambahkan SellerID ke produk
	product.SellerID = objectID

	// Validasi CategoryID dan SubCategoryID
	categoryCollection := config.MongoClient.Database("ecommerce").Collection("categories")
	var category model.Category
	err = categoryCollection.FindOne(context.Background(), bson.M{"_id": product.CategoryID}).Decode(&category)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid Category ID",
		})
	}

	// Periksa apakah SubCategoryID valid
	subCategoryExists := false
	for _, subCat := range category.SubCategories {
		if subCat.ID == product.SubCategoryID {
			subCategoryExists = true
			break
		}
	}
	if !subCategoryExists {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid Sub-Category ID",
		})
	}

	// Simpan produk ke database
	productCollection := config.MongoClient.Database("ecommerce").Collection("products")
	product.ID = primitive.NewObjectID()
	result, err := productCollection.InsertOne(context.Background(), product)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Error saving product to database",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message":    "Product created successfully",
		"product_id": result.InsertedID,
	})
}
package handler

import (
	"be_ecommerce/config"
	"be_ecommerce/model"
	"be_ecommerce/utils"
	"context"
	"fmt"
	"log"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func CreateProduct(c *fiber.Ctx) error {
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

	// Konversi price dari int64 ke int
	price64, err := strconv.ParseInt(form.Value["price"][0], 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid price format",
			"error":   err.Error(),
		})
	}
	price := int(price64) // Konversi ke int

	// Konversi discount dari int64 ke int
	discount64, err := strconv.ParseInt(form.Value["discount"][0], 10, 64)
	if err != nil {
		discount64 = 0 // Default discount jika tidak valid
	}
	discount := int(discount64) // Konversi ke int

	sellerID, err := primitive.ObjectIDFromHex(form.Value["seller_id"][0])
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid Seller ID format",
			"error":   err.Error(),
		})
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
		imagePath = "" // Default jika tidak ada gambar
	}

	// Prepare product data
	product := model.Product{
		ID:            primitive.NewObjectID(),
		Name:          name[0],
		Price:         price,     // Menggunakan konversi ke int
		Discount:      discount,  // Menggunakan konversi ke int
		SellerID:      sellerID,
		CategoryID:    categoryID,
		SubCategoryID: subCategoryID,
		Description:   description[0],
		Image:         imagePath,
	}

	// Save product to database
	productCollection := config.MongoClient.Database("ecommerce").Collection("products")
	result, err := productCollection.InsertOne(context.Background(), product)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to save product",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message":    "Product created successfully",
		"product_id": result.InsertedID,
	})
}


func DeleteProductByID(c *fiber.Ctx) error {
	// Ambil ID produk dari parameter URL
	productID := c.Params("id")

	// Konversi ID ke ObjectID MongoDB
	objectID, err := primitive.ObjectIDFromHex(productID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid product ID format",
			"error":   err.Error(),
		})
	}

	// Ambil koleksi produk dari database
	productCollection := config.MongoClient.Database("ecommerce").Collection("products")

	// Cari dan hapus produk berdasarkan ID
	result, err := productCollection.DeleteOne(context.Background(), bson.M{"_id": objectID})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to delete product",
			"error":   err.Error(),
		})
	}

	// Periksa apakah produk ditemukan dan dihapus
	if result.DeletedCount == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "Product not found",
		})
	}

	// Berikan respons berhasil
	return c.JSON(fiber.Map{
		"message": "Product deleted successfully",
		"status":  "success",
	})
}
func CreateSellerProduct(c *fiber.Ctx) error {
	// Ambil seller_id dari token (middleware JWT harus sudah diterapkan sebelumnya)
	sellerID, err := utils.GetUserIDFromToken(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Unauthorized", "error": err.Error()})
	}

	// Parse multipart form
	form, err := c.MultipartForm()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid form data", "error": err.Error()})
	}

	// Ambil dan validasi field
	name := form.Value["name"]
	if len(name) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Product name is required"})
	}

	price, err := strconv.Atoi(form.Value["price"][0])
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid price format", "error": err.Error()})
	}

	discount, _ := strconv.Atoi(form.Value["discount"][0])

	categoryID, _ := primitive.ObjectIDFromHex(form.Value["category_id"][0])
	subCategoryID, _ := primitive.ObjectIDFromHex(form.Value["sub_category_id"][0])

	description := form.Value["description"][0]

	// Handle file upload
	var imagePath string
	fileHeaders := form.File["image"]
	if len(fileHeaders) > 0 {
		file := fileHeaders[0]
		imagePath = fmt.Sprintf("./uploads/%s", file.Filename)

		// Simpan file gambar
		if err := c.SaveFile(file, imagePath); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to save image", "error": err.Error()})
		}
	} else {
		imagePath = ""
	}

	// Simpan produk ke database
	product := model.Product{
		ID:            primitive.NewObjectID(),
		Name:          name[0],
		Price:         price,
		Discount:      discount,
		SellerID:      sellerID,
		CategoryID:    categoryID,
		SubCategoryID: subCategoryID,
		Description:   description,
		Image:         imagePath,
	}

	productCollection := config.MongoClient.Database("ecommerce").Collection("products")
	result, err := productCollection.InsertOne(context.Background(), product)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to save product", "error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "Product created successfully", "product_id": result.InsertedID})
}

// **2. Update Product for Seller**
func UpdateSellerProductByID(c *fiber.Ctx) error {
	// Ambil seller_id dari token
	sellerID, err := utils.GetUserIDFromToken(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Unauthorized", "error": err.Error()})
	}

	// Ambil ID produk dari parameter URL
	productID := c.Params("id")
	objectID, err := primitive.ObjectIDFromHex(productID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid product ID format", "error": err.Error()})
	}

	// Parse form data
	form, err := c.MultipartForm()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid form data", "error": err.Error()})
	}

	name := form.Value["name"][0]
	price, _ := strconv.Atoi(form.Value["price"][0])
	discount, _ := strconv.Atoi(form.Value["discount"][0])
	categoryID, _ := primitive.ObjectIDFromHex(form.Value["category_id"][0])
	subCategoryID, _ := primitive.ObjectIDFromHex(form.Value["sub_category_id"][0])
	description := form.Value["description"][0]

	// Handle file upload
	var imagePath string
	fileHeaders := form.File["image"]
	if len(fileHeaders) > 0 {
		file := fileHeaders[0]
		imagePath = fmt.Sprintf("./uploads/%s", file.Filename)
		c.SaveFile(file, imagePath)
	}

	updateData := bson.M{
		"name":            name,
		"price":           price,
		"discount":        discount,
		"category_id":     categoryID,
		"sub_category_id": subCategoryID,
		"description":     description,
	}

	if imagePath != "" {
		updateData["image"] = imagePath
	}

	productCollection := config.MongoClient.Database("ecommerce").Collection("products")

	// Pastikan produk dimiliki oleh seller yang sedang login
	result, err := productCollection.UpdateOne(context.Background(), bson.M{"_id": objectID, "seller_id": sellerID}, bson.M{"$set": updateData})
	if err != nil || result.MatchedCount == 0 {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"message": "Failed to update product or product not found"})
	}

	return c.JSON(fiber.Map{"message": "Product updated successfully"})
}

// **3. Delete Product for Seller**
func DeleteSellerProductByID(c *fiber.Ctx) error {
	// Ambil seller_id dari token
	sellerID, err := utils.GetUserIDFromToken(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Unauthorized", "error": err.Error()})
	}

	// Ambil ID produk dari parameter URL
	productID := c.Params("id")
	objectID, err := primitive.ObjectIDFromHex(productID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid product ID format", "error": err.Error()})
	}

	productCollection := config.MongoClient.Database("ecommerce").Collection("products")

	// Pastikan produk dimiliki oleh seller yang sedang login
	result, err := productCollection.DeleteOne(context.Background(), bson.M{"_id": objectID, "seller_id": sellerID})
	if err != nil || result.DeletedCount == 0 {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"message": "Failed to delete product or product not found"})
	}

	return c.JSON(fiber.Map{"message": "Product deleted successfully"})
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

	// Pipeline agregasi
	pipeline := mongo.Pipeline{
		// Lookup kategori berdasarkan category_id
		{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "categories"},          // Koleksi yang di-lookup
			{Key: "localField", Value: "category_id"},   // Field di koleksi "products"
			{Key: "foreignField", Value: "_id"},         // Field di koleksi "categories"
			{Key: "as", Value: "category"},              // Alias hasil lookup
		}}},
		// Lookup sub-kategori berdasarkan sub_category_id
		{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "categories"},          // Koleksi yang di-lookup
			{Key: "localField", Value: "sub_category_id"},
			{Key: "foreignField", Value: "_id"},
			{Key: "as", Value: "sub_category"},
		}}},
		// Hilangkan array kategori dan sub-kategori (ubah menjadi objek tunggal)
		{{Key: "$unwind", Value: bson.D{{Key: "path", Value: "$category"}, {Key: "preserveNullAndEmptyArrays", Value: true}}}},
		{{Key: "$unwind", Value: bson.D{{Key: "path", Value: "$sub_category"}, {Key: "preserveNullAndEmptyArrays", Value: true}}}},
	}

	// Jalankan agregasi
	cursor, err := collection.Aggregate(c.Context(), pipeline)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to fetch products with categories",
		})
	}
	defer cursor.Close(c.Context())

	// Decode hasil agregasi ke slice
	var products []bson.M
	if err := cursor.All(c.Context(), &products); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to parse products",
		})
	}

	// Kembalikan daftar produk dengan kategori dan sub-kategori
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

	// Parse multipart form untuk upload gambar
	form, err := c.MultipartForm()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid form data",
			"error":   err.Error(),
		})
	}

	// Ambil data dari form
	name := form.Value["name"][0]
	price, _ := strconv.Atoi(form.Value["price"][0])
	discount, _ := strconv.Atoi(form.Value["discount"][0])
	stock, _ := strconv.Atoi(form.Value["stock"][0])
	categoryID, _ := primitive.ObjectIDFromHex(form.Value["category_id"][0])
	subCategoryID, _ := primitive.ObjectIDFromHex(form.Value["sub_category_id"][0])
	description := form.Value["description"][0]

	// Validasi kategori dan subkategori
	categoryCollection := config.MongoClient.Database("ecommerce").Collection("categories")
	var category model.Category
	err = categoryCollection.FindOne(context.Background(), bson.M{"_id": categoryID}).Decode(&category)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid Category ID",
		})
	}

	subCategoryExists := false
	for _, subCat := range category.SubCategories {
		if subCat.ID == subCategoryID {
			subCategoryExists = true
			break
		}
	}
	if !subCategoryExists {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid Sub-Category ID",
		})
	}

	// Handle file upload (jika ada)
	var imagePath string
	fileHeaders := form.File["image"]
	if len(fileHeaders) > 0 {
		file := fileHeaders[0]
		imagePath = fmt.Sprintf("uploads/%s", file.Filename)

		// Simpan file ke folder `uploads/`
		if err := c.SaveFile(file, imagePath); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": "Failed to save image",
				"error":   err.Error(),
			})
		}
	} else {
		imagePath = "uploads/default.png" // Gambar default jika tidak ada gambar diunggah
	}

	// Simpan produk ke database
	product := model.Product{
		ID:            primitive.NewObjectID(),
		Name:          name,
		Price:         price,
		Discount:      discount,
		Stock:         stock,
		SellerID:      objectID,
		CategoryID:    categoryID,
		SubCategoryID: subCategoryID,
		Description:   description,
		Image:         imagePath,
	}

	productCollection := config.MongoClient.Database("ecommerce").Collection("products")
	result, err := productCollection.InsertOne(context.Background(), product)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Error saving product to database",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message":    "Product created successfully",
		"product_id": result.InsertedID,
		"image_url":  fmt.Sprintf("%s/%s", "http://localhost:3000", imagePath),
	})
}

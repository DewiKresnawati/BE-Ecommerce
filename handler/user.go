package handler

import (
	"be_ecommerce/config"
	"be_ecommerce/model"
	"context"
	"fmt"
	"log"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

// Helper function to get user collection
func getUserCollection() *mongo.Collection {
	return config.MongoClient.Database("ecommerce").Collection("users")
}

// CRUD for Customers
func GetCustomers(c *fiber.Ctx) error {
	collection := getUserCollection()

	// Query untuk mendapatkan semua pelanggan dengan role "customer"
	filter := bson.M{"roles": "customer"}
	cursor, err := collection.Find(context.Background(), filter)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Error fetching customers",
		})
	}
	defer cursor.Close(context.Background())

	// Parsing hasil query ke dalam slice
	var customers []bson.M
	if err := cursor.All(context.Background(), &customers); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Error parsing customers",
		})
	}

	// Transform data untuk menyederhanakan output
	transformedCustomers := make([]map[string]interface{}, 0)
	for _, customer := range customers {
		transformed := map[string]interface{}{
			"id":        customer["_id"],
			"username":  customer["username"],
			"email":     customer["email"],
			"roles":     customer["roles"],
			"suspended": false, // Default jika tidak ada informasi tentang suspend
		}

		// Hanya tambahkan informasi `suspended` jika relevan
		if suspended, ok := customer["suspended"].(bool); ok {
			transformed["suspended"] = suspended
		}

		transformedCustomers = append(transformedCustomers, transformed)
	}

	// Kembalikan data ke frontend
	return c.JSON(transformedCustomers)
}

// Fungsi untuk hashing password
func hashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

func CreateCustomer(c *fiber.Ctx) error {
	var user model.User
	if err := c.BodyParser(&user); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid request body",
		})
	}

	// Hashing password
	hashedPassword, err := hashPassword(user.Password)
	if err != nil {
		log.Println("Error hashing password:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Error processing password",
		})
	}
	user.Password = hashedPassword

	// Set roles and ID
	user.Roles = []string{"customer"}
	user.ID = primitive.NewObjectID()

	// Simpan ke database
	collection := getUserCollection()
	_, err = collection.InsertOne(context.Background(), user)
	if err != nil {
		log.Println("Error creating customer:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Error creating customer",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Customer created successfully",
	})
}
func UpdateCustomer(c *fiber.Ctx) error {
	// Parsing body request
	var body struct {
		UserID  string                 `json:"user_id"` // ID pengguna
		Updates map[string]interface{} `json:"updates"` // Data yang akan diupdate
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid request body",
		})
	}

	// Validasi ID pengguna
	if body.UserID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "User ID is required",
		})
	}

	// Konversi UserID ke ObjectID
	userID, err := primitive.ObjectIDFromHex(body.UserID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid User ID format",
		})
	}

	// Validasi: pastikan tidak ada field sensitif yang diperbarui
	disallowedFields := []string{"_id", "password", "roles"}
	for _, field := range disallowedFields {
		if _, exists := body.Updates[field]; exists {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": fmt.Sprintf("Field '%s' cannot be updated", field),
			})
		}
	}

	// Pastikan data yang akan diupdate tidak kosong
	if len(body.Updates) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "No updates provided",
		})
	}

	// Update pengguna di database
	collection := getUserCollection()
	filter := bson.M{"_id": userID, "roles": "customer"}
	update := bson.M{"$set": body.Updates}

	result, err := collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		log.Println("Error updating customer:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Error updating customer",
		})
	}

	// Periksa apakah ada dokumen yang diperbarui
	if result.MatchedCount == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "Customer not found or no changes made",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Customer updated successfully",
	})
}

func DeleteCustomer(c *fiber.Ctx) error {
	id := c.Params("id")
	userID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid ID",
		})
	}

	collection := getUserCollection()
	_, err = collection.DeleteOne(context.Background(), bson.M{"_id": userID, "roles": "customer"})
	if err != nil {
		log.Println("Error deleting customer:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Error deleting customer",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Customer deleted successfully",
	})
}

// CRUD for Sellers
func GetSellers(c *fiber.Ctx) error {
	collection := getUserCollection()

	// Query to fetch all users with the role "seller"
	filter := bson.M{"roles": "seller"}
	cursor, err := collection.Find(context.Background(), filter)
	if err != nil {
		log.Println("Error fetching sellers:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Error fetching sellers",
		})
	}
	defer cursor.Close(context.Background())

	// Parse query results into a slice
	var sellers []bson.M
	if err := cursor.All(context.Background(), &sellers); err != nil {
		log.Println("Error decoding sellers:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Error decoding sellers",
		})
	}

	// Transform the data for simplified output
	transformedSellers := make([]map[string]interface{}, 0)
	for _, seller := range sellers {
		transformed := map[string]interface{}{
			"id":           seller["_id"],
			"username":     seller["username"],
			"email":        seller["email"],
			"roles":        seller["roles"],
			"store_status": seller["store_status"],
			"store_info": map[string]interface{}{
				"store_name":   getStringOrDefault(seller, "store_info", "store_name"),
				"full_address": getStringOrDefault(seller, "store_info", "full_address"),
				"nik":          getStringOrDefault(seller, "store_info", "nik"),
			},
		}

		transformedSellers = append(transformedSellers, transformed)
	}

	// Return data to the frontend
	return c.JSON(fiber.Map{
		"data":    transformedSellers,
		"message": "Sellers fetched successfully",
	})
}

// Helper function to safely get nested strings
func getStringOrDefault(doc bson.M, key string, nestedKey string) string {
	if outer, ok := doc[key].(bson.M); ok {
		if value, ok := outer[nestedKey].(string); ok {
			return value
		}
	}
	return ""
}

func GetSellerByID(c *fiber.Ctx) error {
	id := c.Params("id")

	// Konversi ID ke ObjectID MongoDB
	userID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid seller ID",
		})
	}

	// Cari seller berdasarkan ID dan role "seller"
	collection := getUserCollection()
	var seller model.User
	err = collection.FindOne(context.Background(), bson.M{"_id": userID, "roles": "seller"}).Decode(&seller)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "Seller not found",
		})
	}

	// Siapkan respons
	response := fiber.Map{
		"id":       seller.ID.Hex(),
		"username": seller.Username,
		"email":    seller.Email,
		"store_name": func() string {
			if seller.StoreInfo != nil {
				return seller.StoreInfo.StoreName
			}
			return ""
		}(),
		"full_address": func() string {
			if seller.StoreInfo != nil {
				return seller.StoreInfo.FullAddress
			}
			return ""
		}(),
		"nik": func() string {
			if seller.StoreInfo != nil {
				return seller.StoreInfo.NIK
			}
			return ""
		}(),
		"store_status": func() string {
			if seller.StoreStatus != nil {
				return *seller.StoreStatus
			}
			return "unknown"
		}(),
		"suspended": func() bool {
			if seller.StoreStatus != nil && *seller.StoreStatus == "suspended" {
				return true
			}
			return false
		}(),
	}

	return c.JSON(response)
}

func CreateSeller(c *fiber.Ctx) error {
	var user model.User
	if err := c.BodyParser(&user); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid request body",
		})
	}

	user.Roles = []string{"seller"}
	user.ID = primitive.NewObjectID()
	collection := getUserCollection()
	_, err := collection.InsertOne(context.Background(), user)
	if err != nil {
		log.Println("Error creating seller:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Error creating seller",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Seller created successfully",
	})
}

func UpdateSeller(c *fiber.Ctx) error {
	id := c.Params("id")
	userID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid ID",
		})
	}

	var updates bson.M
	if err := c.BodyParser(&updates); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid request body",
		})
	}

	collection := getUserCollection()
	_, err = collection.UpdateOne(context.Background(), bson.M{"_id": userID, "roles": "seller"}, bson.M{"$set": updates})
	if err != nil {
		log.Println("Error updating seller:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Error updating seller",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Seller updated successfully",
	})
}

func DeleteSeller(c *fiber.Ctx) error {
	id := c.Params("id")
	userID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid ID",
		})
	}

	collection := getUserCollection()
	_, err = collection.DeleteOne(context.Background(), bson.M{"_id": userID, "roles": "seller"})
	if err != nil {
		log.Println("Error deleting seller:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Error deleting seller",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Seller deleted successfully",
	})
}

// CRUD for Users with Both Roles (Customer and Seller)
func GetCustomerSellers(c *fiber.Ctx) error {
	collection := getUserCollection()
	cursor, err := collection.Find(context.Background(), bson.M{"roles": bson.M{"$all": []string{"customer", "seller"}}})
	if err != nil {
		log.Println("Error fetching customer-sellers:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Error fetching customer-sellers",
		})
	}

	var customerSellers []model.User
	if err = cursor.All(context.Background(), &customerSellers); err != nil {
		log.Println("Error decoding customer-sellers:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Error decoding customer-sellers",
		})
	}

	return c.JSON(customerSellers)
}

func CreateCustomerSeller(c *fiber.Ctx) error {
	var user model.User
	if err := c.BodyParser(&user); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid request body",
		})
	}

	user.Roles = []string{"customer", "seller"}
	user.ID = primitive.NewObjectID()
	collection := getUserCollection()
	_, err := collection.InsertOne(context.Background(), user)
	if err != nil {
		log.Println("Error creating customer-seller:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Error creating customer-seller",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Customer-seller created successfully",
	})
}

func UpdateCustomerSeller(c *fiber.Ctx) error {
	id := c.Params("id")
	userID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid ID",
		})
	}

	var updates bson.M
	if err := c.BodyParser(&updates); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid request body",
		})
	}

	collection := getUserCollection()
	_, err = collection.UpdateOne(context.Background(), bson.M{"_id": userID, "roles": bson.M{"$all": []string{"customer", "seller"}}}, bson.M{"$set": updates})
	if err != nil {
		log.Println("Error updating customer-seller:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Error updating customer-seller",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Customer-seller updated successfully",
	})
}

func DeleteCustomerSeller(c *fiber.Ctx) error {
	id := c.Params("id")
	userID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid ID",
		})
	}

	collection := getUserCollection()
	_, err = collection.DeleteOne(context.Background(), bson.M{"_id": userID, "roles": bson.M{"$all": []string{"customer", "seller"}}})
	if err != nil {
		log.Println("Error deleting customer-seller:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Error deleting customer-seller",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Customer-seller deleted successfully",
	})
}

// Suspend User Account
func SuspendUser(c *fiber.Ctx) error {
	id := c.Params("id")
	userID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid ID",
		})
	}

	collection := getUserCollection()
	_, err = collection.UpdateOne(context.Background(), bson.M{"_id": userID}, bson.M{"$set": bson.M{"suspended": true}})
	if err != nil {
		log.Println("Error suspending user:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Error suspending user",
		})
	}

	return c.JSON(fiber.Map{
		"message": "User account suspended successfully",
	})
}

func UnsuspendUser(c *fiber.Ctx) error {
	id := c.Params("id")
	userID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid ID",
		})
	}

	collection := getUserCollection()
	_, err = collection.UpdateOne(context.Background(), bson.M{"_id": userID}, bson.M{"$set": bson.M{"suspended": false}})
	if err != nil {
		log.Println("Error unsuspending user:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Error unsuspending user",
		})
	}

	return c.JSON(fiber.Map{
		"message": "User account unsuspended successfully",
	})
}
func SuspendSeller(c *fiber.Ctx) error {
	sellerID := c.Params("id")
	log.Println("Suspend request received for seller ID:", sellerID)

	objectID, err := primitive.ObjectIDFromHex(sellerID)
	if err != nil {
		log.Println("Invalid seller ID format:", sellerID)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid seller ID format",
		})
	}

	collection := getUserCollection()
	filter := bson.M{"_id": objectID, "roles": bson.M{"$all": []string{"seller"}}}
	update := bson.M{"$set": bson.M{"suspended": true}}

	result, err := collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		log.Println("Error suspending seller:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to suspend seller",
		})
	}

	if result.ModifiedCount == 0 {
		log.Println("Seller not found or already suspended:", sellerID)
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "Seller not found or already suspended",
		})
	}

	log.Println("Seller suspended successfully:", sellerID)
	return c.JSON(fiber.Map{
		"message": "Seller suspended successfully",
	})
}

func UnsuspendSeller(c *fiber.Ctx) error {
	sellerID := c.Params("id")
	log.Println("Received unsuspend request for seller ID:", sellerID)

	objectID, err := primitive.ObjectIDFromHex(sellerID)
	if err != nil {
		log.Println("Invalid ObjectID:", sellerID)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid seller ID format",
		})
	}

	collection := getUserCollection()
	filter := bson.M{"_id": objectID, "roles": bson.M{"$all": []string{"seller"}}}
	update := bson.M{"$set": bson.M{"suspended": false}}

	result, err := collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		log.Println("Database error:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to unsuspend seller",
		})
	}

	if result.ModifiedCount == 0 {
		log.Println("No matching seller found or already unsuspended:", sellerID)
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "Seller not found or already unsuspended",
		})
	}

	log.Println("Seller unsuspended successfully:", sellerID)
	return c.JSON(fiber.Map{
		"message": "Seller unsuspended successfully",
	})
}

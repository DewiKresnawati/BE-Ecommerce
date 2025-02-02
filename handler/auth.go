package handler

import (
	"be_ecommerce/config"
	"be_ecommerce/model"
	"be_ecommerce/utils"
	"context"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

// Register handles user registration
func Register(c *fiber.Ctx) error {
	var user model.User
	if err := c.BodyParser(&user); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Error parsing request body",
		})
	}

	// Validasi email
	collection := config.MongoClient.Database("ecommerce").Collection("users")
	existingUser := collection.FindOne(context.Background(), bson.M{"email": user.Email})
	if existingUser.Err() == nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Email already registered",
		})
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Error hashing password",
		})
	}
	user.Password = string(hashedPassword)

	// Validasi dan tetapkan role
	if len(user.Roles) == 0 {
		user.Roles = []string{"customer"}
	} else {
		validRoles := map[string]bool{"admin": true, "seller": true, "customer": true}
		for _, role := range user.Roles {
			if !validRoles[role] {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"message": "Invalid role provided",
				})
			}
		}
	}

	// Simpan pengguna ke database
	user.ID = primitive.NewObjectID()
	_, err = collection.InsertOne(context.Background(), user)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Error saving user to database",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "User registered successfully",
	})
}

// Login handles user login
func Login(c *fiber.Ctx) error {
	var req model.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid request body",
		})
	}

	fmt.Println("Email:", req.Email)
	fmt.Println("Password:", req.Password)

	// Cari user di database
	var user model.User
	collection := config.MongoClient.Database("ecommerce").Collection("users")
	err := collection.FindOne(c.Context(), bson.M{"email": req.Email}).Decode(&user)
	if err != nil {
		fmt.Println("User not found for email:", req.Email)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Invalid email or password",
		})
	}

	fmt.Println("User found:", user.Email, "Roles:", user.Roles)

	// Verifikasi password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		fmt.Println("Password mismatch for email:", req.Email)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Invalid email or password",
		})
	}

	// Generate token, tambahkan seller_id jika ada
	var sellerID string
	if user.SellerID != nil {
		sellerID = user.SellerID.Hex()
	}

	token, err := utils.GenerateJWT(user.ID.Hex(), user.Roles[0], sellerID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Could not generate token",
		})
	}

	fmt.Println("Generated token:", token)

	// Response login
	response := fiber.Map{
		"status":  "success",
		"message": "Login successful",
		"role":    user.Roles[0],
		"token":   token,
		"user_id": user.ID.Hex(),
	}

	// Jika user memiliki seller_id, tambahkan seller info
	if sellerID != "" {
		response["seller_id"] = sellerID
	}

	return c.JSON(response)
}
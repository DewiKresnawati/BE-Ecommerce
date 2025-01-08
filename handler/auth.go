package handler

import (
	"be_ecommerce/config"
	"be_ecommerce/model"
	"be_ecommerce/utils"
	"context"

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

	// Check if email is already registered
	collection := config.MongoClient.Database("ecommerce").Collection("users")
	existingUser := collection.FindOne(context.Background(), bson.M{"email": user.Email})
	if existingUser.Err() == nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Email already registered",
		})
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Error hashing password",
		})
	}
	user.Password = string(hashedPassword)

	// Set default fields
	user.ID = primitive.NewObjectID()
	user.Roles = []string{"customer"} // Default role is "customer"

	// Check if user wants to register as admin
	if c.Query("role") == "admin" { // Use query parameter to specify role
		user.Roles = []string{"admin"}
	} else if c.Query("role") != "" && c.Query("role") != "customer" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid role provided",
		})
	}

	// Insert user into MongoDB
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
	// Mendapatkan data dari body request
	var req model.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid request body",
		})
	}

	// Validasi email dan password dari database
	var user model.User
	collection := config.MongoClient.Database("ecommerce").Collection("users")
	err := collection.FindOne(c.Context(), bson.M{"email": req.Email}).Decode(&user)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid email or password",
		})
	}

	// Memeriksa apakah password cocok
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid email or password",
		})
	}

	// Generate JWT token
	token, err := utils.GenerateJWT(user.ID.Hex())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Could not generate token",
		})
	}

	// Mengembalikan response dengan token JWT
	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Login successful",
		"token":   token,
	})
}

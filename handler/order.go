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

// CheckoutHandler menangani proses checkout dan menyimpan order ke database
func CheckoutHandler(c *fiber.Ctx) error {
	var input struct {
		UserID         string              `json:"user_id"`
		Shipping       string              `json:"shipping"`
		Amount         int                 `json:"amount"`
		Items          []model.OrderItem   `json:"items"`
		ShippingCost   int                 `json:"shipping_cost"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	userID, err := primitive.ObjectIDFromHex(input.UserID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid User ID"})
	}

	order := model.Order{
		ID:             primitive.NewObjectID(),
		UserID:         userID,
		Items:          input.Items,
		TotalAmount:    input.Amount,
		ShippingCost:   input.ShippingCost,
		ShippingAddress: input.Shipping,
		Status:        "Pending",
		CreatedAt:      time.Now(),
	}

	collection := config.MongoClient.Database("ecommerce").Collection("orders")
	_, err = collection.InsertOne(context.TODO(), order)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to place order"})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Order placed successfully",
		"order_id": order.ID.Hex(),
	})
}

// GetOrdersHandler mengambil daftar order berdasarkan user_id
func GetOrdersHandler(c *fiber.Ctx) error {
	userID := c.Query("user_id")
	if userID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "User ID is required"})
	}

	objID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid User ID"})
	}

	var orders []model.Order
	collection := config.MongoClient.Database("ecommerce").Collection("orders")

	cursor, err := collection.Find(context.TODO(), bson.M{"user_id": objID})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch orders"})
	}
	defer cursor.Close(context.TODO())

	for cursor.Next(context.TODO()) {
		var order model.Order
		if err := cursor.Decode(&order); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Error decoding order data"})
		}
		orders = append(orders, order)
	}

	return c.JSON(orders)
}

// GetOrderDetailsHandler mengambil detail order berdasarkan order_id
func GetOrderDetailsHandler(c *fiber.Ctx) error {
	orderID := c.Params("order_id")
	objID, err := primitive.ObjectIDFromHex(orderID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid Order ID"})
	}

	var order model.Order
	collection := config.MongoClient.Database("ecommerce").Collection("orders")

	err = collection.FindOne(context.TODO(), bson.M{"_id": objID}).Decode(&order)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Order not found"})
	}

	return c.JSON(order)
}

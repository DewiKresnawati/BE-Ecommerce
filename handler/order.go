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
		UserID       string            `json:"user_id"`
		Shipping     string            `json:"shipping"`
		Amount       int               `json:"amount"`
		Items        []model.OrderItem `json:"items"`
		ShippingCost int               `json:"shipping_cost"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	userID, err := primitive.ObjectIDFromHex(input.UserID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid User ID"})
	}

	order := model.Order{
		ID:              primitive.NewObjectID(),
		UserID:          userID,
		Items:           input.Items,
		TotalAmount:     input.Amount,
		ShippingCost:    input.ShippingCost,
		ShippingAddress: input.Shipping,
		Status:          "Pending",
		CreatedAt:       time.Now(),
	}

	collection := config.MongoClient.Database("ecommerce").Collection("orders")
	_, err = collection.InsertOne(context.TODO(), order)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to place order"})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message":  "Order placed successfully",
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

	return c.JSON(fiber.Map{"message": "Orders fetched successfully", "data": orders})
}

func GetOrdersBySellerHandler(c *fiber.Ctx) error {
	sellerID := c.Query("seller_id")
	if sellerID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Seller ID is required"})
	}

	objID, err := primitive.ObjectIDFromHex(sellerID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid Seller ID"})
	}

	var orders []model.Order
	collection := config.MongoClient.Database("ecommerce").Collection("orders")

	// Menemukan semua pesanan untuk seller
	cursor, err := collection.Find(context.TODO(), bson.M{"seller_id": objID})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch orders"})
	}
	defer cursor.Close(context.TODO())

	// Decode data ke dalam array orders
	for cursor.Next(context.TODO()) {
		var order model.Order
		if err := cursor.Decode(&order); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Error decoding order data"})
		}
		orders = append(orders, order)
	}

	// Mengembalikan data order
	return c.JSON(fiber.Map{"message": "Orders fetched successfully", "data": orders})
}

// **GET /orders/:order_id** → Ambil detail pesanan berdasarkan ID untuk seller
func GetSellerOrderDetailsHandler(c *fiber.Ctx) error {
	orderID := c.Params("order_id")
	objID, err := primitive.ObjectIDFromHex(orderID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid Order ID"})
	}

	var order model.Order
	collection := config.MongoClient.Database("ecommerce").Collection("orders")

	// Cari order berdasarkan orderID
	err = collection.FindOne(context.TODO(), bson.M{"_id": objID}).Decode(&order)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Order not found"})
	}

	// Mengembalikan data lengkap order
	return c.JSON(fiber.Map{"message": "Order details fetched successfully", "data": order})
}

// **PUT /orders/:order_id** → Update status pesanan oleh seller
func UpdateSellerOrderHandler(c *fiber.Ctx) error {
	// Ambil orderID dari params
	orderID := c.Params("order_id")
	objID, err := primitive.ObjectIDFromHex(orderID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid Order ID"})
	}

	// Struct untuk data yang ingin diperbarui
	var updateData struct {
		Status          string `json:"status"`
		ShippingAddress string `json:"shipping_address"`
		Items           []struct {
			ProductID string `json:"product_id"`
			Name      string `json:"name"`
			Quantity  int    `json:"quantity"`
			Price     int    `json:"price"`
			SellerID  string `json:"seller_id"`
		} `json:"items"`
		TotalAmount    int    `json:"total_amount"`
		ShippingCost   int    `json:"shipping_cost"`
		PaymentToken   string `json:"payment_token"`
	}

	// Parsing data request body
	if err := c.BodyParser(&updateData); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Cek apakah status valid (opsional, tergantung kebutuhan)
	validStatuses := []string{"Pending", "Confirmed", "Shipped", "Delivered"}
	statusValid := false
	for _, validStatus := range validStatuses {
		if updateData.Status == validStatus {
			statusValid = true
			break
		}
	}
	if !statusValid {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid status"})
	}

	// Collection orders
	collection := config.MongoClient.Database("ecommerce").Collection("orders")

	// Data yang akan diupdate
	updateFields := bson.M{
		"$set": bson.M{
			"status":           updateData.Status,
			"shipping_address": updateData.ShippingAddress,
			"items":            updateData.Items,
			"total_amount":     updateData.TotalAmount,
			"shipping_cost":    updateData.ShippingCost,
			"payment_token":    updateData.PaymentToken,
		},
	}

	// Melakukan update data berdasarkan orderID
	_, err = collection.UpdateOne(context.TODO(), bson.M{"_id": objID}, updateFields)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update order"})
	}

	// Mengembalikan pesan sukses
	return c.JSON(fiber.Map{"message": "Order updated successfully"})
}

// PUT /orders/status/:order_id → Update status order
func UpdateOrderStatusHandler(c *fiber.Ctx) error {
	orderID := c.Params("order_id")
	objID, err := primitive.ObjectIDFromHex(orderID)
	if err != nil {
	  return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid Order ID"})
	}
  
	var statusUpdate struct {
	  Status string `json:"status"`
	}
  
	if err := c.BodyParser(&statusUpdate); err != nil {
	  return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}
  
	if statusUpdate.Status == "" {
	  return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Status is required"})
	}
  
	collection := config.MongoClient.Database("ecommerce").Collection("orders")
	_, err = collection.UpdateOne(
	  context.TODO(),
	  bson.M{"_id": objID},
	  bson.M{"$set": bson.M{"status": statusUpdate.Status}},
	)
  
	if err != nil {
	  return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update order status"})
	}
  
	return c.JSON(fiber.Map{"message": "Order status updated successfully"})
  }
  
// **DELETE /orders/:order_id** → Hapus pesanan oleh seller
func DeleteSellerOrderHandler(c *fiber.Ctx) error {
	orderID := c.Params("order_id")
	objID, err := primitive.ObjectIDFromHex(orderID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid Order ID"})
	}

	collection := config.MongoClient.Database("ecommerce").Collection("orders")
	// Menghapus order berdasarkan orderID
	_, err = collection.DeleteOne(context.TODO(), bson.M{"_id": objID})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete order"})
	}

	// Mengembalikan pesan sukses
	return c.JSON(fiber.Map{"message": "Order deleted successfully"})
}

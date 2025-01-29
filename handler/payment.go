package handler

import (
	"be_ecommerce/config"
	"be_ecommerce/model"
	"be_ecommerce/services"
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/veritrans/go-midtrans"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// CreatePaymentHandler menangani proses pembayaran menggunakan Midtrans
func CreatePaymentHandler(c *fiber.Ctx) error {
	var input struct {
		UserID       string            `json:"user_id"`
		Shipping     string            `json:"shipping"`
		Amount       int               `json:"amount"`
		ShippingCost int               `json:"shipping_cost"`
		Items        []model.OrderItem `json:"items"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"message": "Invalid request body"})
	}

	if input.UserID == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"message": "User ID is required"})
	}

	// Konversi userID ke ObjectID
	objUserID, err := primitive.ObjectIDFromHex(input.UserID)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"message": "Invalid User ID"})
	}

	// Hitung total amount berdasarkan items
	var totalAmount int64 = 0
	for _, item := range input.Items {
		totalAmount += int64(item.Price * item.Quantity)
	}

	// Tambahkan shipping cost
	if input.ShippingCost > 0 {
		totalAmount += int64(input.ShippingCost)
	}

	// Pastikan total transaksi tidak 0
	if totalAmount < 1 {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"message": "Total transaction amount must be greater than 0"})
	}

	// Buat order baru
	order := model.Order{
		ID:              primitive.NewObjectID(),
		UserID:          objUserID,
		Items:           input.Items,
		TotalAmount:     int(totalAmount),
		ShippingCost:    input.ShippingCost,
		ShippingAddress: input.Shipping,
		Status:          "Pending",
		CreatedAt:       time.Now(),
	}

	// Simpan order ke database
	orderCollection := config.MongoClient.Database("ecommerce").Collection("orders")
	_, err = orderCollection.InsertOne(context.Background(), order)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to place order"})
	}

	// Hapus cart setelah pembayaran
	cartCollection := config.MongoClient.Database("ecommerce").Collection("carts")
	_, err = cartCollection.DeleteOne(context.Background(), bson.M{"user_id": objUserID})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to clear cart after order placement"})
	}

	// Konfigurasi Midtrans
	orderID := "order-" + time.Now().Format("20060102150405")
	midtransClient := services.MidtransClient()
	snapGateway := midtrans.SnapGateway{Client: *midtransClient}

	snapReq := &midtrans.SnapReq{
		TransactionDetails: midtrans.TransactionDetails{
			OrderID:  orderID,
			GrossAmt: totalAmount,
		},
		Items: &[]midtrans.ItemDetail{},
	}

	// Tambahkan produk ke daftar Midtrans
	for _, item := range order.Items {
		*snapReq.Items = append(*snapReq.Items, midtrans.ItemDetail{
			ID:    item.Name,
			Name:  item.Name,
			Qty:   int32(item.Quantity),
			Price: int64(item.Price),
		})
	}

	// Kirim permintaan ke Midtrans
	snapResp, err := snapGateway.GetToken(snapReq)
	if err != nil {
		fmt.Printf("Midtrans Error: %+v\n", err)
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to create payment",
			"details": err.Error(),
		})
	}

	// Perbarui order dengan payment token
	_, err = orderCollection.UpdateOne(context.Background(), bson.M{"_id": order.ID}, bson.M{
		"$set": bson.M{"payment_token": snapResp.Token},
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to update order with payment token"})
	}

	return c.JSON(fiber.Map{
		"token":        snapResp.Token,
		"redirect_url": snapResp.RedirectURL,
	})
}

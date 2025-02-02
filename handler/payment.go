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

	// 🔥 1. Parse Request Body
	if err := c.BodyParser(&input); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"message": "Invalid request body"})
	}

	// 🔥 2. Validasi User ID
	if input.UserID == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"message": "User ID is required"})
	}
	objUserID, err := primitive.ObjectIDFromHex(input.UserID)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"message": "Invalid User ID"})
	}

	// 🔥 3. Ambil Seller ID dari Produk
	productCollection := config.MongoClient.Database("ecommerce").Collection("products")
	for i, item := range input.Items {
		var product model.Product
		err := productCollection.FindOne(context.TODO(), bson.M{"_id": item.ProductID}).Decode(&product)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid product ID"})
		}
		// ✅ Tambahkan seller_id dari produk ke item
		input.Items[i].SellerID = product.SellerID
	}

	// 🔥 4. Hitung Total Amount dan Siapkan Data Midtrans
	var totalAmount int64 = 0
	var midtransItems []midtrans.ItemDetail

	for _, item := range input.Items {
		totalAmount += int64(item.Price * item.Quantity)

		// ✅ Tambahkan ke Midtrans Items
		midtransItems = append(midtransItems, midtrans.ItemDetail{
			ID:    item.ProductID.Hex(),
			Name:  item.Name,
			Qty:   int32(item.Quantity),
			Price: int64(item.Price),
		})
	}

	// ✅ Tambahkan Shipping Cost sebagai item terpisah di Midtrans
	if input.ShippingCost > 0 {
		midtransItems = append(midtransItems, midtrans.ItemDetail{
			ID:    "SHIPPING",
			Name:  "Shipping Cost",
			Qty:   1,
			Price: int64(input.ShippingCost),
		})
		totalAmount += int64(input.ShippingCost)
	}

	// 🔥 5. Pastikan Total Amount Tidak 0
	if totalAmount < 1 {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"message": "Total transaction amount must be greater than 0"})
	}

	// 🔥 6. Simpan Order ke Database
	order := model.Order{
		ID:              primitive.NewObjectID(),
		UserID:          objUserID,
		SellerID:        input.Items[0].SellerID, // Ambil Seller ID dari item pertama
		Items:           input.Items,
		TotalAmount:     int(totalAmount),
		ShippingCost:    input.ShippingCost,
		ShippingAddress: input.Shipping,
		Status:          "Pending",
		CreatedAt:       time.Now(),
	}

	orderCollection := config.MongoClient.Database("ecommerce").Collection("orders")
	_, err = orderCollection.InsertOne(context.Background(), order)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to place order"})
	}

	// 🔥 7. Hapus Cart Setelah Pembayaran
	cartCollection := config.MongoClient.Database("ecommerce").Collection("carts")
	_, err = cartCollection.DeleteOne(context.Background(), bson.M{"user_id": objUserID})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to clear cart after order placement"})
	}

	// 🔥 8. Konfigurasi Midtrans
	orderID := "order-" + time.Now().Format("20060102150405")
	midtransClient := services.MidtransClient()
	snapGateway := midtrans.SnapGateway{Client: *midtransClient}

	snapReq := &midtrans.SnapReq{
		TransactionDetails: midtrans.TransactionDetails{
			OrderID:  orderID,
			GrossAmt: totalAmount, // ✅ Sekarang sudah sesuai dengan item Midtrans
		},
		Items: &midtransItems,
	}

	// 🔥 9. Kirim Permintaan ke Midtrans
	snapResp, err := snapGateway.GetToken(snapReq)
	if err != nil {
		fmt.Printf("Midtrans Error: %+v\n", err)
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to create payment",
			"details": err.Error(),
		})
	}

	// 🔥 10. Perbarui Order dengan Payment Token
	_, err = orderCollection.UpdateOne(context.Background(), bson.M{"_id": order.ID}, bson.M{
		"$set": bson.M{"payment_token": snapResp.Token},
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to update order with payment token"})
	}

	// ✅ 11. Kirim Response ke FE
	return c.JSON(fiber.Map{
		"token":        snapResp.Token,
		"redirect_url": snapResp.RedirectURL,
	})
}

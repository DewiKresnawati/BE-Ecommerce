package handler

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/veritrans/go-midtrans"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"be_ecommerce/config"
	"be_ecommerce/model"
	"be_ecommerce/services"
)

func CreatePayment(c *fiber.Ctx) error {
	// Ambil user_id dari query parameter
	userID := c.Query("user_id")
	if userID == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"message": "user_id is required"})
	}

	// Ambil koleksi cart dan produk dari database
	cartCollection := config.MongoClient.Database("ecommerce").Collection("carts")
	productCollection := config.MongoClient.Database("ecommerce").Collection("products")

	// Cari data keranjang berdasarkan user_id
	var cart model.Cart
	err := cartCollection.FindOne(context.Background(), bson.M{"user_id": userID}).Decode(&cart)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return c.Status(http.StatusNotFound).JSON(fiber.Map{"message": "Cart not found"})
		}
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to fetch cart"})
	}

	// Siapkan data untuk pembayaran Midtrans
	orderID := "order-" + time.Now().Format("20060102150405")
	midtransClient := services.MidtransClient()
	snapGateway := midtrans.SnapGateway{Client: *midtransClient}
	snapReq := &midtrans.SnapReq{
		TransactionDetails: midtrans.TransactionDetails{
			OrderID:  orderID,
			GrossAmt: 0, // Akan dihitung berdasarkan produk di keranjang
		},
		Items: &[]midtrans.ItemDetail{},
	}

	// Iterasi setiap item di keranjang untuk menambahkan ke transaksi
	for _, cartItem := range cart.Products {
		var product model.Product
		productID, _ := primitive.ObjectIDFromHex(cartItem.ProductID)

		// Ambil detail produk berdasarkan ProductID
		err := productCollection.FindOne(context.Background(), bson.M{"_id": productID}).Decode(&product)
		if err != nil {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to fetch product details"})
		}

		// Hitung total harga dengan diskon (jika ada)
		priceAfterDiscount := int64(product.Price - product.Discount) // Pastikan tipe data int64
		totalPrice := priceAfterDiscount * int64(cartItem.Quantity)   // Konversi Quantity ke int64

		// Tambahkan detail item ke permintaan Midtrans
		*snapReq.Items = append(*snapReq.Items, midtrans.ItemDetail{
			ID:    product.ID.Hex(),
			Name:  product.Name,
			Qty:   int32(cartItem.Quantity),
			Price: priceAfterDiscount,
		})

		// Tambahkan harga produk ke total transaksi
		snapReq.TransactionDetails.GrossAmt += totalPrice
	}

	// Kirim permintaan ke Midtrans untuk mendapatkan token pembayaran
	snapResp, err := snapGateway.GetToken(snapReq)
	if err != nil {
		fmt.Printf("Midtrans Error: %+v\n", err)
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to create payment",
			"details": err.Error(),
		})
	}

	// Kembalikan token dan redirect URL untuk pembayaran
	return c.JSON(fiber.Map{
		"token":        snapResp.Token,
		"redirect_url": snapResp.RedirectURL,
	})
}

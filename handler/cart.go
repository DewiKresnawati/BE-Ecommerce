package handler

import (
	"be_ecommerce/config"
	"be_ecommerce/model"
	"context"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// AddToCart menambahkan produk ke keranjang pengguna
func AddToCart(c *fiber.Ctx) error {
	var cartItem model.CartItem
	if err := c.BodyParser(&cartItem); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid request body"})
	}

	// Ambil user_id dari context
	userID := c.Locals("user_id").(string)

	// Ambil koleksi keranjang
	collection := config.MongoClient.Database("ecommerce").Collection("carts")

	// Periksa apakah keranjang sudah ada
	var cart model.Cart
	err := collection.FindOne(context.Background(), bson.M{"user_id": userID}).Decode(&cart)
	if err == mongo.ErrNoDocuments {
		// Jika tidak ada keranjang, buat baru
		cart = model.Cart{
			UserID:   userID,
			Products: []model.CartItem{cartItem},
		}
		_, err := collection.InsertOne(context.Background(), cart)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to add product to cart"})
		}
	} else if err == nil {
		// Jika keranjang ada, tambahkan produk
		found := false
		for i, item := range cart.Products {
			if item.ProductID == cartItem.ProductID {
				cart.Products[i].Quantity += cartItem.Quantity
				found = true
				break
			}
		}
		if !found {
			cart.Products = append(cart.Products, cartItem)
		}

		// Perbarui keranjang
		_, err := collection.UpdateOne(context.Background(), bson.M{"user_id": userID}, bson.M{"$set": bson.M{"products": cart.Products}})
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to update cart"})
		}
	}

	return c.JSON(fiber.Map{"message": "Product added to cart successfully"})
}

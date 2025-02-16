package handler

import (
	"be_ecommerce/config"
	"be_ecommerce/model"
	"context"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// AddToCart menambahkan produk ke keranjang pengguna
func AddToCart(c *fiber.Ctx) error {
	var cartItem model.CartItem
	if err := c.BodyParser(&cartItem); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid request body"})
	}

	// Validasi input
	if cartItem.UserID == "" || cartItem.ProductID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "User ID and Product ID are required"})
	}

	// Konversi ProductID ke ObjectID
	productObjectID, err := primitive.ObjectIDFromHex(cartItem.ProductID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid Product ID format"})
	}

	cartItem.Quantity = 1                      // Default jumlah jika tidak diberikan
	cartItem.ProductID = productObjectID.Hex() // Pastikan format string

	// Ambil koleksi keranjang
	collection := config.MongoClient.Database("ecommerce").Collection("carts")

	// Periksa apakah keranjang sudah ada untuk user
	var cart model.Cart
	err = collection.FindOne(context.Background(), bson.M{"user_id": cartItem.UserID}).Decode(&cart)
	if err == mongo.ErrNoDocuments {
		// Jika tidak ada keranjang, buat baru
		cart = model.Cart{
			UserID:   cartItem.UserID,
			Products: []model.CartItem{cartItem},
		}
		_, err = collection.InsertOne(context.Background(), cart)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to create cart"})
		}
	} else if err == nil {
		// Jika keranjang sudah ada, tambahkan produk
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
		_, err = collection.UpdateOne(context.Background(), bson.M{"user_id": cartItem.UserID}, bson.M{"$set": bson.M{"products": cart.Products}})
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to update cart"})
		}
	} else {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Error fetching cart"})
	}

	return c.JSON(fiber.Map{"message": "Product added to cart successfully"})
}

// FetchCart mengambil data keranjang berdasarkan user_id
func FetchCart(c *fiber.Ctx) error {
	// Ambil user_id dari query
	userID := c.Query("user_id")
	if userID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "user_id is required"})
	}

	// Ambil koleksi keranjang
	cartCollection := config.MongoClient.Database("ecommerce").Collection("carts")
	productCollection := config.MongoClient.Database("ecommerce").Collection("products")

	// Cari keranjang berdasarkan user_id
	var cart model.Cart
	err := cartCollection.FindOne(context.Background(), bson.M{"user_id": userID}).Decode(&cart)
	if err == mongo.ErrNoDocuments {
		// Jika keranjang tidak ditemukan, kembalikan keranjang kosong
		return c.JSON(fiber.Map{
			"products": []fiber.Map{},
		})
	} else if err != nil {
		// Jika terjadi kesalahan, kembalikan status error
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to fetch cart"})
	}

	// Format ulang data produk sebelum dikembalikan
	products := make([]fiber.Map, len(cart.Products))
	for i, item := range cart.Products {
		// Ambil data produk berdasarkan product_id
		var product model.Product
		productID, err := primitive.ObjectIDFromHex(item.ProductID)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid product_id"})
		}

		err = productCollection.FindOne(context.Background(), bson.M{"_id": productID}).Decode(&product)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				// Jika produk tidak ditemukan, gunakan data default
				products[i] = fiber.Map{
					"product_id":  item.ProductID,
					"name":        "Unknown Product",
					"image":       "/images/default.jpg", // Gambar default jika tidak ditemukan
					"price":       item.Price,
					"quantity":    item.Quantity,
					"total_price": item.Price * item.Quantity,
				}
				continue
			}

			// Jika terjadi kesalahan lain, kembalikan status error
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to fetch product details"})
		}

		// Jika produk ditemukan, gunakan data dari database
		products[i] = fiber.Map{
			"product_id":  product.ID.Hex(),
			"name":        product.Name,
			"image":       product.Image,
			"price":       product.Price,
			"quantity":    item.Quantity,
			"total_price": product.Price * item.Quantity,
		}
	}

	// Kembalikan hanya properti "products"
	return c.JSON(fiber.Map{
		"products": products,
	})
}

// UpdateCartItem memperbarui kuantitas produk dalam keranjang
func UpdateCartItem(c *fiber.Ctx) error {
	var request struct {
		UserID    string `json:"user_id"`
		ProductID string `json:"product_id"`
		Quantity  int    `json:"quantity"`
	}

	// Parsing Body
	if err := c.BodyParser(&request); err != nil {
		fmt.Println("Error: Failed to parse request body")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid request body"})
	}

	// Debugging Input Data
	fmt.Printf("Request Data: %+v\n", request)

	// Validasi Input
	if request.UserID == "" {
		fmt.Println("Error: Missing user_id")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "User ID is required"})
	}
	if request.ProductID == "" {
		fmt.Println("Error: Missing product_id")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Product ID is required"})
	}
	if request.Quantity < 1 {
		fmt.Println("Error: Invalid quantity")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Quantity must be greater than 0"})
	}

	// Proses Update Cart
	collection := config.MongoClient.Database("ecommerce").Collection("carts")
	var cart model.Cart
	err := collection.FindOne(context.Background(), bson.M{"user_id": request.UserID}).Decode(&cart)
	if err == mongo.ErrNoDocuments {
		fmt.Println("Error: Cart not found for user_id", request.UserID)
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Cart not found"})
	} else if err != nil {
		fmt.Printf("Error: Failed to fetch cart: %v\n", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Error fetching cart"})
	}

	// Update Quantity
	updated := false
	for i, item := range cart.Products {
		if item.ProductID == request.ProductID {
			cart.Products[i].Quantity = request.Quantity
			updated = true
			fmt.Printf("Updated product %s with new quantity %d\n", request.ProductID, request.Quantity)
			break
		}
	}

	if !updated {
		fmt.Println("Error: Product not found in cart")
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Product not found in cart"})
	}

	// Simpan Perubahan
	_, err = collection.UpdateOne(context.Background(), bson.M{"user_id": request.UserID}, bson.M{"$set": bson.M{"products": cart.Products}})
	if err != nil {
		fmt.Printf("Error: Failed to update cart: %v\n", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to update cart"})
	}

	return c.JSON(fiber.Map{"message": "Cart updated successfully"})
}
func RemoveFromCart(c *fiber.Ctx) error {
	var request struct {
		UserID    string `json:"user_id"`
		ProductID string `json:"product_id"`
	}

	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid request body",
		})
	}

	if request.UserID == "" || request.ProductID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "User ID and Product ID are required",
		})
	}

	collection := config.MongoClient.Database("ecommerce").Collection("carts")
	filter := bson.M{"user_id": request.UserID}
	update := bson.M{"$pull": bson.M{"products": bson.M{"product_id": request.ProductID}}}

	result, err := collection.UpdateOne(context.Background(), filter, update)
	if err != nil || result.ModifiedCount == 0 {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to remove product from cart",
		})
	}

	return c.JSON(fiber.Map{"message": "Product removed successfully"})
}

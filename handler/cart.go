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

	// Validasi user_id
	if cartItem.UserID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "user_id is required"})
	}

	// Ambil koleksi keranjang
	collection := config.MongoClient.Database("ecommerce").Collection("carts")

	// Periksa apakah keranjang sudah ada
	var cart model.Cart
	err := collection.FindOne(context.Background(), bson.M{"user_id": cartItem.UserID}).Decode(&cart)
	if err == mongo.ErrNoDocuments {
		// Jika tidak ada keranjang, buat baru
		cart = model.Cart{
			UserID:   cartItem.UserID,
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
		_, err := collection.UpdateOne(context.Background(), bson.M{"user_id": cartItem.UserID}, bson.M{"$set": bson.M{"products": cart.Products}})
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to update cart"})
		}
	}

	return c.JSON(fiber.Map{"message": "Product added to cart successfully"})
}

// FetchCart mengambil data keranjang berdasarkan user_id
func FetchCart(c *fiber.Ctx) error {
    // Ambil user_id dari query atau body
    userID := c.Query("user_id")
    if userID == "" {
        var body struct {
            UserID string `json:"user_id"`
        }
        if err := c.BodyParser(&body); err != nil || body.UserID == "" {
            return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "user_id is required"})
        }
        userID = body.UserID
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
        err := productCollection.FindOne(context.Background(), bson.M{"_id": item.ProductID}).Decode(&product)
        if err != nil {
            // Jika produk tidak ditemukan, gunakan data default
            products[i] = fiber.Map{
                "product_id":   item.ProductID,
                "product_name": "Unknown Product",
                "price":        item.Price,
                "quantity":     item.Quantity,
            }
        } else {
            // Jika produk ditemukan, gunakan data dari database
            products[i] = fiber.Map{
				"product_id":   product.ID.Hex(),
				"product_name": product.Name,
				"price":        product.Price,
				"quantity":     item.Quantity,
				"image":        product.Image,
			}			
        }
    }

    // Kembalikan hanya properti "products"
    return c.JSON(fiber.Map{
        "products": products,
    })
}


func AddFavoriteToCart(c *fiber.Ctx) error {
	var request struct {
		UserID    string `json:"user_id"`
		ProductID string `json:"product_id"`
		Quantity  int    `json:"quantity"`
		Price     int    `json:"price"`
	}

	// Parsing body request
	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid request body",
		})
	}

	// Validasi input
	if request.UserID == "" || request.ProductID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "User ID and Product ID are required",
		})
	}
	if request.Quantity <= 0 {
		request.Quantity = 1 // Default quantity jika tidak diberikan
	}

	// Ambil koleksi favorit
	favoritesCollection := config.MongoClient.Database("ecommerce").Collection("favorites")

	// Periksa apakah produk ada di favorit
	var favorite model.Favorite
	err := favoritesCollection.FindOne(context.Background(), bson.M{
		"user_id": request.UserID,
	}).Decode(&favorite)
	if err == mongo.ErrNoDocuments {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "Product not found in favorites",
		})
	} else if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to fetch favorite products",
		})
	}

	// Periksa apakah produk ada di `product_ids` favorit
	found := false
	for _, productID := range favorite.ProductIDs {
		if productID == request.ProductID {
			found = true
			break
		}
	}
	if !found {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "Product not found in user's favorites",
		})
	}

	// Tambahkan produk ke keranjang
	cartCollection := config.MongoClient.Database("ecommerce").Collection("carts")
	var cart model.Cart
	err = cartCollection.FindOne(context.Background(), bson.M{"user_id": request.UserID}).Decode(&cart)
	if err == mongo.ErrNoDocuments {
		// Jika keranjang belum ada, buat baru
		cart = model.Cart{
			UserID: request.UserID,
			Products: []model.CartItem{
				{
					UserID:    request.UserID,
					ProductID: request.ProductID,
					Quantity:  request.Quantity,
					Price:     request.Price,
				},
			},
		}
		_, err := cartCollection.InsertOne(context.Background(), cart)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": "Failed to create cart",
			})
		}
	} else if err == nil {
		// Jika keranjang sudah ada, tambahkan produk atau tingkatkan jumlahnya
		productFound := false
		for i, item := range cart.Products {
			if item.ProductID == request.ProductID {
				cart.Products[i].Quantity += request.Quantity
				productFound = true
				break
			}
		}
		if !productFound {
			cart.Products = append(cart.Products, model.CartItem{
				UserID:    request.UserID,
				ProductID: request.ProductID,
				Quantity:  request.Quantity,
				Price:     request.Price,
			})
		}

		// Perbarui keranjang
		_, err = cartCollection.UpdateOne(context.Background(), bson.M{"user_id": request.UserID}, bson.M{
			"$set": bson.M{"products": cart.Products},
		})
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": "Failed to update cart",
			})
		}
	} else {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to fetch cart",
		})
	}

	// Hapus produk dari favorit
	_, err = favoritesCollection.UpdateOne(context.Background(), bson.M{"user_id": request.UserID}, bson.M{
		"$pull": bson.M{"product_ids": request.ProductID},
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to remove product from favorites",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Product moved from favorites to cart successfully",
	})
}

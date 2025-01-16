package router

import (
	"be_ecommerce/handler"

	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(app *fiber.App) {
	// Auth routes
	app.Post("/register", handler.Register)
	app.Post("/login", handler.Login)

	// Product routes
	app.Post("/products", handler.CreateProduct)
	app.Get("/products", handler.GetAllProducts)
	app.Get("/products/under100k", handler.GetProductsUnderPrice)
	app.Get("/products/best-sellers", handler.GetBestSellers)
	app.Post("/products", handler.CreateProduct)
	app.Get("/products/:id", handler.GetProductDetail)
	app.Get("/products/:product_id/rating", handler.GetProductRating)

	app.Post("/categories", handler.AddCategory)        // Tambahkan kategori baru
	app.Post("/categories/sub", handler.AddSubCategory) // Tambahkan sub-kategori ke kategori
	app.Get("/categories", handler.GetCategories)       // Dapatkan semua kategori dan sub-kategori

	app.Post("/reviews", handler.AddReview)                 // Tambahkan review baru
	app.Get("/reviews/:product_id", handler.GetReviews)     // Ambil semua review untuk produk
	app.Put("/reviews/:review_id", handler.UpdateReview)    // Perbarui review
	app.Delete("/reviews/:review_id", handler.DeleteReview) // Hapus review

	app.Post("/cart", handler.AddToCart)
	app.Get("/cart", handler.FetchCart)
	app.Post("/favorites-to-cart", handler.AddFavoriteToCart)

	app.Post("/favorites", handler.AddToFavorites)
	app.Get("/favorites", handler.GetFavorites)

	// Customer applies as seller
	app.Post("/apply-as-seller", handler.ApplyAsSeller)


	// Admin approves/rejects seller application
	app.Post("/admin/approve-seller", handler.ApproveSeller)

}

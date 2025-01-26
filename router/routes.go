package router

import (
	"be_ecommerce/handler"

	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(app *fiber.App) {
	// Auth routes
	app.Post("/register", handler.Register)
	app.Post("/login", handler.Login)
	app.Get("/users/me", handler.GetUserProfile)
	app.Put("/users/update-profile", handler.EditProfile)
	app.Put("/users/reset-password", handler.ResetPassword)
	app.Post("/users/send-password-reset-email", handler.SendPasswordResetEmail)
	app.Post("/users/verify-otp", handler.VerifyOTP)
	// Product routes
	app.Post("/products", handler.CreateProduct)
	app.Get("/products", handler.GetAllProducts)
	app.Get("/products/:id", handler.GetProductDetail)
	app.Get("/products/:product_id/rating", handler.GetProductRating)
	// Endpoint untuk mendapatkan produk berdasarkan ID
	app.Get("/products/:id", handler.GetProductByID)
	app.Put("/products/:id", handler.UpdateProductByID)
	app.Delete("/products/:id", handler.DeleteProductByID)

	app.Static("/uploads", "./uploads")

	app.Post("/categories", handler.AddCategory)                 // Tambahkan kategori baru
	app.Post("/categories/sub", handler.AddSubCategory)          // Tambahkan sub-kategori ke kategori
	app.Get("/categories", handler.GetCategories)                // Dapatkan semua kategori dan sub-kategori
	app.Put("/categories/:id", handler.UpdateCategory)           // Update kategori berdasarkan ID
	app.Put("/categories/sub/:id", handler.UpdateSubCategory)    // Update sub-kategori berdasarkan ID
	app.Delete("/categories/:id", handler.DeleteCategory)        // Hapus kategori berdasarkan ID
	app.Delete("/categories/sub/:id", handler.DeleteSubCategory) // Hapus sub-kategori berdasarkan ID

	app.Post("/reviews", handler.AddReview)                 // Tambahkan review baru
	app.Get("/reviews/:product_id", handler.GetReviews)     // Ambil semua review untuk produk
	app.Put("/reviews/:review_id", handler.UpdateReview)    // Perbarui review
	app.Delete("/reviews/:review_id", handler.DeleteReview) // Hapus review

	app.Post("/cart", handler.AddToCart)
	app.Get("/cart", handler.FetchCart)
	app.Post("/cart/update", handler.UpdateCartItem)
	app.Post("/cart/delete", handler.RemoveFromCart)

	// Customer applies as seller
	app.Post("/apply-as-seller", handler.ApplyAsSeller)

	// Admin approves/rejects seller application
	app.Post("/admin/approve-seller", handler.ApproveSeller)
	app.Post("/admin/reject-seller", handler.RejectSeller)

	app.Get("/users/:id", handler.GetUserByID)

	// Customer Routes
	app.Get("/customers", handler.GetCustomers)
	app.Post("/customers", handler.CreateCustomer)
	app.Put("/customers/update", handler.UpdateCustomer)
	app.Delete("/customers/:id", handler.DeleteCustomer)

	// seller Routes
	app.Get("/sellers", handler.GetSellers)
	app.Post("/sellers", handler.CreateSeller)
	app.Put("/sellers/:id", handler.UpdateSeller)
	app.Delete("/sellers/:id", handler.DeleteSeller)
	app.Get("/seller/products", handler.GetProductsByUserID)
	app.Post("/seller/products", handler.CreateProductForSeller)

	// Customer-Seller Routes
	app.Get("/customer-sellers", handler.GetCustomerSellers)
	app.Post("/customer-sellers", handler.CreateCustomerSeller)
	app.Put("/customer-sellers/:id", handler.UpdateCustomerSeller)
	app.Delete("/customer-sellers/:id", handler.DeleteCustomerSeller)

	// routes untuk pembayaran
	app.Post("/payment", handler.CreatePayment)

	app.Get("/sellers/:id", handler.GetSellerByID)

	app.Post("/become-seller", handler.BecomeSeller)

	// Endpoint untuk store
	app.Get("/stores/:id", handler.GetStoreDetails) // Mendapatkan detail store dan produk terkait

}

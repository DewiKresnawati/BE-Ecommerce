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
}

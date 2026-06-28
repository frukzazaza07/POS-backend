package router

import (
	"pos-backend/internal/handler"
	"pos-backend/internal/middleware"

	"github.com/gofiber/fiber/v2"
)

type Handlers struct {
	Auth       *handler.AuthHandler
	Product    *handler.POSProductHandler
	Order      *handler.OrderHandler
	Stock      *handler.StockHandler
	Webhook    *handler.WebhookHandler
}

func Setup(app *fiber.App, h Handlers) {
	// Health check
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	// Auth (public)
	auth := app.Group("/auth")
	auth.Post("/login", h.Auth.Login)

	// Webhook receiver from Inventory system (public, HMAC-verified internally)
	app.Post("/webhook/inventory", h.Webhook.InventoryEvent)

	// Protected API
	api := app.Group("/api/v1", middleware.JWT())

	// Products
	products := api.Group("/products")
	products.Get("/", h.Product.List)
	products.Get("/:id", h.Product.Get)
	products.Post("/", middleware.AdminOnly(), h.Product.Create)
	products.Put("/:id", middleware.AdminOnly(), h.Product.Update)
	products.Delete("/:id", middleware.AdminOnly(), h.Product.Delete)

	// Orders
	orders := api.Group("/orders")
	orders.Post("/", h.Order.Create)
	orders.Get("/", h.Order.List)
	orders.Get("/:id", h.Order.Get)
	orders.Post("/:id/cancel", h.Order.Cancel)

	// Stock
	stock := api.Group("/stock")
	stock.Get("/", h.Stock.GetCachedStock)
	stock.Post("/sync", middleware.AdminOnly(), h.Stock.SyncStock)
	stock.Get("/availability/:pos_product_id", h.Stock.CheckAvailability)

	// Admin: register new users
	api.Post("/users/register", middleware.AdminOnly(), h.Auth.Register)
}

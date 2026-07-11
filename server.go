package main

import (
	"log"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/joho/godotenv"

	"pos-backend/internal/database"
	"pos-backend/internal/handler"
	"pos-backend/internal/middleware"
	"pos-backend/internal/repository"
	"pos-backend/internal/router"
	"pos-backend/internal/service"
	"pos-backend/pkg/response"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file, reading from environment")
	}

	db, err := database.Connect()
	if err != nil {
		log.Fatalf("db connect: %v", err)
	}
	if err := database.Migrate(db); err != nil {
		log.Fatalf("db migrate: %v", err)
	}
	database.Seed(db)

	// Repositories
	userRepo := repository.NewUserRepository(db)
	productRepo := repository.NewPOSProductRepository(db)
	orderRepo := repository.NewOrderRepository(db)
	stockRepo := repository.NewStockCacheRepository(db)
	bankQRRepo := repository.NewBankQRConfigRepository(db)
	reportRepo := repository.NewReportRepository(db)
	vatConfigRepo := repository.NewVatConfigRepository(db)

	// Inventory client
	invClient := service.NewInventoryClient()

	// Services
	authSvc := service.NewAuthService(userRepo)
	productSvc := service.NewPOSProductService(productRepo)
	orderSvc := service.NewOrderService(orderRepo, productRepo, vatConfigRepo, invClient)
	syncSvc := service.NewStockSyncService(stockRepo, invClient)
	alertSvc := service.NewPayLaterAlertService(orderRepo)
	reportSvc := service.NewReportService(reportRepo)
	vatConfigSvc := service.NewVatConfigService(vatConfigRepo)

	// Initial stock sync on startup
	go func() {
		if err := syncSvc.SyncFromInventory(); err != nil {
			log.Printf("startup stock sync failed: %v (inventory may be offline)", err)
		}
	}()
	syncSvc.StartPeriodicSync(5 * time.Minute)

	// Pay-later overdue alert loop (every 1 hour)
	alertSvc.StartAlertLoop(1 * time.Hour)

	// Handlers
	handlers := router.Handlers{
		Auth:    handler.NewAuthHandler(authSvc),
		Product: handler.NewPOSProductHandler(productSvc),
		Order:   handler.NewOrderHandler(orderSvc),
		Stock:   handler.NewStockHandler(syncSvc, invClient),
		Webhook: handler.NewWebhookHandler(syncSvc),
		BankQR:  handler.NewBankQRHandler(bankQRRepo),
		Report:  handler.NewReportHandler(reportSvc),
		Vat:     handler.NewVatConfigHandler(vatConfigSvc),
	}

	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return response.Error(c, fiber.StatusInternalServerError, err.Error())
		},
	})

	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(middleware.Language())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
		AllowMethods: "GET, POST, PUT, DELETE, OPTIONS",
	}))

	router.Setup(app, handlers)

	port := os.Getenv("PORT") // Vercel ใส่ให้
	if port == "" {
		port = os.Getenv("APP_PORT") // ของคุณตอนรัน local
	}
	if port == "" {
		port = "4000"
	}

	certFile := os.Getenv("TLS_CERT")
	keyFile := os.Getenv("TLS_KEY")

	if certFile != "" && keyFile != "" {
		log.Printf("POS Backend running on https://localhost:%s", port)
		// log.Fatal(app.ListenTLS(":"+port, certFile, keyFile))
	} else {
		log.Printf("POS Backend running on http://localhost:%s", port)
		log.Fatal(app.Listen(":" + port))
	}
}

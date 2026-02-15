package main

import (
	"fmt"
	"jwt-auth/config"
	"jwt-auth/models"
	"jwt-auth/routes"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/swagger"
	"github.com/joho/godotenv"
	_ "jwt-auth/docs"
)

// @title JWT Auth API
// @version 1.0
// @description Go Fiber, GORM ve JWT ile hazırlanmış iskelet proje.
// @termsOfService http://swagger.io/terms/

// @contact.name API Destek
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:3000
// @BasePath /api/auth
// @schemes http

// @securityDefinitions.apikey Bearer
// @in header
// @name Authorization
// @description "Bearer " boşluk ve sonra token şeklinde giriniz. Örn: Bearer eyJhb...

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using system env vars.")
	}

	config.ConnectDatabase()

	if err := config.DB.AutoMigrate(&models.User{}, &models.RefreshToken{}); err != nil {
		log.Fatal("Migration failed:", err)
	}
	log.Println("Migrations completed successfully")

	app := fiber.New(fiber.Config{
		AppName: "JWT Auth Api",
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{"error": err.Error()})
		},
	})

	app.Use(logger.New())
	app.Use(helmet.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "http://localhost:3000,http://localhost:5173",
		AllowHeaders:     "Origin, Content-Type, Accept",
		AllowMethods:     "GET, POST, PUT, DELETE, OPTIONS",
		AllowCredentials: true,
	}))

	routes.SetupAuthRoutes(app)
	
	app.Get("/swagger/*", swagger.HandlerDefault)

	app.Use(func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Route not found"})
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	// Graceful Shutdown Setup

	// Start server in a separate goroutine
	go func() {
		log.Printf("🚀 Server started on port %s\n", port)
		if err := app.Listen(":" + port); err != nil {
			log.Panic(err)
		}
	}()

	// Wait for interrupt signal SIGINT SIGTERM
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	fmt.Println("\n🛑 Shutdown signal received, closing connections...")

	// Shutdown Fiber app
	if err := app.Shutdown(); err != nil {
		log.Fatal("Fiber shutdown error:", err)
	}

	// Close Database connection
	sqlDB, err := config.DB.DB()
	if err != nil {
		log.Fatal("DB instance error:", err)
	}
	if err := sqlDB.Close(); err != nil {
		log.Fatal("DB close error:", err)
	}

	fmt.Println("👋 Graceful shutdown completed.")
}

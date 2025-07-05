package main

import (
	"context"
	"crypto-gateway/config"
	"crypto-gateway/internal/analytics"
	"crypto-gateway/internal/web/db"
	"crypto-gateway/internal/web/routes"
	"log"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
)

func main() {
	config.LoadConfig()
	app := fiber.New()

	analytics.StController = analytics.NewBinanceAPIController(5700)
	analytics.StBinanceApi = analytics.NewBinanceAPI(analytics.StController)
	analytics.StOrchestrator = analytics.NewBinanceAPIOrchestrator(analytics.StBinanceApi)

	app.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "http://localhost:3001", "http://localhost:80", "https://localhost:443"},
		AllowMethods:     []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		AllowCredentials: true,
	}))

	app.Use(func(c fiber.Ctx) error {
		start := time.Now()

		err := c.Next()
		duration := time.Since(start)

		log.Printf("Request to %s %s took %v", c.Method(), c.OriginalURL(), duration)
		return err
	})

	db.InitDB()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	analytics.StOrchestrator.Start(ctx)

	routes.SetupNotificationRoutes(app)
	routes.SetupAuthRoutes(app)
	routes.SetupTriggersRoutes(app)
	routes.SetupSettingsRoutes(app)

	log.Fatal(app.Listen(":8001"))
}

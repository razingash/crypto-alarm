package main

import (
	"crypto-gateway/config"
	"crypto-gateway/internal/db"
	"crypto-gateway/internal/routes"
	"log"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
)

func main() {
	config.LoadConfig()
	app := fiber.New()

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

	routes.SetupNotificationRoutes(app)
	routes.SetupAuthRoutes(app)
	routes.SetupTriggersRoutes(app)

	log.Fatal(app.Listen(":8001"))
}

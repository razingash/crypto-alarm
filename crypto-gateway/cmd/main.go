package main

import (
	"crypto-gateway/crypto-gateway/config"
	"crypto-gateway/crypto-gateway/internal/db"
	"crypto-gateway/crypto-gateway/internal/routes"
	"log"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
)

func main() {
	config.LoadConfig()

	app := fiber.New()

	app.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "http://localhost:3001"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		AllowCredentials: true,
	}))

	db.InitDB()

	routes.SetupAuthRoutes(app)

	log.Fatal(app.Listen(":8001"))
}

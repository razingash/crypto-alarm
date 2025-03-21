package main

import (
	"crypto-gateway/crypto-gateway/config"
	"crypto-gateway/crypto-gateway/internal/db"
	"crypto-gateway/crypto-gateway/internal/routes"
	"log"

	"github.com/gofiber/fiber/v3"
)

func main() {
	config.LoadConfig()

	app := fiber.New()
	db.InitDB()

	routes.SetupAuthRoutes(app)

	log.Fatal(app.Listen(":8001"))
}

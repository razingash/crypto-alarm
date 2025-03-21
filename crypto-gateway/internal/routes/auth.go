package routes

import (
	//"crypto-gateway/crypto-gateway/internal/auth"
	"crypto-gateway/crypto-gateway/internal/handlers"
	"crypto-gateway/crypto-gateway/internal/middlewares"

	"github.com/gofiber/fiber/v3"
)

func SetupAuthRoutes(app *fiber.App) {
	authGroup := app.Group("/api/v1/auth")

	authGroup.Post("/register", handlers.Register, middlewares.ValidateAuthenticationInfo)
	authGroup.Post("/login", handlers.Login, middlewares.ValidateAuthenticationInfo)
	//authGroup.Post("/refresh", handlers.RefreshToken)
	//authGroup.Post("/validate", handlers.ValidateToken)

	// protected := authGroup.Group("/protected", auth.JWT())
	// protected.Get("/", handlers.Protected)
}

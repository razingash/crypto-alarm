package routes

import (
	"crypto-gateway/crypto-gateway/internal/handlers"
	"crypto-gateway/crypto-gateway/internal/middlewares"

	"github.com/gofiber/fiber/v3"
)

func SetupAuthRoutes(app *fiber.App) {
	authGroup := app.Group("/api/v1/auth")

	authGroup.Post("/register", handlers.Register, middlewares.ValidateAuthenticationInfo)
	authGroup.Post("/token", handlers.Login, middlewares.ValidateAuthenticationInfo)
	authGroup.Post("/token/verify", handlers.ValidateToken)
	authGroup.Post("/token/refresh", handlers.RefreshAccessToken)
	authGroup.Post("/logout", handlers.Logout, middlewares.ValidateAuthorization)
}

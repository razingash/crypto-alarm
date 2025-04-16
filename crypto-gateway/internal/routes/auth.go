package routes

import (
	"crypto-gateway/internal/handlers"
	"crypto-gateway/internal/middlewares"
	"crypto-gateway/internal/middlewares/api_validators"

	"github.com/gofiber/fiber/v3"
)

func SetupAuthRoutes(app *fiber.App) {
	authGroup := app.Group("/api/v1/auth")

	authGroup.Post("/register", handlers.Register, api_validators.ValidateAuthenticationInfo)
	authGroup.Post("/token", handlers.Login, api_validators.ValidateAuthenticationInfo)
	authGroup.Post("/token/verify", handlers.ValidateToken)
	authGroup.Post("/token/refresh", handlers.RefreshAccessToken)
	authGroup.Post("/logout", handlers.Logout, middlewares.IsAuthorized)
}

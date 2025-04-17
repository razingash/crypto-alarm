package routes

import (
	"crypto-gateway/internal/handlers"
	"crypto-gateway/internal/middlewares"
	"crypto-gateway/internal/middlewares/api_validators"

	"github.com/gofiber/fiber/v3"
)

func SetupAuthRoutes(app *fiber.App) {
	group := app.Group("/api/v1/auth")

	group.Post("/register", handlers.Register, api_validators.ValidateAuthenticationInfo)
	group.Post("/token", handlers.Login, api_validators.ValidateAuthenticationInfo)
	group.Post("/token/verify", handlers.ValidateToken)
	group.Post("/token/refresh", handlers.RefreshAccessToken)
	group.Post("/logout", handlers.Logout, middlewares.IsAuthorized)
}

package routes

import (
	"crypto-gateway/internal/web/handlers"
	"crypto-gateway/internal/web/middlewares"
	"crypto-gateway/internal/web/middlewares/api_validators"

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

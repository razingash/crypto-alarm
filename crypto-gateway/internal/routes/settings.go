package routes

import (
	"crypto-gateway/internal/handlers"
	"crypto-gateway/internal/middlewares"
	"crypto-gateway/internal/middlewares/api_validators"

	"github.com/gofiber/fiber/v3"
)

func SetupSettingsRoutes(app *fiber.App) {
	group := app.Group("/api/v1/settings")

	group.Get("/", handlers.GetSettings, middlewares.IsAuthorized)
	group.Patch("/update/", handlers.PatchUpdateSettings, api_validators.ValidatePatchSettings, middlewares.IsAuthorized)
}

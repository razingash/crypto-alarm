package routes

import (
	"crypto-gateway/internal/web/handlers"
	"crypto-gateway/internal/web/middlewares"
	"crypto-gateway/internal/web/middlewares/api_validators"

	"github.com/gofiber/fiber/v3"
)

func SetupSettingsRoutes(app *fiber.App) {
	group := app.Group("/api/v1/settings")

	group.Get("/", handlers.GetSettings, middlewares.IsAuthorized)
	group.Patch("/update/", handlers.PatchUpdateSettings, api_validators.ValidatePatchSettings, middlewares.IsAuthorized)
}

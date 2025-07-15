package routes

import (
	"crypto-gateway/internal/web/handlers"
	"crypto-gateway/internal/web/middlewares/validators"

	"github.com/gofiber/fiber/v3"
)

func SetupSettingsRoutes(app *fiber.App) {
	group := app.Group("/api/v1/settings")

	group.Get("/", handlers.GetSettings)
	group.Patch("/update/", handlers.PatchUpdateSettings, validators.ValidatePatchSettings)
}

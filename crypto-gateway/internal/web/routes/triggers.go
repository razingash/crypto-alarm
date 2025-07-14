package routes

import (
	"crypto-gateway/internal/web/handlers"
	"crypto-gateway/internal/web/middlewares/api_validators"

	"github.com/gofiber/fiber/v3"
)

func SetupTriggersRoutes(app *fiber.App) {
	group := app.Group("/api/v1/triggers")

	group.Get("/keyboard", handlers.Keyboard)
	group.Get("/strategy", handlers.StrategyGet)
	group.Get("/strategy/history/:id", handlers.StrategyHistoryGet)
	group.Post("/strategy", handlers.StrategyPost, api_validators.ValidateStrategyPost)
	group.Patch("/strategy", handlers.StrategyPatch, api_validators.ValidateStrategyPatch)
	group.Delete("/strategy/:id/", handlers.StrategyDelete)
}

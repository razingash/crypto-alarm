package web

import (
	"github.com/gofiber/fiber/v3"
)

func SetupTriggersRoutes(app *fiber.App) {
	group := app.Group("/api/v1/triggers")

	group.Get("/keyboard", Keyboard)
	group.Get("/strategy", StrategyGet)
	group.Get("/strategy/history/:id", StrategyHistoryGet)
	group.Post("/strategy", StrategyPost, ValidateStrategyPost)
	group.Patch("/strategy", StrategyPatch, ValidateStrategyPatch)
	group.Delete("/strategy/:id/", StrategyDelete)
}

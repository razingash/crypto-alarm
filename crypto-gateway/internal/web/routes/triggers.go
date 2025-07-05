package routes

import (
	"crypto-gateway/internal/web/handlers"
	"crypto-gateway/internal/web/middlewares/api_validators"

	"github.com/gofiber/fiber/v3"
)

func SetupTriggersRoutes(app *fiber.App) {
	group := app.Group("/api/v1/triggers")

	group.Get("/keyboard", handlers.Keyboard)
	group.Get("/formula", handlers.FormulaGet)
	group.Get("/formula/history/:id", handlers.FormulaHistoryGet)
	group.Post("/formula", handlers.FormulaPost, api_validators.ValidateFormulaPost)
	group.Patch("/formula", handlers.FormulaPatch, api_validators.ValidateFormulaPatch)
	group.Delete("/formula", handlers.FormulaDelete)
}

package routes

import (
	"crypto-gateway/internal/handlers"
	"crypto-gateway/internal/middlewares"
	"crypto-gateway/internal/middlewares/api_validators"

	"github.com/gofiber/fiber/v3"
)

func SetupTriggersRoutes(app *fiber.App) {
	group := app.Group("/api/v1/triggers")

	group.Get("/keyboard", handlers.Keyboard, middlewares.IsAuthorized)
	group.Get("/formula", handlers.FormulaGet, middlewares.IsAuthorized)
	group.Post("/formula", handlers.FormulaPost, middlewares.IsAuthorized, api_validators.ValidateFormulaPost)
	group.Patch("/formula", handlers.FormulaPatch, middlewares.IsAuthorized, api_validators.ValidateFormulaPatch)
	group.Delete("/formula", handlers.FormulaDelete, middlewares.IsAuthorized)
}

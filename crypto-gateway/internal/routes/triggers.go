package routes

import (
	"crypto-gateway/crypto-gateway/internal/handlers"
	"crypto-gateway/crypto-gateway/internal/middlewares"

	"github.com/gofiber/fiber/v3"
)

func SetupTriggersRoutes(app *fiber.App) {
	authGroup := app.Group("/api/v1/triggers")

	authGroup.Get("/keyboard", handlers.Keyboard, middlewares.ValidateAuthorization)
	authGroup.Get("/formulas", handlers.Formulas, middlewares.ValidateAuthorization)
	authGroup.Post("/formula", handlers.FormulaPost, middlewares.ValidateAuthorization, middlewares.ValidateFormula)
	authGroup.Patch("/formula", handlers.FormulaPatch, middlewares.ValidateAuthorization, middlewares.ValidateFormulaId)
}

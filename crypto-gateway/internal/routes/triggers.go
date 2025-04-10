package routes

import (
	"crypto-gateway/crypto-gateway/internal/handlers"
	"crypto-gateway/crypto-gateway/internal/middlewares"
	"crypto-gateway/crypto-gateway/internal/middlewares/api_validators"

	"github.com/gofiber/fiber/v3"
)

func SetupTriggersRoutes(app *fiber.App) {
	authGroup := app.Group("/api/v1/triggers")

	authGroup.Get("/keyboard", handlers.Keyboard, middlewares.IsAuthorized)
	authGroup.Get("/formula", handlers.FormulaGet, middlewares.IsAuthorized)
	authGroup.Post("/formula", handlers.FormulaPost, middlewares.IsAuthorized, api_validators.ValidateFormulaPost)
	authGroup.Patch("/formula", handlers.FormulaPatch, middlewares.IsAuthorized, api_validators.ValidateFormulaPatch)
	authGroup.Delete("/formula", handlers.FormulaDelete, middlewares.IsAuthorized)

	// only from python service
	authGroup.Post("/push-notifications", handlers.Keyboard, api_validators.ValidatePushNotifications)
}

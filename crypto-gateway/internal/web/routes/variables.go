package routes

import (
	"crypto-gateway/internal/web/handlers"
	"crypto-gateway/internal/web/middlewares/validators"

	"github.com/gofiber/fiber/v3"
)

func SetupVariableRoutes(app *fiber.App) {
	group := app.Group("/api/v1/variable")

	group.Get("/", handlers.VariableGet)
	group.Post("/", handlers.VariablePost, validators.ValidateVariablePost)
	group.Patch("/:id", handlers.VariablePatch, validators.ValidateVariablePatch)
	group.Delete("/:id/", handlers.VariableDelete)
}

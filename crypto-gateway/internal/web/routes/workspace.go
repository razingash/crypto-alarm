package routes

import (
	"crypto-gateway/internal/web/handlers"
	"crypto-gateway/internal/web/middlewares/validators"

	"github.com/gofiber/fiber/v3"
)

func SetupWorkspaceRoutes(app *fiber.App) {
	group := app.Group("/api/v1/workspace")

	group.Get("/diagram", handlers.DiagramGet)
	group.Post("/diagram", handlers.DiagramPost, validators.ValidateDiagramPost)
	group.Patch("/diagram/:id", handlers.DiagramPatch, validators.ValidateDiagramPatch)
	group.Delete("/diagram/:id", handlers.DiagramDelete)
}

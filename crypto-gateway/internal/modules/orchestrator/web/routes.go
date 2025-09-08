package web

import (
	"github.com/gofiber/fiber/v3"
)

func SetupWorkspaceWidgetRoutes(app *fiber.App) {
	group := app.Group("/api/v1/workspace/widgets")

	// список правил оркестраторов нет смысла получать пока что
	group.Post("/orchestrator", OrchestratorPost, ValidateOrchestrator)
	group.Get("/orchestrator/parts", OrchestratorPartsGet) // нужен id workflow
	group.Get("/orchestrator/:id", OrchestratorGet)
	group.Patch("/orchestrator/:id", OrchestratorPatch, ValidateOrchestrator)
	group.Delete("/orchestrator/:id", OrchestratorDelete)
}

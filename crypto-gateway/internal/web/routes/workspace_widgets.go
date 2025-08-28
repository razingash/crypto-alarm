package routes

import (
	W_ochestrator "crypto-gateway/internal/modules/ochestrator/web"

	"github.com/gofiber/fiber/v3"
)

func SetupWorkspaceWidgetRoutes(app *fiber.App) {
	group := app.Group("/api/v1/workspace/widgets")

	// список правил оркестраторов нет смысла получать пока что
	group.Post("/orchestrator", W_ochestrator.OrchestratorPost, W_ochestrator.ValidateOrchestrator)
	group.Get("/orchestrator/parts", W_ochestrator.OrchestratorPartsGet) // нужен id workflow
	group.Get("/orchestrator/:id", W_ochestrator.OrchestratorGet)
	group.Patch("/orchestrator/:id", W_ochestrator.OrchestratorPatch, W_ochestrator.ValidateOrchestrator)
	group.Delete("/orchestrator/:id", W_ochestrator.OrchestratorDelete)
}

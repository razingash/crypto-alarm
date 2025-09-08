package web

import (
	"crypto-gateway/internal/modules/orchestrator/repo"
	"encoding/json"
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx"
)

func OrchestratorPost(c fiber.Ctx) error {
	inputsRaw := c.Locals("inputs")
	if inputsRaw == nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "no inputs provided"})
	}

	inputs, ok := inputsRaw.([]repo.OrchestratorInput)
	if !ok {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "invalid inputs format"})
	}

	if id, err := repo.CreateOrchestrator(inputs); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	} else {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"pk": id,
		})
	}
}

func OrchestratorGet(c fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid id"})
	}

	orchestrator, err := repo.GetOrchestratorByID(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(orchestrator)
}

func OrchestratorPartsGet(c fiber.Ctx) error {
	workflowIDStr := c.Query("workflowId")
	workflowID, err := strconv.ParseInt(workflowIDStr, 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid workflowId"})
	}

	nodeID := c.Query("nodeId")
	if nodeID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "missing nodeId"})
	}

	cell, err := repo.GetOrchestratorParts(workflowID, nodeID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(cell)
}

func OrchestratorPatch(c fiber.Ctx) error {
	idStr := c.Params("id")
	orchestratorID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid id"})
	}

	body := c.Body()
	var req repo.Orchestrator
	if err := json.Unmarshal(body, &req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid JSON"})
	}

	if err := repo.UpdateOrchestrator(c.Context(), orchestratorID, req); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"status": "ok"})
}

func OrchestratorDelete(c fiber.Ctx) error {
	idStr := c.Params("id")
	orchestratorID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid id"})
	}

	if err := repo.DeleteOrchestrator(c.Context(), orchestratorID); err != nil {
		if err == pgx.ErrNoRows {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "orchestrator not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"status": "deleted"})
}

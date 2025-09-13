package handlers

import (
	"crypto-gateway/internal/web/repositories"
	"fmt"
	"strconv"

	"github.com/gofiber/fiber/v3"
)

func DiagramGet(c fiber.Ctx) error {
	defaultLimit := 10
	defaultPage := 1

	limit, err := strconv.Atoi(c.Query("limit"))
	if err != nil || limit <= 0 {
		limit = defaultLimit
	}

	page, err := strconv.Atoi(c.Query("page"))
	if err != nil || page <= 0 {
		page = defaultPage
	}

	diagramId := c.Query("id")
	diagrams, hasNext, err := repositories.GetDiagrams(limit, page, diagramId)

	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	if diagramId == "" {
		if diagrams == nil {
			diagrams = []repositories.Diagram{}
		}
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"data":     diagrams,
			"has_next": hasNext,
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"data": diagrams[0],
	})
}

func DiagramPost(c fiber.Ctx) error {
	name := c.Locals("name").(string)

	id, err := repositories.CreateDiagram(name)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"id": id,
	})
}

func DiagramPatch(c fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid diagram id",
		})
	}

	name := c.Locals("name").(*string)
	diagram := c.Locals("diagram").(*string)

	err = repositories.UpdateDiagram(id, name, diagram)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update variable",
		})
	}

	return c.SendStatus(fiber.StatusOK)
}

// сделать универсальнее, конкретно, вообще убрать и перенести этот функционал к конкретным эндпоинтам
func DiagramPatchNode(c fiber.Ctx) error {
	action := c.Locals("action").(string)
	diagramID := c.Locals("diagramID").(int)
	nodeID := c.Locals("nodeID").(string)
	fmt.Println(action)
	switch action {
	case "attachStrategy":
		strategyID := c.Locals("itemID").(string)
		err := repositories.AttachEntityToNode(diagramID, nodeID, strategyID, "strategyId")
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "failed to attach strategy",
			})
		}
		return c.SendStatus(fiber.StatusOK)
	case "attachOrchestrator":
		orchestratorId := c.Locals("itemID").(string)
		err := repositories.AttachEntityToNode(diagramID, nodeID, orchestratorId, "orchestratorId")
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "failed to attach orchestrator",
			})
		}
		return c.SendStatus(fiber.StatusOK)
	case "attachNotificationTelegram":
		notificationId := c.Locals("itemID").(string)
		err := repositories.AttachEntityToNode(diagramID, nodeID, notificationId, "notificationTelegramId")
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "failed to attach telegram notification",
			})
		}
		return c.SendStatus(fiber.StatusOK)
	default:
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "unknown action",
		})
	}
}

func DiagramDelete(c fiber.Ctx) error {
	IdStr := c.Params("id")
	id, err := strconv.Atoi(IdStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid diagram id",
		})
	}
	err = repositories.DeleteDiagramById(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.SendStatus(fiber.StatusOK)
}

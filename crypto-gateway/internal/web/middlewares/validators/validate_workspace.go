package validators

import (
	"encoding/json"
	"strconv"

	"github.com/gofiber/fiber/v3"
)

func ValidateDiagramPost(c fiber.Ctx) error {
	var body struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal(c.Body(), &body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid JSON",
		})
	}
	if len(body.Name) < 5 || len(body.Name) > 255 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Name must be between 5 and 255 characters",
		})
	}

	c.Locals("name", body.Name)

	return c.Next()
}

func ValidateDiagramPatch(c fiber.Ctx) error {
	var body struct {
		Name    *string `json:"name"`
		Diagram *string `json:"diagram"`
	}
	if err := json.Unmarshal(c.Body(), &body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid JSON",
		})
	}
	if body.Name != nil {
		if len(*body.Name) < 5 || len(*body.Name) > 255 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Name must be between 5 and 255 characters",
			})
		}
	}

	c.Locals("name", body.Name)
	c.Locals("diagram", body.Diagram)

	return c.Next()
}

func ValidateDiagramPatchNode(c fiber.Ctx) error {
	diagramID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid diagram id",
		})
	}

	var body map[string]interface{}
	if err := json.Unmarshal(c.Body(), &body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid JSON",
		})
	}

	var actionKey string
	if _, ok := body["attachStrategy"]; ok { // сделать универсальнее
		actionKey = "attachStrategy"
	} else if _, ok := body["attachOrchestrator"]; ok {
		actionKey = "attachOrchestrator"
	} else if _, ok := body["attachNotificationTelegram"]; ok {
		actionKey = "attachNotificationTelegram"
	} else {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "unsupported action",
		})
	}

	attach, ok := body[actionKey].(map[string]interface{})
	if !ok {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid action format",
		})
	}

	nodeID, _ := attach["nodeId"].(string)
	itemID, _ := attach["itemID"].(string)
	if nodeID == "" || itemID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "nodeId and itemID are required",
		})
	}

	c.Locals("diagramID", diagramID)
	c.Locals("nodeID", nodeID)
	c.Locals("itemID", itemID)
	c.Locals("action", actionKey)
	return c.Next()
}

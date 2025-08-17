package validators

import (
	"encoding/json"

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

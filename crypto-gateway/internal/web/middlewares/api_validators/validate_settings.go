package api_validators

import (
	"encoding/json"

	"github.com/gofiber/fiber/v3"
)

func ValidatePatchSettings(c fiber.Ctx) error {
	var body struct {
		Id       int `json:"id"`
		Cooldown int `json:"cooldown"`
	}

	if err := json.Unmarshal(c.Body(), &body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid JSON",
		})
	}

	if body.Id == 0 || body.Cooldown == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "id and cooldown are required",
		})
	}

	c.Locals("id", body.Id)
	c.Locals("cooldown", body.Cooldown)
	return c.Next()
}

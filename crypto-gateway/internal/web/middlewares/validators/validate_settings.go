package validators

import (
	"crypto-gateway/internal/analytics"
	"encoding/json"

	"github.com/gofiber/fiber/v3"
)

func ValidatePatchSettings(c fiber.Ctx) error {
	var body []analytics.ApiUpdate

	if err := json.Unmarshal(c.Body(), &body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid JSON format",
		})
	}

	for i, item := range body {
		if item.Endpoint == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "Each item must have non-empty endpoint",
				"invalid": i,
			})
		}
	}

	c.Locals("updates", body)
	return c.Next()
}

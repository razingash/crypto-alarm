package validators

import (
	"encoding/json"

	"crypto-gateway/internal/modules/strategy/repo"

	"github.com/gofiber/fiber/v3"
)

type PatchSettingsRequest struct {
	Api    []repo.ApiUpdate    `json:"api,omitempty"`
	Config []repo.ConfigUpdate `json:"config,omitempty"`
}

func ValidatePatchSettings(c fiber.Ctx) error {
	var body PatchSettingsRequest

	if err := json.Unmarshal(c.Body(), &body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid JSON format",
		})
	}

	for _, item := range body.Api {
		if item.Endpoint == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Each API item must have 'endpoint'",
			})
		}
	}

	for _, item := range body.Config {
		if item.ID == 0 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Each Config item must have non-zero id",
			})
		}
	}

	c.Locals("updates", body)
	return c.Next()
}

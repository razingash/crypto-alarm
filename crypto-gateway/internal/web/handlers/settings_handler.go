package handlers

import (
	"context"
	"crypto-gateway/internal/analytics"
	"crypto-gateway/internal/web/repositories"

	"github.com/gofiber/fiber/v3"
)

func GetSettings(c fiber.Ctx) error {
	settings, err := repositories.FetchSettings()

	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "something went wrong",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"data": settings,
	})
}

func PatchUpdateSettings(c fiber.Ctx) error {
	updates := c.Locals("updates").([]analytics.ApiUpdate)

	updatedIds, err := analytics.UpdateEndpointsSettings(updates)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Failed to update endpoints",
		})
	}

	go func() {
		for _, id := range updatedIds {
			api, cooldown := repositories.GetApiAndCooldownByID(id)
			analytics.StOrchestrator.AdjustAPITaskCooldown(context.Background(), api, cooldown)
			analytics.StOrchestrator.LaunchNeededAPI(context.Background())
		}
	}()

	return c.SendStatus(fiber.StatusOK)
}

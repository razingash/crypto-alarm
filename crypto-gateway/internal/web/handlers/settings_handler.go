package handlers

import (
	"context"
	"crypto-gateway/internal/analytics"
	"crypto-gateway/internal/web/repositories"

	"github.com/gofiber/fiber/v3"
)

func GetSettings(c fiber.Ctx) error {
	apiSettings, err := repositories.FetchApiSettings()

	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "something went wrong",
		})
	}

	config, err2 := repositories.FetchConfigSettings()

	if err2 != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "something went wrong",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"data": fiber.Map{
			"api":    apiSettings,
			"config": config,
		},
	})
}

func PatchUpdateSettings(c fiber.Ctx) error {
	updates := c.Locals("updates").(analytics.PatchSettingsRequest)

	updatedIds, err := analytics.UpdateSettings(updates)
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

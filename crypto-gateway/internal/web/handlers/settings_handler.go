package handlers

import (
	"context"
	"crypto-gateway/internal/modules/strategy/repo"
	"crypto-gateway/internal/modules/strategy/service"
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
	updates := c.Locals("updates").(service.PatchSettingsRequest)

	updatedIds, err := service.UpdateSettings(updates)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Failed to update endpoints",
		})
	}

	go func() {
		for _, id := range updatedIds {
			api, cooldown := repo.GetApiAndCooldownByID(id)
			service.StOrchestrator.AdjustAPITaskCooldown(context.Background(), api, cooldown)
			service.StOrchestrator.LaunchNeededAPI(context.Background())
		}
	}()

	return c.SendStatus(fiber.StatusOK)
}

package handlers

import (
	"crypto-gateway/internal/db"

	"github.com/gofiber/fiber/v3"
)

func GetSettings(c fiber.Ctx) error {
	settings, err := db.FetchSettings()

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
	id := c.Locals("id").(int)
	cooldown := c.Locals("cooldown").(int)

	err := db.UpdateCooldown(id, cooldown)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "something went wrong",
		})
	}

	go updateApiCooldown(id)

	return c.SendStatus(fiber.StatusOK)
}

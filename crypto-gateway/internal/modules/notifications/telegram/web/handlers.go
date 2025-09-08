package web

import (
	"crypto-gateway/internal/modules/notifications/telegram/repo"
	"strconv"

	"github.com/gofiber/fiber/v3"
)

func TelegramNotificationGet(c fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid telegram notification id",
		})
	}

	info, err := repo.GetTelegramNotificationInfo(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid telegram notification id",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"data": info,
	})
}

func TelegramNotificationPost(c fiber.Ctx) error {
	raw := c.Locals("validatedBody")
	info := raw.(repo.TelegramNotificationInfo)

	if err := repo.SaveTelegramNotificationInfo(info); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Failed to save notification: " + err.Error(),
		})
	}

	return c.SendStatus(fiber.StatusOK)
}

func TelegramNotificationPatch(c fiber.Ctx) error {
	raw := c.Locals("validatedBody")
	if raw == nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Missing validated body: ",
		})
	}

	patch, ok := raw.(repo.TelegramNotificationPatch)
	if !ok {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Invalid body type",
		})
	}

	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid telegram notification id",
		})
	}

	if err := repo.UpdateTelegramNotificationInfo(id, patch); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Update failed: "+err.Error())
	}

	return c.SendStatus(fiber.StatusOK)
}

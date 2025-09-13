package web

import (
	"crypto-gateway/internal/modules/notifications/telegram/repo"
	"fmt"
	"strconv"

	"github.com/gofiber/fiber/v3"
)

func TelegramNotificationGet(c fiber.Ctx) error {
	idStr := c.Query("id", "0")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		id = 0
	}

	info, err := repo.GetTelegramNotificationInfo(id)
	if err != nil {
		fmt.Println(err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid telegram message id",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"data": info,
	})
}

func TelegramNotificationPost(c fiber.Ctx) error {
	raw := c.Locals("validatedBody")
	input := raw.(repo.TelegramNotificationCreate)

	res, err := repo.SaveTelegramNotificationInfo(input)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Failed to save notification: " + err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"pk": res,
	})
}

func TelegramNotificationPatch(c fiber.Ctx) error {
	raw := c.Locals("validatedBody")
	if raw == nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Missing validated body"})
	}

	var patch repo.TelegramNotificationPatch
	switch v := raw.(type) {
	case repo.TelegramNotificationPatch:
		patch = v
	case *repo.TelegramNotificationPatch:
		patch = *v
	default:
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Invalid body type"})
	}

	idStr := c.Query("id", "0")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": fmt.Sprintf("invalid telegram notification id: %v", idStr)})
	}

	if err := repo.UpdateTelegramNotificationInfo(id, patch); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Update failed: "+err.Error())
	}

	return c.SendStatus(fiber.StatusOK)
}

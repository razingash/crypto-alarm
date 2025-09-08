package web

import (
	"crypto-gateway/internal/modules/notifications/telegram/repo"
	"encoding/json"
	"unicode/utf8"

	"github.com/gofiber/fiber/v3"
)

func ValidateTelegramNotificationPost(c fiber.Ctx) error {
	var input repo.TelegramNotificationInfo
	if err := json.Unmarshal(c.Body(), &input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	if input.Token == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Missing field: token")
	}

	if input.ChatId == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Missing field: chat_id")
	}

	if input.Message == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Missing field: message")
	}

	if utf8.RuneCountInString(input.Message) > 1000 {
		return fiber.NewError(fiber.StatusBadRequest, "Message too long (max 1000 characters)")
	}

	if input.Cooldown < 0 {
		return fiber.NewError(fiber.StatusBadRequest, "Cooldown must be non-negative")
	}

	c.Locals("validatedBody", input)

	return c.Next()
}

func ValidateTelegramNotificationPatch(c fiber.Ctx) error {
	var input repo.TelegramNotificationPatch
	if err := json.Unmarshal(c.Body(), &input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	if input.Message != nil && utf8.RuneCountInString(*input.Message) > 1000 {
		return fiber.NewError(fiber.StatusBadRequest, "Message too long (max 1000 characters)")
	}

	if input.Cooldown != nil && *input.Cooldown < 0 {
		return fiber.NewError(fiber.StatusBadRequest, "Cooldown must be non-negative")
	}

	c.Locals("validatedBody", input)

	return c.Next()
}

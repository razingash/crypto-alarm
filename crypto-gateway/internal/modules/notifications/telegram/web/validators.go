package web

import (
	"crypto-gateway/internal/modules/notifications/telegram/repo"
	"encoding/json"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/gofiber/fiber/v3"
)

func ValidateTelegramNotificationPost(c fiber.Ctx) error {
	var input repo.TelegramNotificationCreate
	if err := json.Unmarshal(c.Body(), &input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	if input.BotID == nil && input.Bot == nil {
		return fiber.NewError(fiber.StatusBadRequest, "Either bot_id or bot must be provided")
	}

	if input.BotID != nil {
		if input.Message.ElementId == "" {
			return fiber.NewError(fiber.StatusBadRequest, "Missing field: message.element_id")
		}
		if input.Message.Message == "" {
			return fiber.NewError(fiber.StatusBadRequest, "Missing field: message.message")
		}
		if utf8.RuneCountInString(input.Message.Message) > 1000 {
			return fiber.NewError(fiber.StatusBadRequest, "Message too long (max 1000 characters)")
		}
	}

	if input.Bot.Name == "new" {
		if input.Bot.Token == "" {
			return fiber.NewError(fiber.StatusBadRequest, "Missing field: bot.token")
		}
		if input.Bot.ChatId == "" {
			return fiber.NewError(fiber.StatusBadRequest, "Missing field: bot.chat_id")
		}
	} else if input.Bot.Token == "" && input.Bot.ChatId == "" {
		if input.Bot.Name == "" {
			return fiber.NewError(fiber.StatusBadRequest, "Missing field: bot.name")
		}
		isExists, err := repo.IsBotExists(input.Bot.Name)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "Failed to check bot existence")
		}
		if !isExists {
			return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("Bot %v does not exist", input.Bot.Name))
		}
	}

	c.Locals("validatedBody", input)
	return c.Next()
}

func ValidateTelegramNotificationPatch(c fiber.Ctx) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(c.Body(), &raw); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	if v, ok := raw["params"]; ok {
		var inner map[string]json.RawMessage
		if err := json.Unmarshal(v, &inner); err == nil {
			raw = inner
		}
	} else if v, ok := raw["data"]; ok {
		var inner map[string]json.RawMessage
		if err := json.Unmarshal(v, &inner); err == nil {
			raw = inner
		}
	}

	var input repo.TelegramNotificationPatch

	if v, ok := raw["bot"]; ok {
		var botObj struct {
			Name   string `json:"name"`
			Token  string `json:"token,omitempty"`
			ChatId string `json:"chat_id,omitempty"`
		}
		if err := json.Unmarshal(v, &botObj); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid bot object"})
		}
		if botObj.Name != "" {
			input.BotName = &botObj.Name
		}
	} else if v, ok := raw["bot_name"]; ok {
		var s string
		if err := json.Unmarshal(v, &s); err == nil && s != "" {
			input.BotName = &s
		}
	}

	if v, ok := raw["message"]; ok {
		var msgObj struct {
			ElementId *string `json:"element_id,omitempty"`
			Message   *string `json:"message,omitempty"`
			Signal    *bool   `json:"signal,omitempty"`
		}
		if err := json.Unmarshal(v, &msgObj); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid message object"})
		}
		if msgObj.ElementId != nil {
			input.ElementId = msgObj.ElementId
		}
		if msgObj.Message != nil {
			input.Message = msgObj.Message
		}
		if msgObj.Signal != nil {
			input.Signal = msgObj.Signal
		}
	}

	if input.Message != nil && utf8.RuneCountInString(*input.Message) > 1000 {
		return fiber.NewError(fiber.StatusBadRequest, "Message too long (max 1000 characters)")
	}
	if input.ElementId != nil && strings.TrimSpace(*input.ElementId) == "" {
		return fiber.NewError(fiber.StatusBadRequest, "element_id cannot be empty")
	}
	if input.BotName != nil && strings.TrimSpace(*input.BotName) == "" {
		return fiber.NewError(fiber.StatusBadRequest, "bot name cannot be empty")
	}

	c.Locals("validatedBody", input)

	fmt.Printf("validated patch: %+v\n", input)
	return c.Next()
}

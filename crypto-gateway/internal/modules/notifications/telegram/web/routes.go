package web

import (
	"github.com/gofiber/fiber/v3"
)

func SetupTriggersRoutes(app *fiber.App) {
	group := app.Group("/api/v1/modules/notification-telegram")

	group.Get("/", TelegramNotificationGet)
	group.Post("/", TelegramNotificationPost, ValidateTelegramNotificationPost)
	group.Patch("/", TelegramNotificationPatch, ValidateTelegramNotificationPatch)
}

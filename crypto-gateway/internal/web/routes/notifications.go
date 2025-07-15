package routes

import (
	"crypto-gateway/internal/web/handlers"
	"crypto-gateway/internal/web/middlewares/validators"

	"github.com/gofiber/fiber/v3"
)

func SetupNotificationRoutes(app *fiber.App) {
	group := app.Group("/api/v1/notifications")

	group.Get("/vapid-key", handlers.GetVAPIDKey)
	group.Post("/subscribe", handlers.SavePushSubscription, validators.ValidatePostPushSubscriptions)
}

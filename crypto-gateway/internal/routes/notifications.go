package routes

import (
	"crypto-gateway/internal/handlers"
	"crypto-gateway/internal/middlewares"
	"crypto-gateway/internal/middlewares/api_validators"

	"github.com/gofiber/fiber/v3"
)

func SetupNotificationRoutes(app *fiber.App) {
	group := app.Group("/api/v1/notifications")

	group.Get("/vapid-key", handlers.GetVAPIDKey)
	group.Post("/subscribe", handlers.SavePushSubscription, api_validators.ValidatePostPushSubscriptions, middlewares.IsAuthorized)

	// only from python service
	group.Post("/push", handlers.PushNotificationsPost, api_validators.ValidatePushNotifications)
}

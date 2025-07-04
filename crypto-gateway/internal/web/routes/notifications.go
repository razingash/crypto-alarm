package routes

import (
	"crypto-gateway/internal/web/handlers"
	"crypto-gateway/internal/web/middlewares"
	"crypto-gateway/internal/web/middlewares/api_validators"

	"github.com/gofiber/fiber/v3"
)

func SetupNotificationRoutes(app *fiber.App) {
	group := app.Group("/api/v1/notifications")

	group.Get("/vapid-key", handlers.GetVAPIDKey)
	group.Post("/subscribe", handlers.SavePushSubscription, api_validators.ValidatePostPushSubscriptions, middlewares.IsAuthorized)

	// only from python service
	group.Post("/push", handlers.PushNotificationsPost, api_validators.ValidatePushNotifications)
}

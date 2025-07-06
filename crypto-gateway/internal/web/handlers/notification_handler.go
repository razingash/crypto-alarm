package handlers

import (
	"crypto-gateway/config"
	"crypto-gateway/internal/web/db"
	"fmt"

	"github.com/gofiber/fiber/v3"
)

func GetVAPIDKey(c fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"vapidPublicKey": config.Vapid_Public_Key,
	})
}

func SavePushSubscription(c fiber.Ctx) error {
	endpoint := c.Locals("endpoint").(string)
	p256dh := c.Locals("p256dh").(string)
	auth := c.Locals("auth").(string)

	err := db.SaveSubscription(endpoint, p256dh, auth)
	if err != nil {
		fmt.Println(err)
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	return c.SendStatus(fiber.StatusOK)
}

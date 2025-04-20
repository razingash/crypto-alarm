package handlers

import (
	"crypto-gateway/config"
	"crypto-gateway/internal/db"
	"fmt"

	"github.com/gofiber/fiber/v3"
)

func GetVAPIDKey(c fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"vapidPublicKey": config.Vapid_Public_Key,
	})
}

func SavePushSubscription(c fiber.Ctx) error {
	userUUID := c.Locals("userUUID").(string)
	userID, err := db.GetIdbyUuid(userUUID)

	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	endpoint := c.Locals("endpoint").(string)
	p256dh := c.Locals("p256dh").(string)
	auth := c.Locals("auth").(string)

	err = db.SaveSubscription(endpoint, p256dh, auth, userID)
	if err != nil {
		fmt.Println(err)
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	return c.SendStatus(fiber.StatusOK)
}

func PushNotificationsPost(c fiber.Ctx) error {
	message := c.Locals("message").(string)
	formulas := c.Locals("formulas").([]int)

	err := db.SendPushNotifications(formulas, message)

	if err == nil {
		return c.SendStatus(fiber.StatusOK)
	}

	return c.SendStatus(fiber.StatusInternalServerError)
}

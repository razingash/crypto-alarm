package validators

import (
	"encoding/json"

	"github.com/gofiber/fiber/v3"
)

func ValidatePostPushSubscriptions(c fiber.Ctx) error {
	var body struct {
		Endpoint string `json:"endpoint"`
		P256dh   string `json:"p256dh"`
		Auth     string `json:"auth"`
	}

	if err := json.Unmarshal(c.Body(), &body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid JSON",
		})
	}

	c.Locals("endpoint", body.Endpoint)
	c.Locals("p256dh", body.P256dh)
	c.Locals("auth", body.Auth)
	return c.Next()
}

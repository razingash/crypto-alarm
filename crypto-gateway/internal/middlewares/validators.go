package middlewares

import (
	"encoding/json"

	"github.com/gofiber/fiber/v3"
)

func ValidateAuthenticationInfo(c fiber.Ctx) error {
	var body struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := json.Unmarshal(c.Body(), &body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid JSON",
		})
	}

	if len(body.Username) < 5 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"username": "Username must be at least 6 characters long",
		})
	}

	if len(body.Password) < 5 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"password": "Password must be at least 6 characters long",
		})
	}

	c.Locals("body", body)

	return c.Next()
}

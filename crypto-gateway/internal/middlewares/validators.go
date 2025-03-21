package middlewares

import (
	"github.com/gofiber/fiber/v3"
)

func ValidateAuthenticationInfo(c fiber.Ctx) error {
	username := c.FormValue("username")
	password := c.FormValue("password")

	if len(username) < 6 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Username must be at least 6 characters long",
		})
	}

	if len(password) < 6 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Password must be at least 6 characters long",
		})
	}

	return c.Next()
}

package handlers

import (
	"context"
	"crypto-gateway/crypto-gateway/internal/db"
	"crypto-gateway/crypto-gateway/internal/triggers"

	"github.com/gofiber/fiber/v3"
)

// отправляет датасет актуальных параметров клавиатуры
func Keyboard(c fiber.Ctx) error {
	keyboard := make(map[string]interface{})

	rows, err := db.DB.Query(context.Background(), `
		SELECT crypto_api.api, crypto_params.parameter
		FROM crypto_api
		JOIN crypto_params ON crypto_api.id = crypto_params.crypto_api_id
		WHERE crypto_api.is_actual=true AND crypto_params.is_active=true;
	`)

	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}
	defer rows.Close()

	apiData := make(map[string][]string)
	for rows.Next() {
		var api string
		var parameter string

		err := rows.Scan(&api, &parameter)
		if err != nil {
			return c.SendStatus(fiber.StatusInternalServerError)
		}

		apiData[api] = append(apiData[api], parameter)
	}

	if err := rows.Err(); err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	rows2, err := db.DB.Query(context.Background(), `
		SELECT crypto_currencies.currency
		FROM crypto_currencies
		WHERE crypto_currencies.is_available=true;
	`)

	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	defer rows2.Close()

	var currencies []string

	for rows2.Next() {
		var currency string

		err := rows2.Scan(&currency)
		if err != nil {
			return c.SendStatus(fiber.StatusInternalServerError)
		}

		currencies = append(currencies, currency)
	}

	if err := rows2.Err(); err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	keyboard["api"] = apiData
	keyboard["currencies"] = currencies

	return c.JSON(keyboard)
}

func Formula(c fiber.Ctx) error {
	expression := c.Locals("formula").(string)

	errCode := triggers.Analys(expression)
	switch errCode {
	case 1:
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unknown symbol",
		})
	default:
		return c.SendStatus(fiber.StatusOK)
	}
}

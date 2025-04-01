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
	// Analys("(BOTBTC_price+BOTBTC_price24hr)*2==10.2+3^2")
	errCode := triggers.Analys(expression)
	switch errCode {
	case 0:
		userUUID := c.Locals("userUUID").(string)
		err := db.SaveFormula(expression, userUUID)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "error during saving formula",
			})
		}
		return c.SendStatus(fiber.StatusOK)
	case 1:
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unknown symbol",
		})
	case 2:
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Incorrect variable",
		})
	case 3: // unused
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "there is no such variable in the database",
		})
	case 4:
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "The variable is outdated",
		})
	case 5:
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Incorrect sequence of symbols",
		})
	case 6:
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Incorrect brackets",
		})
	case 7:
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "There are no comparison operators",
		})
	case 10:
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "database error",
		})
	default:
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unprocessed error",
		})
	}
}

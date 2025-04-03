package handlers

import (
	"context"
	"crypto-gateway/crypto-gateway/internal/db"

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

func FormulaPost(c fiber.Ctx) error {
	expression := c.Locals("formula").(string)
	userUUID := c.Locals("userUUID").(string)

	err := db.SaveFormula(expression, userUUID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "error during saving formula",
		})
	}
	return c.SendStatus(fiber.StatusOK)
}

func FormulaPatch(c fiber.Ctx) error {
	formulaId := c.Locals("formulaId").(string)
	data := c.Locals("updateData").(map[string]interface{})

	errCode := db.UpdateUserFormula(formulaId, data)

	switch errCode {
	case 0:
		return c.SendStatus(fiber.StatusOK)
	default:
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "unprocessed error",
		})
	}
}

func Formulas(c fiber.Ctx) error {
	userUUID := c.Locals("userUUID").(string)

	formulas, err := db.GetUserFormulas(userUUID)

	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "something went wrong",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"data": formulas,
	})
}

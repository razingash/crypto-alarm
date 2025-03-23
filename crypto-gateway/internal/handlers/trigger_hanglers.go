package handlers

import (
	"context"
	"crypto-gateway/crypto-gateway/internal/db"

	"github.com/gofiber/fiber/v3"
)

// отправляет датасет актуальных параметров клавиатуры
func Keyboard(c fiber.Ctx) error {
	keyboard := make(map[string][]string)

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

	for rows.Next() {
		var api string
		var parameter string

		err := rows.Scan(&api, &parameter)
		if err != nil {
			return c.SendStatus(fiber.StatusInternalServerError)
		}

		keyboard[api] = append(keyboard[api], parameter)
	}

	if err := rows.Err(); err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	return c.JSON(keyboard)
}

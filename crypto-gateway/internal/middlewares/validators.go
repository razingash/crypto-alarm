package middlewares

import (
	"context"
	"crypto-gateway/crypto-gateway/internal/db"
	"crypto-gateway/crypto-gateway/internal/triggers"
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

	if len(body.Username) < 6 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"username": "Username must be at least 6 characters long",
		})
	}

	if len(body.Password) < 6 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"password": "Password must be at least 6 characters long",
		})
	}

	c.Locals("body", body)

	return c.Next()
}

func ValidateFormulaId(c fiber.Ctx) error {
	var payload map[string]interface{}

	if err := json.Unmarshal(c.Body(), &payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid JSON",
		})
	}

	formulaId, ok := payload["formula_id"].(string)
	if !ok || formulaId == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "formula_id is required",
		})
	}
	delete(payload, "formula_id")

	userUUID := c.Locals("userUUID").(string)
	userId, err := db.GetIdbyUuid(userUUID)

	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "user does not exist",
		})
	}

	var count int
	err2 := db.DB.QueryRow(context.Background(), `
		SELECT COUNT(*) 
		FROM trigger_formula
		WHERE id = $1 AND owner_id = $2
	`, formulaId, userId).Scan(&count)

	if err2 != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Database error",
		})
	}

	if count == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "formula does not exists",
		})
	}

	var allowedFields = map[string]struct{}{
		"formula":        {},
		"name":           {},
		"description":    {},
		"is_notified":    {},
		"is_active":      {},
		"is_history_on":  {},
		"is_shutted_off": {},
		"last_triggered": {},
	}

	validData := make(map[string]interface{})
	for key, value := range payload {
		if _, exists := allowedFields[key]; exists {
			validData[key] = value
		}
	}

	if len(validData) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "no valid fields provided for update",
		})
	}

	c.Locals("formulaId", formulaId)
	c.Locals("updateData", validData)
	return c.Next()
}

func ValidateFormula(c fiber.Ctx) error {
	var body struct {
		Formula string `json:"formula"`
	}

	if err := json.Unmarshal(c.Body(), &body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid JSON",
		})
	}

	errCode := triggers.Analys(body.Formula)
	switch errCode {
	case 0:
		c.Locals("formula", body.Formula)
		return c.Next()
	case 1:
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Unknown symbol",
		})
	case 2:
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Incorrect variable",
		})
	case 3: // unused
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "there is no such variable in the database",
		})
	case 4:
		return c.Status(fiber.StatusNotAcceptable).JSON(fiber.Map{
			"error": "The variable is outdated",
		})
	case 5:
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Incorrect sequence of symbols",
		})
	case 6:
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Incorrect brackets",
		})
	case 7:
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "There are no comparison operators",
		})
	case 10:
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "database error",
		})
	default:
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "unprocessed error",
		})
	}
}

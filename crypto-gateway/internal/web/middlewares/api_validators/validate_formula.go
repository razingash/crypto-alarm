package api_validators

import (
	"crypto-gateway/internal/web/middlewares/field_validators"
	"encoding/json"

	"github.com/gofiber/fiber/v3"
)

func ValidateFormulaPost(c fiber.Ctx) error {
	var body struct {
		Formula    string `json:"formula"`
		FormulaRaw string `json:"formula_raw"` // нет нормального валидатора!
		Name       string `json:"name"`        // идентично
	}

	if err := json.Unmarshal(c.Body(), &body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid JSON",
		})
	}

	variables, errCode := field_validators.ValidateTriggerFormulaFormula(body.Formula)
	switch errCode {
	case 0:
		c.Locals("formula", body.Formula)
		c.Locals("formula_raw", body.FormulaRaw)
		c.Locals("name", body.Name)
		c.Locals("variables", variables)
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

func ValidateFormulaPatch(c fiber.Ctx) error {
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

	errCode := field_validators.ValidateTriggerFormulaId(formulaId)
	switch errCode {
	case 0: // дальше по коду
	case 2:
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Database error",
		})
	case 3:
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "formula does not exists",
		})
	}

	validator := field_validators.FormulaValidator{
		Formula:     field_validators.ValidateText(3, 50000),
		FormulaRaw:  field_validators.ValidateText(3, 50000),
		Name:        field_validators.ValidateText(0, 150),
		Description: field_validators.ValidateText(0, 1500),
		IsNotified:  field_validators.ValidateBool,
		IsActive:    field_validators.ValidateBool,
		IsHistoryOn: field_validators.ValidateBool,
		Cooldown:    field_validators.ValidateCooldown,
	}

	fieldValidators := map[string]func(interface{}) string{
		"formula":       validator.Formula,
		"formula_raw":   validator.FormulaRaw,
		"name":          validator.Name,
		"description":   validator.Description,
		"is_notified":   validator.IsNotified,
		"is_active":     validator.IsActive,
		"is_history_on": validator.IsHistoryOn,
		"cooldown":      validator.Cooldown,
	}

	validData := make(map[string]interface{})
	errors := make(map[string]string)
	for key, value := range payload {
		if validatorFunc, exists := fieldValidators[key]; exists {
			if err := validatorFunc(value); err != "" {
				errors[key] = err
			} else {
				validData[key] = value
			}
		}
	}

	if len(errors) > 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"errors": errors,
		})
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

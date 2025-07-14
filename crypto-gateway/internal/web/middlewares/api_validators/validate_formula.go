package api_validators

import (
	"crypto-gateway/internal/web/middlewares/field_validators"
	"crypto-gateway/internal/web/repositories"
	"encoding/json"

	"github.com/gofiber/fiber/v3"
)

func ValidateStrategyPost(c fiber.Ctx) error {
	var body struct {
		Name        string                            `json:"name"`
		Description string                            `json:"description"`
		Expressions []repositories.StrategyExpression `json:"conditions"`
	}

	if err := json.Unmarshal(c.Body(), &body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid JSON",
		})
	}

	if body.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Missing name",
		})
	}

	if len(body.Expressions) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Conditions cannot be empty",
		})
	}

	var allVariables []repositories.CryptoVariable
	unique := make(map[string]struct{})

	for _, condition := range body.Expressions {
		if condition.Formula == "" || condition.FormulaRaw == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Each condition must have formula and formula_raw",
			})
		}

		variables, err := field_validators.ValidateStrategyExpression(condition.Formula)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		for _, v := range variables {
			key := v.Currency + ":" + v.Variable
			if _, exists := unique[key]; !exists {
				unique[key] = struct{}{}
				allVariables = append(allVariables, v)
			}
		}
	}

	c.Locals("name", body.Name)
	c.Locals("description", body.Description)
	c.Locals("expressions", body.Expressions)
	c.Locals("variables", allVariables)

	return c.Next()
}

func ValidateStrategyPatch(c fiber.Ctx) error {
	var payload map[string]interface{}

	if err := json.Unmarshal(c.Body(), &payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid JSON",
		})
	}

	strategyID, ok := payload["strategy_id"].(string)
	if !ok || strategyID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "strategy_id is required",
		})
	}
	delete(payload, "strategy_id")

	validator := field_validators.StrategyValidator{
		Name:        field_validators.ValidateText(0, 150),
		Description: field_validators.ValidateText(0, 1500),
		IsNotified:  field_validators.ValidateBool,
		IsActive:    field_validators.ValidateBool,
		IsHistoryOn: field_validators.ValidateBool,
		Cooldown:    field_validators.ValidateCooldown,
		Conditions: func(value interface{}) string {
			arr, ok := value.([]interface{})
			if !ok {
				return "Invalid conditions format"
			}

			for _, item := range arr {
				cond, ok := item.(map[string]interface{})
				if !ok {
					return "Invalid condition entry format"
				}

				if err := field_validators.ValidateText(3, 50000)(cond["formula"]); err != "" {
					return "Invalid formula"
				}
				if err := field_validators.ValidateText(3, 50000)(cond["formula_raw"]); err != "" {
					return "Invalid formula_raw"
				}
			}
			return ""
		},
	}

	fieldValidators := map[string]func(interface{}) string{
		"name":          validator.Name,
		"description":   validator.Description,
		"is_notified":   validator.IsNotified,
		"is_active":     validator.IsActive,
		"is_history_on": validator.IsHistoryOn,
		"cooldown":      validator.Cooldown,
		"conditions":    validator.Conditions,
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

	c.Locals("strategyID", strategyID)
	c.Locals("updateData", validData)
	return c.Next()
}

package handlers

import (
	"crypto-gateway/internal/web/repositories"
	"strconv"

	"github.com/gofiber/fiber/v3"
)

func VariableGet(c fiber.Ctx) error {
	defaultLimit := 10
	defaultPage := 1

	limit, err := strconv.Atoi(c.Query("limit"))
	if err != nil || limit <= 0 {
		limit = defaultLimit
	}

	page, err := strconv.Atoi(c.Query("page"))
	if err != nil || page <= 0 {
		page = defaultPage
	}

	variableID := c.Query("id")
	variables, hasNext, err := repositories.GetVariables(limit, page, variableID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	if variableID == "" {
		if variables == nil {
			variables = []repositories.Variable{}
		}
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"data":     variables,
			"has_next": hasNext,
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"data": variables[0],
	})
}

func VariablePost(c fiber.Ctx) error {
	symbol := c.Locals("symbol").(string)
	name := c.Locals("name").(string)
	description := c.Locals("description").(string)
	formula := c.Locals("formula").(string)
	formulaRaw := c.Locals("formula_raw").(string)
	tokens := c.Locals("tokens").([]repositories.Token)

	id, err := repositories.CreateVariable(symbol, name, description, formula, formulaRaw, tokens)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"variable_id": id,
	})
}

func VariablePatch(c fiber.Ctx) error {
	variableId := c.Locals("variable_id").(int)
	input := c.Locals("input").(repositories.UpdateVariableStruct)
	tokens := c.Locals("tokens").([]repositories.Token)

	isFormulaUpdated, err := repositories.UpdateVariable(variableId, &input, tokens)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update variable",
		})
	}
	if isFormulaUpdated {
		go updateStrategiesRelatedToVariable(variableId)
	}
	return c.SendStatus(fiber.StatusOK)
}

func VariableDelete(c fiber.Ctx) error {
	variableIdStr := c.Params("id")
	variableId, err := strconv.Atoi(variableIdStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid variable ID",
		})
	}
	err = repositories.DeleteVariableById(variableId)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.SendStatus(fiber.StatusOK)
}

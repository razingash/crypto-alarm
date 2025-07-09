package handlers

import (
	"context"
	"crypto-gateway/internal/web/db"
	"crypto-gateway/internal/web/middlewares/field_validators"
	"crypto-gateway/internal/web/repositories"
	"fmt"
	"strconv"

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
	raw_expression := c.Locals("formula_raw").(string)
	name := c.Locals("name").(string)
	variables := c.Locals("variables").([]repositories.CryptoVariable)

	id, err := repositories.SaveFormula(expression, raw_expression, name)
	if err != nil {
		fmt.Println(err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "error during saving formula",
		})
	}

	err2 := repositories.SaveCryptoVariables(id, variables)
	if err2 != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "error during saving formula variables",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"id": id,
	})
}

func FormulaPatch(c fiber.Ctx) error {
	formulaId := c.Locals("formulaId").(string)
	data := c.Locals("updateData").(map[string]interface{})

	err := repositories.UpdateUserFormula(formulaId, data)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	if _, hasFormula := data["formula"]; hasFormula {
		go updateFormulaInGraph(formulaId)
	}

	if isActiveRaw, hasIsActive := data["is_active"]; hasIsActive {
		if isActive, ok := isActiveRaw.(bool); ok {
			id, _ := strconv.Atoi(formulaId)
			if isActive {
				go addFormulaToGraph(id)
			} else {
				go deleteFormulaFromGraph(id)
			}
		}
	}
	return c.SendStatus(fiber.StatusOK)
}

func FormulaDelete(c fiber.Ctx) error {
	formulaId := c.Query("formula_id")
	if formulaId == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "formula_id is required",
		})
	}

	err := field_validators.ValidateTriggerFormulaId(formulaId)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	err2 := repositories.DeleteUserFormula(formulaId)
	if err2 != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err2.Error(),
		})
	}

	id, _ := strconv.Atoi(formulaId)
	go deleteFormulaFromGraph(id)
	return c.SendStatus(fiber.StatusOK)
}

func FormulaHistoryGet(c fiber.Ctx) error {
	formulaIDStr := c.Params("id")
	formulaID, err := strconv.Atoi(formulaIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid formula ID",
		})
	}

	prevCursor, err := strconv.Atoi(c.Query("prevCursor"))
	if err != nil || prevCursor <= 0 {
		prevCursor = 0
	}

	limit, err := strconv.Atoi(c.Query("limit"))
	if err != nil || limit <= 0 {
		limit = 100
	}

	page, err := strconv.Atoi(c.Query("page"))
	if err != nil || page <= 0 {
		page = 1
	}

	hasNext, rawRows, err := repositories.GetFormulaHistory(formulaID, limit, page, prevCursor)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	resultMap := make(map[int64]map[string]interface{})
	var timestamps []int64

	for _, r := range rawRows {
		timestamp := r.Timestamp.Unix()

		if _, ok := resultMap[timestamp]; !ok {
			resultMap[timestamp] = map[string]interface{}{
				"timestamp": timestamp,
			}
			timestamps = append(timestamps, timestamp)
		}

		if r.VarName != "" && r.Value != "" {
			resultMap[timestamp][r.VarName] = r.Value
		}
	}

	var response []map[string]interface{}
	for _, item := range timestamps {
		response = append(response, resultMap[item])
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"data":     response,
		"has_next": hasNext,
	})
}

func FormulaGet(c fiber.Ctx) error {
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

	formulaID := c.Query("id")

	formulas, hasNext, err := repositories.GetFormulas(limit, page, formulaID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "something went wrong",
		})
	}

	if formulaID == "" {
		if formulas == nil {
			formulas = []repositories.UserFormula{}
		}
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"data":     formulas,
			"has_next": hasNext,
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"data": formulas[0],
	})
}

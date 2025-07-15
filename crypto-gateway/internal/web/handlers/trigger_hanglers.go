package handlers

import (
	"context"
	"crypto-gateway/internal/web/db"
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

func StrategyPost(c fiber.Ctx) error {
	name := c.Locals("name").(string)
	description := c.Locals("description").(string)
	expressions := c.Locals("expressions").([]repositories.StrategyExpression)
	variables := c.Locals("variables").([]repositories.CryptoVariable)

	id, err := repositories.SaveStrategy(name, description, expressions, variables)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "error during saving formula",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"id": id,
	})
}

func StrategyPatch(c fiber.Ctx) error {
	strategyID := c.Locals("strategyID").(int)
	data := c.Locals("updateData").(map[string]interface{})

	err := repositories.UpdateStrategy(strategyID, data)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	if rawConditions, ok := data["conditions"]; ok {
		if conditionsSlice, ok := rawConditions.([]interface{}); ok {
			var conditions []map[string]interface{}
			for _, item := range conditionsSlice {
				if condMap, ok := item.(map[string]interface{}); ok {
					conditions = append(conditions, condMap)
				} else {
					return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
						"error": "Invalid condition format",
					})
				}
			}

			err = repositories.UpdateStrategyConditions(strategyID, conditions)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error": err.Error(),
				})
			}
		} else {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "conditions should be an array",
			})
		}
	}

	if _, hasFormula := data["formula"]; hasFormula {
		go updateStrategyInGraph(strategyID)
	}
	if isActiveRaw, hasIsActive := data["is_active"]; hasIsActive {
		if isActive, ok := isActiveRaw.(bool); ok {
			if isActive {
				go updateStrategyInGraph(strategyID)
			} else {
				go deleteStrategyFromGraph(strategyID)
			}
		}
	}
	return c.SendStatus(fiber.StatusOK)
}

func StrategyDelete(c fiber.Ctx) error {
	strategyIDStr := c.Params("id")
	strategyID, err := strconv.Atoi(strategyIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid strategy ID",
		})
	}

	formulaIDstr := c.Query("formula_id")
	if formulaIDstr == "" {
		err := repositories.DeleteStrategyById(strategyID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
	} else {
		formulaID, err2 := strconv.Atoi(formulaIDstr)
		if err2 != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid formula ID",
			})
		}
		err := repositories.DeleteFormulaById(formulaID)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid formula ID",
			})
		}
	}

	if formulaIDstr == "" { // удаление одной формулы из стратегии(обновление стратегии)
		go updateStrategyInGraph(strategyID)
	} else { // удаление стратегии
		go deleteStrategyFromGraph(strategyID)
	}

	return c.SendStatus(fiber.StatusOK)
}

func StrategyHistoryGet(c fiber.Ctx) error {
	strategyIDStr := c.Params("id")
	strategyID, err := strconv.Atoi(strategyIDStr)
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

	hasNext, rawRows, err := repositories.GetStrategyHistory(strategyID, limit, page, prevCursor)
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

func StrategyGet(c fiber.Ctx) error {
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

	strategyID := c.Query("id")

	strategies, hasNext, err := repositories.GetStrategiesWithFormulas(limit, page, strategyID)
	if err != nil {
		fmt.Println(err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "something went wrong",
		})
	}

	if strategyID == "" {
		if strategies == nil {
			strategies = []repositories.Strategy{}
		}
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"data":     strategies,
			"has_next": hasNext,
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"data": strategies[0],
	})
}

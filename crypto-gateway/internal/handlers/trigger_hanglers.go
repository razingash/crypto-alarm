package handlers

import (
	"context"
	"crypto-gateway/internal/db"
	"crypto-gateway/internal/middlewares/field_validators"
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
	name := c.Locals("name").(string)
	userUUID := c.Locals("userUUID").(string)
	variables := c.Locals("variables").([]db.CryptoVariable)

	id, err := db.SaveFormula(expression, name, userUUID)
	if err != nil {
		fmt.Println(err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "error during saving formula",
		})
	}

	err2 := db.SaveCryptoVariables(id, variables)
	if err2 != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "error during saving formula variables",
		})
	}
	go addFormulaToGraph(id)
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"id": id,
	})
}

func FormulaPatch(c fiber.Ctx) error {
	formulaId := c.Locals("formulaId").(string)
	data := c.Locals("updateData").(map[string]interface{})

	errCode := db.UpdateUserFormula(formulaId, data)

	switch errCode {
	case 0:
		if _, hasFormula := data["formula"]; hasFormula {
			go updateFormulaInGraph(formulaId)
		}

		if isActiveRaw, hasIsActive := data["is_active"]; hasIsActive {
			if isActive, ok := isActiveRaw.(bool); ok {
				if isActive {
					id, _ := strconv.Atoi(formulaId)
					go addFormulaToGraph(id)
				} else {
					go deleteFormulaFromGraph(formulaId)
				}
			}
		}
		return c.SendStatus(fiber.StatusOK)
	default:
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "unprocessed error",
		})
	}
}

func FormulaDelete(c fiber.Ctx) error {
	formulaId := c.Query("formula_id")
	if formulaId == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "formula_id is required",
		})
	}

	userUUID := c.Locals("userUUID").(string)
	errCode := field_validators.ValidateTriggerFormulaId(userUUID, formulaId)
	switch errCode {
	case 0: // дальше по коду
	case 1:
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "user does not exist",
		})
	case 2:
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Database error",
		})
	case 3:
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "formula does not exists",
		})
	}

	code := db.DeleteUserFormula(formulaId)

	switch code {
	case 2:
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Database error",
		})
	default:
		go deleteFormulaFromGraph(formulaId)
		return c.SendStatus(fiber.StatusOK)
	}
}

func FormulaGet(c fiber.Ctx) error {
	userUUID := c.Locals("userUUID").(string)

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

	formulas, hasNext, err := db.GetUserFormulas(userUUID, limit, page, formulaID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "something went wrong",
		})
	}

	if formulaID == "" {
		if formulas == nil {
			formulas = []db.UserFormula{}
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

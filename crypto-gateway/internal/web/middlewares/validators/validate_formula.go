package validators

import (
	"crypto-gateway/internal/web/repositories"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v3"
)

const (
	NUMBER     = "NUMBER"
	OPERATOR   = "OPERATOR"
	VARIABLE   = "VARIABLE"
	FUNCTION   = "FUNCTION"
	LPAREN     = "LPAREN"
	RPAREN     = "RPAREN"
	COMPARISON = "COMPARISON"
	UNKNOWN    = "UNKNOWN"
)

type Token struct {
	Type  string
	Value string
}

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

		variables, err := ValidateStrategyExpression(condition.Formula)
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

	strategyId, ok := payload["strategy_id"].(string)
	if !ok || strategyId == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "strategy_id is required",
		})
	}
	strategyID, err := strconv.Atoi(strategyId)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid strategy_id",
		})
	}
	delete(payload, "strategy_id")

	validator := StrategyValidator{
		Name:        ValidateText(0, 150),
		Description: ValidateText(0, 1500),
		IsNotified:  ValidateBool,
		IsActive:    ValidateBool,
		IsHistoryOn: ValidateBool,
		Cooldown:    ValidateCooldown,
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

				if err := ValidateText(3, 50000)(cond["formula"]); err != "" {
					return "Invalid formula"
				}
				if err := ValidateText(3, 50000)(cond["formula_raw"]); err != "" {
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

// ниже функционал для проверки синтаксиса formula из crypto_strategy

// проверяет формулу на валидность
func ValidateStrategyExpression(formula string) ([]repositories.CryptoVariable, error) {
	tokens, variables, err := tokenize(formula)

	if err != nil {
		return nil, err
	}

	err2 := validateTokens(tokens)

	return variables, err2
}

// проверка на синтаксис
func tokenize(expression string) ([]Token, []repositories.CryptoVariable, error) {
	var tokens []Token
	var variables []repositories.CryptoVariable

	tokenPatterns := []struct {
		Type    string
		Pattern string
	}{
		{NUMBER, `^\d+(\.\d+)?`},                  // Числа
		{OPERATOR, `^[+\-*/^]`},                   // Операторы
		{COMPARISON, `^(<=|>=|==|<|>)`},           // Операторы сравнения
		{VARIABLE, `^([a-zA-Z]+)_([a-zA-Z0-9]+)`}, // Переменные в формате crypto_variable
		{FUNCTION, `^(sqrt|abs)`},                 // Функции
		{LPAREN, `^\(`},                           // Левая скобка
		{RPAREN, `^\)`},                           // Правая скобка
	}

	for len(expression) > 0 {
		matched := false
		for _, pattern := range tokenPatterns {
			re := regexp.MustCompile(pattern.Pattern)
			match := re.FindString(expression)
			if match != "" {
				if pattern.Type == VARIABLE {
					parts := strings.Split(match, "_")
					if len(parts) != 2 { // неправильная переменная
						return nil, nil, fmt.Errorf("incorrect variable")
					}

					isValid, err := repositories.IsValidCryptoCurrency(parts[0])
					if err != nil {
						fmt.Println(err)
						return nil, nil, fmt.Errorf("database error")
					}
					if !isValid {
						fmt.Println("недопустимая переменная:", match)
						/*
							3 - переменной нет в базе данных
							4 - переменная не актуальна
						*/
						return nil, nil, fmt.Errorf("variable %v is outdated", VARIABLE)
					}

					isValid, err = repositories.IsValidVariable(parts[1])
					if err != nil {
						fmt.Println(err)
						return nil, nil, fmt.Errorf("database error")
					}
					if !isValid {
						fmt.Println("недопустимая переменная:", match)
						/*
							3 - переменной нет в базе данных
							4 - переменная не актуальна
						*/
						return nil, nil, fmt.Errorf("variable %v is outdated", VARIABLE)
					}
					variables = append(variables, repositories.CryptoVariable{Currency: parts[0], Variable: parts[1]})
				}

				tokens = append(tokens, Token{Type: pattern.Type, Value: match})
				expression = expression[len(match):]
				matched = true
				break
			}
		}

		if !matched { // неизвестный символ
			return nil, nil, fmt.Errorf("unknown symbol")
		}
	}

	return tokens, variables, nil
}

// проверка на правильность
func validateTokens(tokens []Token) error {
	stack := []Token{}
	lastTokenType := ""
	comparisonFound := false

	for i, token := range tokens {
		switch token.Type {
		/*
			1) Два числа, переменные, оператора, сравнения подряд - недопустимы
			2) Оператор, сравнение не могут стоять в начале или после другого оператора
			3) Оператор сравнения должен быть хоть один раз
			4) проверка скобок с помощью стека - при встрече '(' ложить в стек, при ')' убирать из стэка
				- если в конце стек не пуст или при встрече ')' то будет ошибка
			5) операторы требуют после себя выражений(скорее всего доработать стоит)
		*/
		case NUMBER, VARIABLE:
			// Два числа переменные подряд недопустимы
			if lastTokenType == NUMBER || lastTokenType == VARIABLE {
				return fmt.Errorf("incorrect sequence of symbols")
			}
		case OPERATOR:
			//Оператор не может стоять в начале или после другого оператора
			if i == 0 || lastTokenType == OPERATOR || lastTokenType == LPAREN || lastTokenType == COMPARISON {
				return fmt.Errorf("incorrect sequence of symbols")
			}
		case COMPARISON:
			// Оператор сравнения не может стоять в начале, после другого сравнения или скобки
			comparisonFound = true
			if i == 0 || lastTokenType == COMPARISON || lastTokenType == LPAREN {
				return fmt.Errorf("incorrect sequence of symbols")
			}
		case FUNCTION:
			// Функция требует открывающую скобку сразу после
			//stack = append(stack, token)
		case LPAREN:
			stack = append(stack, token)
		case RPAREN:
			if len(stack) == 0 {
				return fmt.Errorf("incorrect brackets")
			}
			stack = stack[:len(stack)-1]
		}

		lastTokenType = token.Type
		if (lastTokenType == RPAREN) &&
			i+1 < len(tokens) {
			nextToken := tokens[i+1]
			if nextToken.Type == NUMBER || nextToken.Type == VARIABLE || nextToken.Type == FUNCTION || nextToken.Type == LPAREN {
				return fmt.Errorf("missing operator between ')' and next token")
			}
		}
	}

	if !comparisonFound {
		return fmt.Errorf("there are no comparison operators")
	}
	if len(stack) > 0 {
		return fmt.Errorf("incorrect brackets")
	}

	return nil
}

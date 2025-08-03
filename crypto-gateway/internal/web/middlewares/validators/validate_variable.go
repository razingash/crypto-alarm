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

func ValidateVariablePost(c fiber.Ctx) error {
	var body struct {
		Symbol      string `json:"symbol"`
		Name        string `json:"name"`
		Description string `json:"description"`
		Formula     string `json:"formula"`
		FormulaRaw  string `json:"formula_raw"`
	}

	if err := json.Unmarshal(c.Body(), &body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid JSON",
		})
	}

	if len(body.Symbol) < 2 || len(body.Symbol) > 40 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Symbol must be between 2 and 40 characters",
		})
	}
	if !regexp.MustCompile(`^[a-zA-Z0-9_]+$`).MatchString(body.Symbol) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Symbol can only contain letters, digits, and underscores",
		})
	}

	if len(body.Name) < 5 || len(body.Name) > 255 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Name must be between 5 and 255 characters",
		})
	}

	if body.Formula == "" || body.FormulaRaw == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Formula and formula_raw cannot be empty",
		})
	}
	if strings.Contains(body.Name, "_") {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Formula name cannot contain '_', you can use camelCase to name them",
		})
	}

	tokens, err := ValidateVariable(body.Formula)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	c.Locals("symbol", body.Symbol)
	c.Locals("name", body.Name)
	c.Locals("description", body.Description)
	c.Locals("formula", body.Formula)
	c.Locals("formula_raw", body.FormulaRaw)
	c.Locals("tokens", tokens)

	return c.Next()
}

func ValidateVariablePatch(c fiber.Ctx) error {
	var body repositories.UpdateVariableStruct

	variableIdStr := c.Params("id")
	variableId, err := strconv.Atoi(variableIdStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid variable ID",
		})
	}

	if err := json.Unmarshal(c.Body(), &body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid JSON",
		})
	}

	if body.Symbol != nil {
		if len(*body.Symbol) < 2 || len(*body.Symbol) > 40 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Symbol must be between 2 and 40 characters",
			})
		}
		re := regexp.MustCompile(`^[a-zA-Z0-9_]+$`)
		if !re.MatchString(*body.Symbol) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Symbol can only contain Latin letters, digits, and underscores",
			})
		}
	}

	if body.Name != nil {
		if len(*body.Name) < 5 || len(*body.Name) > 255 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Name must be between 5 and 255 characters",
			})
		}
	}

	if (body.Formula != nil && body.FormulaRaw == nil) || (body.Formula == nil && body.FormulaRaw != nil) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Formula and formula_raw must be provided together",
		})
	}
	var tokens []repositories.Token
	if body.Formula != nil && body.FormulaRaw != nil {
		if *body.Formula == "" || *body.FormulaRaw == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Formula and formula_raw cannot be empty",
			})
		}
		ts, err := ValidateVariable(*body.Formula)
		tokens = ts
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
	}

	c.Locals("variable_id", variableId)
	c.Locals("input", body)
	c.Locals("tokens", tokens)

	return c.Next()
}

func ValidateVariable(variable string) ([]repositories.Token, error) {
	tokens, err := tokenizeVariable(variable)

	if err != nil {
		return nil, err
	}

	err2 := validateVariableTokens(tokens)

	return tokens, err2
}

// проверка на синтаксис
func tokenizeVariable(expression string) ([]repositories.Token, error) {
	var tokens []repositories.Token

	tokenPatterns := []struct {
		Type    string
		Pattern string
	}{
		{NUMBER, `^\d+(\.\d+)?`},                 // Числа
		{OPERATOR, `^[+\-*/^]`},                  // Операторы
		{COMPARISON, `^(<=|>=|==|<|>)`},          // Операторы сравнения ЗАПРЕЩЕНЫ
		{FUNCTION, `^(sqrt|abs)`},                // Функции
		{USER_VARIABLE, `^[a-zA-Z][a-zA-Z0-9]*`}, // пользовательские переменные, изначально не содержат криптовалюты и являются универсальными
		{VARIABLE, `^[a-zA-Z_][a-zA-Z0-9_]*`},    // Переменные бинанс - цифры, буквы и '_' на всякий случай
		{LPAREN, `^\(`},                          // Левая скобка
		{RPAREN, `^\)`},                          // Правая скобка
	}

	for len(expression) > 0 {
		matched := false
		for _, pattern := range tokenPatterns {
			re := regexp.MustCompile(pattern.Pattern)
			match := re.FindString(expression)
			if match != "" {
				tokens = append(tokens, repositories.Token{Type: pattern.Type, Value: match})
				expression = expression[len(match):]
				matched = true
				break
			}
		}

		if !matched { // неизвестный символ
			return nil, fmt.Errorf("unknown symbol")
		}
	}

	return tokens, nil
}

// проверка на правильность
func validateVariableTokens(tokens []repositories.Token) error {
	stack := []repositories.Token{}
	lastTokenType := ""

	for i, token := range tokens {
		switch token.Type {
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
			// Оператор сравнения должен отсутсвовать(по крайней мере пока нет булевых переменных)
			return fmt.Errorf("variable shouldn't have a comparison operator")
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

	if len(stack) > 0 {
		return fmt.Errorf("incorrect brackets")
	}

	return nil
}

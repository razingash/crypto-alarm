package field_validators

import (
	"context"
	"crypto-gateway/crypto-gateway/internal/db"
	"fmt"
	"regexp"
	"strings"
)

// добавить когда-нибудь тесты

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

// проверяет формулу на валидность
func ValidateTriggerFormulaFormula(formula string) ([]db.CryptoVariable, int) {
	tokens, variables, errCode := tokenize(formula)

	if errCode != 0 {
		return nil, errCode
	}

	errCode = validateTokens(tokens)

	return variables, errCode
}

// проверяет существует ли формула с таким id, и является ли пользователь её автором
func ValidateTriggerFormulaId(userUUID string, formulaId string) int {
	userId, err := db.GetIdbyUuid(userUUID)
	if err != nil {
		return 1
	}

	var count int
	err2 := db.DB.QueryRow(context.Background(), `
		SELECT COUNT(*) 
		FROM trigger_formula
		WHERE id = $1 AND owner_id = $2
	`, formulaId, userId).Scan(&count)

	if err2 != nil {
		return 2
	}

	if count == 0 {
		return 3
	}

	return 0
}

// проверка на синтаксис
func tokenize(expression string) ([]Token, []db.CryptoVariable, int) {
	var tokens []Token
	var variables []db.CryptoVariable

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
					// Проверяем, является ли переменная допустимой
					parts := strings.Split(match, "_")
					if len(parts) != 2 { // неправильная переменная
						return nil, nil, 2
					}

					isValid, err := db.IsValidCryptoCurrency(parts[0])
					if err != nil {
						fmt.Println(err)
						return nil, nil, 10
					}
					if !isValid {
						fmt.Println("недопустимая переменная:", match)
						/*
							3 - переменной нет в базе данных
							4 - переменная не актуальна
						*/
						return nil, nil, 4
					}

					isValid, err = db.IsValidVariable(parts[1])
					if err != nil {
						fmt.Println(err)
						return nil, nil, 10
					}
					if !isValid {
						fmt.Println("недопустимая переменная:", match)
						/*
							3 - переменной нет в базе данных
							4 - переменная не актуальна
						*/
						return nil, nil, 4
					}
					variables = append(variables, db.CryptoVariable{Currency: parts[0], Variable: parts[1]})
				}

				tokens = append(tokens, Token{Type: pattern.Type, Value: match})
				expression = expression[len(match):]
				matched = true
				break
			}
		}

		if !matched { // неизвестный символ
			return nil, nil, 1
		}
	}

	return tokens, variables, 0
}

// проверка на правильность
func validateTokens(tokens []Token) int {
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
				return 5
			}
		case OPERATOR:
			//Оператор не может стоять в начале или после другого оператора
			if i == 0 || lastTokenType == OPERATOR || lastTokenType == LPAREN || lastTokenType == COMPARISON {
				return 5
			}
		case COMPARISON:
			// Оператор сравнения не может стоять в начале, после другого сравнения или скобки
			comparisonFound = true
			if i == 0 || lastTokenType == COMPARISON || lastTokenType == LPAREN {
				return 5
			}
		case FUNCTION:
			// Функция требует открывающую скобку сразу после
			//stack = append(stack, token)
		case LPAREN:
			stack = append(stack, token)
		case RPAREN:
			if len(stack) == 0 {
				return 6
			}
			stack = stack[:len(stack)-1]
		}

		lastTokenType = token.Type
	}

	if !comparisonFound {
		return 7
	}
	if len(stack) > 0 {
		return 6
	}

	return 0
}

package field_validators

import (
	"context"
	"crypto-gateway/internal/web/db"
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
func ValidateTriggerFormulaFormula(formula string) ([]db.CryptoVariable, error) {
	tokens, variables, err := tokenize(formula)

	if err != nil {
		return nil, err
	}

	err2 := validateTokens(tokens)

	return variables, err2
}

// проверяет существует ли формула с таким id, и является ли пользователь её автором
func ValidateTriggerFormulaId(formulaId string) error {
	var count int
	err := db.DB.QueryRow(context.Background(), `
		SELECT COUNT(*) 
		FROM trigger_formula
		WHERE id = $1
	`, formulaId).Scan(&count)
	if err != nil {
		return fmt.Errorf("database error")
	}

	if count == 0 {
		return fmt.Errorf("formula does not exists")
	}

	return nil
}

// проверка на синтаксис
func tokenize(expression string) ([]Token, []db.CryptoVariable, error) {
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
						return nil, nil, fmt.Errorf("incorrect variable")
					}

					isValid, err := db.IsValidCryptoCurrency(parts[0])
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

					isValid, err = db.IsValidVariable(parts[1])
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
					variables = append(variables, db.CryptoVariable{Currency: parts[0], Variable: parts[1]})
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

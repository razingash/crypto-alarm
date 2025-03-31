package triggers

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
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

func Analys(expression string) int {
	tokens, errCode := tokenize(expression)

	if errCode != 0 {
		return errCode
	}

	_, errCode = parse(tokens)

	return errCode
}

// проверка на синтаксис
func tokenize(expression string) ([]Token, int) {
	var tokens []Token

	expression = strings.ReplaceAll(expression, " ", "") // убрать пробелы(хотя их не должно быть,
	// и возможно лучше сразу ошибку скинуть)

	tokenPatterns := []struct {
		Type    string
		Pattern string
	}{
		{NUMBER, `\d+(\.\d+)?`},              // Числа
		{OPERATOR, `[+\-*/^]`},               // Операторы
		{VARIABLE, `[A-Za-z]+[A-Za-z0-9_]*`}, // Переменные
		{FUNCTION, `sqrt\(|abs\(`},           // Функции
		{LPAREN, `\(`},                       // Левые скобки
		{RPAREN, `\)`},                       // Правые скобки
		{COMPARISON, `[<>]=?|>=?`},           // Операторы сравнения
	}

	for _, pattern := range tokenPatterns {
		re := regexp.MustCompile("^" + pattern.Pattern)
		match := re.FindString(expression)
		if match != "" {
			tokens = append(tokens, Token{Type: pattern.Type, Value: match})
			expression = strings.TrimPrefix(expression, match)
		}
	}

	if len(expression) > 0 {
		return nil, 1 // unknown symbol
	}

	return tokens, 0
}

// проверка на правильность
func parse(tokens []Token) (float64, int) {
	// Пример простого обхода
	stack := []float64{}

	for _, token := range tokens {
		switch token.Type {
		case NUMBER:
			num, err := strconv.ParseFloat(token.Value, 64)
			if err != nil {
				fmt.Printf("не удалось преобразовать число: %s", token.Value)
				return 0, 2
			}
			stack = append(stack, num)
		case OPERATOR:
			// добавить обработку приоритетов операторов
		case FUNCTION: // sqrt | abs
			// обработка функций
		case COMPARISON:
			// проверка на сравнение
		case LPAREN, RPAREN:
			// обработка скобок
		default:
			fmt.Printf("неизвестный токен: %s", token.Value)
			return 0, 3
		}
	}

	return stack[0], 0
}

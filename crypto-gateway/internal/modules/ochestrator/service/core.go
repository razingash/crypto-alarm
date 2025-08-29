package service

import (
	"fmt"
	"strings"

	"github.com/Knetic/govaluate"
)

// возвращает результат булевого выражения
func analyzeSygnal(formula string) (error, bool) {
	// формулы вида (true AND false) AND NOT true
	exprStr := strings.ReplaceAll(formula, "AND", "&&")
	exprStr = strings.ReplaceAll(exprStr, "OR", "||")
	exprStr = strings.ReplaceAll(exprStr, "NOT", "!")
	exprStr = strings.ReplaceAll(exprStr, "! ", "!")
	exprStr = strings.TrimSpace(exprStr)

	expr, err := govaluate.NewEvaluableExpression(exprStr)
	if err != nil {
		return err, false
	}

	result, err := expr.Evaluate(nil)
	if err != nil {
		return err, false
	}

	boolRes, ok := result.(bool)
	if !ok {
		return fmt.Errorf("result is not boolean"), false
	}

	return nil, boolRes
}

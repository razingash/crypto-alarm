package analytics

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"go/ast"
	"go/parser"
	"log"
	"slices"
	"sort"
	"strconv"
	"strings"

	"github.com/Knetic/govaluate"
)

type DependencyGraph struct {
	Graph           map[string][]int                          // переменная -> формулы
	Strategies      map[int][]int                             // ID стратегии -> список формул
	Formulas        map[int]string                            // ID -> формула
	Compiled        map[int]*govaluate.EvaluableExpression    // ID -> функция
	Variables       map[string]float64                        // переменная -> значение
	Cache           map[string]float64                        // кэш для промежуточных результатов - числовой
	TriggerCache    map[string]bool                           // кэш для промежуточных результатов - булевый
	SubexprCompiled map[string]*govaluate.EvaluableExpression // подвыражение -> компилированные подвыражения (сомнительная херня, можно сделать её списком ведь разницы нет?)
	SubexprWeights  map[string]int                            // подвыражение -> кол-во повторов
}

func NewDependencyGraph() *DependencyGraph {
	return &DependencyGraph{
		Graph:           make(map[string][]int),
		Strategies:      make(map[int][]int),
		Formulas:        make(map[int]string),
		Compiled:        make(map[int]*govaluate.EvaluableExpression),
		Variables:       make(map[string]float64),
		Cache:           make(map[string]float64),
		TriggerCache:    make(map[string]bool),
		SubexprCompiled: make(map[string]*govaluate.EvaluableExpression),
		SubexprWeights:  make(map[string]int),
	}
}

// Добавляет стратегию с её формулами в граф зависимостей
func (dg *DependencyGraph) AddStrategy(strategyID int, formulas map[int]string) error {
	if strategyID <= 0 {
		return fmt.Errorf("strategyID should be greater than 0")
	}
	if _, exists := dg.Strategies[strategyID]; exists {
		return fmt.Errorf("strategy with id %d already exists", strategyID)
	}

	formulaIDs := make([]int, 0, len(formulas))

	for formulaID, formula := range formulas {
		if err := dg.AddFormula(formula, formulaID); err != nil {
			return fmt.Errorf("failed to add formula %d for strategy %d: %w", formulaID, strategyID, err)
		}
		formulaIDs = append(formulaIDs, formulaID)
	}

	dg.Strategies[strategyID] = formulaIDs
	fmt.Printf("Added strategy %d with formulas: %v\n", strategyID, formulaIDs)

	return nil
}

// Добавляет формулу в граф зависимостей
func (dg *DependencyGraph) AddFormula(formula string, formulaID int) error {
	if formulaID <= 0 {
		return fmt.Errorf("formulaID should be greater than 0")
	}
	if _, exists := dg.Formulas[formulaID]; exists {
		return fmt.Errorf("formula with id %d already exists", formulaID)
	}
	if !isFormulaContainsComparisonOperator(formula) {
		return fmt.Errorf("formula must contain a comparison operator (>, <, ==, >=, <=)")
	}

	expr, err := govaluate.NewEvaluableExpression(formula)
	if err != nil {
		return fmt.Errorf("parsing formula failed: %w", err)
	}

	// AST pasring for subexpresions
	node, err := parser.ParseExpr(formula)
	if err != nil {
		return fmt.Errorf("AST parsing failed: %w", err)
	}

	subexprs := make(map[string]struct{})
	ast.Inspect(node, func(n ast.Node) bool {
		switch n := n.(type) {
		case *ast.BinaryExpr, *ast.ParenExpr:
			start := n.Pos() - 1
			end := n.End() - 1
			if int(start) >= 0 && int(end) <= len(formula) {
				s := strings.TrimSpace(formula[start:end])
				subexprs[s] = struct{}{}
			}
		}
		return true
	})

	dg.Formulas[formulaID] = formula
	dg.Compiled[formulaID] = expr

	// dependencies indexation by alternating
	for _, v := range expr.Vars() {
		if !slices.Contains(dg.Graph[v], formulaID) {
			dg.Graph[v] = append(dg.Graph[v], formulaID)
		}
	}

	// subexpressions compilation
	for sub := range subexprs {
		if _, ok := dg.SubexprCompiled[sub]; !ok {
			subExprCompiled, err := govaluate.NewEvaluableExpression(sub)
			if err != nil {
				fmt.Printf("warning: failed to compile subexpression '%s': %v\n", sub, err)
				continue
			}
			dg.SubexprCompiled[sub] = subExprCompiled
			dg.SubexprWeights[sub] = 1
		} else {
			dg.SubexprWeights[sub]++
		}
	}

	fmt.Printf("Added formula %d: %s\n", formulaID, formula)
	return nil
}

func isFormulaContainsComparisonOperator(s string) bool {
	ops := []string{"==", "!=", ">=", "<=", ">", "<"}
	for _, op := range ops {
		if strings.Contains(s, op) {
			return true
		}
	}
	return false
}

// удаляет стратегию вместе с связанными формулами и подвыражениями
func (dg *DependencyGraph) RemoveStrategy(strategyID int) error {
	formulaIDs, ok := dg.Strategies[strategyID]
	if !ok {
		return fmt.Errorf("strategy with id '%d' doesn't found", strategyID)
	}

	for _, formulaID := range formulaIDs {
		usedElsewhere := false // на всякий случай проверка на использование формулы в других стратегиях
		for sID, formulas := range dg.Strategies {
			if sID == strategyID {
				continue
			}
			for _, fID := range formulas {
				if fID == formulaID {
					usedElsewhere = true
					break
				}
			}
			if usedElsewhere {
				break
			}
		}

		if !usedElsewhere {
			err := dg.RemoveFormula(formulaID)
			if err != nil {
				return fmt.Errorf("failed to remove formula %d for strategy %d: %w", formulaID, strategyID, err)
			}
		}
	}

	delete(dg.Strategies, strategyID)

	fmt.Printf("Strategy with id %v deleted\n", strategyID)
	return nil
}

// Удаляет формулу из графа и все связанные подвыражения
func (dg *DependencyGraph) RemoveFormula(formulaID int) error {
	formula, ok := dg.Formulas[formulaID]
	if !ok {
		return fmt.Errorf("formula with id '%d' doesn't found", formulaID)
	}

	// list of variables that were used in the formula
	expr, err := govaluate.NewEvaluableExpression(formula)
	if err == nil {
		for _, v := range expr.Vars() {
			formulaIDs := dg.Graph[v]
			newList := make([]int, 0, len(formulaIDs))
			for _, id := range formulaIDs {
				if id != formulaID {
					newList = append(newList, id)
				}
			}
			if len(newList) > 0 {
				dg.Graph[v] = newList
			} else {
				delete(dg.Graph, v)
			}
		}
	}

	// subexpressions processing
	node, err := parser.ParseExpr(formula)
	if err == nil {
		subexprs := make(map[string]struct{})
		ast.Inspect(node, func(n ast.Node) bool {
			switch n := n.(type) {
			case *ast.BinaryExpr, *ast.ParenExpr:
				start := n.Pos() - 1
				end := n.End() - 1
				if int(start) >= 0 && int(end) <= len(formula) {
					s := strings.TrimSpace(formula[start:end])
					subexprs[s] = struct{}{}
				}
			}
			return true
		})

		for sub := range subexprs {
			if count, ok := dg.SubexprWeights[sub]; ok {
				if count <= 1 {
					delete(dg.SubexprWeights, sub)
					delete(dg.SubexprCompiled, sub)
				} else {
					dg.SubexprWeights[sub]--
				}
			}
		}
	}

	fmt.Printf("Formula with id %v deleted\n", formulaID)
	dg.RemoveVariablesIfNeeded(formulaID)
	delete(dg.Formulas, formulaID)
	delete(dg.Compiled, formulaID)

	return nil
}

// удаляет переменную и все подвязанные формулы и подвыражения используется в случае если на Binance
// уберут какой-нибудь параметр или валюту возвращает список ID удаленных формул
func (dg *DependencyGraph) RemoveVariablesIfNeeded(formulaID int) []string {
	removed := []string{}

	formula, ok := dg.Formulas[formulaID]
	if !ok {
		return removed
	}

	// получение переменных
	expr, err := govaluate.NewEvaluableExpression(formula)
	if err != nil {
		return removed
	}

	for _, variable := range expr.Vars() {
		stillUsed := false
		for id, compiled := range dg.Compiled {
			if id == formulaID {
				continue // на случай если сразу была удалена
			}
			for _, v := range compiled.Vars() {
				if v == variable {
					stillUsed = true
					break
				}
			}
			if stillUsed {
				break
			}
		}

		if !stillUsed {
			delete(dg.Variables, variable)
			removed = append(removed, variable)
			fmt.Printf("Переменная '%s' больше не используется и удалена.\n", variable)
		}
	}

	return removed
}

// алгоритм Кана (не учитывает циклические зависимости)
// Обновляет сразу несколько переменных и пересчитывает только необходимые формулы
func (dg *DependencyGraph) UpdateVariablesTopologicalKahn(updates map[string]float64) []int {
	fmt.Println("Updating variables with data:", updates)
	for k, v := range updates {
		dg.Variables[k] = v
	}

	affected := make(map[int]struct{})
	queue := make([]int, 0)

	// формулы, напрямую зависящие от обновлённых переменных
	for varName := range updates {
		if deps, ok := dg.Graph[varName]; ok {
			for _, fid := range deps {
				if _, seen := affected[fid]; !seen {
					affected[fid] = struct{}{}
					queue = append(queue, fid)
				}
			}
		}
	}

	// рсширение заражённых формул - если одна формула зависит от другой
	for i := 0; i < len(queue); i++ {
		curr := queue[i]
		currExpr := dg.Compiled[curr]
		for _, sym := range currExpr.Vars() {
			if deps, ok := dg.Graph[sym]; ok {
				for _, next := range deps {
					if _, seen := affected[next]; !seen {
						affected[next] = struct{}{}
						queue = append(queue, next)
					}
				}
			}
		}
	}

	if len(affected) == 0 {
		return nil
	}

	inDegree := make(map[int]int, len(affected))
	dependents := make(map[int][]int, len(affected))

	for fid := range affected {
		inDegree[fid] = 0
	}

	for fid := range affected {
		expr := dg.Compiled[fid]
		for _, sym := range expr.Vars() {
			for depFid := range affected {
				if depFid == fid {
					continue
				}

				depExpr := dg.Compiled[depFid]
				for _, definedSym := range depExpr.Vars() {
					if definedSym == sym {
						// fid зависит от depFid
						inDegree[fid]++
						dependents[depFid] = append(dependents[depFid], fid)
					}
				}
			}
		}
	}

	for _, formulaID := range queue {
		if _, err := dg.EvaluateFormula(formulaID); err != nil {
			if strings.Contains(err.Error(), "No parameter") { // еще не все переменные доступны
				continue // формула использует данные из разных апи, и не все из них установлены на момент EvaluateFormula()
			} else {
				log.Printf("Ошибка вычисления формулы %d: %v\n", formulaID, err)
			}
		}
	}

	triggered := dg.GetTriggeredFormulas(queue)
	fmt.Println("Triggered formulas:", triggered)
	return triggered
}

// подставляет значение формулы с учетом кэша
func (dg *DependencyGraph) EvaluateFormula(formulaID int) (interface{}, error) {
	expr, ok := dg.Compiled[formulaID]
	if !ok {
		return nil, fmt.Errorf("formula %d not compiled", formulaID)
	}

	subexprValues := make(map[string]interface{})
	for subexpr := range dg.SubexprWeights {
		formulaIDs, ok := dg.Graph[subexpr]
		if !ok {
			continue
		}

		for _, fid := range formulaIDs {
			if fid == formulaID {
				val, err := dg.EvaluateSubexpression(subexpr)
				if err != nil {
					return nil, fmt.Errorf("error in subexpression %s: %v", subexpr, err)
				}
				subexprValues[subexpr] = val
				break
			}
		}
	}

	for k, v := range dg.Variables {
		subexprValues[k] = v
	}

	result, err := expr.Evaluate(subexprValues)
	if err != nil {
		return nil, fmt.Errorf("evaluation failed for formula %d: %v", formulaID, err)
	}

	// Если это обычная числовая формула — обновляет переменную
	switch v := result.(type) {
	case float64:
		return v, nil
	case bool:
		return v, nil
	default:
		return nil, fmt.Errorf("formula %d evaluated to unsupported type %T", formulaID, result)
	}
}

// Вычисляет значение подвыражения с учетом кэша
func (dg *DependencyGraph) EvaluateSubexpression(subexpr string) (float64, error) {
	// использует заранее скомпилированное выражение, если оно доступно
	expr, ok := dg.SubexprCompiled[subexpr]
	if !ok {
		var err error
		expr, err = govaluate.NewEvaluableExpression(subexpr)
		if err != nil {
			return 0, fmt.Errorf("invalid subexpression: %v", err)
		}
		dg.SubexprCompiled[subexpr] = expr // кэшируем компиляцию
	}

	varsValues := make(map[string]interface{})
	for _, v := range expr.Vars() {
		if val, ok := dg.Variables[v]; ok {
			varsValues[v] = val
		} else {
			varsValues[v] = 0 // по умолчанию !!! МОЖЕТ БЫТЬ БАГ
		}
	}

	cacheKey := getCanonicalCacheKey(varsValues, nil)
	if val, ok := dg.Cache[cacheKey]; ok {
		return val, nil
	}

	result, err := expr.Evaluate(varsValues)
	if err != nil {
		return 0, fmt.Errorf("error evaluating subexpression %q: %v", subexpr, err)
	}

	floatResult, ok := result.(float64)
	if !ok {
		return 0, fmt.Errorf("subexpression %q did not evaluate to float64", subexpr)
	}

	dg.Cache[cacheKey] = floatResult
	return floatResult, nil
}

// проверяет, какие из формул сработали после обновления переменных, и возвращает их id
func (dg *DependencyGraph) GetTriggeredFormulas(formulaIDs []int) []int {
	var triggered []int
	for _, fid := range formulaIDs {
		ok, err := dg.IsFormulaTriggered(fid)
		if err != nil {
			fmt.Printf("warning: failed to check trigger for %d: %v\n", fid, err)
			continue
		}
		if ok {
			triggered = append(triggered, fid)
		}
	}
	return triggered
}

// возвращает булево значение результата выражения
func (dg *DependencyGraph) IsFormulaTriggered(formulaID int) (bool, error) {
	expr, ok := dg.Compiled[formulaID]
	if !ok {
		return false, fmt.Errorf("formula %d not compiled", formulaID)
	}

	args := make(map[string]interface{}, len(expr.Vars()))
	for _, v := range expr.Vars() {
		val, ok := dg.Variables[v]
		if !ok {
			return false, fmt.Errorf("variable %q not found for formula %d", v, formulaID)
		}
		args[v] = val
	}

	key := getCanonicalCacheKey(args, &formulaID)

	if cached, found := dg.TriggerCache[key]; found {
		return cached, nil
	}

	raw, err := expr.Evaluate(args)
	if err != nil {
		return false, fmt.Errorf("error evaluating formula %d: %w", formulaID, err)
	}
	triggered, ok := raw.(bool)
	if !ok {
		return false, fmt.Errorf("formula %d did not evaluate to bool, got %T", formulaID, raw)
	}

	dg.TriggerCache[key] = triggered
	return triggered, nil
}

// возвращает подвязанные переменные для каждой формулы из списка ID
func (dg *DependencyGraph) GetFormulasVariables(formulaIDs []int) (map[int][]string, map[string]float64) {
	result := make(map[int][]string)
	variableValues := make(map[string]float64)

	for _, formulaID := range formulaIDs {
		compiledExpr, ok := dg.Compiled[formulaID]
		if !ok || compiledExpr == nil {
			continue
		}

		vars := compiledExpr.Vars()
		result[formulaID] = vars

		for _, v := range vars {
			if val, ok := dg.Variables[v]; ok {
				if _, exists := variableValues[v]; !exists {
					variableValues[v] = val
				}
			}
		}
	}

	return result, variableValues
}

// Создаёт уникальный кэш-ключ, не зависящий от порядка переменных
// РАБОТАЕТ ПРИМИТИВНО - не учитывает трудные случаи
func getCanonicalCacheKey(args map[string]interface{}, formulaID *int) string {
	keys := make([]string, 0, len(args))
	for k := range args {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// созданте строки по типу "formulaID|k1=v1;k2=v2;..." чтобы не дублировать идентичные выражения
	var b strings.Builder
	if formulaID != nil {
		b.WriteString(strconv.Itoa(*formulaID))
		b.WriteByte('|')
	}
	for _, k := range keys {
		b.WriteString(k)
		b.WriteByte('=')
		b.WriteString(fmt.Sprint(args[k]))
		b.WriteByte(';')
	}
	sum := md5.Sum([]byte(b.String()))
	return hex.EncodeToString(sum[:])
}

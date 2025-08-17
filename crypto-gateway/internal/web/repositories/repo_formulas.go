package repositories

import (
	"context"
	"crypto-gateway/internal/web/db"
	"database/sql"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

type Strategy struct {
	Id            string               `json:"id"`
	Name          string               `json:"name"`
	Description   string               `json:"description"`
	IsNotified    bool                 `json:"is_notified"`
	IsActive      bool                 `json:"is_active"`
	IsHistoryOn   bool                 `json:"is_history_on"`
	IsShuttedOff  bool                 `json:"is_shutted_off"`
	LastTriggered string               `json:"last_triggered"`
	Cooldown      int                  `json:"cooldown"`
	Conditions    []StrategyExpression `json:"conditions"`
}

type StrategyExpression struct {
	FormulaID  string `json:"formula_id"`
	Formula    string `json:"formula"`
	FormulaRaw string `json:"formula_raw"`
}

type UserFormula struct {
	Id            string `json:"id"`
	Formula       string `json:"formula"`
	FormulaRaw    string `json:"formula_raw"`
	Name          string `json:"name"`
	Description   string `json:"description"`
	IsNotified    bool   `json:"is_notified"`
	IsActive      bool   `json:"is_active"`
	IsHistoryOn   bool   `json:"is_history_on"`
	IsShuttedOff  bool   `json:"is_shutted_off"`
	LastTriggered string `json:"last_triggered"`
	Cooldown      int    `json:"cooldown"`
}

// both user-defined and ordinary
type CryptoVariable struct {
	Currency string
	Variable string
}

type tempRow struct {
	Timestamp time.Time
	VarName   string
	Value     string
}

func IsValidCryptoCurrency(name string) (bool, error) {
	var isAvailable bool

	err := db.DB.QueryRow(context.Background(), `
		SELECT EXISTS (
			SELECT 1 
			FROM crypto_currencies 
			WHERE currency = $1 AND is_available = true
		)
	`, name).Scan(&isAvailable)

	if err != nil {
		return false, err
	}

	return isAvailable, nil
}

func IsValidVariable(name string) (bool, error) {
	var isAvailable bool
	err := db.DB.QueryRow(context.Background(), `
		SELECT EXISTS (
			SELECT 1 
			FROM crypto_params 
			WHERE parameter = $1 AND is_active = true
		)
	`, name).Scan(&isAvailable)

	if err != nil {
		return false, err
	}

	return isAvailable, nil
}

func IsValidUserVariable(name string) (bool, error) {
	var isAvailable bool
	err := db.DB.QueryRow(context.Background(), `
		SELECT EXISTS (
			SELECT 1 
			FROM crypto_variables 
			WHERE symbol = $1
		)
	`, name).Scan(&isAvailable)

	if err != nil {
		return false, err
	}

	return isAvailable, nil
}

func GetApiAndCooldownByID(id int) (string, int) {
	var api string
	var cooldown int
	err := db.DB.QueryRow(context.Background(), `
        SELECT api, cooldown FROM crypto_api WHERE id = $1
    `, id).Scan(&api, &cooldown)
	if err != nil {
		fmt.Println(err)
	}
	return api, cooldown
}

func GetFormulaById(formulaID int) string {
	// удалить функцию, сейчас это заглушка, возможно лучше сделать чтобы бралась из графа
	var formula string
	err := db.DB.QueryRow(context.Background(), `
        SELECT formula FROM trigger_formula WHERE id = $1
    `, formulaID).Scan(&formula)
	if err != nil {
		fmt.Println(err)
	}
	return formula
}

// получает формулы, без переменных
func GetStrategyFullFormulasById(strategyID int) map[int]string {
	ctx := context.Background()

	variables := make(map[string]string)
	rowsVars, err := db.DB.Query(ctx, `
		SELECT cv.symbol, cv.formula
		FROM crypto_strategy_variable csv
		JOIN crypto_variables cv ON csv.crypto_variable_id = cv.id
		WHERE csv.strategy_id = $1
	`, strategyID)
	if err != nil {
		fmt.Printf("failed to query variables for strategy %v: %v\n", strategyID, err)
		return nil
	}
	defer rowsVars.Close()

	for rowsVars.Next() {
		var symbol, formula string
		if err := rowsVars.Scan(&symbol, &formula); err != nil {
			fmt.Printf("failed to scan variable: %v\n", err)
			continue
		}
		variables[symbol] = formula
	}
	if err := rowsVars.Err(); err != nil {
		fmt.Printf("error after scanning variables: %v\n", err)
		return nil
	}

	rows, err := db.DB.Query(ctx, `
		SELECT tf.id, tf.formula
		FROM crypto_strategy_formula csf
		JOIN trigger_formula tf ON csf.formula_id = tf.id
		WHERE csf.strategy_id = $1
	`, strategyID)
	if err != nil {
		fmt.Printf("failed to query formulas for strategy %v: %v\n", strategyID, err)
		return nil
	}
	defer rows.Close()

	result := make(map[int]string)
	for rows.Next() {
		var id int
		var formula string
		if err := rows.Scan(&id, &formula); err != nil {
			fmt.Printf("failed to scan formula row: %v\n", err)
			continue
		}

		finalFormula := formula
		for symbol, replacement := range variables {
			re := regexp.MustCompile(`\b` + regexp.QuoteMeta(symbol) + `\b`)
			finalFormula = re.ReplaceAllString(finalFormula, replacement)
		}

		result[id] = finalFormula
	}
	if err := rows.Err(); err != nil {
		fmt.Printf("error after scanning formulas: %v\n", err)
		return nil
	}

	return result
}

func GetStrategiesWithFormulas(limit int, page int, strategyID string) ([]Strategy, bool, error) {
	var strategies []Strategy
	var hasNext bool

	if strategyID != "" {
		rows, err := db.DB.Query(context.Background(), `
		SELECT 
			cs.id, cs.name, COALESCE(cs.description, ''), cs.is_notified, cs.is_active, cs.is_shutted_off, 
			cs.is_history_on, cs.cooldown, COALESCE(TO_CHAR(cs.last_triggered, 'YYYY-MM-DD HH24:MI:SS'), ''),
			tf.formula, tf.formula_raw, tf.id
		FROM crypto_strategy cs
		LEFT JOIN crypto_strategy_formula csf ON cs.id = csf.strategy_id
		LEFT JOIN trigger_formula tf ON tf.id = csf.formula_id
		WHERE cs.id = $1
	`, strategyID)
		if err != nil {
			return nil, false, err
		}
		defer rows.Close()

		strategies, _, err := scanStrategies(rows)
		if err != nil {
			return nil, false, err
		}

		if len(strategies) == 0 {
			return nil, false, fmt.Errorf("strategy with id %s not found", strategyID)
		}

		return strategies, false, nil
	}

	offset := (page - 1) * limit
	rows, err := db.DB.Query(context.Background(), `
		SELECT 
			cs.id, cs.name, COALESCE(cs.description, ''), cs.is_notified, cs.is_active, cs.is_shutted_off, 
			cs.is_history_on, cs.cooldown, COALESCE(TO_CHAR(cs.last_triggered, 'YYYY-MM-DD HH24:MI:SS'), ''),
			tf.formula, tf.formula_raw, tf.id
		FROM crypto_strategy cs
		LEFT JOIN crypto_strategy_formula csf ON cs.id = csf.strategy_id
		LEFT JOIN trigger_formula tf ON tf.id = csf.formula_id
		ORDER BY cs.id DESC
		LIMIT $1 OFFSET $2
	`, limit+1, offset)
	if err != nil {
		return nil, false, err
	}
	defer rows.Close()

	strategies, _, err = scanStrategies(rows)
	if err != nil {
		return nil, false, err
	}

	if len(strategies) > limit {
		hasNext = true
		strategies = strategies[:limit]
	}

	return strategies, hasNext, nil
}

func scanStrategies(rows pgx.Rows) ([]Strategy, bool, error) {
	strategiesMap := make(map[string]*Strategy)

	for rows.Next() {
		var s Strategy
		var formulaID, formula, formulaRaw sql.NullString

		err := rows.Scan(
			&s.Id, &s.Name, &s.Description, &s.IsNotified, &s.IsActive, &s.IsShuttedOff, &s.IsHistoryOn, &s.Cooldown,
			&s.LastTriggered, &formula, &formulaRaw, &formulaID,
		)
		if err != nil {
			return nil, false, err
		}

		existing, ok := strategiesMap[s.Id]
		if !ok {
			existing = &Strategy{
				Id:            s.Id,
				Name:          s.Name,
				Description:   s.Description,
				IsNotified:    s.IsNotified,
				IsActive:      s.IsActive,
				IsShuttedOff:  s.IsShuttedOff,
				IsHistoryOn:   s.IsHistoryOn,
				Cooldown:      s.Cooldown,
				LastTriggered: s.LastTriggered,
				Conditions:    []StrategyExpression{},
			}
			strategiesMap[s.Id] = existing
		}

		if formulaID.Valid {
			existing.Conditions = append(existing.Conditions, StrategyExpression{
				FormulaID:  formulaID.String,
				Formula:    formula.String,
				FormulaRaw: formulaRaw.String,
			})
		}
	}

	var strategies []Strategy
	for _, s := range strategiesMap {
		strategies = append(strategies, *s)
	}

	return strategies, false, nil
}

func GetStrategyHistory(strategyID int, limit int, page int, prevCursor int) (bool, []tempRow, error) {
	var rows pgx.Rows
	var err error

	if prevCursor == 0 {
		rows, err = db.DB.Query(context.Background(), `
			SELECT sh.timestamp,
				   STRING_AGG(tc.name, ', ') AS names,
				   STRING_AGG(CAST(tch.value AS TEXT), ', ') AS values
			FROM strategy_history sh
			LEFT JOIN trigger_component_history tch ON sh.id = tch.expression_id
			LEFT JOIN trigger_component tc ON tc.id = tch.component_id
			WHERE sh.formula_id IN (
				SELECT formula_id FROM crypto_strategy_formula WHERE strategy_id = $1
			)
			GROUP BY sh.timestamp
			ORDER BY sh.timestamp DESC
			LIMIT $2;
		`, strategyID, limit+1)
	} else {
		cursorTime := time.Unix(int64(prevCursor), 0).UTC()
		rows, err = db.DB.Query(context.Background(), `
			SELECT sh.timestamp,
				   STRING_AGG(tc.name, ', ') AS names,
				   STRING_AGG(CAST(tch.value AS TEXT), ', ') AS values
			FROM strategy_history sh
			LEFT JOIN trigger_component_history tch ON sh.id = tch.expression_id
			LEFT JOIN trigger_component tc ON tc.id = tch.component_id
			WHERE sh.formula_id IN (
				SELECT formula_id FROM crypto_strategy_formula WHERE strategy_id = $1
			)
			AND sh.timestamp < $2
			GROUP BY sh.timestamp
			ORDER BY sh.timestamp DESC
			LIMIT $3;
		`, strategyID, cursorTime, limit+1)
	}

	if err != nil {
		return false, nil, fmt.Errorf("failed to query strategy history: %w", err)
	}
	defer rows.Close()

	var rawRows []tempRow
	var hasNext bool

	for rows.Next() {
		var r tempRow
		var names, values string

		if err := rows.Scan(&r.Timestamp, &names, &values); err != nil {
			return false, nil, fmt.Errorf("failed to scan row: %w", err)
		}
		r.VarName = names
		r.Value = values
		rawRows = append(rawRows, r)
	}
	if err := rows.Err(); err != nil {
		return false, nil, fmt.Errorf("error iterating over rows: %w", err)
	}

	if len(rawRows) > limit {
		hasNext = true
		rawRows = rawRows[:limit]
	}

	return hasNext, rawRows, nil
}

func SaveStrategy(
	name, description string, expressions []StrategyExpression, variables []CryptoVariable, userVariables []CryptoVariable,
) (int, error) {
	tx, err := db.DB.Begin(context.Background())
	if err != nil {
		return 0, err
	}
	defer func() {
		if err != nil {
			tx.Rollback(context.Background())
		} else {
			tx.Commit(context.Background())
		}
	}()

	var strategyId int
	err = tx.QueryRow(context.Background(), `
        INSERT INTO crypto_strategy (name, description, is_notified, is_active, is_shutted_off, is_history_on, cooldown)
        VALUES ($1, $2, false, false, false, false, 3600)
        RETURNING id
    `, name, description).Scan(&strategyId)
	if err != nil {
		return 0, err
	}

	for _, expr := range expressions {
		var formulaId int
		err = tx.QueryRow(context.Background(), `
            INSERT INTO trigger_formula (formula, formula_raw)
            VALUES ($1, $2)
            RETURNING id
        `, expr.Formula, expr.FormulaRaw).Scan(&formulaId)
		if err != nil {
			return 0, err
		}

		_, err = tx.Exec(context.Background(), `
            INSERT INTO crypto_strategy_formula (strategy_id, formula_id)
            VALUES ($1, $2)
        `, strategyId, formulaId)
		if err != nil {
			return 0, err
		}

		for _, v := range variables {
			var triggerComponentID int
			err = tx.QueryRow(context.Background(), `
				SELECT tc.id
				FROM trigger_component tc
				JOIN crypto_currencies cc ON tc.currency_id = cc.id
				JOIN crypto_params cp ON tc.parameter_id = cp.id
				WHERE cc.currency = $1 AND cp.parameter = $2
				LIMIT 1
			`, v.Currency, v.Variable).Scan(&triggerComponentID)
			if err != nil {
				return 0, err
			}

			_, err = tx.Exec(context.Background(), `
				INSERT INTO trigger_formula_component (component_id, formula_id)
				VALUES ($1, $2)
				ON CONFLICT DO NOTHING
			`, triggerComponentID, formulaId)
			if err != nil {
				return 0, err
			}
		}
	}

	// сохранение переменных... позже попробовать упростить userVariables
	if len(userVariables) > 0 {
		variableSet := make(map[string]struct{})
		currencySet := make(map[string]struct{})
		for _, v := range userVariables {
			variableSet[v.Variable] = struct{}{}
			currencySet[v.Currency] = struct{}{}
		}

		variableNames := make([]interface{}, 0, len(variableSet))
		variablePlaceholders := make([]string, 0, len(variableSet))
		i := 1
		for name := range variableSet {
			variablePlaceholders = append(variablePlaceholders, fmt.Sprintf("$%d", i))
			variableNames = append(variableNames, name)
			i++
		}

		variableQuery := fmt.Sprintf(`
			SELECT id, symbol FROM crypto_variables
			WHERE symbol IN (%s)
		`, strings.Join(variablePlaceholders, ", "))
		rows, err := tx.Query(context.Background(), variableQuery, variableNames...)
		if err != nil {
			return 0, err
		}
		defer rows.Close()

		symbolToVarId := make(map[string]int)
		for rows.Next() {
			var id int
			var symbol string
			if err := rows.Scan(&id, &symbol); err != nil {
				return 0, err
			}
			symbolToVarId[symbol] = id
		}
		if err := rows.Err(); err != nil {
			return 0, err
		}

		currencyNames := make([]interface{}, 0, len(currencySet))
		currencyPlaceholders := make([]string, 0, len(currencySet))
		j := 1
		for c := range currencySet {
			currencyPlaceholders = append(currencyPlaceholders, fmt.Sprintf("$%d", j))
			currencyNames = append(currencyNames, c)
			j++
		}
		currencyQuery := fmt.Sprintf(`
			SELECT id, currency FROM crypto_currencies
			WHERE currency IN (%s)
		`, strings.Join(currencyPlaceholders, ", "))
		currencyRows, err := tx.Query(context.Background(), currencyQuery, currencyNames...)
		if err != nil {
			return 0, err
		}
		defer currencyRows.Close()

		currencyToId := make(map[string]int)
		for currencyRows.Next() {
			var id int
			var name string
			if err := currencyRows.Scan(&id, &name); err != nil {
				return 0, err
			}
			currencyToId[name] = id
		}
		if err := currencyRows.Err(); err != nil {
			return 0, err
		}

		var variableInserts []string
		var insertArgs []interface{}
		argIndex := 1

		for _, variable := range userVariables {
			varId, ok1 := symbolToVarId[variable.Variable]
			currencyId, ok2 := currencyToId[variable.Currency]
			if !ok1 || !ok2 {
				continue
			}

			variableInserts = append(variableInserts, fmt.Sprintf("($%d, $%d, $%d)", argIndex, argIndex+1, argIndex+2))
			insertArgs = append(insertArgs, strategyId, varId, currencyId)
			argIndex += 3
		}

		if len(variableInserts) > 0 {
			insertQuery := fmt.Sprintf(`
				INSERT INTO crypto_strategy_variable (strategy_id, crypto_variable_id, crypto_currency_id)
				VALUES %s
				ON CONFLICT DO NOTHING
			`, strings.Join(variableInserts, ", "))
			_, err := tx.Exec(context.Background(), insertQuery, insertArgs...)
			if err != nil {
				return 0, err
			}
		}
	}

	return strategyId, nil
}

func UpdateStrategyConditions(strategyID int, conditions []map[string]interface{}) (err error) {
	tx, err := db.DB.Begin(context.Background())
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback(context.Background())
		} else {
			tx.Commit(context.Background())
		}
	}()

	for _, cond := range conditions {
		formulaID, ok := cond["formula_id"].(string)
		if !ok {
			return fmt.Errorf("invalid formula_id")
		}
		formula, _ := cond["formula"].(string)
		formulaRaw, _ := cond["formula_raw"].(string)

		cmdTag, errExec := tx.Exec(context.Background(), `
			UPDATE trigger_formula
			SET formula = $1, formula_raw = $2
			WHERE id = $3 AND EXISTS (
				SELECT 1 FROM crypto_strategy_formula
				WHERE formula_id = $3 AND strategy_id = $4
			)
		`, formula, formulaRaw, formulaID, strategyID)
		if errExec != nil {
			return errExec
		}
		if cmdTag.RowsAffected() == 0 {
			return fmt.Errorf("formula_id %v does not belong to strategy_id %v", formulaID, strategyID)
		}
	}

	return nil
}

func UpdateStrategy(strategyID int, data map[string]interface{}) error {
	var setClauses []string
	var args []interface{}
	argIndex := 1

	allowedFields := map[string]bool{
		"name":          true,
		"description":   true,
		"is_notified":   true,
		"is_active":     true,
		"is_history_on": true,
		"cooldown":      true,
	}

	for field, value := range data {
		if !allowedFields[field] {
			continue
		}
		setClauses = append(setClauses, fmt.Sprintf("%s = $%d", field, argIndex))
		args = append(args, value)
		argIndex++
	}

	if len(setClauses) == 0 {
		return nil
	}

	query := fmt.Sprintf("UPDATE crypto_strategy SET %s WHERE id = $%d", strings.Join(setClauses, ", "), argIndex)
	args = append(args, strategyID)

	_, err := db.DB.Exec(context.Background(), query, args...)
	if err != nil {
		return fmt.Errorf("database error")
	}

	return nil
}

func DeleteStrategyById(strategyID int) error {
	tx, err := db.DB.Begin(context.Background())
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback(context.Background())

	_, err = tx.Exec(context.Background(), `DELETE FROM crypto_strategy WHERE id = $1`, strategyID)
	if err != nil {
		return fmt.Errorf("failed to delete strategy: %w", err)
	}

	_, err = tx.Exec(context.Background(), `
		DELETE FROM trigger_formula tf
		WHERE NOT EXISTS (
			SELECT 1 FROM crypto_strategy_formula csf
			WHERE csf.formula_id = tf.id
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to delete orphan formulas: %w", err)
	}

	err = tx.Commit(context.Background())
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func DeleteFormulaById(formulaID int) error {
	_, err := db.DB.Exec(context.Background(), `DELETE FROM trigger_formula WHERE id = $1`, formulaID)

	if err != nil {
		fmt.Println(err)
		return fmt.Errorf("database error")
	}
	return nil
}

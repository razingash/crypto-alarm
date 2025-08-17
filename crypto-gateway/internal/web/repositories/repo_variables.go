package repositories

import (
	"context"
	"crypto-gateway/internal/web/db"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
)

type UpdateVariableStruct struct {
	Symbol      *string `json:"symbol"`
	Name        *string `json:"name"`
	Description *string `json:"description"`
	Formula     *string `json:"formula"`
	FormulaRaw  *string `json:"formula_raw"`
}

type Variable struct {
	ID          int64  `json:"id"`
	Symbol      string `json:"symbol"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Formula     string `json:"formula"`
	FormulaRaw  string `json:"formula_raw"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

type VariableKeyboard struct {
	ID         int64  `json:"id"`
	Symbol     string `json:"symbol"`
	Formula    string `json:"formula"`
	FormulaRaw string `json:"formula_raw"`
}

type Token struct {
	Type  string
	Value string
}

func GetVariables(limit int, page int, variableID string) ([]Variable, bool, error) {
	var variables []Variable
	var hasNext bool

	if variableID != "" {
		rows, err := db.DB.Query(context.Background(), `
		SELECT 
			id, symbol, name, COALESCE(description, ''), formula, formula_raw,
			TO_CHAR(created_at, 'YYYY-MM-DD HH24:MI:SS'),
			TO_CHAR(updated_at, 'YYYY-MM-DD HH24:MI:SS')
		FROM crypto_variables
		WHERE id = $1
	`, variableID)
		if err != nil {
			return nil, false, err
		}
		defer rows.Close()

		variables, _, err := scanVariables(rows)
		if err != nil {
			return nil, false, err
		}

		if len(variables) == 0 {
			return nil, false, fmt.Errorf("variable with id %s not found", variableID)
		}

		return variables, false, nil
	}

	offset := (page - 1) * limit

	rows, err := db.DB.Query(context.Background(), `
		SELECT 
			id, symbol, name, COALESCE(description, ''), formula, formula_raw,
			TO_CHAR(created_at, 'YYYY-MM-DD HH24:MI:SS'),
			TO_CHAR(updated_at, 'YYYY-MM-DD HH24:MI:SS')
		FROM crypto_variables
		ORDER BY id DESC
		LIMIT $1 OFFSET $2
	`, limit+1, offset)
	if err != nil {
		return nil, false, err
	}
	defer rows.Close()

	variables, _, err = scanVariables(rows)
	if err != nil {
		return nil, false, err
	}

	if len(variables) > limit {
		hasNext = true
		variables = variables[:limit]
	}

	return variables, hasNext, nil
}

func GetVariablesForKeyboard() ([]VariableKeyboard, error) {
	var variables []VariableKeyboard
	rows, err := db.DB.Query(context.Background(), `
		SELECT id, symbol, formula, formula_raw
		FROM crypto_variables
		ORDER BY id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var v VariableKeyboard
		err := rows.Scan(
			&v.ID, &v.Symbol, &v.Formula, &v.FormulaRaw,
		)
		if err != nil {
			return nil, err
		}
		variables = append(variables, v)
	}

	return variables, nil
}

// не учитывает циклические переменные, но это легко будет добавить используя tokens
func CreateVariable(symbol, name, description, formula, formulaRaw string, tokens []Token) (int64, error) {
	ctx := context.Background()
	tx, err := db.DB.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		} else {
			_ = tx.Commit(ctx)
		}
	}()

	var variableID int64
	err = tx.QueryRow(ctx, `
        INSERT INTO crypto_variables (symbol, name, description, formula, formula_raw)
        VALUES ($1, $2, $3, $4, $5)
        RETURNING id
    `, symbol, name, description, formula, formulaRaw).Scan(&variableID)
	if err != nil {
		return 0, err
	}

	varSet := make(map[string]struct{})
	for _, token := range tokens {
		if token.Type == "USER_VARIABLE" {
			varSet[token.Value] = struct{}{}
		}
	}
	if len(varSet) == 0 {
		return variableID, nil
	}

	paramNames := make([]string, 0, len(varSet))
	args := make([]interface{}, 0, len(varSet))
	i := 1
	for name := range varSet {
		paramNames = append(paramNames, fmt.Sprintf("$%d", i))
		args = append(args, name)
		i++
	}

	query := fmt.Sprintf(`
		SELECT cp.id, cp.parameter, cp.crypto_api_id
		FROM crypto_params cp
		WHERE cp.parameter IN (%s)
	`, strings.Join(paramNames, ", "))
	rows, err := tx.Query(ctx, query, args...)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	type paramInfo struct {
		paramID int
		apiID   int
	}
	unique := make(map[string]paramInfo)

	for rows.Next() {
		var p paramInfo
		var paramName string
		if err := rows.Scan(&p.paramID, &paramName, &p.apiID); err != nil {
			return 0, err
		}
		key := fmt.Sprintf("%d_%d", p.paramID, p.apiID)
		unique[key] = p
	}
	if err = rows.Err(); err != nil {
		return 0, err
	}

	if len(unique) == 0 {
		return variableID, nil
	}

	valueStrings := make([]string, 0, len(unique))
	valueArgs := make([]interface{}, 0, len(unique)*3)
	idx := 1
	for _, p := range unique {
		valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d, $%d)", idx, idx+1, idx+2))
		valueArgs = append(valueArgs, p.apiID, variableID, p.paramID)
		idx += 3
	}

	insertQuery := fmt.Sprintf(`
		INSERT INTO crypto_variables_api (api_id, variable_id, parameter_id)
		VALUES %s
		ON CONFLICT DO NOTHING
	`, strings.Join(valueStrings, ", "))

	_, err = tx.Exec(ctx, insertQuery, valueArgs...)
	if err != nil {
		return 0, err
	}

	return variableID, nil
}

func UpdateVariable(id int, input *UpdateVariableStruct, tokens []Token) (bool, error) {
	ctx := context.Background()
	tx, err := db.DB.Begin(ctx)
	if err != nil {
		return false, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		} else {
			_ = tx.Commit(ctx)
		}
	}()

	setParts := []string{}
	args := []interface{}{}
	pos := 1

	if input.Symbol != nil {
		setParts = append(setParts, fmt.Sprintf("symbol = $%d", pos))
		args = append(args, *input.Symbol)
		pos++
	}
	if input.Name != nil {
		setParts = append(setParts, fmt.Sprintf("name = $%d", pos))
		args = append(args, *input.Name)
		pos++
	}
	if input.Description != nil {
		setParts = append(setParts, fmt.Sprintf("description = $%d", pos))
		args = append(args, *input.Description)
		pos++
	}
	if input.Formula != nil && input.FormulaRaw != nil {
		setParts = append(setParts,
			fmt.Sprintf("formula = $%d", pos),
			fmt.Sprintf("formula_raw = $%d", pos+1),
		)
		args = append(args, *input.Formula, *input.FormulaRaw)
		pos += 2
	}

	if len(setParts) > 0 {
		setParts = append(setParts, "updated_at = now()")
		query := fmt.Sprintf("UPDATE crypto_variables SET %s WHERE id = $%d",
			strings.Join(setParts, ", "), pos,
		)
		args = append(args, id)

		if _, err = tx.Exec(ctx, query, args...); err != nil {
			return false, err
		}
	}

	varSet := map[string]struct{}{}
	for _, t := range tokens {
		if t.Type == "USER_VARIABLE" {
			varSet[t.Value] = struct{}{}
		}
	}

	if len(varSet) > 0 {
		if _, err = tx.Exec(ctx,
			`DELETE FROM crypto_variables_api WHERE variable_id = $1`, id,
		); err != nil {
			return false, err
		}

		names := make([]string, 0, len(varSet))
		for name := range varSet {
			names = append(names, name)
		}

		_, err = tx.Exec(ctx, `
            INSERT INTO crypto_variables_api (api_id, variable_id, parameter_id)
            SELECT DISTINCT
                cp.crypto_api_id,
                $2::bigint        AS variable_id,
                cp.id              AS parameter_id
            FROM crypto_params cp
            WHERE cp.parameter = ANY($1::text[])
            ON CONFLICT DO NOTHING
        `, names, id)
		if err != nil {
			return false, err
		}
	}

	return true, nil
}

func DeleteVariableById(variableID int) error {
	var count int
	err := db.DB.QueryRow(context.Background(), `
		SELECT COUNT(*) 
		FROM crypto_strategy_variable 
		WHERE crypto_variable_id = $1
	`, variableID).Scan(&count)

	if err != nil {
		fmt.Println(err)
		return fmt.Errorf("database error")
	}

	if count > 0 {
		return fmt.Errorf("cannot delete variable: used in %d strategies", count)
	}

	_, err = db.DB.Exec(context.Background(), `
		DELETE FROM crypto_variables 
		WHERE id = $1
	`, variableID)

	if err != nil {
		return fmt.Errorf("database error")
	}

	return nil
}

func scanVariables(rows pgx.Rows) ([]Variable, bool, error) {
	var result []Variable
	for rows.Next() {
		var v Variable
		err := rows.Scan(
			&v.ID, &v.Symbol, &v.Name, &v.Description, &v.Formula, &v.FormulaRaw,
			&v.CreatedAt, &v.UpdatedAt,
		)
		if err != nil {
			return nil, false, err
		}
		result = append(result, v)
	}
	return result, true, nil
}

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

func CreateVariable(symbol, name, description, formula, formulaRaw string) (int64, error) {
	var id int64
	query := `
        INSERT INTO crypto_variables (symbol, name, description, formula, formula_raw)
        VALUES ($1, $2, $3, $4, $5)
        RETURNING id;
    `
	err := db.DB.QueryRow(context.Background(), query, symbol, name, description, formula, formulaRaw).Scan(&id)
	return id, err
}

func UpdateVariable(id int, input *UpdateVariableStruct) error {
	setParts := []string{}
	args := []any{}
	argPos := 1

	if input.Symbol != nil {
		setParts = append(setParts, fmt.Sprintf("symbol = $%d", argPos))
		args = append(args, *input.Symbol)
		argPos++
	}
	if input.Name != nil {
		setParts = append(setParts, fmt.Sprintf("name = $%d", argPos))
		args = append(args, *input.Name)
		argPos++
	}
	if input.Description != nil {
		setParts = append(setParts, fmt.Sprintf("description = $%d", argPos))
		args = append(args, *input.Description)
		argPos++
	}
	if input.Formula != nil && input.FormulaRaw != nil {
		setParts = append(setParts, fmt.Sprintf("formula = $%d, formula_raw = $%d", argPos, argPos+1))
		args = append(args, *input.Formula, *input.FormulaRaw)
		argPos += 2
	}

	if len(setParts) == 0 {
		return nil
	}

	setParts = append(setParts, "updated_at = now()")

	query := fmt.Sprintf(`UPDATE crypto_variables SET %s WHERE id = $%d`, strings.Join(setParts, ", "), argPos)
	args = append(args, id)

	_, err := db.DB.Exec(context.Background(), query, args...)
	return err
}

func DeleteVariableById(variableID int) error {
	_, err := db.DB.Exec(context.Background(), `DELETE FROM crypto_variables WHERE id = $1`, variableID)

	if err != nil {
		fmt.Println(err)
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

package db

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

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
}

type CryptoVariable struct {
	Currency string
	Variable string
	//
}

type tempRow struct {
	Timestamp time.Time
	VarName   string
	Value     string
}

func IsValidCryptoCurrency(name string) (bool, error) {
	var isAvailable bool

	err := DB.QueryRow(context.Background(), `
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

	err := DB.QueryRow(context.Background(), `
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

func GetUserFormulas(uuid string, limit int, page int, formulaID string) ([]UserFormula, bool, error) {
	ownerID, err := GetIdbyUuid(uuid)
	if err != nil {
		return nil, false, err
	}

	var formulas []UserFormula
	var hasNext bool

	if formulaID != "" {
		row := DB.QueryRow(context.Background(), `
            SELECT id, formula_raw, COALESCE(name, ''), COALESCE(description, ''), is_notified, is_active,
                is_shutted_off, is_history_on, COALESCE(TO_CHAR(last_triggered, 'YYYY-MM-DD HH24:MI:SS'), '') AS last_triggered
            FROM trigger_formula
            WHERE id = $1 AND owner_id = $2;
        `, formulaID, ownerID)

		var formula UserFormula
		err := row.Scan(
			&formula.Id, &formula.FormulaRaw, &formula.Name, &formula.Description, &formula.IsNotified,
			&formula.IsActive, &formula.IsShuttedOff, &formula.IsHistoryOn, &formula.LastTriggered,
		)
		if err != nil {
			return nil, false, err
		}

		formulas = append(formulas, formula)
		return formulas, false, nil
	}

	offset := (page - 1) * limit

	rows, err := DB.Query(context.Background(), `
        SELECT id, formula_raw, COALESCE(name, ''), COALESCE(description, ''), is_notified, is_active,
            is_shutted_off, is_history_on, COALESCE(TO_CHAR(last_triggered, 'YYYY-MM-DD HH24:MI:SS'), '') AS last_triggered
        FROM trigger_formula
        WHERE owner_id = $1
        ORDER BY id DESC
        LIMIT $2 OFFSET $3;
    `, ownerID, limit+1, offset)

	if err != nil {
		return nil, false, err
	}
	defer rows.Close()

	for rows.Next() {
		var formula UserFormula
		err := rows.Scan(
			&formula.Id, &formula.FormulaRaw, &formula.Name, &formula.Description, &formula.IsNotified,
			&formula.IsActive, &formula.IsShuttedOff, &formula.IsHistoryOn, &formula.LastTriggered,
		)
		if err != nil {
			return nil, false, err
		}
		formulas = append(formulas, formula)
	}

	if err := rows.Err(); err != nil {
		return nil, false, err
	}

	if len(formulas) > limit {
		hasNext = true
		formulas = formulas[:limit]
	}

	return formulas, hasNext, nil
}

func GetFormulaHistory(formulaID int, limit int, page int, prevCursor int) (int, bool, []tempRow) {
	var rows pgx.Rows
	var err error

	if prevCursor == 0 {
		rows, err = DB.Query(context.Background(), `
			SELECT th.timestamp,
				   STRING_AGG(tc.name, ', ') AS names,
				   STRING_AGG(CAST(tch.value AS TEXT), ', ') AS values
			FROM trigger_history th
			LEFT JOIN trigger_component_history tch ON th.id = tch.trigger_history_id
			LEFT JOIN trigger_component tc ON tc.id = tch.component_id
			WHERE th.formula_id = $1
			GROUP BY th.timestamp
			ORDER BY th.timestamp DESC
			LIMIT $2;
		`, formulaID, limit+1)
	} else {
		cursorTime := time.Unix(int64(prevCursor), 0).UTC()
		rows, err = DB.Query(context.Background(), `
			SELECT th.timestamp,
				   STRING_AGG(tc.name, ', ') AS names,
				   STRING_AGG(CAST(tch.value AS TEXT), ', ') AS values
			FROM trigger_history th
			LEFT JOIN trigger_component_history tch ON th.id = tch.trigger_history_id
			LEFT JOIN trigger_component tc ON tc.id = tch.component_id
			WHERE th.formula_id = $1 AND th.timestamp < $2
			GROUP BY th.timestamp
			ORDER BY th.timestamp DESC
			LIMIT $3;
		`, formulaID, cursorTime, limit+1)
	}

	if err != nil {
		return 1, false, nil
	}

	defer rows.Close()

	var rawRows []tempRow
	var hasNext bool

	for rows.Next() {
		var r tempRow
		var names, values string

		if err := rows.Scan(&r.Timestamp, &names, &values); err != nil {
			return 2, false, nil
		}
		r.VarName = names
		r.Value = values
		rawRows = append(rawRows, r)
	}
	if err := rows.Err(); err != nil {
		return 3, false, nil
	}

	if len(rawRows) > limit {
		hasNext = true
		rawRows = rawRows[:limit]
	}

	return 0, hasNext, rawRows
}

func SaveFormula(formula string, formula_raw string, name string, uuid string) (int, error) {
	owner_id, err := GetIdbyUuid(uuid)
	if err != nil {
		return 0, err
	}

	var formulaId int
	err = DB.QueryRow(context.Background(), `
        INSERT INTO trigger_formula (formula, formula_raw, name, owner_id, is_notified, is_active, is_shutted_off, is_history_on, cooldown) 
        VALUES ($1, $2, $3, $4, false, false, false, false, 3600)
        RETURNING id
    `, formula, formula_raw, name, owner_id).Scan(&formulaId)

	if err != nil {
		return 0, err
	}

	return formulaId, nil
}

func SaveCryptoVariables(formulaID int, variables []CryptoVariable) error {
	for _, v := range variables {
		var triggerComponentID int

		err := DB.QueryRow(context.Background(), `
			SELECT tc.id
			FROM trigger_component tc
			JOIN crypto_currencies cc ON tc.currency_id = cc.id
			JOIN crypto_params cp ON tc.parameter_id = cp.id
			WHERE cc.currency = $1 AND cp.parameter = $2
			LIMIT 1
		`, v.Currency, v.Variable).Scan(&triggerComponentID)

		if err != nil {
			return err
		}

		_, err = DB.Exec(context.Background(), `
			INSERT INTO trigger_formula_component (component_id, formula_id)
			VALUES ($1, $2)
			ON CONFLICT DO NOTHING
		`, triggerComponentID, formulaID)

		if err != nil {
			return err
		}
	}

	return nil
}

// позже переделать обработку ошибок
func UpdateUserFormula(formulaId string, data map[string]interface{}) int {
	var setClauses []string
	var args []interface{}
	argIndex := 1

	for field, value := range data {
		setClauses = append(setClauses, fmt.Sprintf("%s = $%d", field, argIndex))
		args = append(args, value)
		argIndex++
	}

	if len(setClauses) == 0 {
		return 1
	}

	query := fmt.Sprintf("UPDATE trigger_formula SET %s WHERE id = $%d", strings.Join(setClauses, ", "), argIndex)
	args = append(args, formulaId)

	_, err := DB.Exec(context.Background(), query, args...)
	if err != nil {
		return 2
	}

	return 0
}

func DeleteUserFormula(formulaId string) int {
	_, err := DB.Exec(context.Background(), `DELETE FROM trigger_formula WHERE id = $1`, formulaId)

	if err != nil {
		fmt.Println(err)
		return 2
	}

	return 0
}

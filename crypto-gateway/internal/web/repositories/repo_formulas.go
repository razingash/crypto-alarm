package repositories

import (
	"context"
	"crypto-gateway/internal/web/db"
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
	Cooldown      int    `json:"cooldown"`
}

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

func GetFormulas(limit int, page int, formulaID string) ([]UserFormula, bool, error) {
	var formulas []UserFormula
	var hasNext bool

	if formulaID != "" {
		row := db.DB.QueryRow(context.Background(), `
            SELECT id, formula_raw, COALESCE(name, ''), COALESCE(description, ''), is_notified, is_active, is_shutted_off,
				is_history_on, cooldown, COALESCE(TO_CHAR(last_triggered, 'YYYY-MM-DD HH24:MI:SS'), '') AS last_triggered
            FROM trigger_formula
            WHERE id = $1;
        `, formulaID)

		var formula UserFormula
		err := row.Scan(
			&formula.Id, &formula.FormulaRaw, &formula.Name, &formula.Description, &formula.IsNotified,
			&formula.IsActive, &formula.IsShuttedOff, &formula.IsHistoryOn, &formula.Cooldown, &formula.LastTriggered,
		)
		if err != nil {
			return nil, false, err
		}

		formulas = append(formulas, formula)
		return formulas, false, nil
	}

	offset := (page - 1) * limit

	rows, err := db.DB.Query(context.Background(), `
        SELECT id, formula_raw, COALESCE(name, ''), COALESCE(description, ''), is_notified, is_active, is_shutted_off,
			is_history_on, cooldown, COALESCE(TO_CHAR(last_triggered, 'YYYY-MM-DD HH24:MI:SS'), '') AS last_triggered
        FROM trigger_formula
        ORDER BY id DESC
        LIMIT $1 OFFSET $2;
    `, limit+1, offset)

	if err != nil {
		return nil, false, err
	}
	defer rows.Close()

	for rows.Next() {
		var formula UserFormula
		err := rows.Scan(
			&formula.Id, &formula.FormulaRaw, &formula.Name, &formula.Description, &formula.IsNotified,
			&formula.IsActive, &formula.IsShuttedOff, &formula.IsHistoryOn, &formula.Cooldown, &formula.LastTriggered,
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

func GetFormulaHistory(formulaID int, limit int, page int, prevCursor int) (bool, []tempRow, error) {
	var rows pgx.Rows
	var err error

	if prevCursor == 0 {
		rows, err = db.DB.Query(context.Background(), `
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
		rows, err = db.DB.Query(context.Background(), `
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
		return false, nil, fmt.Errorf("failed to query history")
	}

	defer rows.Close()

	var rawRows []tempRow
	var hasNext bool

	for rows.Next() {
		var r tempRow
		var names, values string

		if err := rows.Scan(&r.Timestamp, &names, &values); err != nil {
			return false, nil, fmt.Errorf("failed to scan row")
		}
		r.VarName = names
		r.Value = values
		rawRows = append(rawRows, r)
	}
	if err := rows.Err(); err != nil {
		return false, nil, fmt.Errorf("error iterating over rows")
	}

	if len(rawRows) > limit {
		hasNext = true
		rawRows = rawRows[:limit]
	}

	return hasNext, rawRows, nil
}

func SaveFormula(formula string, formula_raw string, name string) (int, error) {
	var formulaId int
	err := db.DB.QueryRow(context.Background(), `
        INSERT INTO trigger_formula (formula, formula_raw, name, is_notified, is_active, is_shutted_off, is_history_on, cooldown) 
        VALUES ($1, $2, $3, false, false, false, false, 3600)
        RETURNING id
    `, formula, formula_raw, name).Scan(&formulaId)

	if err != nil {
		return 0, err
	}

	return formulaId, nil
}

func SaveCryptoVariables(formulaID int, variables []CryptoVariable) error {
	for _, v := range variables {
		var triggerComponentID int

		err := db.DB.QueryRow(context.Background(), `
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

		_, err = db.DB.Exec(context.Background(), `
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

func UpdateUserFormula(formulaId string, data map[string]interface{}) error {
	var setClauses []string
	var args []interface{}
	argIndex := 1

	for field, value := range data {
		setClauses = append(setClauses, fmt.Sprintf("%s = $%d", field, argIndex))
		args = append(args, value)
		argIndex++
	}

	if len(setClauses) == 0 {
		return fmt.Errorf("unprocessed error")
	}

	query := fmt.Sprintf("UPDATE trigger_formula SET %s WHERE id = $%d", strings.Join(setClauses, ", "), argIndex)
	args = append(args, formulaId)

	_, err := db.DB.Exec(context.Background(), query, args...)
	if err != nil {
		return fmt.Errorf("database error")
	}

	return nil
}

func DeleteUserFormula(formulaId string) error {
	_, err := db.DB.Exec(context.Background(), `DELETE FROM trigger_formula WHERE id = $1`, formulaId)

	if err != nil {
		fmt.Println(err)
		return fmt.Errorf("database error")
	}
	return nil
}

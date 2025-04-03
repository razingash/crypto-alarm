package db

import (
	"context"
	"fmt"
	"strings"
)

type UserFormula struct {
	Id            string `json:"id"`
	Formula       string `json:"formula"`
	Name          string `json:"name"`
	Description   string `json:"description"`
	IsNotified    bool   `json:"is_notified"`
	IsActive      bool   `json:"is_active"`
	IsHistoryOn   bool   `json:"is_history_on"`
	IsShuttedOff  bool   `json:"is_shutted_off"`
	LastTriggered string `json:"last_triggered"`
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

func GetUserFormulas(uuid string) ([]UserFormula, error) {
	owner_id, err := GetIdbyUuid(uuid)

	if err != nil {
		return nil, err
	}

	rows, err := DB.Query(context.Background(), `
	    SELECT id, formula, COALESCE(name, ''), COALESCE(description, ''), is_notified, is_active,
			is_shutted_off, is_history_on, COALESCE(TO_CHAR(last_triggered, 'YYYY-MM-DD HH24:MI:SS'), '') AS last_triggered
	    FROM trigger_formula
	    WHERE owner_id=$1;
	`, owner_id)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var formulas []UserFormula

	for rows.Next() {
		var formula UserFormula
		err := rows.Scan(
			&formula.Id, &formula.Formula, &formula.Name, &formula.Description, &formula.IsNotified,
			&formula.IsActive, &formula.IsShuttedOff, &formula.IsHistoryOn, &formula.LastTriggered,
		)
		if err != nil {
			return nil, err
		}
		formulas = append(formulas, formula)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return formulas, nil
}

func SaveFormula(formula string, uuid string) error {
	owner_id, err := GetIdbyUuid(uuid)

	if err != nil {
		return err
	}

	_, err2 := DB.Exec(context.Background(), `
		INSERT INTO trigger_formula (formula, owner_id, is_notified, is_active, is_shutted_off, is_history_on) 
		VALUES ($1, $2, false, false, false, false)`,
		formula, owner_id)

	if err2 != nil {
		return err2
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

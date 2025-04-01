package db

import (
	"context"
)

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

func getIdbyUuid(uuid string) (string, error) {
	var id string

	err := DB.QueryRow(context.Background(), `
		SELECT id 
		FROM user_user 
		WHERE uuid = $1
	`, uuid).Scan(&id)

	if err != nil {
		return "", err
	}

	return id, nil
}

func SaveFormula(formula string, uuid string) error {
	owner_id, err := getIdbyUuid(uuid)

	if err != nil {
		return err
	}

	_, err2 := DB.Exec(context.Background(), `
		INSERT INTO trigger_formula (formula, owner_id, is_notified, is_active) 
		VALUES ($1, $2, $3, $4)`,
		formula, owner_id, false, false)

	if err2 != nil {
		return err2
	}

	return nil
}

package db

import (
	"context"
	"time"
)

type CryptoApi struct {
	Id         int
	Api        string
	Cooldown   int
	IsActual   bool
	LastUpdate time.Time
}

func FetchSettings() ([]CryptoApi, error) {
	var apiSettings []CryptoApi

	rows, err := DB.Query(context.Background(), `
        SELECT id, api, cooldown, is_actual, last_updated::timestamptz AS last_updated
        FROM crypto_api
    `)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var setting CryptoApi
		err := rows.Scan(
			&setting.Id, &setting.Api, &setting.Cooldown, &setting.IsActual, &setting.LastUpdate,
		)
		if err != nil {
			return nil, err
		}
		apiSettings = append(apiSettings, setting)
	}

	return apiSettings, nil
}

func UpdateCooldown(id int, newCooldown int) error {
	_, err := DB.Exec(context.Background(), `
        UPDATE crypto_api
        SET cooldown = $1, last_updated = NOW()
        WHERE id = $2
    `, newCooldown, id)
	if err != nil {
		return err
	}
	return nil
}

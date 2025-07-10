package repositories

import (
	"context"
	"crypto-gateway/internal/web/db"
	"time"
)

type CryptoApi struct {
	Id          int
	Api         string
	Cooldown    int
	IsActual    bool
	IsHistoryOn bool
	LastUpdate  time.Time
}

func FetchSettings() ([]CryptoApi, error) {
	var apiSettings []CryptoApi

	rows, err := db.DB.Query(context.Background(), `
        SELECT id, api, cooldown, is_actual, is_history_on, last_updated::timestamptz AS last_updated
        FROM crypto_api
    `)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var setting CryptoApi
		err := rows.Scan(
			&setting.Id, &setting.Api, &setting.Cooldown, &setting.IsActual, &setting.IsHistoryOn, &setting.LastUpdate,
		)
		if err != nil {
			return nil, err
		}
		apiSettings = append(apiSettings, setting)
	}

	return apiSettings, nil
}

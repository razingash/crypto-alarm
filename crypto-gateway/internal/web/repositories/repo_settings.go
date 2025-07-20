package repositories

import (
	"context"
	"crypto-gateway/internal/web/db"
	"time"
)

type CryptoApi struct {
	Id          int       `json:"id"`
	Api         string    `json:"api"`
	Cooldown    int       `json:"cooldown"`
	IsActual    bool      `json:"is_actual"`
	IsHistoryOn bool      `json:"is_history_on"`
	LastUpdate  time.Time `json:"last_update"`
}

type Settings struct {
	Id       int    `json:"id"`
	Name     string `json:"name"`
	IsActive bool   `json:"is_active"`
}

func FetchApiSettings() ([]CryptoApi, error) {
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

func FetchConfigSettings() ([]Settings, error) {
	var settings []Settings

	rows, err := db.DB.Query(context.Background(), `
        SELECT id, name, is_active
        FROM settings
    `)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var setting Settings
		err := rows.Scan(
			&setting.Id, &setting.Name, &setting.IsActive,
		)
		if err != nil {
			return nil, err
		}
		settings = append(settings, setting)
	}

	return settings, nil
}

// используется только для setup
func GetInitialLoadSystem(ctx context.Context) ([]Settings, error) {
	rows, err := db.DB.Query(ctx, `
        SELECT id, name, is_active
        FROM settings
    `)
	if err != nil {
		return nil, err
	}

	var settings []Settings

	for rows.Next() {
		var s Settings
		if err := rows.Scan(&s.Id, &s.Name, &s.IsActive); err != nil {
			return nil, err
		}
		settings = append(settings, s)
	}

	defer rows.Close()

	return settings, nil
}

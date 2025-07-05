package main

import (
	"context"
	"crypto-gateway/config"
	"crypto-gateway/internal/analytics"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type apiParam struct {
	apiID     int
	paramID   int
	paramName string
}

func main() {
	config.LoadConfig()

	if err := fillCryptoModels(); err != nil {
		fmt.Println("fillCryptoModels failed: %w", err)
	} else {
		fmt.Println("Initialization completed")
	}
}

// заполняет CryptoApi и CryptoParams модели полученными данными относительно лучших апи
func fillCryptoModels() error {
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, config.Database_Url)
	if err != nil {
		return fmt.Errorf("failed to connect to target DB: %w", err)
	}
	defer pool.Close()

	controller := analytics.NewBinanceAPIController(5700)
	binAPI := analytics.NewBinanceAPI(controller)

	data, err := getInitialDataParams(ctx, binAPI)
	if err != nil {
		return fmt.Errorf("getInitialDataParams: %w", err)
	}
	if err := initializeCryptoModels(ctx, pool, data); err != nil {
		return fmt.Errorf("initializeCryptoModels: %w", err)
	}

	if err := getValidCurrencies(ctx, pool, binAPI); err != nil {
		return fmt.Errorf("getValidCurrencies: %w", err)
	}

	if err := createTriggerComponents(ctx, pool); err != nil {
		return fmt.Errorf("createTriggerComponents: %w", err)
	}
	return nil
}

// receives the keys to the datasets regarding the API Binance to initialize them in the database
// and use in the keyboard on the client side
func getInitialDataParams(ctx context.Context, binAPI *analytics.BinanceAPI) (map[string]map[string]interface{}, error) {
	data := make(map[string]map[string]interface{})
	var endpoints = map[string]int{
		"/v3/ticker/price": 2,
		"/v3/ticker/24hr":  80,
	}

	for ep, weight := range endpoints {
		res, err := binAPI.Get(ctx, ep, weight, map[string]string{"symbol": "ETHBTC"})
		if err != nil {
			return nil, fmt.Errorf("binAPI.Get %s: %w", ep, err)
		}

		var parsed map[string]interface{}
		if err := json.Unmarshal(res, &parsed); err != nil {
			return nil, fmt.Errorf("json.Unmarshal for %s: %w", ep, err)
		}

		data[ep] = parsed
	}
	return data, nil
}

// Initialization of variables for user strategies
func initializeCryptoModels(ctx context.Context, pool *pgxpool.Pool, dataset map[string]map[string]interface{}) error {
	for ep, kv := range dataset {
		var cryptoApiID int
		err := pool.QueryRow(ctx, `
			INSERT INTO crypto_api(api) VALUES($1) RETURNING id
		`, ep).Scan(&cryptoApiID)
		if err != nil {
			return fmt.Errorf("failed to insert crypto_api %s: %w", ep, err)
		}

		for param := range kv {
			if param == "symbol" {
				continue
			}
			if _, err := pool.Exec(ctx,
				`INSERT INTO crypto_params(parameter, crypto_api_id) VALUES($1, $2)`,
				param, cryptoApiID); err != nil {
				return fmt.Errorf("failed to insert param %s: %w", param, err)
			}
		}
	}
	return nil
}

// getValidCurrencies — формирует список пересечения доступных валют чтобы использовать только те которые наверняка поддерживаются
func getValidCurrencies(ctx context.Context, pool *pgxpool.Pool, binAPI *analytics.BinanceAPI) error {
	symbolSets := make([][]string, 3)
	eps := []string{"/v3/ticker/price", "/v3/ticker/24hr", "/v3/exchangeInfo"}

	for i, ep := range eps {
		resBytes, err := binAPI.Get(ctx, ep, 80, nil)
		if err != nil {
			return fmt.Errorf("binAPI.Get %s: %w", ep, err)
		}

		switch i {
		case 0, 1:
			var arr []map[string]interface{}
			if err := json.Unmarshal(resBytes, &arr); err != nil {
				return fmt.Errorf("json.Unmarshal for %s: %w", ep, err)
			}
			for _, item := range arr {
				if s, ok := item["symbol"].(string); ok {
					symbolSets[i] = append(symbolSets[i], s)
				}
			}
		case 2:
			var obj map[string]interface{}
			if err := json.Unmarshal(resBytes, &obj); err != nil {
				return fmt.Errorf("json.Unmarshal for %s: %w", ep, err)
			}
			if arr, ok := obj["symbols"].([]interface{}); ok {
				for _, it := range arr {
					if m, mok := it.(map[string]interface{}); mok {
						if s, sok := m["symbol"].(string); sok {
							symbolSets[2] = append(symbolSets[2], s)
						}
					}
				}
			}
		}
	}

	set0 := make(map[string]struct{})
	for _, s := range symbolSets[0] {
		set0[s] = struct{}{}
	}
	for _, s := range symbolSets[1] {
		if _, ok := set0[s]; ok {
			set0[s] = struct{}{}
		} else {
			delete(set0, s)
		}
	}
	for _, s := range symbolSets[2] {
		if _, ok := set0[s]; !ok {
			delete(set0, s)
		}
	}

	for currency := range set0 {
		if _, err := pool.Exec(ctx,
			`INSERT INTO crypto_currencies(currency) VALUES($1) ON CONFLICT DO NOTHING`, currency); err != nil {
			return fmt.Errorf("insert crypto_currency %s: %w", currency, err)
		}
	}
	return nil
}

// заполняет модель TriggerComponent исходя из имеющихся данных
func createTriggerComponents(ctx context.Context, pool *pgxpool.Pool) error {
	apiParamsRows, err := pool.Query(ctx, `
        SELECT api.id AS api_id, p.id AS param_id, p.parameter
        FROM crypto_api AS api
        JOIN crypto_params p ON api.id = p.crypto_api_id
    `)
	if err != nil {
		return fmt.Errorf("select crypto_api+params: %w", err)
	}
	defer apiParamsRows.Close()

	var apiParams []apiParam
	for apiParamsRows.Next() {
		var ap apiParam
		if err := apiParamsRows.Scan(&ap.apiID, &ap.paramID, &ap.paramName); err != nil {
			return fmt.Errorf("scan api/param: %w", err)
		}
		apiParams = append(apiParams, ap)
	}

	type currency struct {
		id       int
		currency string
	}

	currRows, err := pool.Query(ctx, `SELECT id, currency FROM crypto_currencies`)
	if err != nil {
		return fmt.Errorf("select currencies: %w", err)
	}
	defer currRows.Close()

	var currencies []currency
	for currRows.Next() {
		var c currency
		if err := currRows.Scan(&c.id, &c.currency); err != nil {
			return fmt.Errorf("scan curr: %w", err)
		}
		currencies = append(currencies, c)
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin_tx: %w", err)
	}
	defer tx.Rollback(ctx)

	stmt, err := tx.Prepare(ctx, "insert_trigger_component", `
        INSERT INTO trigger_component(api_id, parameter_id, currency_id, name) VALUES ($1, $2, $3, $4)
    `)
	if err != nil {
		return fmt.Errorf("prepare insert trigger_component: %w", err)
	}

	for _, ap := range apiParams {
		for _, c := range currencies {
			name := fmt.Sprintf("%s_%s", c.currency, ap.paramName) // исправленный name
			if _, err := tx.Exec(ctx, stmt.Name, ap.apiID, ap.paramID, c.id, name); err != nil {
				return fmt.Errorf("insert trigger_component: %w", err)
			}
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("tx commit: %w", err)
	}
	return nil
}

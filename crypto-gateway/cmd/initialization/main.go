package main

import (
	"context"
	"crypto-gateway/config"
	"crypto-gateway/internal/appmetrics"
	"crypto-gateway/internal/modules/strategy/service"
	"crypto-gateway/internal/web/db"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type apiParam struct {
	apiID     int
	paramID   int
	paramName string
}

// contain endpoints with standard weight
var baseEndpointsWeights = map[string]int{
	"/v3/ping":         1,
	"/v3/ticker/price": 2,
	"/v3/klines":       4,
	"/v3/ticker/24hr":  80,
	// позже добавить и его, это важный эндпоинт, просто ему нужны атрибуты, и нужно будет шаманить с клавиатурой
	//"/v3/exchangeInfo": 20,
}

func main() {
	config.LoadConfig()

	if err := fillCryptoModels(); err != nil {
		appmetrics.AnalyticsServiceLogging(4, "fillCryptoModels failed", err)
		fmt.Println("fillCryptoModels failed: %w", err)
	} else {
		fmt.Println("Initialization completed")
	}
}

// заполняет CryptoApi и CryptoParams модели полученными данными относительно лучших апи
func fillCryptoModels() error {
	ctx := context.Background()
	db.InitDB()

	controller := service.NewBinanceAPIController(5700)
	binAPI := service.NewBinanceAPI(controller)

	data, err := getInitialDataParams(ctx, binAPI)
	if err != nil {
		appmetrics.AnalyticsServiceLogging(4, "getInitialDataParams error", err)
		return fmt.Errorf("getInitialDataParams: %w", err)
	}
	if err := initializeCryptoModels(ctx, db.DB, data); err != nil {
		appmetrics.AnalyticsServiceLogging(4, "initializeCryptoModels error", err)
		return fmt.Errorf("initializeCryptoModels: %w", err)
	}
	if err := getValidCurrencies(ctx, db.DB, binAPI); err != nil {
		appmetrics.AnalyticsServiceLogging(4, "getValidCurrencies error", err)
		return fmt.Errorf("getValidCurrencies: %w", err)
	}
	if err := createTriggerComponents(ctx, db.DB); err != nil {
		appmetrics.AnalyticsServiceLogging(4, "createTriggerComponents error", err)
		return fmt.Errorf("createTriggerComponents: %w", err)
	}
	if err := createInitialSettings(ctx, db.DB); err != nil {
		appmetrics.AnalyticsServiceLogging(4, "createInitialSettings error", err)
		return fmt.Errorf("createInitialSettings: %w", err)
	}
	return nil
}

// receives the keys to the datasets regarding the API Binance to initialize them in the database
// and use in the keyboard on the client side
func getInitialDataParams(ctx context.Context, binAPI *service.BinanceAPI) (map[string]map[string]interface{}, error) {
	data := make(map[string]map[string]interface{})

	for ep, weight := range baseEndpointsWeights {
		if ep == "/v3/ping" {
			continue
		}
		if ep == "/v3/klines" {
			_, err := binAPI.Get(ctx, ep, weight, map[string]string{"symbol": "ETHBTC", "interval": "1h"})
			if err != nil {
				return nil, fmt.Errorf("binAPI.Get %s: %w", ep, err)
			}

			data[ep] = map[string]interface{}{
				"OpenTime":                 nil,
				"Open":                     nil,
				"High":                     nil,
				"Low":                      nil,
				"Close":                    nil,
				"Volume":                   nil,
				"CloseTime":                nil,
				"QuoteAssetVolume":         nil,
				"NumberOfTrades":           nil,
				"TakerBuyBaseAssetVolume":  nil,
				"TakerBuyQuoteAssetVolume": nil,
				"Ignore":                   nil,
			}
			continue
		}

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
	now := time.Now().UTC()
	for ep, weight := range baseEndpointsWeights {
		var cryptoApiID int
		isAccessible := true
		if ep == "/v3/ping" {
			isAccessible = false
		}

		err := pool.QueryRow(ctx, `
			INSERT INTO crypto_api (api, is_accessible)
			VALUES ($1, $2)
			RETURNING id
		`, ep, isAccessible).Scan(&cryptoApiID)
		if err != nil {
			return fmt.Errorf("failed to insert into crypto_api for %s: %w", ep, err)
		}

		_, err = pool.Exec(ctx, `
			INSERT INTO crypto_api_history (crypto_api_id, weight, created_at)
			VALUES ($1, $2, $3)
		`, cryptoApiID, weight, now)
		if err != nil {
			return fmt.Errorf("failed to insert into crypto_api_history for %s: %w", ep, err)
		}

		if params, ok := dataset[ep]; ok {
			for param := range params {
				if param == "symbol" {
					continue
				}
				_, err := pool.Exec(ctx,
					`INSERT INTO crypto_params (parameter, crypto_api_id)
					 VALUES ($1, $2)`,
					param, cryptoApiID)
				if err != nil {
					return fmt.Errorf("failed to insert param %s for %s: %w", param, ep, err)
				}
			}
		}
		// v3/klines initialization

	}

	return nil
}

// getValidCurrencies — формирует список пересечения доступных валют чтобы использовать только те которые наверняка поддерживаются
func getValidCurrencies(ctx context.Context, pool *pgxpool.Pool, binAPI *service.BinanceAPI) error {
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

// заполняет настройки
func createInitialSettings(ctx context.Context, pool *pgxpool.Pool) error {
	settings := []struct {
		name     string
		isActive bool
	}{
		{"Average System Load", true},
	}

	if len(settings) == 0 {
		return nil
	}

	args := []interface{}{}
	valueStrings := []string{}
	argIndex := 1
	for _, setting := range settings {
		valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d)", argIndex, argIndex+1))
		args = append(args, setting.name, setting.isActive)
		argIndex += 2
	}

	query := fmt.Sprintf(`
		INSERT INTO settings (name, is_active)
		VALUES %s
	`, strings.Join(valueStrings, ", "))

	_, err := pool.Exec(ctx, query, args...)
	return err
}

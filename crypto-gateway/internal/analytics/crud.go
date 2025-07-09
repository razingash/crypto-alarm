package analytics

import (
	"context"
	"crypto-gateway/internal/web/db"
	"database/sql"
	"fmt"
	"time"
)

// перенести в репозиторий после того как блок будет готов
type ActualComponentInfo struct {
	Cooldown int
	Count    int
}

type FormulaRecord struct {
	ID      int
	Formula string
}

type ApiUpdate struct {
	Endpoint string `json:"endpoint"`
	Cooldown *int   `json:"cooldown,omitempty"`
	History  *bool  `json:"history,omitempty"`
}

func GetActiveFormulas(ctx context.Context) ([]FormulaRecord, error) {
	rows, err := db.DB.Query(ctx, `
        SELECT id, formula FROM trigger_formula WHERE is_active=true
    `)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var formulas []FormulaRecord

	for rows.Next() {
		var rec FormulaRecord
		if err := rows.Scan(&rec.ID, &rec.Formula); err != nil {
			return nil, err
		}
		formulas = append(formulas, rec)
	}

	return formulas, nil
}

// получает необходимые апи, к которым нужно делать запрос в зависимости от актальности формул и компонентов
func GetActualComponents(ctx context.Context) (map[string]ActualComponentInfo, error) {
	rows, err := db.DB.Query(ctx, `
        SELECT ca.api, ca.cooldown, COUNT(*) AS count
        FROM crypto_api ca
        JOIN trigger_component tc ON ca.id = tc.api_id
        JOIN crypto_params cp ON tc.parameter_id = cp.id
        JOIN trigger_formula_component tfc ON tfc.component_id = tc.id
        JOIN trigger_formula tf ON tf.id = tfc.formula_id
        WHERE ca.is_actual = true
          AND cp.is_active = true
          AND tf.is_active = true
        GROUP BY ca.api, ca.cooldown
    `)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]ActualComponentInfo)
	for rows.Next() {
		var api string
		var cooldown, count int
		if err := rows.Scan(&api, &cooldown, &count); err != nil {
			return nil, err
		}
		result[api] = ActualComponentInfo{Cooldown: cooldown, Count: count}
	}

	return result, nil
}

// получает список необходимых параметров для конкретного эндпоинта
func GetNeededFieldsFromEndpoint(ctx context.Context, endpoint string) (map[string][]string, error) {
	rows, err := db.DB.Query(ctx, `
        SELECT cc.currency, cp.parameter
        FROM crypto_params cp
        JOIN trigger_component tc ON tc.parameter_id = cp.id
        JOIN crypto_currencies cc ON tc.currency_id = cc.id
        JOIN crypto_api ca ON ca.id = tc.api_id
        JOIN trigger_formula_component tfc ON tfc.component_id = tc.id
        JOIN trigger_formula tf ON tf.id = tfc.formula_id
        WHERE ca.api = $1
          AND ca.is_actual = true
          AND tf.is_active = true
          AND tf.is_shutted_off = false
          AND cp.is_active = true
    `, endpoint)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string][]string)
	for rows.Next() {
		var symbol, param string
		if err := rows.Scan(&symbol, &param); err != nil {
			return nil, err
		}
		result[symbol] = append(result[symbol], param)
	}

	return result, nil
}

// записывает в историю сопутствующие данные сработавших триггеров
func AddTriggerHistory(ctx context.Context, data map[int][]string, formulasValues map[string]float64) error {
	tx, err := db.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	now := time.Now().UTC()
	for formulaID, variables := range data {
		var triggerHistoryID int
		err := tx.QueryRow(ctx, `
            INSERT INTO trigger_history (formula_id, timestamp, status)
            VALUES ($1, $2, true)
            RETURNING id
        `, formulaID, now).Scan(&triggerHistoryID)
		if err != nil {
			return err
		}

		rows, err := tx.Query(ctx, `
            SELECT tfc.id, tfc.component_id, c.name
            FROM trigger_formula_component tfc
            JOIN trigger_component c ON tfc.component_id = c.id
            WHERE tfc.formula_id = $1
        `, formulaID)
		if err != nil {
			return err
		}

		componentMap := make(map[string]struct {
			ComponentID int
			TFCID       int
		})

		for rows.Next() {
			var tfcID, compID int
			var name string
			if err := rows.Scan(&tfcID, &compID, &name); err != nil {
				return err
			}
			componentMap[name] = struct {
				ComponentID int
				TFCID       int
			}{ComponentID: compID, TFCID: tfcID}
		}
		rows.Close()

		for _, varName := range variables {
			value, ok := formulasValues[varName]
			if !ok {
				continue
			}

			comp, found := componentMap[varName]
			if !found {
				continue
			}

			_, err := tx.Exec(ctx, `
                INSERT INTO trigger_components_history (trigger_history_id, component_id, value)
                VALUES ($1, $2, $3)
            `, triggerHistoryID, comp.ComponentID, value)
			if err != nil {
				return err
			}
		}
	}

	return tx.Commit(ctx)
}

// возвращает список эндпоинтов которые нужно замерять, если они есть
func GetRecordedEndpoints(ctx context.Context) ([]string, error) {
	rows, err := db.DB.Query(ctx, `
		SELECT api FROM crypto_api WHERE is_actual = true AND is_history_on = true
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var endpoints []string
	for rows.Next() {
		var api string
		if err := rows.Scan(&api); err != nil {
			return nil, err
		}
		endpoints = append(endpoints, api)
	}
	return endpoints, nil
}

// добавляет в бд новые весы эндпоинта Binance, если он был изменен
func SaveEndpointWeightIfChanged(ctx context.Context, endpoint string, weight int) error {
	// 3 запроса на одну проверку, как-то слишком
	var apiID int
	err := db.DB.QueryRow(ctx, `
		SELECT id FROM crypto_api WHERE api = $1 AND is_actual = true AND is_history_on = true
	`, endpoint).Scan(&apiID)
	if err != nil {
		return err
	}

	var lastWeight int
	err = db.DB.QueryRow(ctx, `
		SELECT weight FROM crypto_api_history 
		WHERE crypto_api_id = $1 
		ORDER BY created_at DESC 
		LIMIT 1
	`, apiID).Scan(&lastWeight)

	if err == sql.ErrNoRows || lastWeight != weight {
		_, err = db.DB.Exec(ctx, `
			INSERT INTO crypto_api_history (crypto_api_id, weight) 
			VALUES ($1, $2)
		`, apiID, weight)
		return err
	}

	return nil
}

func UpdateEndpointsSettings(updates []ApiUpdate) ([]int, error) { // если меняется история то добавлять в Recorded новый эндпоинт
	if len(updates) == 0 {
		return nil, nil
	}
	updatedIds := make([]int, 0)

	// синхронизация, без неё будет менее безопасно(хотя мб без багов всеравно)
	stApi := StBinanceApi
	if stApi == nil {
		return nil, fmt.Errorf("StBinanceApi is nil")
	}

	stApi.Controller.Mu.Lock()
	defer stApi.Controller.Mu.Unlock()

	recordedSet := make(map[string]struct{}, len(stApi.RecordedAPI))
	for _, e := range stApi.RecordedAPI {
		recordedSet[e] = struct{}{}
	}

	for _, item := range updates {
		var id int
		err := db.DB.QueryRow(context.Background(), `
			UPDATE crypto_api
	    	SET 
	    	    cooldown = COALESCE($2, cooldown),
	    	    is_history_on = COALESCE($3, is_history_on),
	    	    last_updated = now()
	    	WHERE api = $1
	    	RETURNING id;
		`, item.Endpoint, item.Cooldown, item.History).Scan(&id)
		if err != nil {
			return nil, err
		}
		updatedIds = append(updatedIds, id)

		if item.History != nil {
			if *item.History {
				if _, exists := recordedSet[item.Endpoint]; !exists {
					StBinanceApi.RecordedAPI = append(stApi.RecordedAPI, item.Endpoint)
					recordedSet[item.Endpoint] = struct{}{}
				}
			} else {
				if _, exists := recordedSet[item.Endpoint]; exists {
					newList := make([]string, 0, len(stApi.RecordedAPI))
					for _, e := range stApi.RecordedAPI {
						if e != item.Endpoint {
							newList = append(newList, e)
						}
					}
					stApi.RecordedAPI = newList
					delete(recordedSet, item.Endpoint)
				}
			}
		}
	}

	return updatedIds, nil
}

package repo

import (
	"context"
	"crypto-gateway/internal/web/db"
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

type ConfigUpdate struct {
	ID       int  `json:"id"`
	IsActive bool `json:"is_active"`
}

// получает необходимые апи, к которым нужно делать запрос в зависимости от актальности формул и компонентов
func GetActualComponents(ctx context.Context) (map[string]ActualComponentInfo, error) {
	result := make(map[string]ActualComponentInfo)

	rows, err := db.DB.Query(ctx, `
        SELECT ca.api, ca.cooldown, COUNT(*) AS count
        FROM crypto_api ca
        JOIN trigger_component tc       ON ca.id = tc.api_id
        JOIN crypto_params cp           ON tc.parameter_id    = cp.id
        JOIN trigger_formula_component tfc ON tfc.component_id  = tc.id
        JOIN trigger_formula tf         ON tf.id               = tfc.formula_id
        JOIN crypto_strategy_formula csf ON csf.formula_id     = tf.id
        JOIN crypto_strategy cs         ON cs.id               = csf.strategy_id
        WHERE ca.is_actual = true
          AND cp.is_active = true
          AND cs.is_active = true
        GROUP BY ca.api, ca.cooldown
    `)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var api string
		var cooldown, count int
		if err := rows.Scan(&api, &cooldown, &count); err != nil {
			return nil, err
		}
		result[api] = ActualComponentInfo{Cooldown: cooldown, Count: count}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	rows2, err := db.DB.Query(ctx, `
        SELECT ca.api, ca.cooldown, COUNT(*) AS count
        FROM crypto_api ca
        JOIN crypto_variables_api cva ON ca.id  = cva.api_id
        JOIN crypto_params cp ON cp.id = cva.parameter_id
        JOIN crypto_strategy_variable csv ON csv.crypto_variable_id = cva.variable_id
        JOIN crypto_strategy cs ON cs.id = csv.strategy_id
        WHERE ca.is_actual = true
          AND cp.is_active = true
          AND cs.is_active = true
        GROUP BY ca.api, ca.cooldown
    `)
	if err != nil {
		return nil, err
	}
	defer rows2.Close()

	for rows2.Next() {
		var api string
		var cooldown, count int
		if err := rows2.Scan(&api, &cooldown, &count); err != nil {
			return nil, err
		}
		if prev, ok := result[api]; ok {
			prev.Count += count
			result[api] = prev
		} else {
			result[api] = ActualComponentInfo{Cooldown: cooldown, Count: count}
		}
	}
	if err := rows2.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

// получает список необходимых параметров для конкретного эндпоинта
func GetNeededFieldsFromEndpoint(ctx context.Context, endpoint string) (map[string][]string, error) {
	result := make(map[string][]string)

	rows, err := db.DB.Query(ctx, `
        SELECT cc.currency, cp.parameter
        FROM crypto_params cp
        JOIN trigger_component tc ON tc.parameter_id = cp.id
        JOIN crypto_currencies cc ON tc.currency_id = cc.id
        JOIN crypto_api ca ON ca.id = tc.api_id
        JOIN trigger_formula_component tfc ON tfc.component_id = tc.id
        JOIN trigger_formula tf ON tf.id = tfc.formula_id
        JOIN crypto_strategy_formula csf ON csf.formula_id = tf.id
        JOIN crypto_strategy cs ON cs.id = csf.strategy_id
        WHERE ca.api = $1
          AND ca.is_actual = true
          AND cs.is_active = true
          AND cs.is_shutted_off = false
          AND cp.is_active = true;
    `, endpoint)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var symbol, param string
		if err := rows.Scan(&symbol, &param); err != nil {
			return nil, err
		}
		result[symbol] = append(result[symbol], param)
	}

	rows2, err := db.DB.Query(ctx, `
        SELECT cc.currency, cp.parameter
		FROM crypto_params cp
		JOIN crypto_variables_api cva ON cva.parameter_id = cp.id
		JOIN crypto_api ca ON ca.id = cva.api_id
		JOIN crypto_strategy_variable csv ON csv.crypto_variable_id = cva.variable_id
		JOIN crypto_strategy cs ON cs.id = csv.strategy_id
		JOIN crypto_currencies cc ON csv.crypto_currency_id = cc.id
		WHERE ca.api = $1
		  AND ca.is_actual = true
		  AND cs.is_active = true
		  AND cs.is_shutted_off = false
		  AND cp.is_active = true;
    `, endpoint)
	if err != nil {
		return nil, err
	}
	defer rows2.Close()

	for rows2.Next() {
		var symbol, param string
		if err := rows2.Scan(&symbol, &param); err != nil {
			return nil, err
		}
		result[symbol] = append(result[symbol], param)
	}

	return result, nil
}

// записывает в историю сопутствующие данные сработавших триггеров
func AddTriggerHistory(ctx context.Context, data map[int][]string, formulasValues map[string]float64) error {
	fmt.Println(1, data, formulasValues)
	tx, err := db.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	now := time.Now().UTC()
	for formulaID, variables := range data {
		var strategyHistoryID int
		err := tx.QueryRow(ctx, `
            INSERT INTO strategy_history (formula_id, timestamp, status)
            VALUES ($1, $2, true)
            RETURNING id
        `, formulaID, now).Scan(&strategyHistoryID)
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
                INSERT INTO trigger_component_history (expression_id, component_id, value)
                VALUES ($1, $2, $3)
            `, strategyHistoryID, comp.ComponentID, value)
			if err != nil {
				fmt.Println(err)
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

// возвращает словарь эндпоинтов к их актуальной стоимости
func GetActualEndpointsWeight(ctx context.Context) (map[string]int, error) {
	endpoints := make(map[string]int)

	rows, err := db.DB.Query(ctx, `
		SELECT ca.api,
			COALESCE((
			    SELECT cah.weight
			    FROM crypto_api_history cah
			    WHERE cah.crypto_api_id = ca.id
			    ORDER BY cah.created_at DESC
			    LIMIT 1
			), ca.base_weight) AS final_weight
		FROM crypto_api ca
		WHERE ca.is_actual = true
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to query actual endpoints with weights: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var api string
		var weight int
		if err := rows.Scan(&api, &weight); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		endpoints[api] = weight
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return endpoints, nil
}

// добавляет в бд новые весы эндпоинта Binance. Нужно использовать только если апи на самом деле изменен
func SaveEndpointWeight(ctx context.Context, endpoint string, weight int) error {
	// неприятно что нужно знать id Api - это добавляет лишний запрос.
	// Как вариант можно использовать позицию endpoint'а + 1 относительно endpoints
	var apiID int
	err := db.DB.QueryRow(ctx, `
		SELECT id FROM crypto_api WHERE api = $1
	`, endpoint).Scan(&apiID)
	if err != nil {
		return err
	}

	// вернуть этот участок если все будет плохо или понадобится строгий режим
	/*
		var lastWeight int
		err = db.DB.QueryRow(ctx, `
			SELECT weight FROM crypto_api_history
			WHERE crypto_api_id = $1
			ORDER BY created_at DESC
			LIMIT 1
		`, apiID).Scan(&lastWeight)
		if err == sql.ErrNoRows { || lastWeight != weight {
	*/

	_, err = db.DB.Exec(ctx, `
			INSERT INTO crypto_api_history (crypto_api_id, weight) 
			VALUES ($1, $2)
		`, apiID, weight)
	return err
}

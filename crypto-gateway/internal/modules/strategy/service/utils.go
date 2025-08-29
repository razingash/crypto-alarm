package service

import (
	"context"
	"crypto-gateway/internal/modules/strategy/repo"
	"crypto-gateway/internal/web/db"
	"fmt"
)

type PatchSettingsRequest struct {
	Api    []repo.ApiUpdate    `json:"api,omitempty"`
	Config []repo.ConfigUpdate `json:"config,omitempty"`
}

func UpdateSettings(updates PatchSettingsRequest) ([]int, error) { // если меняется история то добавлять в Recorded новый эндпоинт
	updatedIds := make([]int, 0)

	stApi := StBinanceApi
	alm := AverageLoadMetrics

	if stApi == nil {
		return nil, fmt.Errorf("StBinanceApi is nil")
	}
	if alm == nil {
		return nil, fmt.Errorf("AverageLoadMetrics is nil")
	}

	stApi.Controller.Mu.Lock()
	defer stApi.Controller.Mu.Unlock()

	recordedSet := make(map[string]struct{}, len(stApi.RecordedAPI))
	for _, e := range stApi.RecordedAPI {
		recordedSet[e] = struct{}{}
	}

	for _, item := range updates.Api {
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
					stApi.RecordedAPI = append(stApi.RecordedAPI, item.Endpoint)
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

	for _, item := range updates.Config {
		_, err := db.DB.Exec(context.Background(), `
			UPDATE settings
			SET is_active = $2
			WHERE id = $1
		`, item.ID, item.IsActive)
		if err != nil {
			return nil, err
		}
		if item.ID == 1 {
			alm.ToggleAverageLoadMetrics(item.IsActive)
			Collector.SwitchCollect(item.IsActive)
		}
	}

	return updatedIds, nil
}

package engine

import (
	"context"
	"crypto-gateway/internal/web/db"
	"encoding/json"
	"fmt"
)

type DiagramStrategyIDs struct {
	DiagramID   int64
	StrategyIDs []int64
}

type Diagram struct {
	ID   int64           `json:"id"`
	Data json.RawMessage `json:"data"`
}

func ExtractStrategiesIDs(diagrams []Diagram) ([]DiagramStrategyIDs, error) {
	var result []DiagramStrategyIDs

	for _, d := range diagrams {
		var payload struct {
			Strategies []struct {
				ID int64 `json:"id"`
			} `json:"strategies"`
		}

		if err := json.Unmarshal(d.Data, &payload); err != nil {
			return nil, fmt.Errorf("failed to unmarshal diagram %d: %w", d.ID, err)
		}

		strategyIDs := make([]int64, 0, len(payload.Strategies))
		for _, s := range payload.Strategies {
			strategyIDs = append(strategyIDs, s.ID)
		}

		result = append(result, DiagramStrategyIDs{
			DiagramID:   d.ID,
			StrategyIDs: strategyIDs,
		})
	}

	return result, nil
}

func GetActiveDiagrams() ([]Diagram, error) {
	ctx := context.Background()
	rows, err := db.DB.Query(ctx, `SELECT id, data FROM diagrams WHERE data IS NOT NULL AND isActive=true`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var diagrams []Diagram
	for rows.Next() {
		var diagram Diagram
		if err := rows.Scan(&diagram.ID, &diagram.Data); err != nil {
			return nil, err
		}
		diagrams = append(diagrams, diagram)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return diagrams, nil
}

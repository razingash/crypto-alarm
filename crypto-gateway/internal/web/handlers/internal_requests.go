package handlers

import (
	"context"
	"crypto-gateway/internal/analytics"
	"crypto-gateway/internal/web/db"
	"crypto-gateway/internal/web/repositories"
	"fmt"
)

// всю эту фигню после тестов убрать, это лишний слой, + запросы в бд можно не делать(про запросы актуально только в deleteStrategyFromGraph)
func deleteStrategyFromGraph(strategyID int) {
	analytics.StOrchestrator.DependencyGraph.RemoveStrategy(strategyID)
	analytics.StOrchestrator.LaunchNeededAPI(context.Background())
}

func updateStrategyInGraph(strategyID int) {
	analytics.StOrchestrator.DependencyGraph.RemoveStrategy(strategyID)
	analytics.StOrchestrator.DependencyGraph.AddStrategy(strategyID, repositories.GetStrategyFullFormulasById(strategyID))
	analytics.StOrchestrator.LaunchNeededAPI(context.Background())
}

// при обновлении переменной вызывать данную функцию
func updateStrategiesRelatedToVariable(variableId int) {
	ctx := context.Background()
	rows, err := db.DB.Query(ctx, `
		SELECT cs.id
		FROM crypto_strategy_variable csv
		JOIN crypto_strategy cs ON csv.strategy_id = cs.id
		WHERE csv.crypto_variable_id = $1 AND cs.is_active = true
	`, variableId)
	if err != nil {
		fmt.Printf("failed to fetch related strategies for variable %d: %v\n", variableId, err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var strategyID int
		if err := rows.Scan(&strategyID); err != nil {
			fmt.Printf("failed to scan strategy id: %v\n", err)
			continue
		}

		analytics.StOrchestrator.DependencyGraph.RemoveStrategy(strategyID)
		analytics.StOrchestrator.DependencyGraph.AddStrategy(strategyID, repositories.GetStrategyFullFormulasById(strategyID))
	}
	if err := rows.Err(); err != nil {
		fmt.Printf("error after scanning strategy ids: %v\n", err)
	}

	analytics.StOrchestrator.LaunchNeededAPI(ctx)
}

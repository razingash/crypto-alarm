package web

import (
	"context"
	"crypto-gateway/internal/modules/strategy/repo"
	"crypto-gateway/internal/modules/strategy/service"
	"crypto-gateway/internal/web/db"
	"fmt"
)

// всю эту фигню после тестов убрать, это лишний слой, + запросы в бд можно не делать(про запросы актуально только в deleteStrategyFromGraph)
func deleteStrategyFromGraph(strategyID int) {
	service.StOrchestrator.DependencyGraph.RemoveStrategy(strategyID)
	service.StOrchestrator.LaunchNeededAPI(context.Background())
}

func updateStrategyInGraph(strategyID int) {
	service.StOrchestrator.DependencyGraph.RemoveStrategy(strategyID)
	service.StOrchestrator.DependencyGraph.AddStrategy(strategyID, repo.GetStrategyFullFormulasById(strategyID))
	service.StOrchestrator.LaunchNeededAPI(context.Background())
}

// при обновлении переменной вызывать данную функцию
func UpdateStrategiesRelatedToVariable(variableId int) {
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

		service.StOrchestrator.DependencyGraph.RemoveStrategy(strategyID)
		service.StOrchestrator.DependencyGraph.AddStrategy(strategyID, repo.GetStrategyFullFormulasById(strategyID))
	}
	if err := rows.Err(); err != nil {
		fmt.Printf("error after scanning strategy ids: %v\n", err)
	}

	service.StOrchestrator.LaunchNeededAPI(ctx)
}

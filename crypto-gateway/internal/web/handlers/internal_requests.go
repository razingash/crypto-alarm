package handlers

import (
	"context"
	"crypto-gateway/internal/analytics"
	"crypto-gateway/internal/web/repositories"
)

// всю эту фигню после тестов убрать, это лишний слой, + запросы в бд можно не делать
func deleteStrategyFromGraph(strategyID int) {
	analytics.StOrchestrator.DependencyGraph.RemoveStrategy(strategyID)
	analytics.StOrchestrator.LaunchNeededAPI(context.Background())
}

func addStrategyToGraph(strategyID int) {
	analytics.StOrchestrator.DependencyGraph.AddStrategy(strategyID, repositories.GetStrategyFormulasById(strategyID))
	analytics.StOrchestrator.LaunchNeededAPI(context.Background())
}

func updateStrategyInGraph(strategyID int) {
	analytics.StOrchestrator.DependencyGraph.RemoveStrategy(strategyID)
	analytics.StOrchestrator.DependencyGraph.AddStrategy(strategyID, repositories.GetStrategyFormulasById(strategyID))
	analytics.StOrchestrator.LaunchNeededAPI(context.Background())
}

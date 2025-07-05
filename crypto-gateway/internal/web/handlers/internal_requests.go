package handlers

import (
	"context"
	"crypto-gateway/internal/analytics"
	"crypto-gateway/internal/web/db"
	"strconv"
)

// всю эту фигню после тестов убрать, это лишний слой, + запросы в бд можно не делать
func deleteFormulaFromGraph(formulaID int) {
	analytics.StOrchestrator.DependencyGraph.RemoveFormula(formulaID)
	analytics.StOrchestrator.LaunchNeededAPI(context.Background())
}

func addFormulaToGraph(formulaID int) {
	analytics.StOrchestrator.DependencyGraph.AddFormula(db.GetFormulaById(formulaID), formulaID)
	analytics.StOrchestrator.LaunchNeededAPI(context.Background())
}

func updateFormulaInGraph(formulaID string) {
	id, _ := strconv.Atoi(formulaID)
	analytics.StOrchestrator.DependencyGraph.RemoveFormula(id)
	analytics.StOrchestrator.DependencyGraph.AddFormula(db.GetFormulaById(id), id)
	analytics.StOrchestrator.LaunchNeededAPI(context.Background())
}

func updateApiCooldown(apiId int) {
	api, cooldown := db.GetApiAndCooldownByID(apiId)
	analytics.StOrchestrator.AdjustAPITaskCooldown(context.Background(), api, cooldown)
	analytics.StOrchestrator.LaunchNeededAPI(context.Background())
}

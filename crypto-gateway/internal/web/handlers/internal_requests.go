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

/*
func sendHTTPRequest(method, url string) string {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		log.Printf("Failed to create %s request: %v", method, err)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Failed to perform %s request: %v", method, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("Failed to read response body:", err)
	}

	return string(body)
}

func deleteFormulaFromGraph(formulaID string) {
	addr := fmt.Sprintf("%v/analytics/formula/%v/", config.Internal_Server_Api, formulaID)
	response := sendHTTPRequest(http.MethodDelete, addr)
	log.Println("Internal Response:", response)
}

func addFormulaToGraph(formulaID int) {
	addr := fmt.Sprintf("%v/analytics/formula/%v/", config.Internal_Server_Api, formulaID)
	response := sendHTTPRequest(http.MethodPost, addr)
	log.Println("Internal Response:", response)
}

func updateFormulaInGraph(formulaID string) {
	addr := fmt.Sprintf("%v/analytics/formula/%v/", config.Internal_Server_Api, formulaID)
	response := sendHTTPRequest(http.MethodPut, addr)
	log.Println("Internal Response:", response)
}

func updateApiCooldown(apiId int) {
	addr := fmt.Sprintf("%v/analytics/endpoint/%v/", config.Internal_Server_Api, apiId)
	response := sendHTTPRequest(http.MethodPut, addr)
	log.Println("Internal Response:", response)
}
*/

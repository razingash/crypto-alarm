package handlers

import (
	"crypto-gateway/config"
	"fmt"
	"io"
	"log"
	"net/http"
)

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

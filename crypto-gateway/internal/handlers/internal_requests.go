package handlers

import (
	"fmt"
	"io"
	"log"
	"net/http"
)

func sendHTTPRequest(method, url string) string {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		log.Fatalf("Failed to create %s request: %v", method, err)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Failed to perform %s request: %v", method, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("Failed to read response body:", err)
	}

	return string(body)
}

func deleteFormulaFromGraph(formulaID string) {
	addr := fmt.Sprintf("http://127.0.0.1:8000/api/v1/analytics/formula/%v/", formulaID)
	response := sendHTTPRequest(http.MethodDelete, addr)
	log.Println("Internal Response:", response)
}

func addFormulaToGraph(formulaID int) {
	addr := fmt.Sprintf("http://127.0.0.1:8000/api/v1/analytics/formula/%v/", formulaID)
	response := sendHTTPRequest(http.MethodPost, addr)
	log.Println("Internal Response:", response)
}

func updateFormulaInGraph(formulaID string) {
	addr := fmt.Sprintf("http://127.0.0.1:8000/api/v1/analytics/formula/%v/", formulaID)
	response := sendHTTPRequest(http.MethodPut, addr)
	log.Println("Internal Response:", response)
}

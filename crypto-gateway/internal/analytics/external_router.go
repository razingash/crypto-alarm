package analytics

import (
	"context"
	"crypto-gateway/internal/appmetrics"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

// структура для получения данных из бинанса путем запуска разных апи
type BinanceAPI struct {
	baseURL     string
	RecordedAPI []string
	httpClient  *http.Client // no need for httpx
	Controller  *BinanceAPIController
}

func NewBinanceAPI(controller *BinanceAPIController) *BinanceAPI {
	ctx := context.Background()
	endpoints, err := GetRecordedEndpoints(ctx)
	if err != nil {
		log.Printf("failed to load recorded endpoints: %v", err)
		endpoints = []string{}
	}
	fmt.Printf("Recorded endpoints: %v\n", endpoints)

	return &BinanceAPI{
		baseURL:     "https://api.binance.com/api",
		RecordedAPI: endpoints,
		httpClient:  &http.Client{Timeout: 10 * time.Second},
		Controller:  controller,
	}
}

// Method for adjusting the weights of endpoints. It isn't periodic task since the Binance weights are updated too
// often, it makes no sense to make a periodic renewal once every few minutes, given the low cost of the operation
func (api *BinanceAPI) checkAndUpdateEndpointWeights(resp *http.Response, endpoint string) (int, error) {
	// позже улучшить механизм обновления весов - сейчас в глобальном endpoints есть словарь из апи: весов
	// его можно использовать чтобы на старте получать самые последние значения из бд и заполнять его ими,
	// и вместо выполнения 3х запросов в бд сравнивать в начале с тем что имеется в endpoints

	usedWeightStr := resp.Header.Get("x-mbx-used-weight-1m")
	if usedWeightStr == "" {
		usedWeightStr = "0"
	}

	var usedWeight int
	_, err := fmt.Sscanf(usedWeightStr, "%d", &usedWeight)
	if err != nil {
		return 0, err
	}

	if api.RecordedAPI != nil {
		var deltaWeight int
		prevWeight := api.Controller.CurrentWeight
		if usedWeight < prevWeight { // вес сбросился со стороны binance
			deltaWeight = usedWeight
		} else {
			deltaWeight = usedWeight - prevWeight
		}

		if deltaWeight > 0 {
			recorded := make(map[string]struct{}, len(api.RecordedAPI))
			for _, e := range api.RecordedAPI {
				recorded[e] = struct{}{}
			}
			if _, ok := recorded[endpoint]; ok {
				ctx := context.Background()
				if err := SaveEndpointWeightIfChanged(ctx, endpoint, deltaWeight); err != nil {
					log.Printf("failed to save weight: %v", err)
				}
			}
		}
	}

	api.Controller.Mu.Lock()
	api.Controller.CurrentWeight = usedWeight
	api.Controller.Mu.Unlock()

	return usedWeight, nil
}

// добавить отлов 500 чтобы устанавливать бибанс на неактивность
func (api *BinanceAPI) Get(ctx context.Context, endpoint string, endpointExpectedWeight int, params map[string]string) ([]byte, error) {
	// добавить обработку остальных ошибок
	var responseBody []byte

	requestFunc := func() error {
		url := api.baseURL + endpoint
		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			return err
		}

		q := req.URL.Query()
		for k, v := range params {
			q.Add(k, v)
		}
		req.URL.RawQuery = q.Encode()

		resp, err := api.httpClient.Do(req)
		if err != nil {
			log.Printf("HTTP request failed: %v", err)
			return err
		}
		defer resp.Body.Close()

		responseBody, err = io.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		if resp.StatusCode != http.StatusOK {
			appmetrics.DefaultLogging(2, fmt.Sprintf("BINANCE ERROR: status %d, body: %s", resp.StatusCode, string(responseBody)))
			log.Printf("BINANCE ERROR: status %d, body: %s", resp.StatusCode, string(responseBody))
			return fmt.Errorf("bad status code: %d", resp.StatusCode)
		}

		_, err = api.checkAndUpdateEndpointWeights(resp, endpoint)
		if err != nil {
			log.Printf("Error parsing weights header: %v", err)
		}

		return nil
	}

	err := api.Controller.RequestWithLimit(endpointExpectedWeight, requestFunc)
	return responseBody, err
}

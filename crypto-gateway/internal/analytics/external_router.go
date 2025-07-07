package analytics

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

// структура для получения данных из бинанса путем запуска разных апи
type BinanceAPI struct {
	baseURL    string
	httpClient *http.Client // no need for httpx
	controller *BinanceAPIController
}

func NewBinanceAPI(controller *BinanceAPIController) *BinanceAPI {
	return &BinanceAPI{
		baseURL:    "https://api.binance.com/api",
		httpClient: &http.Client{Timeout: 10 * time.Second},
		controller: controller,
	}
}

// Method for adjusting the weights of endpoints. It isn't periodic task since the Binance weights are updated too
// often, it makes no sense to make a periodic renewal once every few minutes, given the low cost of the operation
func (api *BinanceAPI) checkAndUpdateEndpointWeights(resp *http.Response, endpoint string) (int, error) {
	// 1) ВАЖНО! потом добавить сбор метрик чтобы можно было также получить график нагрузки на эндпоинты
	// 2) здесь можно сравнить с ожидаемой нагрузкой и залогировать если нужно, чтобы была инфа о нагрузках по разным дням/неделям.
	// - - но зачем это надо? можно сделать дополнительной фичей которая будет по желанию активироватся
	usedWeightStr := resp.Header.Get("x-mbx-used-weight-1m")
	if usedWeightStr == "" {
		usedWeightStr = "0"
	}

	var usedWeight int
	_, err := fmt.Sscanf(usedWeightStr, "%d", &usedWeight)
	if err != nil {
		return 0, err
	}

	api.controller.mu.Lock()
	api.controller.currentWeight = usedWeight
	api.controller.mu.Unlock()

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
			DefaultLogging(2, fmt.Sprintf("BINANCE ERROR: status %d, body: %s", resp.StatusCode, string(responseBody)))
			log.Printf("BINANCE ERROR: status %d, body: %s", resp.StatusCode, string(responseBody))
			return fmt.Errorf("bad status code: %d", resp.StatusCode)
		}

		_, err = api.checkAndUpdateEndpointWeights(resp, endpoint)
		if err != nil {
			log.Printf("Error parsing weights header: %v", err)
		}

		return nil
	}

	err := api.controller.RequestWithLimit(endpointExpectedWeight, requestFunc)
	return responseBody, err
}

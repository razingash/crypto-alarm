package analytics

import (
	"context"
	"crypto-gateway/internal/web/db"
	"encoding/json"
	"fmt"
	"strconv"
	"sync"
	"time"
)

// отвечает за запуск периодических задач
type BinanceAPIOrchestrator struct {
	DependencyGraph *DependencyGraph
	binanceAPI      *BinanceAPI
	isBinanceOnline bool

	mu            sync.Mutex
	tasks         map[string]context.CancelFunc
	taskCooldowns map[string]int
}

func NewBinanceAPIOrchestrator(api *BinanceAPI) *BinanceAPIOrchestrator {
	return &BinanceAPIOrchestrator{
		DependencyGraph: NewDependencyGraph(),
		binanceAPI:      api,
		isBinanceOnline: true,
		tasks:           make(map[string]context.CancelFunc),
		taskCooldowns:   make(map[string]int),
	}
}

// Запуск фоновых задач. первая задача должна быть проверка доступности апи бинанса
func (o *BinanceAPIOrchestrator) Start(ctx context.Context) error {
	response, err := o.binanceAPI.Get(context.Background(), "/v3/ping", endpoints["/v3/ping"], nil) // get Binance accessibility
	if err != nil {
		o.checkBinanceResponse(nil)
	} else {
		o.checkBinanceResponse(response)
	}

	if o.isBinanceOnline {
		o.LaunchNeededAPI(context.Background())
	}

	return nil
}

// выбирает апи, которые нужны для получения актуальных данных в граф, и убирает ненужные
func (o *BinanceAPIOrchestrator) LaunchNeededAPI(ctx context.Context) {
	fmt.Println("Launching needed API tasks...")

	data, err := GetActualComponents(ctx)
	fmt.Println(data)
	if err != nil {
		panic("GetActualComponents returned error in LaunchNeededAPI function")
	}

	o.mu.Lock()
	defer o.mu.Unlock()

	currentAPIs := make(map[string]struct{})
	for k := range data {
		currentAPIs[k] = struct{}{}
	}

	runningAPIs := make(map[string]struct{})
	for k := range o.tasks {
		if k != "weights" {
			runningAPIs[k] = struct{}{}
		}
	}

	// остановка устаревших задач
	for api := range runningAPIs {
		if _, ok := currentAPIs[api]; !ok {
			fmt.Printf("Stopping outdated API: %s\n", api)
			cancelFunc := o.tasks[api]
			cancelFunc()
			delete(o.tasks, api)
			delete(o.taskCooldowns, api)
		}
	}

	// запуск новых, в следствии обновления периодичности или создания формул которым нужны данные других апи
	for api, info := range data {
		if info.Count <= 0 {
			continue
		}

		if _, exists := o.tasks[api]; exists {
			continue
		}

		fmt.Printf("Starting API task: %s (cooldown: %d seconds)\n", api, info.Cooldown)
		o.launchAPI(ctx, api, info.Cooldown)
	}
}

// Updates the frequency of proxification for a particular API, if necessary
func (o *BinanceAPIOrchestrator) AdjustAPITaskCooldown(ctx context.Context, api string, newCooldown int) {
	o.mu.Lock()
	defer o.mu.Unlock()

	oldCooldown, exists := o.taskCooldowns[api]
	if !exists || oldCooldown == newCooldown {
		return
	}

	cancelFunc, exists := o.tasks[api]
	if exists {
		cancelFunc()
		delete(o.tasks, api)
		delete(o.taskCooldowns, api)
	}

	o.launchAPI(ctx, api, newCooldown)
}

func (o *BinanceAPIOrchestrator) launchAPI(ctx context.Context, api string, cooldown int) {
	cctx, cancel := context.WithCancel(ctx)
	o.tasks[api] = cancel
	o.taskCooldowns[api] = cooldown

	switch api {
	case "/v3/ticker/price":
		go o.updateTickerCurrentPrice(cctx, cooldown)
	case "/v3/ticker/24hr":
		go o.updatePriceChange24h(cctx, cooldown)
	case "/v3/ping":
		go o.checkAPIStatusLoop(cctx, cooldown)
	default:
		fmt.Printf("Unknown API task: %s\n", api)
	}
}

// checks Binacne availability
func (o *BinanceAPIOrchestrator) checkAPIStatusLoop(ctx context.Context, cooldown int) {
	ticker := time.NewTicker(time.Duration(cooldown) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			response, err := o.binanceAPI.Get(ctx, "/v3/ping", endpoints["/v3/ping"], nil) // get Binance accessibility
			if err != nil {
				o.checkBinanceResponse(nil)
			} else {
				o.checkBinanceResponse(response)
			}
			if o.isBinanceOnline {
				return
			}
		}
	}
}

func (o *BinanceAPIOrchestrator) updateTickerCurrentPrice(ctx context.Context, cooldown int) {
	ticker := time.NewTicker(time.Duration(cooldown) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			fmt.Println("updateTickerCurrentPrice running...")
			response, err := o.binanceAPI.Get(ctx, "/v3/ticker/price", endpoints["/v3/ticker/price"], nil) // get ticker price
			if err != nil {
				o.checkBinanceResponse(nil)
				continue
			}
			o.checkBinanceResponse(response)

			currencies, err := GetNeededFieldsFromEndpoint(ctx, "/v3/ticker/price")
			if err != nil {
				panic(fmt.Errorf("in updateTickerCurrentPrice - GetNeededFieldsFromEndpoint returned Error %v", err))
			}
			dataForGraph := extractDataFromTickerCurrentPrice(response, currencies)

			triggeredFormulas := o.DependencyGraph.UpdateVariablesTopologicalKahn(dataForGraph)
			if len(triggeredFormulas) > 0 {
				result, variableValues := o.DependencyGraph.GetFormulasVariables(triggeredFormulas)
				AddTriggerHistory(ctx, result, variableValues)
				db.SendPushNotifications(triggeredFormulas, "TEST MESSAGE IN ORCHESTRATOR")
			}
		}
	}
}

func (o *BinanceAPIOrchestrator) updatePriceChange24h(ctx context.Context, cooldown int) {
	ticker := time.NewTicker(time.Duration(cooldown) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			fmt.Println("updatePriceChange24h running...")
			response, err := o.binanceAPI.Get(ctx, "/v3/ticker/24hr", endpoints["/v3/ticker/24hr"], nil) // get price change
			if err != nil {
				o.checkBinanceResponse(nil)
				continue
			}
			o.checkBinanceResponse(response)

			fields, err := GetNeededFieldsFromEndpoint(ctx, "/v3/ticker/24hr")
			if err != nil {
				panic(fmt.Errorf("in updatePriceChange24h - GetNeededFieldsFromEndpoint returned Error %v", err))
			}
			dataForGraph := extractDataFromPriceChange24h(response, fields)

			triggeredFormulas := o.DependencyGraph.UpdateVariablesTopologicalKahn(dataForGraph)
			if len(triggeredFormulas) > 0 {
				result, variableValues := o.DependencyGraph.GetFormulasVariables(triggeredFormulas)
				AddTriggerHistory(ctx, result, variableValues)
				db.SendPushNotifications(triggeredFormulas, "TEST MESSAGE IN ORCHESTRATOR")
			}
		}
	}
}

// проверяет ответы от Binance, попутно вявляя ошибки с его стороны
func (o *BinanceAPIOrchestrator) checkBinanceResponse(response interface{}) {
	if response == nil { // Binance DOWN
		if o.isBinanceOnline {
			DefaultLogging(1, "Binance DOWN")
			o.isBinanceOnline = false
		}

		o.mu.Lock()
		defer o.mu.Unlock()

		// отмена всех задач кроме пинга
		for api, cancelFunc := range o.tasks {
			if api != "/v3/ping" {
				cancelFunc()
				delete(o.tasks, api)
				delete(o.taskCooldowns, api)
			}
		}

		// запуск пинга если его нет(не должно быть)
		if _, exists := o.tasks["/v3/ping"]; !exists {
			o.launchAPI(context.Background(), "/v3/ping", 60)
		}

	} else if !o.isBinanceOnline { // Binance UP
		DefaultLogging(1, "Binance UP")
		o.isBinanceOnline = true

		o.mu.Lock()
		defer o.mu.Unlock()

		// отмена всех задач
		for _, cancelFunc := range o.tasks {
			cancelFunc()
		}
		o.tasks = make(map[string]context.CancelFunc)
		o.taskCooldowns = make(map[string]int)

		// запуск нужные задач заново
		o.LaunchNeededAPI(context.Background())
	}
}

func extractDataFromPriceChange24h(rawData []byte, fields map[string][]string) map[string]float64 {
	var dataset []map[string]interface{}
	err := json.Unmarshal(rawData, &dataset)
	if err != nil {
		panic("ERROR IN extractDataFromPriceChange24h")
	}

	datasetDict := make(map[string]map[string]interface{})
	for _, data := range dataset {
		if symbol, ok := data["symbol"].(string); ok {
			datasetDict[symbol] = data
		}
	}

	result := make(map[string]float64)
	for symbol, fieldList := range fields {
		if data, exists := datasetDict[symbol]; exists {
			for _, field := range fieldList {
				if val, ok := data[field]; ok {
					switch v := val.(type) {
					case float64:
						result[symbol+"_"+field] = v
					case float32:
						result[symbol+"_"+field] = float64(v)
					case int:
						result[symbol+"_"+field] = float64(v)
					case int64:
						result[symbol+"_"+field] = float64(v)
					case string:
						if parsedVal, err := strconv.ParseFloat(v, 64); err == nil {
							result[symbol+"_"+field] = parsedVal
						}
					}
				}
			}
		}
	}
	return result
}

func extractDataFromTickerCurrentPrice(rawData []byte, currencies map[string][]string) map[string]float64 {
	var dataset []map[string]interface{}
	err := json.Unmarshal(rawData, &dataset)
	if err != nil {
		panic("ERROR IN extractDataFromPriceChange24h")
	}
	datasetDict := make(map[string]map[string]interface{})
	for _, data := range dataset {
		if symbol, ok := data["symbol"].(string); ok {
			datasetDict[symbol] = data
		}
	}

	result := make(map[string]float64)
	for symbol := range currencies {
		if data, exists := datasetDict[symbol]; exists {
			if priceVal, ok := data["price"]; ok {
				switch v := priceVal.(type) {
				case float64:
					result[symbol+"_price"] = v
				case float32:
					result[symbol+"_price"] = float64(v)
				case int:
					result[symbol+"_price"] = float64(v)
				case int64:
					result[symbol+"_price"] = float64(v)
				case string:
					if parsedVal, err := strconv.ParseFloat(v, 64); err == nil {
						result[symbol+"_price"] = parsedVal
					}
				}
			}
		}
	}
	return result
}

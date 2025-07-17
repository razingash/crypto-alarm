package analytics

import "crypto-gateway/internal/appmetrics"

var ( // изменить этот позор на адекватную систему
	StController   *BinanceAPIController
	StBinanceApi   *BinanceAPI
	StOrchestrator *BinanceAPIOrchestrator
	Collector      *appmetrics.LoadMetricsCollector
	StartTime      int64
	endpoints      map[string]int
)

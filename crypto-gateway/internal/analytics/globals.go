package analytics

import (
	"crypto-gateway/internal/appmetrics"
)

// позже попробовать перенести инициализацию глобальных переменных в отдельный пакет setup(но возможно со всем не получится из-за зависимостей)
var (
	StController       *BinanceAPIController
	StBinanceApi       *BinanceAPI
	StOrchestrator     *BinanceAPIOrchestrator
	Collector          *appmetrics.LoadMetricsCollector
	AverageLoadMetrics *AverageLoadMetricsManager
	StartTime          int64
	endpoints          map[string]int
)

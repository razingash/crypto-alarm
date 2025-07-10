package analytics

var ( // изменить этот позор на адекватную систему
	StController   *BinanceAPIController
	StBinanceApi   *BinanceAPI
	StOrchestrator *BinanceAPIOrchestrator
	StartTime      int64
	endpoints      map[string]int
)

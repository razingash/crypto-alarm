package handlers

import (
	"bufio"
	"context"
	"crypto-gateway/internal/analytics"
	"crypto-gateway/internal/appmetrics"
	"crypto-gateway/internal/web/db"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"runtime"
	"time"

	"github.com/fasthttp/websocket"
	"github.com/gofiber/fiber/v3"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/process"
)

type RuntimePayload struct {
	Metrics   RuntimeMetrics            `json:"metrics"`
	LoadAvg60 []appmetrics.LoadAverages `json:"load_avg_60"`
}

type RuntimeMetrics struct {
	MemAllocMB         uint64  `json:"mem_alloc_mb"`
	MemSysMB           uint64  `json:"mem_sys_mb"`
	RAMUsedMB          uint64  `json:"ram_used_mb"`
	MemoryUsagePercent float64 `json:"memory_usage_percent"`
	CPUUsedPercent     float64 `json:"cpu_used_percent"`
	CPUUsagePercent    float64 `json:"cpu_usage_percent"`
	CPUAllocation      float64 `json:"cpu_allocation"`
	NumGC              uint32  `json:"num_gc"`
	BinanceOverload    int     `json:"binance_overload"`
}
type AvailabilityMetric struct {
	Timestamp   string `json:"timestamp"`
	Level       string `json:"level"`
	Caller      string `json:"caller"`
	Message     string `json:"message"`
	Type        int    `json:"type"`
	Event       string `json:"event"`
	IsAvailable int    `json:"isAvailable"`
}

// только message и error
type DefaultErrors struct {
	Timestamp string `json:"timestamp"`
	Level     string `json:"level"`
	Caller    string `json:"caller"`
	Message   string `json:"message"`
	Error     string `json:"error"`
}

func GetAvailabilityMetrics(c fiber.Ctx) error {
	file, err := os.Open("logs/AvailabilityMetrics.log")
	if err != nil {
		if os.IsNotExist(err) {
			return c.Status(fiber.StatusOK).JSON(fiber.Map{
				"data": []interface{}{},
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "unable to read log file",
		})
	}
	defer file.Close()

	var metrics []AvailabilityMetric
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		var entry AvailabilityMetric
		if err := json.Unmarshal([]byte(line), &entry); err == nil {
			metrics = append(metrics, entry)
		}
	}

	if err := scanner.Err(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "error reading log file",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"data": metrics,
	})
}

// возвращает детальную информацию о логах
func GetErrorsDetailedInfo(c fiber.Ctx) error {
	logType := c.Query("type")
	var filename string = "logs/AnalyticsServiceErrors.log"
	switch logType {
	case "application":
		filename = "logs/ApplicationErrors.log"
	case "binance":
		filename = "logs/BinanceErrors.log"
	}
	file, err := os.Open(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return c.Status(fiber.StatusOK).JSON(fiber.Map{
				"data": []interface{}{},
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "unable to read log file",
		})
	}
	defer file.Close()

	var metrics []DefaultErrors
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		var entry DefaultErrors
		if err := json.Unmarshal([]byte(line), &entry); err == nil {
			metrics = append(metrics, entry)
		}
	}

	if err := scanner.Err(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "error reading log file",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"data": metrics,
	})
}

func GetErrorsBasicInfo(c fiber.Ctx) error {
	files := []struct {
		Type     string
		Filename string
	}{
		{"analytics", "logs/AnalyticsServiceErrors.log"},
		{"application", "logs/ApplicationErrors.log"},
		{"binance", "logs/BinanceErrors.log"},
	}

	type FileInfo struct {
		Type  string `json:"type"`
		Lines int    `json:"lines"`
	}

	var result []FileInfo

	for _, file := range files {
		lines := 0
		f, err := os.Open(file.Filename)
		if err == nil {
			scanner := bufio.NewScanner(f)
			for scanner.Scan() {
				lines++
			}
			f.Close()
		}
		result = append(result, FileInfo{
			Type:  file.Type,
			Lines: lines,
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"data": result,
	})
}

func GetStaticMetrics(c fiber.Ctx) error {
	type StaticMetrics struct {
		TotalCPU         int    `json:"total_cpu"`
		UsedCPU          int    `json:"used_cpu"` // процессоры, которые использует программа (GOMAXPROCS)
		TotalMemoryMB    uint64 `json:"total_memory_mb"`
		MaxBinanceWeight int    `json:"max_binance_weight"`
		StartTime        int64  `json:"start_time"`
	}
	vmStat, err := mem.VirtualMemory()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("failed to get memory info")
	}

	info := StaticMetrics{
		TotalCPU:         runtime.NumCPU(),
		UsedCPU:          runtime.GOMAXPROCS(0),
		TotalMemoryMB:    vmStat.Total / 1024 / 1024,
		MaxBinanceWeight: analytics.StController.MaxWeight,
		StartTime:        analytics.StartTime,
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"data": info,
	})
}

// отправляет метрики и нагрузке на OS приложением
func SendRuntimeMetricsWS(conn *websocket.Conn) {
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)

		proc, _ := process.NewProcess(int32(os.Getpid()))

		memStats, _ := mem.VirtualMemory()
		ramUsed, _ := proc.MemoryInfo()
		cpuPercent, _ := proc.CPUPercent()
		totalCPUPercent, _ := cpu.Percent(0, false)
		cpuCores, _ := cpu.Counts(true)

		metrics := RuntimeMetrics{
			MemAllocMB:         m.Alloc / 1024 / 1024,
			MemSysMB:           m.Sys / 1024 / 1024,
			RAMUsedMB:          ramUsed.RSS / 1024 / 1024,
			MemoryUsagePercent: memStats.UsedPercent,
			CPUUsedPercent:     cpuPercent,
			CPUUsagePercent:    totalCPUPercent[0],
			CPUAllocation:      (cpuPercent / 100.0) * float64(cpuCores),
			// NumGC:              m.NumGC, // циклы сборщика мусора, позже добавить
			BinanceOverload: analytics.StController.CurrentWeight,
		}

		analytics.Collector.CollectWindowsCPU(cpuPercent) // отрабатывает только для винды
		analytics.Collector.Collect()
		payload := RuntimePayload{
			Metrics:   metrics,
			LoadAvg60: analytics.Collector.Values(),
		}

		data, err := json.Marshal(payload)
		if err != nil {
			log.Println("Marshal error:", err)
			return
		}

		if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
			log.Println("WriteMessage error:", err)
			return
		}
	}
}

func GetBinanceApiWeightMetrics(c fiber.Ctx) error {
	type HistoryEntry struct {
		Timestamp time.Time `json:"created_at"`
		Weight    int       `json:"weight"`
	}

	type ApiMetrics struct {
		Endpoint string         `json:"endpoint"`
		Weights  []HistoryEntry `json:"weights"`
	}

	apisRows, err := db.DB.Query(context.Background(), `
		SELECT id, api FROM crypto_api
		WHERE is_actual = true AND is_history_on = true
	`)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, fmt.Sprintf("failed to query crypto_api: %v", err))
	}
	defer apisRows.Close()

	apiMap := make(map[int]string)
	for apisRows.Next() {
		var id int
		var api string
		if err := apisRows.Scan(&id, &api); err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, fmt.Sprintf("failed to scan api row: %v", err))
		}
		apiMap[id] = api
	}

	historyRows, err := db.DB.Query(context.Background(), `
		SELECT crypto_api_id, weight, created_at
		FROM crypto_api_history
		WHERE crypto_api_id = ANY($1)
		ORDER BY created_at
	`, keys(apiMap))
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, fmt.Sprintf("failed to query history: %v", err))
	}
	defer historyRows.Close()

	historyMap := make(map[int][]HistoryEntry)
	for historyRows.Next() {
		var apiID, weight int
		var createdAt time.Time
		if err := historyRows.Scan(&apiID, &weight, &createdAt); err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, fmt.Sprintf("failed to scan history row: %v", err))
		}
		historyMap[apiID] = append(historyMap[apiID], HistoryEntry{
			Timestamp: createdAt,
			Weight:    weight,
		})
	}

	var result []ApiMetrics
	for id, endpoint := range apiMap {
		weights := historyMap[id]
		if weights == nil {
			weights = []HistoryEntry{}
		}
		result = append(result, ApiMetrics{
			Endpoint: endpoint,
			Weights:  weights,
		})
	}
	return c.Status(fiber.StatusOK).JSON(result)
}

func keys(m map[int]string) []int {
	ids := make([]int, 0, len(m))
	for id := range m {
		ids = append(ids, id)
	}
	return ids
}

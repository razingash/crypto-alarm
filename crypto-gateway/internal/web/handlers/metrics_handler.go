package handlers

import (
	"bufio"
	"crypto-gateway/internal/analytics"
	"encoding/json"
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

type RuntimeMetrics struct {
	MemAllocMB         uint64  `json:"mem_alloc_mb"`
	MemSysMB           uint64  `json:"mem_sys_mb"`
	RAMUsedMB          uint64  `json:"ram_used_mb"`
	MemoryUsagePercent float64 `json:"memory_usage_percent"`
	CPUUsedPercent     float64 `json:"cpu_used_percent"`
	CPUUsagePercent    float64 `json:"cpu_usage_percent"`
	CPUAllocation      float64 `json:"cpu_allocation"`
	NumGC              uint32  `json:"num_gc"`
	BinanceOverload    int     `json:"binanceOverload"`
}

func GetAvailabilityMetrics(c fiber.Ctx) error {
	type AvailabilityMetric struct {
		Timestamp   string `json:"timestamp"`
		Level       string `json:"level"`
		Caller      string `json:"caller"`
		Message     string `json:"message"`
		Type        int    `json:"type"`
		Event       string `json:"event"`
		IsAvailable int    `json:"isAvailable"`
	}

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

// получает информацию о критических ошибках(эта информация необходима для разработчиков)
func GetErrorsInfo(c fiber.Ctx) error {
	type ApplicationCriticalErrors struct {
		Timestamp string `json:"timestamp"`
		Level     string `json:"level"`
		Caller    string `json:"caller"`
		Message   string `json:"message"`
		Event     string `json:"event"`
		Error     string `json:"error"`
	}

	file, err := os.Open("logs/ApplicationCriticalErrors.log")
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

	var metrics []ApplicationCriticalErrors
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		var entry ApplicationCriticalErrors
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

		data, err := json.Marshal(metrics)
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

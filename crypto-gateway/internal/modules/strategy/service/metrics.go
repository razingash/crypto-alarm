package service

import (
	"context"
	"crypto-gateway/internal/appmetrics"
	"crypto-gateway/internal/web/repositories"
	"log"
	"sync"
	"time"
)

type AverageLoadMetricsManager struct {
	mu        sync.Mutex
	IsOn      bool
	cancel    context.CancelFunc
	collector *appmetrics.LoadMetricsCollector
}

func NewAverageLoadMetricsManager(collector *appmetrics.LoadMetricsCollector) *AverageLoadMetricsManager {
	return &AverageLoadMetricsManager{
		collector: collector,
	}
}

// now only metrics
func SetupInitialSettings(ctx context.Context) {
	settings, err := repositories.GetInitialLoadSystem(ctx)
	if err != nil {
		appmetrics.ApplicationErrorsLogging(3, "error while receiving settings from DB", err)
	}
	for _, setting := range settings {
		if setting.Name == "Average System Load" {
			AverageLoadMetrics.ToggleAverageLoadMetrics(setting.IsActive)
		}
	}
}

func (m *AverageLoadMetricsManager) ToggleAverageLoadMetrics(start bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if start {
		if m.IsOn {
			return
		}
		ctx, cancel := context.WithCancel(context.Background())
		m.cancel = cancel
		m.IsOn = true
		go appmetrics.StartLoadMetricsCollector(ctx, m.collector, 5*time.Second)
		log.Println("Metrics collector started")
	} else {
		if !m.IsOn {
			return
		}
		m.cancel()
		m.IsOn = false
		log.Println("Metrics collector stopped")
	}
}

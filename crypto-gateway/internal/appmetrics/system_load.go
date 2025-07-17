package appmetrics

import (
	"math"
	"runtime"
	"sync"

	"github.com/shirou/gopsutil/load"
)

// тут получаются метрики Average(system) load для 1m/5m/15m, стоит ли оно того...

// unix
type LoadAverages struct {
	Load1  float64 `json:"1m"`
	Load5  float64 `json:"5m"`
	Load15 float64 `json:"15m"`
}

type LoadAvgRingBuffer struct {
	data   []LoadAverages
	size   int
	cursor int
	filled bool
	mu     sync.Mutex
}

type LoadMetricsCollector struct {
	Buffer         *LoadAvgRingBuffer
	CollectLoadAvg bool
	windowsLoadAvg *WindowsLoadAvg // windows
	mu             sync.Mutex
}

// windows
type WindowsLoadAvg struct {
	Load    LoadAverages
	alpha1  float64
	alpha5  float64
	alpha15 float64
}

func NewWindowsLoadAvg(sampleIntervalSeconds float64) *WindowsLoadAvg {
	return &WindowsLoadAvg{
		alpha1:  math.Exp(-sampleIntervalSeconds / 60.0),
		alpha5:  math.Exp(-sampleIntervalSeconds / 300.0),
		alpha15: math.Exp(-sampleIntervalSeconds / 900.0),
	}
}

func (w *WindowsLoadAvg) Update(cpuPercent float64) {
	load := cpuPercent / 100.0
	w.Load.Load1 = w.Load.Load1*w.alpha1 + load*(1-w.alpha1)
	w.Load.Load5 = w.Load.Load5*w.alpha5 + load*(1-w.alpha5)
	w.Load.Load15 = w.Load.Load15*w.alpha15 + load*(1-w.alpha15)
}

func (w *WindowsLoadAvg) Current() LoadAverages {
	return w.Load
}

// unix
func NewLoadAvgRingBuffer(size int) *LoadAvgRingBuffer {
	return &LoadAvgRingBuffer{
		data: make([]LoadAverages, size),
		size: size,
	}
}

func NewLoadMetricsCollector(size int) *LoadMetricsCollector {
	c := &LoadMetricsCollector{
		Buffer:         NewLoadAvgRingBuffer(size),
		CollectLoadAvg: true,
	}
	if runtime.GOOS == "windows" {
		c.windowsLoadAvg = NewWindowsLoadAvg(5.0) // периодичность опросов
	}
	return c
}

// unix
func (b *LoadAvgRingBuffer) Add(value LoadAverages) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.data[b.cursor] = value
	b.cursor = (b.cursor + 1) % b.size
	if b.cursor == 0 {
		b.filled = true
	}
}

func (b *LoadAvgRingBuffer) Values() []LoadAverages {
	b.mu.Lock()
	defer b.mu.Unlock()
	if !b.filled {
		return b.data[:b.cursor]
	}
	result := make([]LoadAverages, b.size)
	copy(result, b.data[b.cursor:])
	copy(result[b.size-b.cursor:], b.data[:b.cursor])
	return result
}

func (b *LoadAvgRingBuffer) Reset() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.data = make([]LoadAverages, b.size)
	b.cursor = 0
	b.filled = false
}

func (c *LoadMetricsCollector) SwitchCollect(collect bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !collect && c.CollectLoadAvg {
		c.Buffer.Reset()
	}
	c.CollectLoadAvg = collect
}

func (c *LoadMetricsCollector) Collect() {
	c.mu.Lock()
	if !c.CollectLoadAvg {
		c.mu.Unlock()
		return
	}
	c.mu.Unlock()

	var triple LoadAverages
	if runtime.GOOS == "windows" {
		c.mu.Lock()
		avg := c.windowsLoadAvg.Current()
		c.mu.Unlock()
		triple = LoadAverages{
			Load1:  avg.Load1,
			Load5:  avg.Load5,
			Load15: avg.Load15,
		}
	} else {
		avg, err := load.Avg()
		if err == nil {
			triple = LoadAverages{Load1: avg.Load1, Load5: avg.Load5, Load15: avg.Load15}
		}
	}
	c.Buffer.Add(triple)
}

func (c *LoadMetricsCollector) Values() []LoadAverages {
	return c.Buffer.Values()
}

// windows and unix
func (c *LoadMetricsCollector) CollectWindowsCPU(cpuPercent float64) {
	if runtime.GOOS != "windows" {
		return
	}
	c.mu.Lock()
	c.windowsLoadAvg.Update(cpuPercent)
	c.mu.Unlock()
}

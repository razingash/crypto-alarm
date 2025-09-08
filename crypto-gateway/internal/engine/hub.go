package engine

import (
	"sync"
	"time"
)

type Hub struct {
	Mu             sync.Mutex
	Channels       map[string]map[string]*Channel // workflow -> channel name -> channel
	metrics        map[string]*WorkflowMetrics    // workflow -> metrics
	sampleInterval time.Duration
	stopSampler    chan struct{}
}

type ModuleMetrics struct {
	MsgCount     int64         `json:"msg_count"`
	AvgLatency   time.Duration `json:"avg_latency"`
	TotalLatency time.Duration `json:"-"` // для пересчета avg
	QueueLength  int           `json:"queue_length"`
}

type WorkflowMetrics struct {
	Modules map[string]*ModuleMetrics `json:"modules"`
}

func NewHub(sampleInterval time.Duration) *Hub {
	h := &Hub{
		Channels:       make(map[string]map[string]*Channel),
		metrics:        make(map[string]*WorkflowMetrics),
		sampleInterval: sampleInterval,
		stopSampler:    make(chan struct{}),
	}
	go h.sampler()
	return h
}

func (h *Hub) Close() {
	close(h.stopSampler)
}

// CreateChannel creates or returns existing channel for workflow
func (h *Hub) CreateChannel(workflow, name string, buffer int) *Channel {
	h.Mu.Lock()
	defer h.Mu.Unlock()
	if _, ok := h.Channels[workflow]; !ok {
		h.Channels[workflow] = map[string]*Channel{}
	}
	if ch, ok := h.Channels[workflow][name]; ok {
		return ch
	}
	ch := NewChannel(workflow, name, buffer)
	h.Channels[workflow][name] = ch
	return ch
}

func (h *Hub) GetChannel(workflow, name string) (*Channel, bool) {
	h.Mu.Lock()
	defer h.Mu.Unlock()
	m, ok := h.Channels[workflow]
	if !ok {
		return nil, false
	}
	ch, ok := m[name]
	return ch, ok
}

func (h *Hub) RemoveWorkflowChannels(workflow string) {
	h.Mu.Lock()
	defer h.Mu.Unlock()
	if chans, ok := h.Channels[workflow]; ok {
		for _, c := range chans {
			c.Close()
		}
		delete(h.Channels, workflow)
	}
}

// Connect sets out channel on src module and in channel on dst module
// If channel with this name doesn't exist - create with buffer
func (h *Hub) Connect(workflow string, channelName string, buffer int, src Module, srcOut string, dst Module, dstIn string) {
	ch := h.CreateChannel(workflow, channelName, buffer)
	src.SetOutput(srcOut, ch)
	dst.SetInput(dstIn, ch)
}

// увеличивает счетчик сообщений
func (h *Hub) IncProcessed(workflow, module string) {
	h.RecordMessage(workflow, module, 0)
}

// наблюдает latency
func (h *Hub) ObserveLatency(workflow, module string, dur time.Duration) {
	h.RecordMessage(workflow, module, dur)
}

// устанавливает длину очереди
func (h *Hub) SetQueueLen(workflow, channelName string, l int) {
	h.UpdateQueueLength(workflow, channelName, l)
}

func (h *Hub) RecordMessage(workflow, module string, latency time.Duration) {
	h.Mu.Lock()
	defer h.Mu.Unlock()

	wf, ok := h.metrics[workflow]
	if !ok {
		wf = &WorkflowMetrics{Modules: map[string]*ModuleMetrics{}}
		h.metrics[workflow] = wf
	}

	m, ok := wf.Modules[module]
	if !ok {
		m = &ModuleMetrics{}
		wf.Modules[module] = m
	}

	m.MsgCount++
	m.TotalLatency += latency
	m.AvgLatency = m.TotalLatency / time.Duration(m.MsgCount)
}

func (h *Hub) UpdateQueueLength(workflow, channel string, length int) {
	h.Mu.Lock()
	defer h.Mu.Unlock()

	wf, ok := h.metrics[workflow]
	if !ok {
		wf = &WorkflowMetrics{Modules: map[string]*ModuleMetrics{}}
		h.metrics[workflow] = wf
	}

	m, ok := wf.Modules[channel]
	if !ok {
		m = &ModuleMetrics{}
		wf.Modules[channel] = m
	}

	m.QueueLength = length
}

// periodic sampler of queue lengths
func (h *Hub) sampler() {
	t := time.NewTicker(h.sampleInterval)
	defer t.Stop()
	for {
		select {
		case <-t.C:
			h.Mu.Lock()
			for wf, chans := range h.Channels {
				for name, ch := range chans {
					l := len(ch.Ch)
					h.UpdateQueueLength(wf, name, l)
				}
			}
			h.Mu.Unlock()
		case <-h.stopSampler:
			return
		}
	}
}

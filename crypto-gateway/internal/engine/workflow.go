package engine

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

type Workflow struct {
	ID      string
	Modules []Module
	Hub     *Hub

	runMu    sync.Mutex
	running  bool
	cancel   context.CancelFunc
	runWg    sync.WaitGroup
	started  time.Time
	shutdown chan struct{}
}

// модули добавляются потом
func NewWorkflow(id string, hub *Hub) *Workflow {
	return &Workflow{
		ID:       id,
		Hub:      hub,
		Modules:  make([]Module, 0),
		shutdown: make(chan struct{}),
	}
}

func (wf *Workflow) AddModule(m Module) {
	wf.Modules = append(wf.Modules, m)
}

func (wf *Workflow) Start(parentCtx context.Context) error {
	wf.runMu.Lock()
	defer wf.runMu.Unlock()
	if wf.running {
		return errors.New("workflow already running")
	}

	ctx, cancel := context.WithCancel(parentCtx)
	wf.cancel = cancel
	wf.runWg = sync.WaitGroup{}
	wf.started = time.Now()

	// регистрация метрик workflow
	wf.Hub.Mu.Lock()
	if _, ok := wf.Hub.metrics[wf.ID]; !ok {
		wf.Hub.metrics[wf.ID] = &WorkflowMetrics{Modules: map[string]*ModuleMetrics{}}
	}
	wf.Hub.Mu.Unlock()

	// запуск модулей
	for _, m := range wf.Modules {
		wf.runWg.Add(1)
		go func(mod Module) {
			defer wf.runWg.Done()
			start := time.Now()
			if err := mod.Start(ctx); err != nil {
				fmt.Println(err)
				return
			}

			<-ctx.Done()

			// фиксирует время жизни модуля в этом запуске
			wf.Hub.ObserveLatency(wf.ID, mod.ID(), time.Since(start))
			mod.Stop()
		}(m)
	}

	wf.running = true
	return nil
}

func (wf *Workflow) Stop() {
	wf.runMu.Lock()
	defer wf.runMu.Unlock()
	if !wf.running {
		return
	}
	if wf.cancel != nil {
		wf.cancel()
	}
	wf.runWg.Wait()

	wf.running = false
}

func (wf *Workflow) Restart(parentCtx context.Context) error {
	wf.Stop()
	return wf.Start(parentCtx)
}

package service

import (
	"context"
	"crypto-gateway/internal/engine"
	"fmt"
)

type StrategyModule struct {
	*engine.BaseModule
}

// NewStrategyModule конструктор для FBP-архитектуры
func NewStrategyModule(id string) *StrategyModule {
	return &StrategyModule{
		BaseModule: engine.NewBaseModule(id),
	}
}

// работает немного не по паттерну, потому что выгоднее делать общий пул для датасорсов
func (m *StrategyModule) Start(ctx context.Context) error {
	// инициализация контекста и cancel
	_ = m.BaseModule.Start(ctx)

	out, ok := m.Outputs()["out"]
	if !ok {
		return fmt.Errorf("strategy %s missing output 'out'", m.ID())
	}

	// подключение к Binance через Orchestrator
	if StOrchestrator.isBinanceOnline {
		StOrchestrator.checkBinanceResponse([]int{})
	} else {
		StOrchestrator.checkBinanceResponse(nil)
	}
	StOrchestrator.LoadNeededStrategy(m.Context(), m.ID())

	// пример рабочей горутины, если нужно что-то слушать/отправлять
	m.WaitGroup().Add(1)
	go func() {
		defer m.WaitGroup().Done()
		for {
			select {
			case <-m.Context().Done():
				return
			case msg := <-m.Inputs()["in"].Ch:
				out.Ch <- engine.NewMessage("processed", nil)
				fmt.Println("delete this message later", msg)
			}
		}
	}()

	return nil
}

func (m *StrategyModule) Stop() {
	m.BaseModule.Stop()
}

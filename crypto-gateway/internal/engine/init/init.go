package startup

import (
	"context"
	"crypto-gateway/internal/appmetrics"
	"crypto-gateway/internal/engine"
	orchestrator "crypto-gateway/internal/modules/orchestrator/service"
	strategy "crypto-gateway/internal/modules/strategy/service"
	"encoding/json"
	"fmt"
	"time"
)

type Component struct {
	ID      string                 `json:"id"`
	Type    string                 `json:"type"`
	Options map[string]interface{} `json:"options"`
}

type Edge struct {
	From   string `json:"from"`
	To     string `json:"to"`
	Output string `json:"output"`
	Input  string `json:"input"`
}

type DiagramData struct {
	Components []Component `json:"components"`
	Edges      []Edge      `json:"edges"`
}

// Launches the entire system - initializing and running the active workflow and related components
func StartEngine(ctx context.Context) error {
	strategy.Collector = appmetrics.NewLoadMetricsCollector(60)
	strategy.StController = strategy.NewBinanceAPIController(5700)
	strategy.StBinanceApi = strategy.NewBinanceAPI(strategy.StController)
	strategy.StOrchestrator = strategy.NewBinanceAPIOrchestrator(strategy.StBinanceApi)
	strategy.AverageLoadMetrics = strategy.NewAverageLoadMetricsManager(strategy.Collector)

	strategy.SetupInitialSettings(ctx)

	hub := engine.NewHub(100 * time.Millisecond)

	_, err := LaunchActiveDiagrams(ctx, hub)
	return err
}

// запускает активные диаграмы где diagrams.isActive=true
func LaunchActiveDiagrams(ctx context.Context, hub *engine.Hub) ([]*engine.Workflow, error) {
	diagrams, err := engine.GetActiveDiagrams()
	if err != nil {
		return nil, fmt.Errorf("failed to get active diagrams: %w", err)
	}
	parsed, err := ExtractComponents(diagrams)
	if err != nil {
		return nil, fmt.Errorf("failed to extract components: %w", err)
	}

	var workflows []*engine.Workflow
	for id, data := range parsed {
		wf, err := LaunchWorkflow(ctx, id, data, hub)
		if err != nil {
			return nil, fmt.Errorf("failed to launch workflow %d: %w", id, err)
		}
		workflows = append(workflows, wf)
	}
	return workflows, nil
}

// получение типов компонентов и их id, а также всей последовательности
func ExtractComponents(diagrams []engine.Diagram) (map[int64]DiagramData, error) {
	result := make(map[int64]DiagramData)
	for _, d := range diagrams {
		var data DiagramData
		if err := json.Unmarshal(d.Data, &data); err != nil {
			return nil, fmt.Errorf("failed to parse diagram %d: %w", d.ID, err)
		}
		result[d.ID] = data
	}
	return result, nil
}

// запуск активных рабочих процессов, разделенных по id диаграммы
func LaunchWorkflow(ctx context.Context, diagramID int64, data DiagramData, hub *engine.Hub) (*engine.Workflow, error) {
	wf := engine.NewWorkflow(fmt.Sprintf("wf-%d", diagramID), hub)

	modByID := make(map[string]engine.Module)
	for _, c := range data.Components {
		var m engine.Module
		switch c.Type { // сделать универсальнее - чтобы тип и имя хранились в json файле в модуле
		case "strategy":
			/*
				strategyID, ok := c.Options["strategyID"].(string)
				if !ok {
					return nil, fmt.Errorf("strategy %s missing strategyID in options", c.ID)
				}*/
			fmt.Println(c.ID, c)
			m = strategy.NewStrategyModule(c.ID)
		case "orchestrator":
			formula, _ := c.Options["formula"].(string)
			m = orchestrator.NewOrchestratorModule(c.ID, formula)
		case "notifier":
			//m = NewNotifierModule(c.ID, c.Options)
		default:
			return nil, fmt.Errorf("unknown component type: %s", c.Type)
		}
		modByID[c.ID] = m
		wf.Modules = append(wf.Modules, m)
	}

	// соединение каналов по модулям
	for _, e := range data.Edges {
		src, ok := modByID[e.From]
		if !ok {
			return nil, fmt.Errorf("edge: src %q not found", e.From)
		}
		dst, ok := modByID[e.To]
		if !ok {
			return nil, fmt.Errorf("edge: dst %q not found", e.To)
		}

		chName := fmt.Sprintf("%s.%s -> %s.%s", e.From, e.Output, e.To, e.Input)
		hub.Connect(
			wf.ID,
			chName,
			128, // буфер по умолчанию, можно брать из options
			src, e.Output,
			dst, e.Input,
		)
	}

	// запуск workflow
	if err := wf.Start(ctx); err != nil {
		return nil, err
	}
	return wf, nil
}

func GetActiveDiagramsStrategiesId() ([]engine.DiagramStrategyIDs, error) {
	diagrams, err := engine.GetActiveDiagrams()
	if err != nil {
		return nil, fmt.Errorf("failed to get active diagrams: %w", err)
	}

	strategies, err := engine.ExtractStrategiesIDs(diagrams)
	if err != nil {
		return nil, fmt.Errorf("failed to extract strategies from diagrams: %w", err)
	}

	return strategies, nil
}

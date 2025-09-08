package service

import (
	"context"
	"crypto-gateway/internal/engine"
	"fmt"
)

type OrchestratorModule struct {
	id      string
	formula string

	inputs  map[string]*engine.Channel
	outputs map[string]*engine.Channel
	cancel  context.CancelFunc
}

func NewOrchestratorModule(id, formula string) *OrchestratorModule {
	return &OrchestratorModule{
		id:      id,
		formula: formula,
		inputs:  make(map[string]*engine.Channel),
		outputs: make(map[string]*engine.Channel),
	}
}

func (m *OrchestratorModule) ID() string { return m.id }

func (m *OrchestratorModule) SetInput(name string, ch *engine.Channel) {
	m.inputs[name] = ch
}

func (m *OrchestratorModule) SetOutput(name string, ch *engine.Channel) {
	m.outputs[name] = ch
}

func (m *OrchestratorModule) Inputs() map[string]*engine.Channel  { return m.inputs }
func (m *OrchestratorModule) Outputs() map[string]*engine.Channel { return m.outputs }
func (m *OrchestratorModule) Start(ctx context.Context) error {
	ctx, m.cancel = context.WithCancel(ctx)

	in, ok := m.inputs["in"]
	if !ok {
		return fmt.Errorf("orchestrator %s has no input 'in'", m.id)
	}
	out, ok := m.outputs["out"]
	if !ok {
		return fmt.Errorf("orchestrator %s has no output 'out'", m.id)
	}

	go func() {
		for {
			select {
			case msg := <-in.Ch:
				if msg == nil {
					continue
				}
				// ожидаем что msg.Payload = string (формула)
				formula, ok := msg.Payload.(string)
				if !ok {
					// можно залогировать ошибку
					continue
				}

				err, result := analyzeSignal(formula)
				if err != nil {
					// логируем в Meta ошибку
					errorMsg := engine.NewMessage(nil, map[string]interface{}{
						"error": err.Error(),
						"from":  m.id,
					})
					select {
					case out.Ch <- errorMsg:
					case <-ctx.Done():
						return
					}
					continue
				}

				// новый Message
				newMsg := engine.NewMessage(result, map[string]interface{}{
					"from":      m.id,
					"formula":   formula,
					"parent_id": msg.ID,
				})

				select {
				case out.Ch <- newMsg:
				case <-ctx.Done():
					return
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	return nil
}

func (m *OrchestratorModule) Stop() {
	if m.cancel != nil {
		m.cancel()
	}
}

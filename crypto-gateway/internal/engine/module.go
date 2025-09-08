package engine

import (
	"context"
	"sync"
	"sync/atomic"
)

type Module interface {
	ID() string
	Start(ctx context.Context) error
	Stop() // graceful stop
	SetInput(name string, ch *Channel)
	SetOutput(name string, ch *Channel)
	Inputs() map[string]*Channel
	Outputs() map[string]*Channel
}

type BaseModule struct {
	id      string
	inputs  map[string]*Channel
	outputs map[string]*Channel

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	stopped int32
}

func NewBaseModule(id string) *BaseModule {
	return &BaseModule{
		id:      id,
		inputs:  map[string]*Channel{},
		outputs: map[string]*Channel{},
	}
}

// сеттеры

func (b *BaseModule) SetInput(name string, ch *Channel) {
	b.inputs[name] = ch
}
func (b *BaseModule) SetOutput(name string, ch *Channel) {
	b.outputs[name] = ch
}

// геттеры

func (b *BaseModule) ID() string               { return b.id }
func (b *BaseModule) Context() context.Context { return b.ctx }
func (b *BaseModule) Inputs() map[string]*Channel {
	return b.inputs
}
func (b *BaseModule) Outputs() map[string]*Channel {
	return b.outputs
}
func (b *BaseModule) WaitGroup() *sync.WaitGroup {
	return &b.wg
}

func (b *BaseModule) Start(ctx context.Context) error {
	// Derived modules should override Start. Provide base context and cancel.
	b.ctx, b.cancel = context.WithCancel(ctx)
	return nil
}

func (b *BaseModule) Stop() {
	if atomic.CompareAndSwapInt32(&b.stopped, 0, 1) {
		if b.cancel != nil {
			b.cancel()
		}
		b.wg.Wait()
	}
}

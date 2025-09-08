package engine

import (
	"sync"
	"sync/atomic"
)

type Channel struct {
	Name      string
	Ch        chan *Message
	Buffer    int
	Workflow  string
	closed    int32 // atomic flag
	closeOnce sync.Once
}

func NewChannel(workflow, name string, buffer int) *Channel {
	return &Channel{
		Name:     name,
		Ch:       make(chan *Message, buffer),
		Buffer:   buffer,
		Workflow: workflow,
	}
}

func (c *Channel) Close() {
	c.closeOnce.Do(func() {
		atomic.StoreInt32(&c.closed, 1)
		close(c.Ch)
	})
}

func (c *Channel) IsClosed() bool {
	return atomic.LoadInt32(&c.closed) == 1
}

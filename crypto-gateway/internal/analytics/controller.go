package analytics

import (
	"log"
	"sync"
	"time"
)

type RequestFunc func() error

// данный контроллер следит за тем чтобы не пробить весовой лимит Binance.com и не получить автобан.
// Также он отвечает за выполнение запросов из структуры BinanceApi
type BinanceAPIController struct {
	MaxWeight       int
	CurrentWeight   int
	lastResetTime   time.Time
	mu              sync.Mutex
	queue           chan queuedRequest
	queueEvent      chan struct{}
	pendingRequests map[string]struct{}
}

type queuedRequest struct {
	id     string
	fn     RequestFunc
	weight int
}

func NewBinanceAPIController(maxWeight int) *BinanceAPIController {
	c := &BinanceAPIController{
		MaxWeight:       maxWeight,
		CurrentWeight:   0,
		lastResetTime:   time.Now(),
		queue:           make(chan queuedRequest, 1000), // configurable buffer
		queueEvent:      make(chan struct{}, 1),
		pendingRequests: make(map[string]struct{}),
	}

	go c.resetLoop()
	go c.processQueue()
	return c
}

// автосброс нагрузки на бинанс
func (c *BinanceAPIController) resetLoop() {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()
	for {
		<-ticker.C
		c.mu.Lock()
		c.CurrentWeight = 0
		c.lastResetTime = time.Now()
		c.mu.Unlock()
	}
}

// обрабатывает очередь запросов
func (c *BinanceAPIController) processQueue() {
	for range c.queueEvent {
		for {
			c.mu.Lock()
			if len(c.queue) == 0 {
				c.mu.Unlock()
				break
			}
			req := <-c.queue
			delete(c.pendingRequests, req.id)
			c.mu.Unlock()

			if err := req.fn(); err != nil {
				log.Printf("Error executing queued request: %v", err)
			}

			c.mu.Lock()
			c.CurrentWeight += req.weight
			c.mu.Unlock()
		}
	}
}

// Управляет лимитом и выполняет запрос. Если лимит превышен, запрос ставится в очередь.
func (c *BinanceAPIController) RequestWithLimit(weight int, fn RequestFunc) error {
	c.mu.Lock()
	if c.CurrentWeight+weight > c.MaxWeight {
		log.Printf("[WARN] API limit reached. Current weight: %d", c.CurrentWeight)

		req := queuedRequest{fn: fn, weight: weight}
		c.queue <- req

		select {
		case c.queueEvent <- struct{}{}:
		default:
			// Already triggered
		}

		c.mu.Unlock()
		return nil
	}
	c.CurrentWeight += weight
	c.mu.Unlock()

	return fn()
}

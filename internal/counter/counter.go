package counter

import (
	"sync"
	"time"
)

type Counter struct {
	mu       sync.RWMutex
	value    int
	interval time.Duration
	stop     chan struct{}
}

func New(interval time.Duration) *Counter {
	c := &Counter{
		value:    0,
		interval: interval,
		stop:     make(chan struct{}),
	}

	go c.run()

	return c
}

func (c *Counter) run() {
	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.mu.Lock()
			c.value++
			c.mu.Unlock()
		case <-c.stop:
			return
		}
	}
}

// Value returns the current counter value synchronously
func (c *Counter) Value() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.value
}

func (c *Counter) Stop() {
	close(c.stop)
}

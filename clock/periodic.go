package clock

import "time"

// PeriodicClock generates ticks at fixed intervals.
type PeriodicClock struct {
	interval time.Duration
	ticker   *time.Ticker
	tickChan chan struct{}
	stop     chan struct{}
}

// NewPeriodicClock creates a new clock that ticks at the specified interval.
func NewPeriodicClock(interval time.Duration) *PeriodicClock {
	return &PeriodicClock{
		interval: interval,
		tickChan: make(chan struct{}),
		stop:     make(chan struct{}),
	}
}

// Start begins generating ticks.
func (c *PeriodicClock) Start() {
	c.ticker = time.NewTicker(c.interval)
	go c.run()
}

func (c *PeriodicClock) run() {
	for {
		select {
		case <-c.ticker.C:
			select {
			case c.tickChan <- struct{}{}:
			case <-c.stop:
				return
			}
		case <-c.stop:
			return
		}
	}
}

// Stop stops the clock and closes the tick channel.
func (c *PeriodicClock) Stop() {
	if c.ticker != nil {
		c.ticker.Stop()
	}
	close(c.stop)
	close(c.tickChan)
}

// Subscribe returns the channel that receives tick events.
func (c *PeriodicClock) Subscribe() <-chan struct{} {
	return c.tickChan
}

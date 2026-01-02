package source

import (
	"math/rand/v2"
	"sync"

	"github.com/neox5/simval/internal/clock"
)

// RandomIntSource generates random integers within a range [min, max].
type RandomIntSource struct {
	clock    clock.Clock
	min, max int

	initOnce    sync.Once
	clockChan   <-chan struct{}
	mu          sync.Mutex
	subscribers []chan int
}

// NewRandomIntSource creates a source that generates random integers
// in the inclusive range [min, max].
func NewRandomIntSource(clk clock.Clock, min, max int) *RandomIntSource {
	return &RandomIntSource{
		clock: clk,
		min:   min,
		max:   max,
	}
}

// Subscribe returns a channel that receives random integers on each clock tick.
func (s *RandomIntSource) Subscribe() <-chan int {
	s.initOnce.Do(func() {
		s.clockChan = s.clock.Subscribe()
		go s.run()
	})

	s.mu.Lock()
	defer s.mu.Unlock()

	ch := make(chan int)
	s.subscribers = append(s.subscribers, ch)
	return ch
}

func (s *RandomIntSource) run() {
	for range s.clockChan {
		value := s.min + rand.IntN(s.max-s.min+1)

		s.mu.Lock()
		subs := s.subscribers
		s.mu.Unlock()

		for _, subChan := range subs {
			subChan <- value
		}
	}

	// Clock closed, close all subscriber channels
	s.mu.Lock()
	for _, subChan := range s.subscribers {
		close(subChan)
	}
	s.mu.Unlock()
}

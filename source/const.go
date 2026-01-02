package source

import (
	"sync"

	"github.com/neox5/simval/clock"
)

// ConstSource always returns the same value.
type ConstSource[T any] struct {
	clock clock.Clock
	value T

	initOnce    sync.Once
	clockChan   <-chan struct{}
	mu          sync.Mutex
	subscribers []chan T
}

// NewConstSource creates a source that always returns the given value.
func NewConstSource[T any](clk clock.Clock, value T) *ConstSource[T] {
	return &ConstSource[T]{
		clock: clk,
		value: value,
	}
}

// Subscribe returns a channel that receives constant values on each clock tick.
func (s *ConstSource[T]) Subscribe() <-chan T {
	s.initOnce.Do(func() {
		s.clockChan = s.clock.Subscribe()
		go s.run()
	})

	s.mu.Lock()
	defer s.mu.Unlock()

	ch := make(chan T)
	s.subscribers = append(s.subscribers, ch)
	return ch
}

func (s *ConstSource[T]) run() {
	for range s.clockChan {
		value := s.value

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

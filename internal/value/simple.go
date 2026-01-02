package value

import (
	"sync"

	"github.com/neox5/simval/internal/transform"
)

// SimpleValue is a standard implementation with source and transforms.
type SimpleValue[T any] struct {
	source     Publisher[T]
	transforms []transform.Transformation[T]

	mu      sync.RWMutex
	current T
}

// New creates a new SimpleValue with the given source and optional transforms.
// The SimpleValue automatically starts its internal goroutine.
// The goroutine exits automatically when the source channel closes.
func New[T any](
	src Publisher[T],
	transforms ...transform.Transformation[T],
) *SimpleValue[T] {
	v := &SimpleValue[T]{
		source:     src,
		transforms: transforms,
	}

	go v.run()
	return v
}

func (v *SimpleValue[T]) run() {
	sourceChan := v.source.Subscribe()

	for sourceValue := range sourceChan {
		v.mu.Lock()

		transformed := sourceValue
		for _, t := range v.transforms {
			transformed = t.Apply(transformed, v)
		}

		v.current = transformed

		v.mu.Unlock()
	}
}

// GetState returns the current state.
// Implements transform.State[T].
// Must be called with lock held (from within run()).
func (v *SimpleValue[T]) GetState() T {
	return v.current
}

// Value returns the current value.
func (v *SimpleValue[T]) Value() T {
	v.mu.RLock()
	defer v.mu.RUnlock()
	return v.current
}

// Clone creates a new SimpleValue with same source and transforms but independent state.
func (v *SimpleValue[T]) Clone() Value[T] {
	return New(v.source, v.transforms...)
}

// SetState directly sets the current state (bypasses transforms).
func (v *SimpleValue[T]) SetState(state T) {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.current = state
}

package value

import (
	"sync"

	"github.com/neox5/simval/internal/transform"
)

// Publisher provides a subscription interface for typed values.
type Publisher[T any] interface {
	Subscribe() <-chan T
}

// Value represents a simulated value that changes over time.
type Value[T any] struct {
	source     Publisher[T]
	transforms []transform.Transformation[T]

	mu      sync.RWMutex
	current T
}

// New creates a new Value with the given source and optional transforms.
// The Value automatically starts its internal goroutine.
// The goroutine exits automatically when the source channel closes.
func New[T any](
	src Publisher[T],
	transforms ...transform.Transformation[T],
) *Value[T] {
	v := &Value[T]{
		source:     src,
		transforms: transforms,
	}

	go v.run()
	return v
}

func (v *Value[T]) run() {
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
func (v *Value[T]) GetState() T {
	return v.current
}

// Value returns the current value.
func (v *Value[T]) Value() T {
	v.mu.RLock()
	defer v.mu.RUnlock()
	return v.current
}

// Reset sets the current value to resetValue.
func (v *Value[T]) Reset(resetValue T) {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.current = resetValue
}

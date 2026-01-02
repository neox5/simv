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
		// Apply transforms
		transformed := sourceValue
		for _, t := range v.transforms {
			transformed = t.Apply(transformed)
		}

		// Store result
		v.mu.Lock()
		v.current = transformed
		v.mu.Unlock()
	}
}

// Value returns the current value.
func (v *Value[T]) Value() T {
	v.mu.RLock()
	defer v.mu.RUnlock()
	return v.current
}

package value

import "sync"

// ResetOnRead wraps a value and resets it on each read.
type ResetOnRead[T any] struct {
	inner      Value[T]
	resetValue T
	mu         sync.Mutex
}

// NewResetOnRead creates a value that resets to resetValue on each read.
func NewResetOnRead[T any](v Value[T], resetValue T) *ResetOnRead[T] {
	return &ResetOnRead[T]{
		inner:      v,
		resetValue: resetValue,
	}
}

// Value returns current value and immediately resets.
func (v *ResetOnRead[T]) Value() T {
	v.mu.Lock()
	defer v.mu.Unlock()

	current := v.inner.Value()
	v.inner.SetState(v.resetValue)
	return current
}

// Clone creates a new ResetOnRead with cloned inner value.
func (v *ResetOnRead[T]) Clone() Value[T] {
	return NewResetOnRead(v.inner.Clone(), v.resetValue)
}

// SetState sets the reset value (updates what value resets to).
func (v *ResetOnRead[T]) SetState(state T) {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.resetValue = state
}

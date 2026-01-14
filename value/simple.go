package value

import (
	"sync"
	"sync/atomic"

	"github.com/neox5/simv/transform"
)

// SimpleValue is a standard implementation with source and transforms.
type SimpleValue[T any] struct {
	source     Publisher[T]
	transforms []transform.Transformation[T]

	mu         sync.RWMutex
	current    T
	updateHook atomic.Value // stores UpdateHook[T]
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

		hook := v.getUpdateHook()

		// Notify: input received
		if hook != nil {
			v.safeHookCall(func() { hook.OnInput(sourceValue, v.current) })
		}

		// Apply transforms with notifications
		transformed := sourceValue
		for _, t := range v.transforms {
			input := transformed
			currentState := v.current

			transformed = t.Apply(transformed, v)

			if hook != nil {
				name := t.Name()
				v.safeHookCall(func() {
					hook.OnTransform(name, input, transformed, currentState)
				})
			}
		}

		// Update state (triggers AfterUpdate)
		v.setState(transformed)

		v.mu.Unlock()
	}
}

// setState updates the internal state and triggers AfterUpdate hook.
// Must be called with v.mu held (locked).
func (v *SimpleValue[T]) setState(newState T) {
	v.current = newState

	if hook := v.getUpdateHook(); hook != nil {
		v.safeHookCall(func() { hook.AfterUpdate(newState) })
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

// WithTransforms creates a new SimpleValue with extended transform chain.
// Original value is unchanged.
func (v *SimpleValue[T]) WithTransforms(tfs ...transform.Transformation[T]) Value[T] {
	extended := make([]transform.Transformation[T], len(v.transforms)+len(tfs))
	copy(extended, v.transforms)
	copy(extended[len(v.transforms):], tfs)
	return New(v.source, extended...)
}

// SetState directly sets the current state (bypasses transforms).
func (v *SimpleValue[T]) SetState(state T) {
	v.mu.Lock()
	defer v.mu.Unlock()

	v.setState(state)
}

// SetUpdateHook sets the update hook for this value.
// Pass nil to disable hook.
func (v *SimpleValue[T]) SetUpdateHook(hook UpdateHook[T]) {
	if hook == nil {
		v.updateHook.Store((UpdateHook[T])(nil))
	} else {
		v.updateHook.Store(hook)
	}
}

// getUpdateHook retrieves current hook (internal).
func (v *SimpleValue[T]) getUpdateHook() UpdateHook[T] {
	if h := v.updateHook.Load(); h != nil {
		return h.(UpdateHook[T])
	}
	return nil
}

// safeHookCall executes hook synchronously with panic recovery.
func (v *SimpleValue[T]) safeHookCall(fn func()) {
	defer func() {
		if r := recover(); r != nil {
			// Hook panicked, silently ignore
		}
	}()
	fn()
}

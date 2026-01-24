package value

import (
	"sync"
	"sync/atomic"

	"github.com/neox5/simv/transform"
)

// Publisher provides a subscription interface for typed values.
type Publisher[T any] interface {
	Subscribe() <-chan T
}

// ValueStats contains observable metrics for a Value.
type ValueStats[T any] struct {
	UpdateCount    uint64
	CurrentValue   T
	TransformCount int
}

// Value represents a thread-safe simulated value with configurable behavior.
// Values must be explicitly started via Start() after configuration.
type Value[T any] struct {
	// Configuration (immutable after Start)
	source     Publisher[T]
	transforms []transform.Transformation[T]

	// Reset behavior
	resetOnRead bool
	resetValue  T

	// Lifecycle
	sourceChan <-chan T
	started    atomic.Bool
	stopOnce   sync.Once
	done       chan struct{}

	// State (mutable, protected by mu)
	mu          sync.RWMutex
	current     T
	updateCount atomic.Uint64

	// Observability
	updateHook atomic.Value // stores UpdateHook[T]
}

// New creates a new Value that will receive values from the given source.
// The value must be started via Start() before it begins receiving updates.
func New[T any](src Publisher[T]) *Value[T] {
	return &Value[T]{
		source: src,
		done:   make(chan struct{}),
	}
}

// AddTransform appends a transform to the processing pipeline.
// Returns the value for method chaining.
// Panics if called after Start().
func (v *Value[T]) AddTransform(t transform.Transformation[T]) *Value[T] {
	if v.started.Load() {
		panic("cannot add transform after Start()")
	}
	v.transforms = append(v.transforms, t)
	return v
}

// EnableResetOnRead configures the value to reset to resetValue on each Value() call.
// Returns the value for method chaining.
// Panics if called after Start().
func (v *Value[T]) EnableResetOnRead(resetValue T) *Value[T] {
	if v.started.Load() {
		panic("cannot enable reset-on-read after Start()")
	}
	v.resetOnRead = true
	v.resetValue = resetValue
	return v
}

// SetUpdateHook sets the update hook for this value.
// Pass nil to disable hook.
// Can be called before or after Start().
func (v *Value[T]) SetUpdateHook(hook UpdateHook[T]) *Value[T] {
	if hook == nil {
		v.updateHook.Store((UpdateHook[T])(nil))
	} else {
		v.updateHook.Store(hook)
	}
	return v
}

// Start begins receiving updates from the source.
// Locks configuration - no further AddTransform or EnableResetOnRead calls allowed.
// Returns the value for method chaining.
// Panics if already started.
func (v *Value[T]) Start() *Value[T] {
	if !v.started.CompareAndSwap(false, true) {
		panic("already started")
	}
	v.sourceChan = v.source.Subscribe()
	go v.run()
	return v
}

// Stop stops receiving updates and releases resources.
// Blocks until the update goroutine exits.
// Safe to call multiple times.
func (v *Value[T]) Stop() {
	v.stopOnce.Do(func() {
		// Wait for run() to finish and close done channel
		<-v.done
	})
}

// Value returns the current value.
// If reset-on-read is enabled, atomically reads and resets the value.
func (v *Value[T]) Value() T {
	if v.resetOnRead {
		v.mu.Lock()
		defer v.mu.Unlock()

		current := v.current
		v.current = v.resetValue
		return current
	}

	v.mu.RLock()
	defer v.mu.RUnlock()
	return v.current
}

// Stats returns current value metrics without side effects.
func (v *Value[T]) Stats() ValueStats[T] {
	v.mu.RLock()
	defer v.mu.RUnlock()

	return ValueStats[T]{
		UpdateCount:    v.updateCount.Load(),
		CurrentValue:   v.current,
		TransformCount: len(v.transforms),
	}
}

// GetState returns the current state.
// Implements transform.State[T].
// Must be called with lock held (from within run()).
func (v *Value[T]) GetState() T {
	return v.current
}

// run processes incoming values from the source.
// Runs in its own goroutine, started by Start().
func (v *Value[T]) run() {
	defer close(v.done)
	defer func() {
		if r := recover(); r != nil {
			// Transform panicked - isolate error, don't crash program
			// Future: could call panic hook here for observability
		}
	}()

	for sourceValue := range v.sourceChan {
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

		// Update state
		v.setState(transformed)
		v.updateCount.Add(1)

		v.mu.Unlock()
	}
}

// setState updates the internal state and triggers AfterUpdate hook.
// Must be called with v.mu held (locked).
func (v *Value[T]) setState(newState T) {
	v.current = newState

	if hook := v.getUpdateHook(); hook != nil {
		v.safeHookCall(func() { hook.AfterUpdate(newState) })
	}
}

// getUpdateHook retrieves current hook (internal).
func (v *Value[T]) getUpdateHook() UpdateHook[T] {
	if h := v.updateHook.Load(); h != nil {
		return h.(UpdateHook[T])
	}
	return nil
}

// safeHookCall executes hook synchronously with panic recovery.
func (v *Value[T]) safeHookCall(fn func()) {
	defer func() {
		if r := recover(); r != nil {
			// Hook panicked, silently ignore
		}
	}()
	fn()
}

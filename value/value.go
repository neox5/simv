package value

import "github.com/neox5/simv/transform"

// Publisher provides a subscription interface for typed values.
type Publisher[T any] interface {
	Subscribe() <-chan T
}

// Value represents a readable, clonable, and resettable simulated value.
type Value[T any] interface {
	Value() T
	Clone() Value[T]
	WithTransforms(transforms ...transform.Transformation[T]) Value[T]
	SetState(T)
	SetUpdateHook(UpdateHook[T])
}

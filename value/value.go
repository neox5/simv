package value

// Publisher provides a subscription interface for typed values.
type Publisher[T any] interface {
	Subscribe() <-chan T
}

// Value represents a readable, clonable, and resettable simulated value.
type Value[T any] interface {
	Value() T
	Clone() Value[T]
	SetState(T)
}

package transform

// State provides read-only access to current state.
type State[T any] interface {
	GetState() T
}

// Transformation modifies a value.
type Transformation[T any] interface {
	Apply(incoming T, state State[T]) T
}

// Accumulate adds each value to a running total.
// Requires T to support the + operator (int, int64, float64, etc.).
type Accumulate[T Numeric] struct{}

// NewAccumulate creates a transform that accumulates values.
func NewAccumulate[T Numeric]() *Accumulate[T] {
	return &Accumulate[T]{}
}

// Apply adds the incoming value to the current state and returns the new total.
func (t *Accumulate[T]) Apply(incoming T, state State[T]) T {
	current := state.GetState()
	return current + incoming
}

// Numeric defines types that support arithmetic operations.
type Numeric interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 |
		~float32 | ~float64
}

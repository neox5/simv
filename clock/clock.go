package clock

// Publisher provides a subscription interface for typed values.
type Publisher[T any] interface {
	Subscribe() <-chan T
}

// Clock provides timing signals for value updates.
type Clock interface {
	Publisher[struct{}]
	Start()
	Stop()
}

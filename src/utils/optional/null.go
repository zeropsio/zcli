package optional

import "fmt"

func Empty[T any]() Null[T] {
	return Null[T]{}

}

func New[T any](in T) Null[T] {
	return Null[T]{
		filled: true,
		value:  in,
	}
}

type Null[T any] struct {
	filled bool
	value  T
}

func (n Null[T]) Filled() bool {
	return n.filled
}

func (n Null[T]) Some() T {
	return n.value
}

func (n Null[T]) Get() (T, bool) {
	return n.value, n.filled
}

func (n Null[T]) Unwrap() T {
	if n.filled {
		return n.value
	}
	panic("not filled")
}

func (n Null[T]) String() string {
	if !n.filled {
		return fmt.Sprintf("<nil>[%T]", n.value)
	}
	return fmt.Sprintf("%v", n.value)
}

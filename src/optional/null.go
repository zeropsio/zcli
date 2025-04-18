package optional

import (
	"fmt"

	"github.com/pkg/errors"
	"golang.org/x/exp/constraints"
)

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

func (n Null[T]) OrDefault(d T) T {
	if n.filled {
		return n.value
	}
	return d
}

func (n Null[T]) Get() (T, bool) {
	return n.value, n.filled
}

func (n Null[T]) Expect(errMessage string) (T, error) {
	if n.filled {
		return n.value, nil
	}
	var t T
	return t, errors.New(errMessage)
}

func (n Null[T]) String() string {
	if !n.filled {
		return fmt.Sprintf("<nil>[%T]", n.value)
	}
	return fmt.Sprintf("%v", n.value)
}

// Grpc use only for conversion to protobuf struct
func (n Null[T]) Grpc() *T {
	if n.filled {
		return &n.value
	}
	return nil
}

type Native[U any] interface {
	Native() U
}

type Nullable[T any] interface {
	Get() (T, bool)
}

// Compare compares A to B and returns
//   - +1 if a > b
//   - -1 if a < b
//   - 0 if a = b or EITHER ONE is not set
func Compare[T constraints.Integer | constraints.Float, N Nullable[T]](a, b N) (int, T, T) {
	aVal, aFilled := a.Get()
	bVal, bFilled := b.Get()
	if !aFilled || !bFilled {
		return 0, aVal, bVal
	}
	if aVal > bVal {
		return 1, aVal, bVal
	}
	if aVal < bVal {
		return -1, aVal, bVal
	}
	return 0, aVal, bVal
}

func AreAllFilled[T any, N Nullable[T]](in ...N) bool {
	for _, n := range in {
		if _, filled := n.Get(); !filled {
			return false
		}
	}
	return true
}

func AreAllEmpty[T any, N Nullable[T]](in ...N) bool {
	for _, n := range in {
		if _, filled := n.Get(); filled {
			return false
		}
	}
	return true
}

func FromTyped[T any, N Nullable[T]](nullable N) Null[T] {
	typed, filled := nullable.Get()
	if !filled {
		return Empty[T]()
	}
	return New(typed)
}

func SomeFromTyped[T any, N Nullable[T]](nullable N) T {
	typed, filled := nullable.Get()
	if !filled {
		var empty T
		return empty
	}
	return typed
}

func GrpcToNull[T, K any](in *T, convert func(in *T) K) Null[K] {
	if in == nil {
		return Empty[K]()
	}
	return New(convert(in))
}

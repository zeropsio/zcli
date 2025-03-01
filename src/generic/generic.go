package generic

func Must[T any](in T, err error) T {
	if err != nil {
		panic(err)
	}
	return in
}

func Some[T any](in T, err error) T {
	if err != nil {
		return Empty[T]()
	}
	return in
}

func Pointer[T any](in T) *T {
	return &in
}

func Deref[T any](in *T) T {
	if in == nil {
		var t T
		return t
	}
	return *in
}

type OptionError[T any] func(*T) error

func ApplyOptionsError[K any, T ~func(*K) error](in ...T) (K, []error) {
	var emptyValue K
	return ApplyOptionsErrorWithDefault[K](emptyValue, in...)
}

func ApplyOptionsErrorWithDefault[K any, T ~func(*K) error](k K, in ...T) (K, []error) {
	var errors []error
	for _, o := range in {
		if err := o(&k); err != nil {
			errors = append(errors, err)
		}
	}
	return k, errors
}

type Option[T any] func(*T)

func If[T any](cond bool, in ...Option[T]) Option[T] {
	return func(c *T) {
		if !cond {
			return
		}
		for _, o := range in {
			o(c)
		}
	}
}

func WithOptions[T any](in ...Option[T]) Option[T] {
	return func(c *T) {
		for _, o := range in {
			o(c)
		}
	}
}

func ApplyOptions[K any, T ~func(*K)](in ...T) K {
	var emptyValue K
	return ApplyOptionsWithDefault(emptyValue, in...)
}

func ApplyOptionsWithDefault[K any, T ~func(*K)](k K, in ...T) K {
	for _, o := range in {
		o(&k)
	}
	return k
}

func GetSliceElementByIndexOrDefault[T any](in []T, index int, defaultValue T) T {
	if len(in) <= index {
		return defaultValue
	}
	return in[index]
}

func FilterSlice[T any](in []T, filter func(in T) bool) []T {
	r := make([]T, 0, len(in))
	for _, i := range in {
		if filter(i) {
			r = append(r, i)
		}
	}
	return r
}

func FindOne[T any](in []T, filter func(in T) bool) (r T, _ bool) {
	for _, i := range in {
		if filter(i) {
			return i, true
		}
	}
	return r, false
}

func TransformSlice[T, K any](in []T, transform func(in T) K) []K {
	r := make([]K, len(in))
	for i, v := range in {
		r[i] = transform(v)
	}
	return r
}

func TransformMapToSlice[K comparable, V, T any](in map[K]V, transform func(k K, v V) T) []T {
	r := make([]T, len(in))
	index := 0
	for k, v := range in {
		r[index] = transform(k, v)
		index++
	}
	return r
}

func TransformSliceToMap[K comparable, V, T any](in []T, transform func(e T, i int) (K, V)) map[K]V {
	r := make(map[K]V, len(in))
	for i, e := range in {
		k, v := transform(e, i)
		r[k] = v
	}
	return r
}

func ToAny[T any](in T) any {
	return in
}

func TransformSliceErr[T, K any](in []T, transform func(in T) (K, error)) ([]K, error) {
	r := make([]K, len(in))
	for i, v := range in {
		k, err := transform(v)
		if err != nil {
			return nil, err
		}
		r[i] = k
	}
	return r, nil
}

func SliceContains[T any](in []T, check func(in T) bool) bool {
	for _, v := range in {
		if check(v) {
			return true
		}
	}
	return false
}

func MergeMaps[V comparable, T any](maps ...map[V]T) map[V]T {
	if len(maps) == 0 {
		return make(map[V]T)
	}
	out := make(map[V]T, len(maps[0])*3) // probably `x * 3` is better than `x * len(maps)` in cases where maps have overlapping keys
	for _, m := range maps {
		for k, v := range m {
			out[k] = v
		}
	}
	return out
}

func CopyMap[V comparable, T any](in map[V]T) map[V]T {
	out := make(map[V]T, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

func IsOneOf[T comparable](val T, values ...T) bool {
	for _, v := range values {
		if v == val {
			return true
		}
	}
	return false
}

func IsEmpty[T comparable](in T) bool {
	var emptyVal T
	return in == emptyVal
}

func DefaultOnEmpty[T comparable](in, def T) T {
	if IsEmpty(in) {
		return def
	}
	return in
}

func DefaultOnNilPointer[T any](in *T, def T) T {
	if in == nil {
		return def
	}
	return *in
}

func AreAllEmpty[T comparable](in ...T) bool {
	var emptyVal T
	for _, t := range in {
		if t != emptyVal {
			return false
		}
	}
	return true
}

func AreAllNil[T any](in ...*T) bool {
	for _, t := range in {
		if t != nil {
			return false
		}
	}
	return true
}

func AreAllPointerValuesEqual[T comparable](v ...*T) bool {
	if len(v) <= 1 {
		return true
	}
	val := v[0]
	for _, t := range v[1:] {
		if t == nil && val == nil {
			continue
		}
		if t != nil && val != nil && *t == *val {
			continue
		}
		return false
	}
	return true
}

func Empty[T any]() T {
	var emptyVal T
	return emptyVal
}

func Ternary[T any, C ~bool](condition C, v1, v2 T) T {
	if condition {
		return v1
	}
	return v2
}

func FirstNotNil[T any](v ...*T) *T {
	for _, t := range v {
		if t != nil {
			return t
		}
	}
	return nil
}

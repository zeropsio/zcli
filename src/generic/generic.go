package generic

type FilterFunc[T any] func(in T) bool

func FilterSlice[T any](in []T, filter FilterFunc[T]) []T {
	r := make([]T, 0, len(in))
	for _, i := range in {
		if filter(i) {
			r = append(r, i)
		}
	}
	return r
}

package generic

func TransformSlice[T, K any](in []T, transform func(i T) K) []K {
	r := make([]K, len(in))
	for i, e := range in {
		r[i] = transform(e)
	}
	return r
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

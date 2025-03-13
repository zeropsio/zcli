package generic

// K = key, V = value
func MapKeysToSlice[K comparable, V any](m map[K]V) []K {
	o := make([]K, len(m))
	index := 0
	for k := range m {
		o[index] = k
		index++
	}
	return o
}

// K = key, V = value
func MapToSlice[K comparable, V any](m map[K]V) []V {
	o := make([]V, len(m))
	index := 0
	for _, v := range m {
		o[index] = v
		index++
	}
	return o
}

// K = key, V = value
func MapToPointerSlice[K comparable, V any](m map[K]V) []*V {
	o := make([]*V, len(m))
	index := 0
	for k := range m {
		v := m[k]
		o[index] = &v
		index++
	}
	return o
}

// K = key, V = value
func PointerMapToPointerSlice[K comparable, V any](m map[K]*V) []*V {
	o := make([]*V, len(m))
	index := 0
	for _, v := range m {
		o[index] = v
		index++
	}
	return o
}

// K = key, V = value
func PointerMapToSlice[K comparable, V any](m map[K]*V) []V {
	o := make([]V, len(m))
	index := 0
	for _, v := range m {
		o[index] = *v
		index++
	}
	return o
}

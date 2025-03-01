package generic

// K = key, V = value
func SliceToMap[K comparable, V any](slice []V, keyFromValue func(v V, i int) K) map[K]V {
	m := make(map[K]V, len(slice))
	for i, e := range slice {
		m[keyFromValue(e, i)] = e
	}
	return m
}

// K = key, V = value
func SliceToMapErr[K comparable, V any](slice []V, keyFromValue func(v V, i int) (K, error)) (map[K]V, error) {
	m := make(map[K]V, len(slice))
	for i, e := range slice {
		key, err := keyFromValue(e, i)
		if err != nil {
			return nil, err
		}
		m[key] = e
	}
	return m, nil
}

// K = key
func SliceToMapKeys[K comparable, V any](slice []V, keyFromValue func(v V) K) map[K]any {
	m := make(map[K]any, len(slice))
	for _, e := range slice {
		m[keyFromValue(e)] = nil
	}
	return m
}

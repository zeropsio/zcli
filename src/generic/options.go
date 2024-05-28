package generic

type Option[T any] func(*T)

func ApplyOptions[K any, T ~func(*K)](in ...T) K {
	var k K
	return ApplyOptionsWithDefault(k, in...)
}

func ApplyOptionsWithDefault[K any, T ~func(*K)](k K, in ...T) K {
	for _, o := range in {
		if o == nil {
			continue
		}
		o(&k)
	}
	return k
}

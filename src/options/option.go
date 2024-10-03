package options

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

package cmd

import (
	"fmt"
	"strings"

	"github.com/zeropsio/zcli/src/gn"
)

func enumDefaultForFlag[T ~string](in T) string {
	s := string(in)
	s = strings.ToLower(s)
	return s
}

func enumValuesForFlag[T ~string](values []T) string {
	v := gn.TransformSlice(
		values,
		func(in T) string {
			s := string(in)
			s = strings.ToLower(s)
			return s
		},
	)
	return fmt.Sprintf("[%s]", strings.Join(v, ", "))
}

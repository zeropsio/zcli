package entity

import (
	"fmt"
	"reflect"
)

func entityTemplateFields[T any]() []string {
	t := reflect.TypeFor[T]()
	fields := make([]string, 0, t.NumField())
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if !f.IsExported() {
			continue
		}
		fields = append(fields, fmt.Sprintf("{{.%s}}", f.Name))
	}
	return fields
}

package generic

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

type data struct {
	str string
}

func TestMapToSlice(t *testing.T) {
	assert := assert.New(t)
	m := map[string]data{
		"test1": {str: "data"},
		"test2": {str: "dataxxx"},
		"test3": {str: "datarrrrrrr"},
	}

	s := MapToSlice(m)

	assert.Len(m, len(s))
	assert.Contains(s, m["test1"])
	assert.Contains(s, m["test2"])
	assert.Contains(s, m["test3"])
}

func TestMapToPointerSlice(t *testing.T) {
	assert := assert.New(t)
	m := map[string]data{
		"test1": {str: "data"},
		"test2": {str: "dataxxx"},
		"test3": {str: "datarrrrrrr"},
	}

	s := MapToPointerSlice(m)

	assert.Len(m, len(s))
	assert.Contains(s, Pointer(m["test1"]))
	assert.Contains(s, Pointer(m["test2"]))
	assert.Contains(s, Pointer(m["test3"]))
}

func TestMapKeysToSlice(t *testing.T) {
	assert := assert.New(t)
	m := map[string]data{
		"test1": {str: "data"},
		"test2": {str: "dataxxx"},
		"test3": {str: "datarrrrrrr"},
	}

	s := MapKeysToSlice(m)

	assert.Len(m, len(s))
	assert.Contains(s, "test1")
	assert.Contains(s, "test2")
	assert.Contains(s, "test3")
}

func TestPointerMapToPointerSlice(t *testing.T) {
	assert := assert.New(t)
	m := map[string]*data{
		"test1": {str: "data"},
		"test2": {str: "dataxxx"},
		"test3": {str: "datarrrrrrr"},
	}

	s := PointerMapToPointerSlice(m)

	assert.Len(m, len(s))
	assert.Contains(s, m["test1"])
	assert.Contains(s, m["test2"])
	assert.Contains(s, m["test3"])
}

func TestPointerMapToSlice(t *testing.T) {
	assert := assert.New(t)
	m := map[string]*data{
		"test1": {str: "data"},
		"test2": {str: "dataxxx"},
		"test3": {str: "datarrrrrrr"},
	}

	s := PointerMapToSlice(m)

	assert.Len(m, len(s))
	assert.Contains(s, Deref(m["test1"]))
	assert.Contains(s, Deref(m["test2"]))
	assert.Contains(s, Deref(m["test3"]))
}

type bench struct {
	num int
}

func BenchmarkMapToSlice(b *testing.B) {
	m := map[string]bench{}
	b.StopTimer()
	for i := 0; i < b.N; i++ {
		m[fmt.Sprintf("%d", i)] = bench{num: i}
	}
	b.StartTimer()
	MapToSlice(m)
	b.StopTimer()
}

package generic

import (
	"sort"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTransformSlice(t *testing.T) {
	ints := []int{1, 2, 30, 77, 42, -5}
	floats := TransformSlice(ints, func(in int) float64 { return float64(in) })
	assert := assert.New(t)
	assert.Len(floats, len(ints))
	for i, f := range floats {
		assert.IsType(float64(0), f)
		assert.Equal(f, float64(ints[i]))
	}
}

func TestTransformSliceErr(t *testing.T) {
	ints := []int{1, 2, 30, 77, 42, -5}
	floats, err := TransformSliceErr(ints, func(in int) (float64, error) { return float64(in), nil })
	assert := assert.New(t)
	assert.Nil(err)
	assert.Len(floats, len(ints))
	for i, f := range floats {
		assert.IsType(float64(0), f)
		assert.Equal(f, float64(ints[i]))
	}

	strings := []string{"1", "2", "f"}
	_, err = TransformSliceErr(strings, func(in string) (int64, error) { return strconv.ParseInt(in, 10, 64) })
	assert.Error(err)

	strings = []string{"1", "5", "43", "-500"}
	int64s, err := TransformSliceErr(strings, func(in string) (int64, error) { return strconv.ParseInt(in, 10, 64) })
	assert.Nil(err)
	assert.Len(int64s, len(strings))
	for index, i64 := range int64s {
		assert.IsType(int64(0), i64)
		parsedInt, err := strconv.ParseInt(strings[index], 10, 64)
		assert.Nil(err)
		assert.Equal(parsedInt, i64)
	}
}

func TestTransformMapToSlice(t *testing.T) {
	items := map[string]string{
		"key1": "value1",
		"key2": "value2",
	}
	out := []string{
		"key1value1",
		"key2value2",
	}
	strings := TransformMapToSlice(items, func(key, value string) string {
		return key + value
	})
	sort.Strings(strings)
	assert := assert.New(t)
	assert.Len(strings, len(items))
	for i, s := range strings {
		assert.IsType("", s)
		assert.Equal(s, out[i])
	}
}

func TestFilterSlice(t *testing.T) {
	ints := []int{1, 2, 30, 77, 42, -5}
	filteredInts := FilterSlice(ints, func(in int) bool { return in > 10 && in < 70 })
	assert := assert.New(t)
	assert.Len(filteredInts, 2)
	assert.Equal(filteredInts[0], 30)
	assert.Equal(filteredInts[1], 42)
}

func TestIsEmpty(t *testing.T) {
	assert := assert.New(t)
	{
		assert.True(IsEmpty(""))
		assert.True(IsEmpty(0))
		assert.True(IsEmpty(0.0))
		assert.True(IsEmpty(false))
		var emptyPointer *int
		assert.True(IsEmpty(emptyPointer))
		var emptyArray [4]int
		assert.True(IsEmpty(emptyArray))
	}
	{
		assert.False(IsEmpty("a"))
		assert.False(IsEmpty(1))
		assert.False(IsEmpty(1.0))
		assert.False(IsEmpty(true))
		nonEmptyPointer := new(int)
		assert.False(IsEmpty(nonEmptyPointer))
		var nonEmptyArray [4]int
		nonEmptyArray[0] = 1
		assert.False(IsEmpty(nonEmptyArray))
	}
}

func TestAreAllPointerValuesEqual(t *testing.T) {
	type args[T comparable] struct {
		v []*T
	}
	type testCase[T comparable] struct {
		name string
		args args[T]
		want bool
	}
	tests := []testCase[int]{
		{
			name: "Empty",
			args: args[int]{
				v: []*int{},
			},
			want: true,
		},
		{
			name: "One value",
			args: args[int]{
				v: []*int{
					Pointer(1),
				},
			},
			want: true,
		},
		{
			name: "All nil",
			args: args[int]{
				v: []*int{
					nil,
					nil,
					nil,
				},
			},
			want: true,
		},
		{
			name: "All same",
			args: args[int]{
				v: []*int{
					Pointer(1),
					Pointer(1),
					Pointer(1),
				},
			},
			want: true,
		},
		{
			name: "All different",
			args: args[int]{
				v: []*int{
					Pointer(1),
					Pointer(2),
					Pointer(3),
				},
			},
			want: false,
		},
		{
			name: "Some different",
			args: args[int]{
				v: []*int{
					Pointer(1),
					Pointer(1),
					Pointer(1),
					Pointer(2),
					Pointer(1),
				},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, AreAllPointerValuesEqual(tt.args.v...), "AreAllPointerValuesEqual")
		})
	}
}

//nolint:gosec
package gn

import (
	"sort"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTransformSlice(t *testing.T) {
	ints := []int{1, 2, 30, 77, 42, -5}
	uints := TransformSlice(ints, func(in int) uint { return uint(in) })
	require := require.New(t)
	require.Len(uints, len(ints))
	for i, f := range uints {
		require.IsType(uint(0), f)
		require.Equal(f, uint(ints[i]))
	}
}

func TestTransformSliceErr(t *testing.T) {
	ints := []int{1, 2, 30, 77, 42, -5}
	uints, err := TransformSliceErr(ints, func(in int) (uint, error) { return uint(in), nil })
	require := require.New(t)
	require.NoError(err)
	require.Len(uints, len(ints))
	for i, f := range uints {
		require.IsType(uint(0), f)
		require.Equal(f, uint(ints[i]))
	}

	strings := []string{"1", "2", "f"}
	_, err = TransformSliceErr(strings, func(in string) (int64, error) { return strconv.ParseInt(in, 10, 64) })
	require.Error(err)

	strings = []string{"1", "5", "43", "-500"}
	int64s, err := TransformSliceErr(strings, func(in string) (int64, error) { return strconv.ParseInt(in, 10, 64) })
	require.NoError(err)
	require.Len(int64s, len(strings))
	for index, i64 := range int64s {
		require.IsType(int64(0), i64)
		parsedInt, err := strconv.ParseInt(strings[index], 10, 64)
		require.NoError(err)
		require.Equal(parsedInt, i64)
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
	require := require.New(t)
	require.Len(strings, len(items))
	for i, s := range strings {
		require.IsType("", s)
		require.Equal(s, out[i])
	}
}

func TestFilterSlice(t *testing.T) {
	ints := []int{1, 2, 30, 77, 42, -5}
	filteredInts := FilterSlice(ints, func(in int) bool { return in > 10 && in < 70 })
	require := require.New(t)
	require.Len(filteredInts, 2)
	require.Equal(30, filteredInts[0])
	require.Equal(42, filteredInts[1])
}

func TestIsEmpty(t *testing.T) {
	require := require.New(t)
	{
		require.True(IsEmpty(""))
		require.True(IsEmpty(0))
		require.True(IsEmpty(0.0))
		require.True(IsEmpty(false))
		var emptyPointer *int
		require.True(IsEmpty(emptyPointer))
		var emptyArray [4]int
		require.True(IsEmpty(emptyArray))
	}
	{
		require.False(IsEmpty("a"))
		require.False(IsEmpty(1))
		require.False(IsEmpty(1.0))
		require.False(IsEmpty(true))
		nonEmptyPointer := new(int)
		require.False(IsEmpty(nonEmptyPointer))
		var nonEmptyArray [4]int
		nonEmptyArray[0] = 1
		require.False(IsEmpty(nonEmptyArray))
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
					Ptr(1),
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
					Ptr(1),
					Ptr(1),
					Ptr(1),
				},
			},
			want: true,
		},
		{
			name: "All different",
			args: args[int]{
				v: []*int{
					Ptr(1),
					Ptr(2),
					Ptr(3),
				},
			},
			want: false,
		},
		{
			name: "Some different",
			args: args[int]{
				v: []*int{
					Ptr(1),
					Ptr(1),
					Ptr(1),
					Ptr(2),
					Ptr(1),
				},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equalf(t, tt.want, AreAllPointerValuesEqual(tt.args.v...), "AreAllPointerValuesEqual")
		})
	}
}

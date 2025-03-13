package generic

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

type elem struct {
	id  int
	str string
}

func TestSliceToMap(t *testing.T) {
	assert := assert.New(t)
	{
		s := []elem{
			{
				id:  30,
				str: "xxx",
			},
			{
				id:  -80,
				str: "aaa",
			},
			{
				id:  8000,
				str: "bbb",
			},
		}

		m := SliceToMap(s, func(v elem, _ int) int {
			return v.id
		})

		assert.Len(m, len(s))
		assert.Equal(m[30], s[0])
		assert.Equal(m[-80], s[1])
		assert.Equal(m[8000], s[2])
	}

	{
		s := []data{
			{
				str: "xxx",
			},
			{
				str: "aaa",
			},
			{
				str: "bbb",
			},
		}

		m := SliceToMap(s, func(_ data, i int) int {
			return i
		})
		assert.Len(m, len(s))
		assert.Equal(m[0], s[0])
		assert.Equal(m[1], s[1])
		assert.Equal(m[2], s[2])
	}
}

func TestSliceToMapErr(t *testing.T) {
	assert := assert.New(t)
	s := []data{
		{
			str: "-300",
		},
		{
			str: "5",
		},
		{
			str: "bbb",
		},
	}

	_, err := SliceToMapErr(s, func(v data, _ int) (int64, error) {
		return strconv.ParseInt(v.str, 10, 64)
	})
	assert.Error(err)
}

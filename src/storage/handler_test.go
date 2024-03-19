package storage

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStorage(t *testing.T) {
	type dataObject struct {
		Param string
	}

	{
		storage, err := New[dataObject](Config{
			FilePath: "./test",
			FileMode: 0666,
		})
		require.NoError(t, err)

		{
			d := storage.Data()
			require.Equal(t, "", d.Param)
		}

		{
			d, err := storage.Update(func(data dataObject) dataObject {
				data.Param = "value"
				return data
			})
			require.NoError(t, err)
			require.Equal(t, "value", d.Param)
		}
	}

	{
		storage, err := New[dataObject](Config{
			FilePath: "./test",
		})
		require.NoError(t, err)

		d := storage.Data()
		require.Equal(t, "value", d.Param)
	}

	err := os.Remove("./test")
	require.NoError(t, err)
}

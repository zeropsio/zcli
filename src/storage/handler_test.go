package storage

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	helperProcessFileEnv = "ZCLI_STORAGE_HELPER_PROCESS_FILE"
	helperProcessUpdates = 50
)

type counterData struct {
	Counter int
	X       string
	Y       string
}

func newCounterHandler(t *testing.T, filePath string) *Handler[counterData] {
	t.Helper()
	h, err := New[counterData](Config{
		FilePath: filePath,
		FileMode: 0600,
	})
	require.NoError(t, err)
	return h
}

func increment(data counterData) counterData {
	data.Counter++
	return data
}

func TestStorage(t *testing.T) {
	type dataObject struct {
		Param string
	}
	filePath := filepath.Join(t.TempDir(), "test")

	{
		storage, err := New[dataObject](Config{
			FilePath: filePath,
			FileMode: 0666,
		})
		require.NoError(t, err)

		{
			d := storage.Data()
			require.Empty(t, d.Param)
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
			FilePath: filePath,
		})
		require.NoError(t, err)

		d := storage.Data()
		require.Equal(t, "value", d.Param)
	}
}

func TestStorage_EmptyFile(t *testing.T) {
	filePath := filepath.Join(t.TempDir(), "data")
	require.NoError(t, os.WriteFile(filePath, nil, 0600))

	h := newCounterHandler(t, filePath)
	require.Equal(t, counterData{}, h.Data())
	require.NoFileExists(t, filePath)
}

func TestStorage_Clear(t *testing.T) {
	filePath := filepath.Join(t.TempDir(), "data")
	h := newCounterHandler(t, filePath)
	_, err := h.Update(increment)
	require.NoError(t, err)

	d, err := h.Clear()
	require.NoError(t, err)
	require.Equal(t, counterData{}, d)
	require.Equal(t, counterData{}, h.Data())
	require.Equal(t, counterData{}, newCounterHandler(t, filePath).Data())
}

func TestStorage_ConcurrentHandlers(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "data")
	const handlers = 4
	const updates = 50

	var wg sync.WaitGroup
	for range handlers {
		h := newCounterHandler(t, filePath)
		wg.Go(func() {
			for range updates {
				_, err := h.Update(increment)
				assert.NoError(t, err)
			}
		})
	}
	wg.Wait()

	require.Equal(t, handlers*updates, newCounterHandler(t, filePath).Data().Counter)
	tmpFiles, err := filepath.Glob(filepath.Join(dir, "*.tmp"))
	require.NoError(t, err)
	require.Empty(t, tmpFiles)
	require.FileExists(t, filePath+".lock")
}

func TestStorage_ConcurrentUpdatesOnOneHandler(t *testing.T) {
	filePath := filepath.Join(t.TempDir(), "data")
	h := newCounterHandler(t, filePath)
	const goroutines = 4
	const updates = 50

	var wg sync.WaitGroup
	for range goroutines {
		wg.Go(func() {
			for range updates {
				_, err := h.Update(increment)
				assert.NoError(t, err)
			}
		})
	}
	wg.Wait()

	require.Equal(t, goroutines*updates, h.Data().Counter)
	require.Equal(t, goroutines*updates, newCounterHandler(t, filePath).Data().Counter)
}

func TestStorage_StaleHandlerKeepsOtherWrites(t *testing.T) {
	filePath := filepath.Join(t.TempDir(), "data")
	a := newCounterHandler(t, filePath)
	b := newCounterHandler(t, filePath)

	_, err := b.Update(func(data counterData) counterData {
		data.X = "x"
		return data
	})
	require.NoError(t, err)

	d, err := a.Update(func(data counterData) counterData {
		data.Y = "y"
		return data
	})
	require.NoError(t, err)
	require.Equal(t, counterData{X: "x", Y: "y"}, d)
	require.Equal(t, d, a.Data())
	require.Equal(t, d, newCounterHandler(t, filePath).Data())
}

func TestStorage_ModifyErrorWritesNothing(t *testing.T) {
	filePath := filepath.Join(t.TempDir(), "data")
	a := newCounterHandler(t, filePath)
	b := newCounterHandler(t, filePath)

	_, err := b.Update(func(data counterData) counterData {
		data.X = "x"
		return data
	})
	require.NoError(t, err)
	before, err := os.ReadFile(filePath)
	require.NoError(t, err)

	errCallback := errors.New("callback failed")
	d, err := a.Modify(func(data counterData) (counterData, error) {
		data.Y = "y"
		return data, errCallback
	})
	require.ErrorIs(t, err, errCallback)
	require.Equal(t, counterData{X: "x"}, d)
	require.Equal(t, counterData{X: "x"}, a.Data())

	after, err := os.ReadFile(filePath)
	require.NoError(t, err)
	require.Equal(t, before, after)
}

func TestStorage_ReaderNeverSeesMissingFile(t *testing.T) {
	filePath := filepath.Join(t.TempDir(), "data")
	_, err := newCounterHandler(t, filePath).Update(increment)
	require.NoError(t, err)

	var writersDone atomic.Bool
	var writers sync.WaitGroup
	for range 2 {
		h := newCounterHandler(t, filePath)
		writers.Go(func() {
			for range 100 {
				_, err := h.Update(increment)
				assert.NoError(t, err)
			}
		})
	}

	var reader sync.WaitGroup
	reader.Go(func() {
		for !writersDone.Load() {
			h, err := New[counterData](Config{
				FilePath: filePath,
				FileMode: 0600,
			})
			//nolint:testifylint // Why: require must not stop the test from outside the test goroutine.
			if !assert.NoError(t, err) {
				return
			}
			if !assert.NotZero(t, h.Data().Counter) {
				return
			}
		}
	})

	writers.Wait()
	writersDone.Store(true)
	reader.Wait()
}

func TestStorage_ConcurrentProcesses(t *testing.T) {
	filePath := filepath.Join(t.TempDir(), "data")
	const processes = 4

	var wg sync.WaitGroup
	for range processes {
		wg.Go(func() {
			cmd := exec.CommandContext(t.Context(), os.Args[0], "-test.run=^TestStorage_HelperProcess$")
			cmd.Env = append(os.Environ(), helperProcessFileEnv+"="+filePath)
			out, err := cmd.CombinedOutput()
			assert.NoError(t, err, string(out))
		})
	}
	wg.Wait()

	require.Equal(t, processes*helperProcessUpdates, newCounterHandler(t, filePath).Data().Counter)
}

func TestStorage_HelperProcess(t *testing.T) {
	filePath := os.Getenv(helperProcessFileEnv)
	if filePath == "" {
		t.Skip("runs only as a child of TestStorage_ConcurrentProcesses")
	}
	h := newCounterHandler(t, filePath)
	for range helperProcessUpdates {
		_, err := h.Update(increment)
		require.NoError(t, err)
	}
}

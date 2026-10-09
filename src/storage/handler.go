package storage

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"

	"github.com/gofrs/flock"
	"github.com/pkg/errors"

	"github.com/zeropsio/zcli/src/i18n"
)

type Config struct {
	FilePath string
	FileMode os.FileMode
}

// Handler keeps an in-memory copy of a JSON file that several processes share.
// Every write locks the file exclusively and applies the change to the data on disk, so a process never writes back stale data.
type Handler[T any] struct {
	//nolint:structcheck // Why: `is unused` error is false positive
	config Config
	//nolint:structcheck // Why: `is unused` error is false positive
	data T
	//nolint:structcheck // Why: `is unused` error is false positive
	lock sync.RWMutex
	// writeLock serializes writes within the process, because a Flock that is already locked returns immediately from Lock.
	writeLock sync.Mutex
	fileLock  *flock.Flock
}

func New[T any](config Config) (*Handler[T], error) {
	h := &Handler[T]{
		config: config,
		// The lock file is readable by everyone, so a non-root zcli can still lock it after a sudo run created it.
		fileLock: flock.New(config.FilePath+".lock", flock.SetPermissions(0644)),
	}

	// Windows refuses a rename over a file that another process holds open, so a reader holds the shared lock.
	err := h.withFileLock(h.fileLock.RLock, func() error {
		data, err := h.read()
		if err != nil {
			return err
		}
		h.data = data
		return nil
	})
	return h, err
}

// Update applies the callback to the data currently on disk and saves the result.
func (h *Handler[T]) Update(callback func(T) T) (T, error) {
	return h.Modify(func(data T) (T, error) {
		return callback(data), nil
	})
}

// Modify applies the callback to the data currently on disk and saves the result.
// If the callback returns an error, nothing is written and Modify returns the current data with the error.
// Other processes wait for the file lock while the callback runs.
func (h *Handler[T]) Modify(callback func(T) (T, error)) (T, error) {
	h.writeLock.Lock()
	defer h.writeLock.Unlock()

	var result T
	err := h.withFileLock(h.fileLock.Lock, func() error {
		current, err := h.read()
		if err != nil {
			result = h.Data()
			return err
		}
		h.setData(current)
		result = current

		data, err := callback(current)
		if err != nil {
			return err
		}
		if err := h.save(data); err != nil {
			return err
		}
		h.setData(data)
		result = data
		return nil
	})
	return result, err
}

func (h *Handler[T]) Clear() (T, error) {
	h.writeLock.Lock()
	defer h.writeLock.Unlock()

	var zero T
	if err := h.withFileLock(h.fileLock.Lock, func() error { return h.save(zero) }); err != nil {
		return h.Data(), err
	}
	h.setData(zero)
	return zero, nil
}

func (h *Handler[T]) Data() T {
	h.lock.RLock()
	defer h.lock.RUnlock()
	return h.data
}

func (h *Handler[T]) setData(data T) {
	h.lock.Lock()
	defer h.lock.Unlock()
	h.data = data
}

func (h *Handler[T]) withFileLock(lock func() error, fn func() error) (err error) {
	if err := lock(); err != nil {
		return errors.WithStack(err)
	}
	defer func() {
		if unlockErr := h.fileLock.Unlock(); unlockErr != nil && err == nil {
			err = errors.WithStack(unlockErr)
		}
	}()
	return fn()
}

// read expects the caller to hold the file lock.
func (h *Handler[T]) read() (T, error) {
	var data T

	fileInfo, err := os.Stat(h.config.FilePath)
	if os.IsNotExist(err) {
		return data, nil
	}
	if err != nil {
		return data, errors.WithStack(err)
	}
	if fileInfo.Size() == 0 {
		// Another reader that holds the shared lock may remove the empty file first.
		if err := os.Remove(h.config.FilePath); err != nil && !os.IsNotExist(err) {
			return data, errors.WithStack(err)
		}
		return data, nil
	}

	f, err := os.Open(h.config.FilePath)
	if err != nil {
		return data, errors.WithStack(err)
	}
	defer f.Close()

	if err := json.NewDecoder(f).Decode(&data); err != nil {
		return data, errors.WithMessage(err, i18n.T(i18n.UnableToDecodeJsonFile, h.config.FilePath))
	}
	return data, nil
}

// save expects the caller to hold the exclusive file lock.
// The rename replaces the file atomically, so a reader never finds the file missing or half written.
func (h *Handler[T]) save(data T) (err error) {
	f, err := os.CreateTemp(filepath.Dir(h.config.FilePath), filepath.Base(h.config.FilePath)+".*.tmp")
	if err != nil {
		return errors.WithStack(err)
	}
	defer func() {
		if err != nil {
			_ = f.Close()
			_ = os.Remove(f.Name())
		}
	}()

	if err := os.Chmod(f.Name(), h.config.FileMode); err != nil {
		return errors.WithStack(err)
	}
	if err := json.NewEncoder(f).Encode(data); err != nil {
		return errors.WithStack(err)
	}
	if err := f.Sync(); err != nil {
		return errors.WithStack(err)
	}
	if err := f.Close(); err != nil {
		return errors.WithStack(err)
	}
	if err := rename(f.Name(), h.config.FilePath); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

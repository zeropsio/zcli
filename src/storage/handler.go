package storage

import (
	"encoding/json"
	"os"
	"sync"

	"github.com/pkg/errors"

	"github.com/zeropsio/zcli/src/file"
	"github.com/zeropsio/zcli/src/i18n"
)

type Config struct {
	FilePath string
	FileMode os.FileMode
}

type Handler[T any] struct {
	//nolint:structcheck // Why: `is unused` error is false positive
	config Config
	//nolint:structcheck // Why: `is unused` error is false positive
	data T
	//nolint:structcheck // Why: `is unused` error is false positive
	lock sync.RWMutex
}

func New[T any](config Config) (*Handler[T], error) {
	h := &Handler[T]{
		config: config,
	}

	return h, h.load()
}

func (h *Handler[T]) load() error {
	fileInfo, err := os.Stat(h.config.FilePath)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return errors.WithStack(err)
	}
	if fileInfo.Size() == 0 {
		if err := os.Remove(h.config.FilePath); err != nil {
			return errors.WithStack(err)
		}
		return nil
	}

	f, err := file.Open(h.config.FilePath, os.O_RDONLY, h.config.FileMode)
	if err != nil {
		return errors.WithStack(err)
	}
	defer f.Close()

	if err := json.NewDecoder(f).Decode(&h.data); err != nil {
		return errors.WithMessagef(err, i18n.T(i18n.UnableToDecodeJsonFile, h.config.FilePath))
	}
	return nil
}

func (h *Handler[T]) Update(callback func(T) T) (T, error) {
	h.lock.Lock()
	defer h.lock.Unlock()
	h.data = callback(h.data)
	return h.data, h.save(h.data)
}

func (h *Handler[T]) save(data T) error {
	h.data = data

	if err := func() error {
		f, err := file.Open(h.config.FilePath+".new", os.O_RDWR|os.O_CREATE|os.O_TRUNC, h.config.FileMode)
		if err != nil {
			return errors.WithStack(err)
		}
		defer f.Close()

		if err := json.NewEncoder(f).Encode(data); err != nil {
			return errors.WithStack(err)
		}
		return nil
	}(); err != nil {
		return err
	}
	os.Remove(h.config.FilePath)
	defer os.Remove(h.config.FilePath + ".new")
	if err := os.Rename(h.config.FilePath+".new", h.config.FilePath); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func (h *Handler[T]) Data() T {
	h.lock.RLock()
	defer h.lock.RUnlock()
	return h.data
}

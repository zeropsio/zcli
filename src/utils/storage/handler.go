package storage

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"

	"github.com/zeropsio/zcli/src/utils"
)

type Config struct {
	FilePath string
}

type Handler[T any] struct {
	config Config
	data   T
	lock   sync.RWMutex
}

func New[T any](config Config) (*Handler[T], error) {
	h := &Handler[T]{
		config: config,
	}

	return h, h.load()
}

func (h *Handler[T]) load() error {

	dir := filepath.Dir(h.config.FilePath)
	err := os.MkdirAll(dir, 0700)
	if err != nil {
		return err
	}

	storageFileExists, err := utils.FileExists(h.config.FilePath)
	if err != nil {
		return err
	}
	if !storageFileExists {
		return nil
	}

	f, err := os.Open(h.config.FilePath)
	if err != nil {
		return err
	}
	defer f.Close()

	if err := json.NewDecoder(f).Decode(&h.data); err != nil {
		return err
	}

	return nil
}

func (h *Handler[T]) Clear() error {
	h.lock.Lock()
	defer h.lock.Unlock()
	var data T
	return h.save(data)
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
		f, err := os.Create(h.config.FilePath + ".new")
		if err != nil {
			return err
		}
		defer f.Close()

		if err := json.NewEncoder(f).Encode(data); err != nil {
			return err
		}
		return nil
	}(); err != nil {
		return err
	}
	if err := os.Rename(h.config.FilePath+".new", h.config.FilePath); err != nil {
		return err
	}
	os.Remove(h.config.FilePath + ".new")
	return nil
}

func (h *Handler[T]) Data() T {
	h.lock.RLock()
	defer h.lock.RUnlock()
	return h.data
}

package storage

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"

	"github.com/zerops-io/zcli/src/utils"
)

type Config struct {
	FilePath string
}

type Handler[T any] struct {
	config Config

	lock sync.Mutex
}

func New[T any](config Config) (*Handler[T], error) {
	h := &Handler[T]{
		config: config,
	}

	dir := filepath.Dir(config.FilePath)
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return nil, err
	}

	f, err := os.OpenFile(config.FilePath, os.O_CREATE, 0755)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return h, nil
}

func (h *Handler[T]) Load() *T {
	h.lock.Lock()
	defer h.lock.Unlock()

	var data T

	storageFileExists, err := utils.FileExists(h.config.FilePath)
	if err != nil {
		return &data
	}

	if storageFileExists {
		err := func() error {
			f, err := os.Open(h.config.FilePath)
			if err != nil {
				return err
			}
			defer f.Close()

			bytes, err := ioutil.ReadAll(f)
			if err != nil {
				return err
			}

			err = json.Unmarshal(bytes, &data)
			if err != nil {
				return err
			}

			return nil
		}()
		if err == nil {
			return &data
		}
	}

	return &data
}

func (h *Handler[T]) Save(data *T) error {
	h.lock.Lock()
	defer h.lock.Unlock()

	dataBytes, err := json.Marshal(data)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(h.config.FilePath, dataBytes, 0644)
	if err != nil {
		return err
	}

	return nil
}

func (h *Handler[T]) Data() *T {
	return h.Load()
}

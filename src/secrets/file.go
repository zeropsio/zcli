package secrets

import (
	"encoding/json"
	"os"
	"strings"
	"sync"

	"github.com/pkg/errors"

	"github.com/zeropsio/zcli/src/file"
)

// fileBackend is a plaintext JSON file persisted at config.FilePath. The whole map is loaded
// on open and held in memory; every mutation rewrites the file atomically (write to .new,
// rename). Same pattern as storage.Handler so the failure modes match.
type fileBackend struct {
	mu       sync.RWMutex
	path     string
	fileMode os.FileMode
	m        map[string]string
}

func newFileBackend(path string, fileMode os.FileMode) (*fileBackend, error) {
	b := &fileBackend{
		path:     path,
		fileMode: fileMode,
		m:        map[string]string{},
	}
	if err := b.load(); err != nil {
		return nil, err
	}
	return b, nil
}

func (b *fileBackend) load() error {
	info, err := os.Stat(b.path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return errors.WithStack(err)
	}
	if info.Size() == 0 {
		return nil
	}
	f, err := file.Open(b.path, os.O_RDONLY, b.fileMode)
	if err != nil {
		return errors.WithStack(err)
	}
	defer f.Close()
	if err := json.NewDecoder(f).Decode(&b.m); err != nil {
		return errors.Wrapf(err, "decode secrets file %s", b.path)
	}
	if b.m == nil {
		b.m = map[string]string{}
	}
	return nil
}

func (b *fileBackend) save() error {
	tmp := b.path + ".new"
	if err := func() error {
		f, err := file.Open(tmp, os.O_RDWR|os.O_CREATE|os.O_TRUNC, b.fileMode)
		if err != nil {
			return errors.WithStack(err)
		}
		defer f.Close()
		if err := json.NewEncoder(f).Encode(b.m); err != nil {
			return errors.WithStack(err)
		}
		return nil
	}(); err != nil {
		return err
	}
	os.Remove(b.path)
	defer os.Remove(tmp)
	if err := os.Rename(tmp, b.path); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func (b *fileBackend) Get(key string) (string, bool, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	v, ok := b.m[key]
	return v, ok, nil
}

func (b *fileBackend) Set(key, value string) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.m[key] = value
	return b.save()
}

func (b *fileBackend) Delete(key string) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if _, ok := b.m[key]; !ok {
		return nil
	}
	delete(b.m, key)
	return b.save()
}

func (b *fileBackend) Description() string { return "plaintext file at " + b.path }

func (b *fileBackend) List(prefix string) ([]string, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	out := make([]string, 0)
	for k := range b.m {
		if strings.HasPrefix(k, prefix) {
			out = append(out, k)
		}
	}
	return out, nil
}

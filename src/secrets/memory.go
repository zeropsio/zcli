package secrets

import (
	"strings"
	"sync"
)

// memoryBackend is an in-memory Backend for tests. Exported via NewMemory.
type memoryBackend struct {
	mu sync.RWMutex
	m  map[string]string
}

// NewMemory returns an empty in-memory Backend. Safe for concurrent use.
func NewMemory() Backend {
	return &memoryBackend{m: map[string]string{}}
}

func (b *memoryBackend) Get(key string) (string, bool, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	v, ok := b.m[key]
	return v, ok, nil
}

func (b *memoryBackend) Set(key, value string) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.m[key] = value
	return nil
}

func (b *memoryBackend) Delete(key string) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	delete(b.m, key)
	return nil
}

func (b *memoryBackend) List(prefix string) ([]string, error) {
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

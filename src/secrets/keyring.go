package secrets

import (
	"strings"

	keyringlib "github.com/99designs/keyring"
	"github.com/pkg/errors"
)

// keyringBackend wraps the OS keyring via 99designs/keyring. The library opens the ring lazily
// per call, which is fine for a CLI with bursty access patterns.
type keyringBackend struct {
	ring keyringlib.Keyring
}

func newKeyringBackend(service string) (*keyringBackend, error) {
	ring, err := keyringlib.Open(keyringlib.Config{
		ServiceName:              service,
		KeychainName:             service,
		KeychainTrustApplication: true,
		// Exclude the library's own file backend so a misconfigured machine doesn't silently
		// fall through to a library-managed encrypted file. Our auto-mode handles the fallback
		// explicitly via fileBackend.
		AllowedBackends: osKeyringBackends,
	})
	if err != nil {
		if errors.Is(err, keyringlib.ErrNoAvailImpl) {
			return nil, ErrKeyringUnavailable
		}
		return nil, errors.Wrap(err, "open keyring")
	}
	return &keyringBackend{ring: ring}, nil
}

func (b *keyringBackend) Get(key string) (string, bool, error) {
	item, err := b.ring.Get(key)
	if err != nil {
		if errors.Is(err, keyringlib.ErrKeyNotFound) {
			return "", false, nil
		}
		return "", false, translateKeyringErr(err)
	}
	return string(item.Data), true, nil
}

func (b *keyringBackend) Set(key, value string) error {
	return translateKeyringErr(b.ring.Set(keyringlib.Item{
		Key:  key,
		Data: []byte(value),
	}))
}

func (b *keyringBackend) Delete(key string) error {
	err := b.ring.Remove(key)
	if err != nil && errors.Is(err, keyringlib.ErrKeyNotFound) {
		return nil
	}
	return translateKeyringErr(err)
}

func (b *keyringBackend) List(prefix string) ([]string, error) {
	all, err := b.ring.Keys()
	if err != nil {
		return nil, translateKeyringErr(err)
	}
	out := make([]string, 0, len(all))
	for _, k := range all {
		if strings.HasPrefix(k, prefix) {
			out = append(out, k)
		}
	}
	return out, nil
}

func translateKeyringErr(err error) error {
	if err == nil {
		return nil
	}
	// The library doesn't expose a typed "locked" sentinel; some platforms surface lock state
	// through opaque OS errors. Best-effort string match for the macOS case; otherwise wrap as-is.
	if msg := err.Error(); strings.Contains(msg, "locked") || strings.Contains(msg, "Locked") {
		return ErrLocked
	}
	return errors.WithStack(err)
}

// osKeyringBackends is the allowlist of OS-native keyring backends. The library's file/pass
// backends are excluded - those would conflict with our explicit fileBackend / plaintext mode
// distinction.
var osKeyringBackends = []keyringlib.BackendType{
	keyringlib.KeychainBackend,
	keyringlib.SecretServiceBackend,
	keyringlib.KWalletBackend,
	keyringlib.WinCredBackend,
}

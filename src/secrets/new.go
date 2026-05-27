package secrets

import (
	"slices"

	keyringlib "github.com/99designs/keyring"
	"github.com/pkg/errors"
)

// New constructs a Backend according to cfg.Mode. See the Mode constants for resolution rules.
//
// For ModeAuto: probes whether any OS keyring backend is supported on this host via the
// 99designs/keyring library's AvailableBackends. If one is present, uses it; otherwise falls
// back to the plaintext file backend and emits a one-time notice via cfg.Log.
func New(cfg Config) (Backend, error) {
	switch cfg.Mode {
	case ModeKeyring:
		return newKeyringBackend(cfg.Service)
	case ModePlaintext:
		if cfg.FilePath == "" {
			return nil, errors.New("secrets: ModePlaintext requires Config.FilePath")
		}
		return newFileBackend(cfg.FilePath, cfg.FileMode)
	case ModeAuto, "":
		if HasOSKeyring() {
			return newKeyringBackend(cfg.Service)
		}
		if cfg.Log != nil {
			cfg.Log("Keyring backend unavailable on this system, secrets will be stored in plaintext at " + cfg.FilePath + " - set secrets_backend in your profile to silence this notice")
		}
		if cfg.FilePath == "" {
			return nil, errors.New("secrets: ModeAuto fallback requires Config.FilePath when keyring is unavailable")
		}
		return newFileBackend(cfg.FilePath, cfg.FileMode)
	default:
		return nil, errors.Errorf("secrets: unknown mode %q", cfg.Mode)
	}
}

// HasOSKeyring reports whether the host has any OS-native keyring backend available. Used by
// ModeAuto internally to decide whether to attempt the keyring path, and exposed so callers
// can decide before calling New whether to prompt the user for a backend preference.
func HasOSKeyring() bool {
	for _, b := range keyringlib.AvailableBackends() {
		if slices.Contains(osKeyringBackends, b) {
			return true
		}
	}
	return false
}

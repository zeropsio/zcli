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
		if hasOSKeyring() {
			return newKeyringBackend(cfg.Service)
		}
		if cfg.Log != nil {
			cfg.Log("keyring backend unavailable on this system, secrets will be stored in plaintext at " + cfg.FilePath + " - set secrets_backend in your profile to silence this notice")
		}
		if cfg.FilePath == "" {
			return nil, errors.New("secrets: ModeAuto fallback requires Config.FilePath when keyring is unavailable")
		}
		return newFileBackend(cfg.FilePath, cfg.FileMode)
	default:
		return nil, errors.Errorf("secrets: unknown mode %q", cfg.Mode)
	}
}

// hasOSKeyring reports whether the host has any OS-native keyring backend available. Used by
// ModeAuto to decide whether to attempt the keyring path at all.
func hasOSKeyring() bool {
	for _, b := range keyringlib.AvailableBackends() {
		if slices.Contains(osKeyringBackends, b) {
			return true
		}
	}
	return false
}

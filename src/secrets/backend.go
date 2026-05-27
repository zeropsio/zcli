// Package secrets is an opaque key-value store for sensitive values (auth tokens, VPN private
// keys) that the CLI persists between runs. Backends decide where the values actually live -
// the OS keyring on macOS/Linux/Windows, a plaintext JSON file (CI / headless environments,
// or explicit user preference), or an in-memory map (tests).
//
// Profile data (region, scoped project, etc.) lives in src/profile, in plaintext JSON files.
// Anything that would be a problem to leak goes through this package instead, so the profile
// JSON is safe to share, version-control, copy between machines.
//
// Keys follow a flat namespace with conventional prefixes - see keys.go for the Scope helpers.
// Callers should construct keys via secrets.Key, not by string concatenation, so the layout can
// evolve without sed-ing string literals across the codebase.
package secrets

import (
	"errors"
	"os"
)

// Backend is the storage abstraction. Implementations are constructed via New and are safe for
// concurrent use unless an implementation's own documentation says otherwise.
type Backend interface {
	// Get returns the value for key. The bool reports presence: (value, true, nil) on hit,
	// ("", false, nil) on miss, ("", false, err) on backend failure. Miss is not an error
	// because "is this token set?" is a normal control-flow question, not an exception.
	Get(key string) (string, bool, error)

	// Set writes value for key, overwriting any existing value. Empty values are stored as-is
	// rather than treated as Delete; callers that mean "delete" should call Delete.
	Set(key, value string) error

	// Delete removes the value for key. Idempotent: deleting a missing key is not an error,
	// so callers can call it unconditionally during cleanup paths.
	Delete(key string) error

	// List returns all stored keys whose name has the given prefix. Used by cleanup paths
	// (e.g. "drop all vpn keys for projects no longer in any profile"). Order is unspecified.
	List(prefix string) ([]string, error)
}

// Mode selects how the factory builds a Backend. Modes are resolved in this order at construction
// time: explicit flag > env var > config file field > Auto default.
type Mode string

const (
	// ModeAuto picks the OS keyring if any is available on this machine, otherwise falls back
	// to the plaintext file backend. The fallback is logged once via Config.Log.
	ModeAuto Mode = "auto"
	// ModeKeyring picks the OS keyring only. New returns ErrKeyringUnavailable if no keyring
	// backend is supported on the host.
	ModeKeyring Mode = "keyring"
	// ModePlaintext picks the plaintext file backend unconditionally. The OS keyring is never
	// touched, which is the right choice for CI / containers / users with sync-based portability.
	ModePlaintext Mode = "plaintext"
)

// Config configures the factory. FilePath is required for ModePlaintext or ModeAuto fallback,
// Service is required for ModeKeyring or ModeAuto success.
type Config struct {
	Mode     Mode
	Service  string      // keyring service name, e.g. "zerops-cli"
	FilePath string      // plaintext fallback location
	FileMode os.FileMode // permissions for FilePath, typically 0o600
	Log      func(string) // one-time warnings (fallback notice); nil = silent
}

var (
	// ErrKeyringUnavailable: ModeKeyring asked for but no OS keychain found on this host.
	// Returned only by New, not by Get/Set/Delete on a valid backend.
	ErrKeyringUnavailable = errors.New("secrets: keyring backend not available on this system")
	// ErrLocked: the keyring is present but locked (e.g. macOS Login keychain not unlocked).
	// Callers can prompt the user to unlock and retry.
	ErrLocked = errors.New("secrets: keyring is locked")
)

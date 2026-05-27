// Package profile is the CLI's persisted per-user state: active region, scoped project, and
// the user's preferred secrets backend. Backed by a JSON file on disk; sensitive values
// (auth token, VPN private keys) are NOT in this file - they live in the secrets package's
// chosen backend (keyring or plaintext fallback) so the profile JSON is safe to share,
// version-control, or copy between machines.
//
// # Layout
//
// Each historical schema lives in its own data_vN.go file as a pure struct (dataV0, dataV1, ...).
// The file registers a decoder for its version via init(). If the struct is not the current
// target it also provides an Up(ctx) method returning the next version's value, with access to
// a migration context (the secrets backend) so legacy plaintext fields can be drained on the
// way up. The current shape is reached via the dataCurrent alias and exposed to callers as the
// distinct Data type.
//
// # Reads and writes
//
// Manager.load probes the on-disk "version" key (missing means v0), decodes into the matching
// dataVN, then walks Up(ctx) until reaching dataCurrent. If the source file was at an earlier
// version, the migrated form is saved back so the file is always at the current schema after
// the first run on this machine.
//
// Sensitive accessors (Token, SetToken, VpnKey, SetVpnKey, DeleteVpnKey) delegate to the
// secrets backend. Non-sensitive accessors (Data, Update, Clear) operate on the on-disk JSON.
//
// # Adding a new migration
//
//  1. Create data_vN.go with the new struct and its init() decoder.
//  2. Add an Up(ctx) method on the previous version's struct returning the new dataVN. Use
//     ctx.secrets to drain any newly-sensitive fields, or to pull existing secrets if the
//     field types change. Up is free to reshape fields, drop them, or compute defaults.
//  3. In handler.go bump CurrentDataVersion to N and flip the dataCurrent alias from
//     dataV(N-1) to dataVN. Data follows automatically through the alias.
//
// Add a migration test in handler_test.go that loads a representative v(N-1) file plus
// matching secrets state, and asserts both end up where they should after migration.
package profile

import (
	"encoding/json"
	"os"
	"sync"

	"github.com/pkg/errors"

	"github.com/zeropsio/zcli/src/entity"
	"github.com/zeropsio/zcli/src/file"
	"github.com/zeropsio/zcli/src/secrets"
	"github.com/zeropsio/zerops-go/types/uuid"
)

const CurrentDataVersion = 1

// dataCurrent is an alias for the latest dataVN. Bumping a version means changing only this
// line: Data and the terminal type assertion in Manager.load resolve through it.
type dataCurrent = dataV1

// Data is the non-sensitive on-disk shape. Sensitive values live in the secrets backend and
// are accessed via the Token / VpnKey families on Manager, never via Data.
type Data dataCurrent

// migrationCtx is passed to each Up step. Keeping it as a struct (not a bag of arguments)
// means future migrations can extend the context without changing every existing Up signature.
type migrationCtx struct {
	secrets secrets.Backend
}

// upgrader is implemented by every dataVN that is NOT the current target. Up returns the next
// version's value, type-erased because each version is its own Go type. Up may fail if a
// migration's side effects fail (e.g. the secrets backend is unwriteable).
type upgrader interface {
	Up(ctx migrationCtx) (any, error)
}

var registry = map[int]func([]byte) (any, error){}

func register(version int, decode func([]byte) (any, error)) {
	if _, dup := registry[version]; dup {
		panic(errors.Errorf("profile: duplicate registration for version %d", version))
	}
	registry[version] = decode
}

// Config configures a Manager. Secrets is required; without it tokens and VPN keys can't be
// stored. ProfilePath is the JSON file holding non-sensitive state; missing-file is fine
// (treated as empty state).
type Config struct {
	ProfilePath string
	ProfileMode os.FileMode
	Secrets     secrets.Backend
}

// Manager is the runtime handle on a profile. It composes the on-disk JSON (non-sensitive
// state) with a secrets backend (sensitive state). The combination is presented as one
// coherent API so callers don't have to know which storage layer holds which field.
type Manager struct {
	cfg     Config
	secrets secrets.Backend

	mu   sync.RWMutex
	data Data
}

// NewManager loads the profile file (migrating from an older schema if necessary, draining
// any plaintext secrets it finds into cfg.Secrets) and returns a ready-to-use Manager.
func NewManager(cfg Config) (*Manager, error) {
	if cfg.Secrets == nil {
		return nil, errors.New("profile: Config.Secrets is required")
	}
	m := &Manager{cfg: cfg, secrets: cfg.Secrets}
	if err := m.load(); err != nil {
		return nil, err
	}
	return m, nil
}

func (m *Manager) load() error {
	info, err := os.Stat(m.cfg.ProfilePath)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return errors.WithStack(err)
	}
	if info.Size() == 0 {
		return nil
	}

	raw, err := os.ReadFile(m.cfg.ProfilePath)
	if err != nil {
		return errors.WithStack(err)
	}

	var probe struct {
		Version *int `json:"version,omitempty"`
	}
	if err := json.Unmarshal(raw, &probe); err != nil {
		return errors.Wrapf(err, "decode profile %s", m.cfg.ProfilePath)
	}
	sourceVersion := 0
	if probe.Version != nil {
		sourceVersion = *probe.Version
	}

	decode, ok := registry[sourceVersion]
	if !ok {
		return errors.Errorf("profile: no decoder for on-disk version %d", sourceVersion)
	}
	current, err := decode(raw)
	if err != nil {
		return errors.Wrapf(err, "decode profile %s as version %d", m.cfg.ProfilePath, sourceVersion)
	}

	ctx := migrationCtx{secrets: m.secrets}
	for {
		if final, ok := current.(dataCurrent); ok {
			v := CurrentDataVersion
			final.Version = &v
			m.data = Data(final)
			break
		}
		up, ok := current.(upgrader)
		if !ok {
			return errors.Errorf("profile: %T is not the current version and has no Up()", current)
		}
		next, err := up.Up(ctx)
		if err != nil {
			return err
		}
		current = next
	}

	// If we migrated from a legacy version, rewrite the file in the current schema so the
	// next run skips the migration path entirely.
	if sourceVersion < CurrentDataVersion {
		if err := m.save(m.data); err != nil {
			return errors.Wrap(err, "rewrite profile in current schema after migration")
		}
	}
	return nil
}

func (m *Manager) save(d Data) error {
	v := CurrentDataVersion
	out := dataCurrent(d)
	out.Version = &v
	payload, err := json.Marshal(out)
	if err != nil {
		return errors.WithStack(err)
	}

	tmp := m.cfg.ProfilePath + ".new"
	if err := func() error {
		f, err := file.Open(tmp, os.O_RDWR|os.O_CREATE|os.O_TRUNC, m.cfg.ProfileMode)
		if err != nil {
			return errors.WithStack(err)
		}
		defer f.Close()
		if _, err := f.Write(payload); err != nil {
			return errors.WithStack(err)
		}
		return nil
	}(); err != nil {
		return err
	}
	os.Remove(m.cfg.ProfilePath)
	defer os.Remove(tmp)
	if err := os.Rename(tmp, m.cfg.ProfilePath); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// Data returns a snapshot of the non-sensitive profile state.
func (m *Manager) Data() Data {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.data
}

// Update mutates the non-sensitive profile state via callback and persists the result.
func (m *Manager) Update(cb func(Data) Data) (Data, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data = cb(m.data)
	return m.data, m.save(m.data)
}

// Clear wipes both the profile file and every secret under the global scope (token, all VPN
// keys). Used by `zcli logout`. Returns the prior Data so callers can log what was removed.
func (m *Manager) Clear() (Data, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	prior := m.data
	m.data = Data{}
	if err := m.save(m.data); err != nil {
		return prior, err
	}
	keys, err := m.secrets.List(secrets.Prefix(secrets.GlobalScope))
	if err != nil {
		return prior, errors.Wrap(err, "list global secrets")
	}
	for _, k := range keys {
		if err := m.secrets.Delete(k); err != nil {
			return prior, errors.Wrapf(err, "delete secret %s", k)
		}
	}
	return prior, nil
}

// Token returns the stored auth token. Returns ("", false, nil) when not set.
func (m *Manager) Token() (string, bool, error) {
	return m.secrets.Get(secrets.Key(secrets.GlobalScope, "token"))
}

// SetToken writes the auth token to the secrets backend.
func (m *Manager) SetToken(token string) error {
	return m.secrets.Set(secrets.Key(secrets.GlobalScope, "token"), token)
}

// VpnKey returns the cached VPN key for a project. Returns (zero, false, nil) when not set.
func (m *Manager) VpnKey(projectId uuid.ProjectId) (entity.VpnKey, bool, error) {
	raw, ok, err := m.secrets.Get(vpnKeyPath(projectId))
	if err != nil || !ok {
		return entity.VpnKey{}, ok, err
	}
	var k entity.VpnKey
	if err := json.Unmarshal([]byte(raw), &k); err != nil {
		return entity.VpnKey{}, false, errors.Wrapf(err, "decode cached vpn key for project %s", projectId.Native())
	}
	return k, true, nil
}

// SetVpnKey writes (or overwrites) the cached VPN key for a project.
func (m *Manager) SetVpnKey(projectId uuid.ProjectId, key entity.VpnKey) error {
	payload, err := json.Marshal(key)
	if err != nil {
		return errors.WithStack(err)
	}
	return m.secrets.Set(vpnKeyPath(projectId), string(payload))
}

// DeleteVpnKey removes the cached VPN key for a project. Idempotent.
func (m *Manager) DeleteVpnKey(projectId uuid.ProjectId) error {
	return m.secrets.Delete(vpnKeyPath(projectId))
}

func vpnKeyPath(projectId uuid.ProjectId) string {
	return secrets.Key(secrets.GlobalScope, "vpn", projectId.Native())
}

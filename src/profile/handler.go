// Package profile is the CLI's persisted per-user state. Composed of two layers:
//
//   - Global profile: token / region / preferred secrets backend. Held in a JSON file at the
//     user's CLI config path. Sensitive values (token, VPN private keys) are not in the JSON;
//     they live in the secrets package's chosen backend.
//   - Repo profile (optional): per-repository overrides. Held in <repo>/.zcli/profile.json
//     when initialised. Carries the scoped project for this repo, an optional region override,
//     and a stable Id used to scope per-repo secrets in the secrets backend.
//
// Effective state is the field-level merge of (global, repo). The repo layer wins per field;
// fields absent from the repo layer fall through to global.
//
// # Layout
//
// Global schema versions live in data_v0.go, data_v1.go, data_v2.go, ... Each registers a
// decoder via init() and (when not the current target) provides an Up(ctx) method returning
// the next version. dataCurrent aliases the latest. Repo schema is currently a single shape
// in data_repo_v1.go - the same versioned framework lands there when a v2 is needed.
//
// # Reads and writes
//
// Manager.load reads both layers (global always, repo when discovered via DiscoverRepoProfileDir),
// migrates each up to its current schema, and exposes the merged effective state via Region,
// Token, ScopeProject getters. Writes are explicit per layer: UpdateGlobal vs UpdateRepo, plus
// convenience methods like SetScopeProject (always repo) and SetToken (target argument).
//
// # Adding a new global migration
//
//  1. Create data_vN.go with the new struct and its init() decoder.
//  2. Add Up(ctx) on the previous version's struct returning the new dataVN. Use ctx.secrets
//     for sensitive-field moves, ctx.rescueScope for state that moves to the repo layer,
//     ctx.log for one-time user notices.
//  3. In handler.go bump CurrentDataVersion to N and flip the dataCurrent alias from
//     dataV(N-1) to dataVN. Data follows automatically through the alias.
package profile

import (
	"encoding/json"
	"os"
	"strconv"
	"sync"

	"github.com/google/uuid"
	"github.com/pkg/errors"

	"github.com/zeropsio/zcli/src/entity"
	"github.com/zeropsio/zcli/src/file"
	"github.com/zeropsio/zcli/src/region"
	"github.com/zeropsio/zcli/src/secrets"
	zuuid "github.com/zeropsio/zerops-go/types/uuid"
)

const CurrentDataVersion = 2

// dataCurrent aliases the latest dataVN. Bumping a version means changing only this line.
type dataCurrent = dataV2

// Data is the current global on-disk shape. Distinct type over dataCurrent so any future
// methods hang here, not on the underlying versioned struct.
type Data dataCurrent

// migrationCtx is passed to each Up step. Migrations can drain to the secrets backend, rescue
// scope into a repo profile, or emit user-facing notices via log.
type migrationCtx struct {
	secrets     secrets.Backend
	cwd         string
	rescueScope func(zuuid.ProjectIdNull) (bool, error)
	log         func(string)
}

// upgrader is implemented by every dataVN that is NOT the current target. Up returns the next
// version's value, type-erased because each version is its own Go type.
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

// Config configures a Manager.
type Config struct {
	// GlobalPath is the JSON file holding non-sensitive global state. Required.
	GlobalPath string
	// GlobalMode is the file mode for GlobalPath, typically 0o600.
	GlobalMode os.FileMode
	// Cwd is the starting directory for repo profile discovery. Defaults to os.Getwd().
	Cwd string
	// Secrets is the secrets backend. Required.
	Secrets secrets.Backend
	// Log receives one-time user notices from migrations (e.g. "previous scope cleared").
	// May be nil; messages are dropped if so.
	Log func(string)
}

// Manager composes the global JSON layer, the optional repo JSON layer, and the secrets
// backend into one coherent API. Callers don't have to know which layer holds which field.
type Manager struct {
	cfg     Config
	secrets secrets.Backend

	mu       sync.RWMutex
	global   Data
	repoPath string   // "" when no repo profile was discovered
	repo     RepoData // zero when repoPath is ""
}

// NewManager loads both layers (migrating global from older schemas as needed, draining any
// plaintext secrets it finds) and returns a ready-to-use Manager.
func NewManager(cfg Config) (*Manager, error) {
	if cfg.Secrets == nil {
		return nil, errors.New("profile: Config.Secrets is required")
	}
	if cfg.Cwd == "" {
		wd, err := os.Getwd()
		if err != nil {
			return nil, errors.Wrap(err, "resolve cwd for profile discovery")
		}
		cfg.Cwd = wd
	}

	m := &Manager{cfg: cfg, secrets: cfg.Secrets}
	if err := m.loadGlobal(); err != nil {
		return nil, err
	}
	if err := m.loadRepo(); err != nil {
		return nil, err
	}
	return m, nil
}

func (m *Manager) loadGlobal() error {
	info, err := os.Stat(m.cfg.GlobalPath)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return errors.WithStack(err)
	}
	if info.Size() == 0 {
		return nil
	}

	raw, err := os.ReadFile(m.cfg.GlobalPath)
	if err != nil {
		return errors.WithStack(err)
	}

	var probe struct {
		Version *int `json:"version,omitempty"`
	}
	if err := json.Unmarshal(raw, &probe); err != nil {
		return errors.Wrapf(err, "decode global profile %s", m.cfg.GlobalPath)
	}
	sourceVersion := 0
	if probe.Version != nil {
		sourceVersion = *probe.Version
	}

	decode, ok := registry[sourceVersion]
	if !ok {
		return errors.Errorf("profile: no decoder for on-disk global version %d", sourceVersion)
	}
	current, err := decode(raw)
	if err != nil {
		return errors.Wrapf(err, "decode global profile %s as version %d", m.cfg.GlobalPath, sourceVersion)
	}

	ctx := migrationCtx{
		secrets:     m.secrets,
		cwd:         m.cfg.Cwd,
		rescueScope: m.rescueScope,
		log:         m.cfg.Log,
	}
	for {
		if final, ok := current.(dataCurrent); ok {
			v := CurrentDataVersion
			final.Version = &v
			m.global = Data(final)
			break
		}
		up, ok := current.(upgrader)
		if !ok {
			return errors.Errorf("profile: %T is not the current global version and has no Up()", current)
		}
		next, err := up.Up(ctx)
		if err != nil {
			return err
		}
		current = next
	}

	if sourceVersion < CurrentDataVersion {
		if err := backupProfile(m.cfg.GlobalPath, raw, sourceVersion, m.cfg.GlobalMode); err != nil {
			return errors.Wrap(err, "backup global profile before migration")
		}
		if err := m.saveGlobal(m.global); err != nil {
			return errors.Wrap(err, "rewrite global profile in current schema after migration")
		}
	}
	return nil
}

// backupProfile writes the original on-disk bytes to a sibling file named
// <path>.v<sourceVersion>.bak before the migration overwrites the live file. Lets users
// recover by `mv cli.data.v0.bak cli.data` if a future code change ever regresses on a
// migration we already ran. Overwrites an existing same-named backup, on the theory that the
// most-recent original wins; the live file is still preserved as a backup, just not older
// generations.
func backupProfile(livePath string, original []byte, sourceVersion int, fileMode os.FileMode) error {
	backupPath := livePath + ".v" + strconv.Itoa(sourceVersion) + ".bak"
	if err := os.WriteFile(backupPath, original, fileMode); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// rescueScope is the migration callback: when a legacy global scoped project ID needs to move
// to the per-repo layer, this writes it into a fresh repo profile if cwd is inside a git repo
// without an existing repo profile. Returns (true, nil) when the rescue actually happened.
func (m *Manager) rescueScope(scope zuuid.ProjectIdNull) (bool, error) {
	repoRoot, err := DiscoverRepoRoot(m.cfg.Cwd)
	if err != nil {
		return false, err
	}
	if repoRoot == "" {
		return false, nil
	}
	path := RepoProfilePath(repoRoot)
	if _, statErr := os.Stat(path); statErr == nil {
		return false, nil
	} else if !os.IsNotExist(statErr) {
		return false, statErr
	}
	rd := RepoData{
		Id:             uuid.NewString(),
		ScopeProjectId: scope,
	}
	if err := writeRepoProfile(path, m.cfg.GlobalMode, rd); err != nil {
		return false, err
	}
	m.repoPath = repoRoot
	m.repo = rd
	if m.cfg.Log != nil {
		m.cfg.Log("scoped project moved into " + path + " (run `git ignore .zcli/` or add .zcli/ to .gitignore to keep secrets out of source control)")
	}
	return true, nil
}

func (m *Manager) loadRepo() error {
	// If rescueScope already populated the repo state during a v1 → v2 migration, don't load
	// over it - the file we just wrote IS the source of truth.
	if m.repoPath != "" {
		return nil
	}
	repoDir, err := DiscoverRepoProfileDir(m.cfg.Cwd)
	if err != nil {
		return err
	}
	if repoDir == "" {
		return nil
	}
	path := RepoProfilePath(repoDir)
	raw, err := os.ReadFile(path)
	if err != nil {
		return errors.WithStack(err)
	}
	var rd RepoData
	if err := json.Unmarshal(raw, &rd); err != nil {
		return errors.Wrapf(err, "decode repo profile %s", path)
	}
	if rd.Id == "" {
		return errors.Errorf("profile: repo profile at %s is missing required \"id\" field", path)
	}
	m.repoPath = repoDir
	m.repo = rd
	return nil
}

func (m *Manager) saveGlobal(d Data) error {
	v := CurrentDataVersion
	out := dataCurrent(d)
	out.Version = &v
	return writeJSONAtomic(m.cfg.GlobalPath, m.cfg.GlobalMode, out)
}

func (m *Manager) saveRepo(d RepoData) error {
	if m.repoPath == "" {
		return errors.New("profile: no repo profile is initialised in this directory tree")
	}
	v := CurrentRepoVersion
	out := repoCurrent(d)
	out.Version = &v
	return writeJSONAtomic(RepoProfilePath(m.repoPath), m.cfg.GlobalMode, out)
}

func writeRepoProfile(path string, fileMode os.FileMode, d RepoData) error {
	if err := os.MkdirAll(filepathDir(path), 0o755); err != nil {
		return errors.WithStack(err)
	}
	v := CurrentRepoVersion
	out := repoCurrent(d)
	out.Version = &v
	return writeJSONAtomic(path, fileMode, out)
}

func writeJSONAtomic(path string, fileMode os.FileMode, value any) error {
	payload, err := json.Marshal(value)
	if err != nil {
		return errors.WithStack(err)
	}
	tmp := path + ".new"
	if err := func() error {
		f, err := file.Open(tmp, os.O_RDWR|os.O_CREATE|os.O_TRUNC, fileMode)
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
	os.Remove(path)
	defer os.Remove(tmp)
	if err := os.Rename(tmp, path); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// Data returns a snapshot of the global non-sensitive state. Most callers should use the
// merged-aware getters (Region, ScopeProject, Token) instead - Data exposes only the global
// layer and won't reflect repo-layer overrides.
func (m *Manager) Data() Data {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.global
}

// HasRepo reports whether a repo profile was discovered for this Manager. False means the
// caller is operating purely on global state.
func (m *Manager) HasRepo() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.repoPath != ""
}

// RepoProfile returns the loaded repo profile and a bool reporting whether one was discovered.
// Used by introspection surfaces (e.g. `zcli profile where`); for normal field access prefer
// the merged getters like Region / ScopeProject so callers don't have to handle the layering.
func (m *Manager) RepoProfile() (RepoData, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.repoPath == "" {
		return RepoData{}, false
	}
	return m.repo, true
}

// GlobalPath returns the on-disk path of the global profile file. Used by introspection.
func (m *Manager) GlobalPath() string { return m.cfg.GlobalPath }

// SecretsDescription returns a human-readable label for the secrets backend in use. Used by
// introspection surfaces; the actual Backend isn't exposed to keep the API surface small.
func (m *Manager) SecretsDescription() string { return m.secrets.Description() }

// RepoDir returns the directory containing the active repo profile, or "" when no repo
// profile was discovered.
func (m *Manager) RepoDir() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.repoPath
}

// Region returns the effective region: repo override when present, else global.
func (m *Manager) Region() region.Item {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.repoPath != "" && m.repo.Region != nil {
		return *m.repo.Region
	}
	return m.global.Region
}

// ScopeProject returns the scoped project id from the repo profile (if any). There is no
// global fallback by design: scope is per-repo. Returns (zero, false) outside a repo profile
// or when no scope is set.
func (m *Manager) ScopeProject() (zuuid.ProjectId, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.repoPath == "" || !m.repo.ScopeProjectId.Filled() {
		return "", false
	}
	id, _ := m.repo.ScopeProjectId.Get()
	return id, true
}

// SetScopeProject persists a scoped project to the repo profile. Errors when no repo profile
// is initialised in the current directory tree - call InitRepo first or run from inside a
// repo that already has one.
func (m *Manager) SetScopeProject(id zuuid.ProjectIdNull) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.repoPath == "" {
		return errors.New("no repo profile in this directory tree; run `zcli profile init` here first")
	}
	m.repo.ScopeProjectId = id
	return m.saveRepo(m.repo)
}

// InitRepo creates a fresh repo profile rooted at the nearest .git ancestor of cwd. Errors
// when cwd is not inside a git repo or a repo profile already exists. Returns the directory
// the profile was created in.
func (m *Manager) InitRepo() (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.repoPath != "" {
		return "", errors.Errorf("repo profile already exists at %s", RepoProfilePath(m.repoPath))
	}
	repoRoot, err := DiscoverRepoRoot(m.cfg.Cwd)
	if err != nil {
		return "", err
	}
	if repoRoot == "" {
		return "", errors.New("not inside a git repository; `zcli profile init` needs a repo root to anchor .zcli/")
	}
	rd := RepoData{Id: uuid.NewString()}
	if err := writeRepoProfile(RepoProfilePath(repoRoot), m.cfg.GlobalMode, rd); err != nil {
		return "", err
	}
	m.repoPath = repoRoot
	m.repo = rd
	return repoRoot, nil
}

// UpdateGlobal mutates the global non-sensitive state via callback and persists the result.
func (m *Manager) UpdateGlobal(cb func(Data) Data) (Data, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.global = cb(m.global)
	return m.global, m.saveGlobal(m.global)
}

// UpdateRepo mutates the repo profile state via callback and persists the result. Errors
// when no repo profile is initialised in the current directory tree.
func (m *Manager) UpdateRepo(cb func(RepoData) RepoData) (RepoData, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.repoPath == "" {
		return RepoData{}, errors.New("no repo profile in this directory tree; run `zcli profile init` here first")
	}
	m.repo = cb(m.repo)
	return m.repo, m.saveRepo(m.repo)
}

// Clear wipes the global profile file plus every secret under the global scope (token, all
// VPN keys). Repo profiles are left alone - they belong to the repo, not the user session.
func (m *Manager) Clear() (Data, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	prior := m.global
	m.global = Data{}
	if err := m.saveGlobal(m.global); err != nil {
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

// ValueSource indicates which profile layer produced an effective value.
type ValueSource string

const (
	// SourceNone: value is not set in any layer.
	SourceNone ValueSource = ""
	// SourceGlobal: value comes from the global profile / global-scope secrets.
	SourceGlobal ValueSource = "global"
	// SourceRepo: value comes from the repo profile / repo-scope secrets (overrides global).
	SourceRepo ValueSource = "repo"
)

// Token returns the effective token. Repo-scoped token (when set) wins over the global one.
// Returns ("", false, nil) when neither layer has a token.
func (m *Manager) Token() (string, bool, error) {
	v, src, err := m.TokenWithSource()
	return v, src != SourceNone, err
}

// TokenWithSource returns the effective token along with which layer produced it. Used by
// introspection surfaces that want to attribute the value (`zcli profile where`); regular
// callers should use Token, which discards the source.
func (m *Manager) TokenWithSource() (string, ValueSource, error) {
	m.mu.RLock()
	repoId := ""
	if m.repoPath != "" {
		repoId = m.repo.Id
	}
	m.mu.RUnlock()

	if repoId != "" {
		v, ok, err := m.secrets.Get(secrets.Key(secrets.RepoScope(repoId), "token"))
		if err != nil {
			return "", SourceNone, err
		}
		if ok {
			return v, SourceRepo, nil
		}
	}
	v, ok, err := m.secrets.Get(secrets.Key(secrets.GlobalScope, "token"))
	if err != nil {
		return "", SourceNone, err
	}
	if ok {
		return v, SourceGlobal, nil
	}
	return "", SourceNone, nil
}

// SetToken writes the auth token. Defaults to the global scope; pass the WithRepo option to
// target the current repo profile instead.
func (m *Manager) SetToken(token string, opts ...SetOption) error {
	o := buildSetOptions(opts)
	scope, err := m.tokenScope(o.target)
	if err != nil {
		return err
	}
	return m.secrets.Set(secrets.Key(scope, "token"), token)
}

func (m *Manager) tokenScope(target Target) (secrets.Scope, error) {
	switch target {
	case TargetRepo:
		m.mu.RLock()
		repoId := ""
		if m.repoPath != "" {
			repoId = m.repo.Id
		}
		m.mu.RUnlock()
		if repoId == "" {
			return nil, errors.New("no repo profile in this directory tree; run `zcli profile init` here first")
		}
		return secrets.RepoScope(repoId), nil
	case TargetGlobal, "":
		return secrets.GlobalScope, nil
	default:
		return nil, errors.Errorf("unknown profile target %q", target)
	}
}

// Target picks which profile layer a write operation lands on. Empty defaults to global.
type Target string

const (
	TargetGlobal Target = "global"
	TargetRepo   Target = "repo"
)

// SetOption configures write operations like SetToken.
type SetOption func(*setOptions)

type setOptions struct {
	target Target
}

// WithRepo targets the current repo profile for a write operation.
func WithRepo() SetOption {
	return func(o *setOptions) { o.target = TargetRepo }
}

// WithGlobal targets the global profile for a write operation (default).
func WithGlobal() SetOption {
	return func(o *setOptions) { o.target = TargetGlobal }
}

func buildSetOptions(opts []SetOption) setOptions {
	var o setOptions
	for _, opt := range opts {
		opt(&o)
	}
	return o
}

// VpnKey returns the cached VPN key for a project. VPN keys are global only - a project's key
// is the same regardless of which repo asks for it.
func (m *Manager) VpnKey(projectId zuuid.ProjectId) (entity.VpnKey, bool, error) {
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
func (m *Manager) SetVpnKey(projectId zuuid.ProjectId, key entity.VpnKey) error {
	payload, err := json.Marshal(key)
	if err != nil {
		return errors.WithStack(err)
	}
	return m.secrets.Set(vpnKeyPath(projectId), string(payload))
}

// DeleteVpnKey removes the cached VPN key for a project. Idempotent.
func (m *Manager) DeleteVpnKey(projectId zuuid.ProjectId) error {
	return m.secrets.Delete(vpnKeyPath(projectId))
}

func vpnKeyPath(projectId zuuid.ProjectId) string {
	return secrets.Key(secrets.GlobalScope, "vpn", projectId.Native())
}

// filepathDir is a tiny shim so this file doesn't pull in "path/filepath" alongside the
// "os" import (both define a Dir-like helper). Inlining the well-known stdlib call keeps the
// import list focused.
func filepathDir(p string) string {
	for i := len(p) - 1; i >= 0; i-- {
		if p[i] == '/' || p[i] == '\\' {
			return p[:i]
		}
	}
	return "."
}

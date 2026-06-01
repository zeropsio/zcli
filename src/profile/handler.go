// Package profile is the CLI's persisted per-user state. Composed of two layers, each
// independently versioned through the versionedjson package:
//
//   - Global profile: token / region / preferred secrets backend. Held in a JSON file at the
//     user's CLI config path. Sensitive values (token, VPN private keys) are not in the JSON;
//     they live in the secrets package's chosen backend.
//   - Repo profile (optional): per-repository overrides. Held in <repo>/.zcli/profile.json
//     when initialised. A repo file holds a named list of profile entries plus an active
//     pointer; each entry has a stable Id used to scope its own per-repo secrets, plus an
//     optional region override and saved scope project.
//
// Effective state for the active entry is the field-level merge of (global, active repo
// profile). The repo layer wins per field; fields absent from the repo layer fall through
// to global. The active entry can be switched persistently via UseRepoProfile or for a single
// invocation via SwitchActiveProfile (wired into the root --profile flag).
//
// # Layout
//
// Three independent versioned chains, each backed by a *versionedjson.Schema. Every chain
// has a hub file (constants + Current alias + exported view + schema var) and one per-version
// file holding just the struct + its Up/StampVersion/Chain methods + the init() that
// registers its decoder on the schema:
//
//   - Global chain — hub data.go, versions data_v0.go..data_v2.go.
//   - Repo file wrapper chain — hub data_repo.go, versions data_repo_vN.go.
//   - Repo profile entry chain — hub data_repo_profile.go, versions data_repo_profile_vN.go.
//
// Each non-current version provides Up(migrationCtx); the current target implements
// VersionStamper so Schema.Load auto-stamps Version on the way out. Atomic file IO lives in
// atomicfile.go. The Manager itself is split for clarity:
//
//   - handler.go (this file): Config, Manager struct, NewManager, helpers, surface getters.
//   - load.go: loadGlobal / loadRepo / rescueScope plus migration helpers.
//   - save.go: saveGlobal / saveRepoFile / stampRepoFile.
//   - secrets.go: Token / SetToken / VpnKey / Clear plus ValueSource.
//
// # Reads and writes
//
// NewManager loads both layers, migrates each up to its current schema, and exposes the
// merged effective state via Region, Token, ScopeProject. Writes:
//
//   - UpdateGlobal, Clear (global JSON)
//   - SetScopeProject (active repo entry)
//   - AddRepoProfile, UseRepoProfile, SwitchActiveProfile (repo file structure)
//   - SetToken (global secrets scope), SetVpnKey/DeleteVpnKey (global secrets, per project)
//
// # Adding a new migration
//
// Recipe is the same for any of the three chains - substitute the chain's hub file and its
// schema var as appropriate:
//
//  1. Create data_<chain>_vN.go with the new struct + init() registering on <chain>Schema.
//     Embed versionedjson.Probe so the new struct inherits the "version" JSON field and the
//     StampVersion hook for free.
//  2. Add Up(ctx) on the previous version's struct returning the new vN value.
//  3. Bump the Current<Chain>Version constant and flip the <chain>Current alias in the hub.
//  4. Implement Chain() returning the chain's family marker (dataChain / repoFileChain /
//     repoProfileChain) so the new version can be passed to versionedjson.Migrate alongside
//     its siblings without a compile-time family mismatch.
//
// # Cascading: bumping a child forces a parent bump
//
// The repo-file chain embeds the repo-profile chain statically: repoFileV1.Profiles is
// []repoProfileV1, repoFileVN.Profiles is []repoProfileVN. Whenever you add repoProfileV(N+1)
// you MUST also add repoFileV(N+1) and flip the RepoFile alias to it, otherwise the
// in-memory and on-disk Profiles slice element type still pins to the older child shape and
// fresh writes can't even carry the new fields. The new repoFileV(N+1).Up needs to walk each
// entry through versionedjson.Migrate(repoProfileSchema, ...) to convert old-shape entries
// into new-shape entries (and old wrappers keep []repoProfileVN for backwards-compat reads).
//
// The reverse is not required: a parent can bump without the child bumping, because the
// older parent's Up step keeps producing the same-typed child slice that already matches the
// current child target.
//
// The global chain has no embedded child chain, so this cascade does not apply there.
package profile

import (
	"os"
	"sync"

	"github.com/google/uuid"
	"github.com/pkg/errors"

	"github.com/zeropsio/zcli/src/entity"
	"github.com/zeropsio/zcli/src/optional"
	"github.com/zeropsio/zcli/src/secrets"
	zuuid "github.com/zeropsio/zerops-go/types/uuid"
)

// Config configures a Manager.
type Config struct {
	// GlobalPath is the JSON file holding non-sensitive global state, written to by every
	// save. Required. Today's canonical name is profile.json; older zcli installs used
	// cli.data and that file is migrated in via LegacyGlobalPath.
	GlobalPath string
	// GlobalMode is the file mode for GlobalPath, typically 0o600.
	GlobalMode os.FileMode
	// LegacyGlobalPath is the location of an older zcli's cli.data file, used only when
	// GlobalPath does not yet exist. Reading falls back here once; writing always goes to
	// GlobalPath. Empty means no legacy file to consider. The legacy file is left in place
	// after migration so users can recover the original by hand if needed.
	LegacyGlobalPath string
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
	repoFile RepoFile // zero when repoPath is ""; holds active name + every profile entry
	active   int      // index into repoFile.Profiles for the currently active entry; -1 when no repo
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

	m := &Manager{cfg: cfg, secrets: cfg.Secrets, active: -1}
	if err := m.loadGlobal(); err != nil {
		return nil, err
	}
	if err := m.loadRepo(); err != nil {
		return nil, err
	}
	return m, nil
}

// activeProfile returns the currently active repo profile entry. Returns (zero, false) when
// no repo profile is loaded. Caller must hold m.mu (read or write).
func (m *Manager) activeProfile() (RepoProfile, bool) {
	if m.repoPath == "" {
		return RepoProfile{}, false
	}
	return m.repoFile.Profiles[m.active], true
}

// activeProfilePtr returns a pointer to the active profile entry for in-place mutation, or
// nil when no repo profile is loaded. Caller must hold m.mu.Lock (write lock).
func (m *Manager) activeProfilePtr() *RepoProfile {
	if m.repoPath == "" {
		return nil
	}
	return &m.repoFile.Profiles[m.active]
}

// requireRepo returns the standard "no repo profile" error when one isn't initialised, or
// nil otherwise. Caller must hold m.mu (read or write). Centralises the user-facing message
// so it stays consistent across the public surface.
func (m *Manager) requireRepo() error {
	if m.repoPath == "" {
		return errors.New("no repo profile in this directory tree; run `zcli profile init` here first")
	}
	return nil
}

func lookupProfileIndex(rf RepoFile, name string) (int, error) {
	for i, p := range rf.Profiles {
		if p.Name == name {
			return i, nil
		}
	}
	return 0, errors.Errorf("repo profile %q not found", name)
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

// RepoProfile returns the currently active repo profile entry and a bool reporting whether one
// was discovered. Used by introspection surfaces (e.g. `zcli profile where`); for normal field
// access prefer the merged getters like Region / ScopeProject so callers don't have to handle
// the layering.
func (m *Manager) RepoProfile() (RepoProfile, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.activeProfile()
}

// ActiveRepoProfileName returns the name of the active repo profile, or "" when there is no
// repo profile in this directory tree.
func (m *Manager) ActiveRepoProfileName() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	p, ok := m.activeProfile()
	if !ok {
		return ""
	}
	return p.Name
}

// RepoProfiles returns a snapshot of every profile entry in the current repo file. Empty when
// no repo profile is discovered.
func (m *Manager) RepoProfiles() []RepoProfile {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.repoPath == "" {
		return nil
	}
	out := make([]RepoProfile, len(m.repoFile.Profiles))
	copy(out, m.repoFile.Profiles)
	return out
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

// Region returns the effective region: active repo profile's override when present, else global.
func (m *Manager) Region() Region {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if p, ok := m.activeProfile(); ok && p.Region != nil {
		return *p.Region
	}
	return m.global.Region
}

// ScopeProject returns the scoped project id from the active repo profile (if any). There is
// no global fallback by design: scope is per-repo. Returns (zero, false) outside a repo profile
// or when no scope is set on the active entry.
func (m *Manager) ScopeProject() (entity.ProjectId, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	p, ok := m.activeProfile()
	if !ok || !p.ScopeProjectId.Filled() {
		return "", false
	}
	id, _ := p.ScopeProjectId.Get()
	return entity.ProjectId(id), true
}

// SetScopeProject persists a scoped project to the active repo profile. Errors when no repo
// profile is initialised in the current directory tree - call InitRepo first or run from inside
// a repo that already has one.
func (m *Manager) SetScopeProject(id optional.Null[entity.ProjectId]) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if err := m.requireRepo(); err != nil {
		return err
	}
	var zid zuuid.ProjectIdNull
	if v, ok := id.Get(); ok {
		zid = zuuid.NewProjectIdNull(zuuid.ProjectId(v))
	}
	m.activeProfilePtr().ScopeProjectId = zid
	return m.saveRepoFile()
}

// ScopeService returns the scoped service id from the active repo profile (if any). Like
// ScopeProject, scope is per-repo - there is no global fallback. Returns (zero, false) outside
// a repo profile or when no scope is set on the active entry.
func (m *Manager) ScopeService() (entity.ServiceId, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	p, ok := m.activeProfile()
	if !ok || !p.ScopeServiceId.Filled() {
		return "", false
	}
	id, _ := p.ScopeServiceId.Get()
	return entity.ServiceId(id), true
}

// SetScopeService persists a scoped service to the active repo profile. Errors when no repo
// profile is initialised - same rule as SetScopeProject.
func (m *Manager) SetScopeService(id optional.Null[entity.ServiceId]) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if err := m.requireRepo(); err != nil {
		return err
	}
	var zid zuuid.ServiceStackIdNull
	if v, ok := id.Get(); ok {
		zid = zuuid.NewServiceStackIdNull(zuuid.ServiceStackId(v))
	}
	m.activeProfilePtr().ScopeServiceId = zid
	return m.saveRepoFile()
}

// ScopeClient returns the effective client (org) scope. Unlike ScopeProject / ScopeService -
// which are repo-only with no global fallback - the client scope merges across layers the way
// Region does: the active repo profile's value wins, then the global value, else (zero, false).
// It is optional everywhere; callers fall back to AllOrgs + flag / interactive selector when unset.
func (m *Manager) ScopeClient() (entity.OrgId, bool) {
	id, src := m.ScopeClientWithSource()
	return id, src != SourceNone
}

// ScopeClientWithSource is ScopeClient plus which layer produced the value, for `zcli profile
// where` provenance. Regular callers should use ScopeClient, which discards the source.
func (m *Manager) ScopeClientWithSource() (entity.OrgId, ValueSource) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if p, ok := m.activeProfile(); ok && p.ScopeClientId.Filled() {
		id, _ := p.ScopeClientId.Get()
		return entity.OrgId(id), SourceRepo
	}
	if m.global.ScopeClientId.Filled() {
		id, _ := m.global.ScopeClientId.Get()
		return entity.OrgId(id), SourceGlobal
	}
	return "", SourceNone
}

// SetGlobalScopeClient persists (or clears, on empty input) the global client scope. There is
// no repo-layer counterpart: the per-repo client scope is set only at profile creation through
// AddRepoProfileFull, matching the "selected when creating a new profile" model.
func (m *Manager) SetGlobalScopeClient(id optional.Null[entity.OrgId]) error {
	var zid zuuid.ClientIdNull
	if v, ok := id.Get(); ok {
		zid = zuuid.NewClientIdNull(zuuid.ClientId(v))
	}
	_, err := m.UpdateGlobal(func(data Data) Data {
		data.ScopeClientId = zid
		return data
	})
	return err
}

// InitRepo creates a fresh repo profile rooted at the nearest .git ancestor of cwd, seeded
// with a single profile named DefaultRepoProfileName. Errors when cwd is not inside a git repo
// or a repo profile already exists. Returns the directory the profile file was created in.
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
	rf := RepoFile{
		ActiveProfile: DefaultRepoProfileName,
		Profiles: []RepoProfile{{
			Name: DefaultRepoProfileName,
			Id:   uuid.NewString(),
		}},
	}
	if err := writeRepoFile(RepoProfilePath(repoRoot), m.cfg.GlobalMode, rf); err != nil {
		return "", err
	}
	m.repoPath = repoRoot
	m.repoFile = rf
	m.active = 0
	return repoRoot, nil
}

// UpdateGlobal mutates the global non-sensitive state via callback and persists the result.
func (m *Manager) UpdateGlobal(cb func(Data) Data) (Data, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.global = cb(m.global)
	return m.global, m.saveGlobal(m.global)
}

// AddRepoProfile appends a new named profile to the repo file with a freshly generated Id.
// The new profile starts with no scope and no region override. Errors when the name is empty,
// already exists, or no repo profile is initialised in this directory tree.
func (m *Manager) AddRepoProfile(name string) (RepoProfile, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if err := m.requireRepo(); err != nil {
		return RepoProfile{}, err
	}
	if name == "" {
		return RepoProfile{}, errors.New("profile name must not be empty")
	}
	for _, p := range m.repoFile.Profiles {
		if p.Name == name {
			return RepoProfile{}, errors.Errorf("profile %q already exists", name)
		}
	}
	rp := RepoProfile{Name: name, Id: uuid.NewString()}
	m.repoFile.Profiles = append(m.repoFile.Profiles, rp)
	return rp, m.saveRepoFile()
}

// RemoveRepoProfile drops the named entry from the repo file and deletes every secret in its
// per-profile scope (token, VPN keys, ...). Refuses to remove the currently active profile so
// a stray `remove` can't leave the repo in a no-active-profile state - callers must switch
// elsewhere with `zcli profile use` first. Errors when no repo profile is initialised or the
// name doesn't match an existing entry.
func (m *Manager) RemoveRepoProfile(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if err := m.requireRepo(); err != nil {
		return err
	}
	if m.repoFile.ActiveProfile == name {
		return errors.Errorf("profile %q is currently active; switch with `zcli profile use <other>` first", name)
	}
	idx, err := lookupProfileIndex(m.repoFile, name)
	if err != nil {
		return err
	}
	removed := m.repoFile.Profiles[idx]
	m.repoFile.Profiles = append(m.repoFile.Profiles[:idx], m.repoFile.Profiles[idx+1:]...)
	// Active index may shift if the removed entry sat before it.
	if idx < m.active {
		m.active--
	}
	keys, err := m.secrets.List(secrets.Prefix(secrets.RepoScope(removed.Id)))
	if err != nil {
		return errors.Wrapf(err, "list secrets for profile %q (id %s)", name, removed.Id)
	}
	for _, k := range keys {
		if err := m.secrets.Delete(k); err != nil {
			return errors.Wrapf(err, "delete secret %s for profile %q", k, name)
		}
	}
	return m.saveRepoFile()
}

// AddRepoProfileFull appends a freshly created entry whose non-sensitive fields are taken
// from the template. The template's Id is ignored - a new UUID is generated so per-profile
// secrets are always isolated. Errors when the name is empty, collides, or no repo profile
// is initialised. Useful as the catch-all create path when callers have a fully built
// RepoProfile value (e.g. an interactive builder collecting region / scope / ... up front).
func (m *Manager) AddRepoProfileFull(in NewRepoProfile) (RepoProfile, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if err := m.requireRepo(); err != nil {
		return RepoProfile{}, err
	}
	if in.Name == "" {
		return RepoProfile{}, errors.New("profile name must not be empty")
	}
	for _, p := range m.repoFile.Profiles {
		if p.Name == in.Name {
			return RepoProfile{}, errors.Errorf("profile %q already exists", in.Name)
		}
	}
	var scopeProject zuuid.ProjectIdNull
	if id, ok := in.ScopeProjectId.Get(); ok {
		scopeProject = zuuid.NewProjectIdNull(zuuid.ProjectId(id))
	}
	var scopeService zuuid.ServiceStackIdNull
	if id, ok := in.ScopeServiceId.Get(); ok {
		scopeService = zuuid.NewServiceStackIdNull(zuuid.ServiceStackId(id))
	}
	var scopeClient zuuid.ClientIdNull
	if id, ok := in.ScopeClientId.Get(); ok {
		scopeClient = zuuid.NewClientIdNull(zuuid.ClientId(id))
	}
	template := RepoProfile{
		Name:           in.Name,
		Id:             uuid.NewString(),
		Region:         in.Region,
		ScopeProjectId: scopeProject,
		ScopeServiceId: scopeService,
		ScopeClientId:  scopeClient,
	}
	m.repoFile.Profiles = append(m.repoFile.Profiles, template)
	return template, m.saveRepoFile()
}

// NewRepoProfile carries the domain-typed inputs for AddRepoProfileFull. Scope ids are domain
// types; the Manager converts them to the on-disk uuid shape so callers never touch zerops-go.
type NewRepoProfile struct {
	Name           string
	Region         *Region
	ScopeProjectId optional.Null[entity.ProjectId]
	ScopeServiceId optional.Null[entity.ServiceId]
	ScopeClientId  optional.Null[entity.OrgId]
}

// SetRepoProfileToken stores a token in the secrets backend under the named profile's
// RepoScope(id) prefix. Targets a specific profile by name, so it works for any entry in the
// repo file - not just the currently active one. Errors when no repo profile is initialised
// or the name doesn't match an existing entry.
func (m *Manager) SetRepoProfileToken(name, token string) error {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if err := m.requireRepo(); err != nil {
		return err
	}
	idx, err := lookupProfileIndex(m.repoFile, name)
	if err != nil {
		return err
	}
	return m.secrets.Set(secrets.Key(secrets.RepoScope(m.repoFile.Profiles[idx].Id), "token"), token)
}

// AddRepoProfileCopy is AddRepoProfile plus an initial copy of the non-sensitive override
// fields from another existing entry. The new entry gets a fresh Id (so per-profile secrets
// are isolated from the source) but inherits the source's Region override and persisted scope
// (project + service + client). Tokens / VPN keys are not copied - those live in the secrets backend
// keyed by the entry's Id and stay isolated. Errors when sourceName doesn't exist, on the
// same conditions as AddRepoProfile (empty name, collision, no repo).
func (m *Manager) AddRepoProfileCopy(name, sourceName string) (RepoProfile, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if err := m.requireRepo(); err != nil {
		return RepoProfile{}, err
	}
	if name == "" {
		return RepoProfile{}, errors.New("profile name must not be empty")
	}
	for _, p := range m.repoFile.Profiles {
		if p.Name == name {
			return RepoProfile{}, errors.Errorf("profile %q already exists", name)
		}
	}
	srcIdx, err := lookupProfileIndex(m.repoFile, sourceName)
	if err != nil {
		return RepoProfile{}, err
	}
	src := m.repoFile.Profiles[srcIdx]
	rp := RepoProfile{
		Name:           name,
		Id:             uuid.NewString(),
		ScopeProjectId: src.ScopeProjectId,
		ScopeServiceId: src.ScopeServiceId,
		ScopeClientId:  src.ScopeClientId,
	}
	if src.Region != nil {
		clone := *src.Region
		rp.Region = &clone
	}
	m.repoFile.Profiles = append(m.repoFile.Profiles, rp)
	return rp, m.saveRepoFile()
}

// RenameRepoProfile changes the user-visible Name of an existing entry. The entry's Id (and
// therefore its per-profile secrets scope) is preserved, so tokens / VPN keys stored under the
// repo scope are not touched. If the renamed entry was the active profile, ActiveProfile
// flips along with it so the file's active pointer stays in sync. Errors when no repo profile
// is initialised, oldName doesn't exist, newName is empty, or newName collides with another
// entry. Renaming to the same name is a no-op success.
func (m *Manager) RenameRepoProfile(oldName, newName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if err := m.requireRepo(); err != nil {
		return err
	}
	if newName == "" {
		return errors.New("profile name must not be empty")
	}
	if oldName == newName {
		return nil
	}
	idx, err := lookupProfileIndex(m.repoFile, oldName)
	if err != nil {
		return err
	}
	for i, p := range m.repoFile.Profiles {
		if i != idx && p.Name == newName {
			return errors.Errorf("profile %q already exists", newName)
		}
	}
	m.repoFile.Profiles[idx].Name = newName
	if m.repoFile.ActiveProfile == oldName {
		m.repoFile.ActiveProfile = newName
	}
	return m.saveRepoFile()
}

// UseRepoProfile changes the persisted active profile to name and writes the file. Errors
// when name doesn't match an existing entry or no repo profile is initialised. This is the
// state-changing variant; for a single-invocation override use SwitchActiveProfile.
func (m *Manager) UseRepoProfile(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if err := m.requireRepo(); err != nil {
		return err
	}
	idx, err := lookupProfileIndex(m.repoFile, name)
	if err != nil {
		return err
	}
	m.repoFile.ActiveProfile = name
	m.active = idx
	return m.saveRepoFile()
}

// SwitchActiveProfile changes the in-memory active profile for the rest of the process without
// touching disk. Wired into the root `--profile <name>` flag so a single command can run under
// a different identity without persisting the choice. Errors when name doesn't exist.
//
// Distinct from requireRepo's standard message: the --profile context is specific enough that
// pointing the user at `profile init` deserves its own phrasing.
func (m *Manager) SwitchActiveProfile(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.repoPath == "" {
		return errors.New("--profile requires a repo profile in this directory tree; run `zcli profile init` here first")
	}
	idx, err := lookupProfileIndex(m.repoFile, name)
	if err != nil {
		return err
	}
	m.active = idx
	return nil
}

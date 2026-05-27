package profile

import (
	"os"

	"github.com/google/uuid"
	"github.com/pkg/errors"

	"github.com/zeropsio/zcli/src/secrets"
	zuuid "github.com/zeropsio/zerops-go/types/uuid"
)

// migrationCtx is threaded into every Up step. Carries the secrets backend (for migrations
// that need to drain plaintext fields), cwd (for repo discovery from inside a v1→v2 global
// migration), a rescueScope callback (used by data_v1.Up to relocate the legacy global scope
// into a repo profile), and a log sink for one-time user notices.
type migrationCtx struct {
	secrets     secrets.Backend
	cwd         string
	rescueScope func(zuuid.ProjectIdNull) (bool, error)
	log         func(string)
}

func (m *Manager) loadGlobal() error {
	readPath, fromLegacy, err := m.pickGlobalReadPath()
	if err != nil {
		return err
	}
	if readPath == "" {
		return nil
	}
	info, err := os.Stat(readPath)
	if err != nil {
		return errors.WithStack(err)
	}
	if info.Size() == 0 {
		return nil
	}

	raw, err := os.ReadFile(readPath)
	if err != nil {
		return errors.WithStack(err)
	}

	result, err := globalSchema.Load(raw, migrationCtx{
		secrets:     m.secrets,
		cwd:         m.cfg.Cwd,
		rescueScope: m.rescueScope,
		log:         m.cfg.Log,
	})
	if err != nil {
		return errors.Wrapf(err, "load global profile %s", m.cfg.GlobalPath)
	}
	// Schema.Load already stamped result.Value.Version via VersionStamper.
	m.global = result.Value

	// Two reasons to rewrite: a data-format migration occurred (need to persist the new
	// shape), or the file came from the legacy cli.data path (need to populate the new
	// profile.json so subsequent runs find it directly). When the source was the legacy
	// path the original cli.data is left in place to act as the user's backup; when the
	// source was GlobalPath itself we write a .vN.bak alongside.
	if result.Migrated(CurrentDataVersion) || fromLegacy {
		if !fromLegacy {
			if err := backupProfile(m.cfg.GlobalPath, raw, result.SourceVersion, m.cfg.GlobalMode); err != nil {
				return errors.Wrap(err, "backup global profile before migration")
			}
		}
		if err := m.saveGlobal(m.global); err != nil {
			return errors.Wrap(err, "rewrite global profile in current schema after migration")
		}
		if fromLegacy && m.cfg.Log != nil {
			m.cfg.Log("Migrated " + readPath + " to " + m.cfg.GlobalPath + "; the legacy file is preserved unchanged in case you need to roll back")
		}
	}
	return nil
}

// pickGlobalReadPath chooses which file to load the global profile from. Prefers GlobalPath
// when it has real content; treats an empty GlobalPath as "no file" so the writability-probe
// side effect from constants.checkPath (which leaves an empty profile.json behind) doesn't
// shadow a legitimate legacy cli.data with the user's actual state. Falls back to
// LegacyGlobalPath when GlobalPath is absent or empty. Returns ("", false, nil) when neither
// path has usable content - that's a clean install, not an error. The bool reports whether
// the picked path was the legacy one, which the caller uses to decide between in-place
// backup and legacy-preservation.
func (m *Manager) pickGlobalReadPath() (string, bool, error) {
	if info, err := os.Stat(m.cfg.GlobalPath); err == nil {
		if info.Size() > 0 {
			return m.cfg.GlobalPath, false, nil
		}
	} else if !os.IsNotExist(err) {
		return "", false, errors.WithStack(err)
	}
	if m.cfg.LegacyGlobalPath == "" {
		return "", false, nil
	}
	if _, err := os.Stat(m.cfg.LegacyGlobalPath); err == nil {
		return m.cfg.LegacyGlobalPath, true, nil
	} else if !os.IsNotExist(err) {
		return "", false, errors.WithStack(err)
	}
	return "", false, nil
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
	rf := RepoFile{
		ActiveProfile: DefaultRepoProfileName,
		Profiles: []RepoProfile{{
			Name:           DefaultRepoProfileName,
			Id:             uuid.NewString(),
			ScopeProjectId: scope,
		}},
	}
	if err := writeRepoFile(path, m.cfg.GlobalMode, rf); err != nil {
		return false, err
	}
	m.repoPath = repoRoot
	m.repoFile = rf
	m.active = 0
	if m.cfg.Log != nil {
		m.cfg.Log("Scoped project moved into " + path + " (run `git ignore .zcli/` or add .zcli/ to .gitignore to keep secrets out of source control)")
	}
	return true, nil
}

func (m *Manager) loadRepo() error {
	// If rescueScope already populated the repo state during a global v1 → v2 migration, don't
	// load over it - the file we just wrote IS the source of truth.
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

	ctx := migrationCtx{
		secrets: m.secrets,
		cwd:     m.cfg.Cwd,
		log:     m.cfg.Log,
	}
	fileResult, err := repoFileSchema.Load(raw, ctx)
	if err != nil {
		return errors.Wrapf(err, "load repo profile %s", path)
	}

	// fileResult.Value.Profiles is already []repoProfileCurrent: each repoFileVN statically
	// holds []repoProfileVN, and the cascade rule (see "Cascading" in handler.go's package
	// doc) means bumping the entry version requires bumping the file wrapper, so by the time
	// Load returns we're at the current child shape. Any per-entry promotion happens inside
	// repoFileVN.Up via versionedjson.Migrate(repoProfileSchema, ...), captured in
	// fileResult.Migrated - no second loop needed here.
	rf := RepoFile(fileResult.Value)
	if err := validateRepoFile(rf, path); err != nil {
		return err
	}
	m.repoFile = rf

	if fileResult.Migrated(CurrentRepoVersion) {
		if err := backupProfile(path, raw, fileResult.SourceVersion, m.cfg.GlobalMode); err != nil {
			return errors.Wrap(err, "backup repo profile before migration")
		}
		if err := writeJSONAtomic(path, m.cfg.GlobalMode, m.repoFile); err != nil {
			return errors.Wrap(err, "rewrite repo profile in current schema after migration")
		}
	}

	idx, err := resolveActiveIndex(m.repoFile)
	if err != nil {
		return errors.Wrapf(err, "resolve active profile in %s", path)
	}
	m.repoPath = repoDir
	m.active = idx
	return nil
}

// validateRepoFile checks invariants the rest of the Manager relies on: at least one profile,
// unique non-empty names, every entry has an Id. Run on every load (including post-migration)
// so a hand-edited file fails loudly instead of producing surprising lookups later.
func validateRepoFile(rf RepoFile, path string) error {
	if len(rf.Profiles) == 0 {
		return errors.Errorf("profile: repo file %s has no profiles", path)
	}
	seen := map[string]struct{}{}
	for i, p := range rf.Profiles {
		if p.Name == "" {
			return errors.Errorf("profile: entry %d in %s is missing \"name\"", i, path)
		}
		if p.Id == "" {
			return errors.Errorf("profile: entry %q in %s is missing \"id\"", p.Name, path)
		}
		if _, dup := seen[p.Name]; dup {
			return errors.Errorf("profile: duplicate profile name %q in %s", p.Name, path)
		}
		seen[p.Name] = struct{}{}
	}
	return nil
}

// resolveActiveIndex maps repoFile.ActiveProfile to the index of the matching entry. Empty
// active_profile picks the first entry (covers freshly migrated files where the field hasn't
// been written yet by an external tool). Returns an error when the active name doesn't match
// any entry - we'd rather surface that than silently drift.
func resolveActiveIndex(rf RepoFile) (int, error) {
	target := rf.ActiveProfile
	if target == "" {
		return 0, nil
	}
	for i, p := range rf.Profiles {
		if p.Name == target {
			return i, nil
		}
	}
	return 0, errors.Errorf("active_profile %q does not match any profile entry", target)
}

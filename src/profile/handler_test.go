package profile

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/zeropsio/zcli/src/secrets"
)

// TestLoadV0File_DrainsTokenAndVpnKeys verifies that a legacy on-disk file (no "version" key,
// PascalCase fields, plaintext token + VPN keys inline) loads as v0, drains its sensitive
// fields into the secrets backend, and gets rewritten on disk in the current snake_case shape
// with the sensitive fields gone.
func TestLoadV0File_DrainsTokenAndVpnKeys(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "profile.json")

	v0 := []byte(`{
		"Token":"legacy-token",
		"RegionData":{"name":"prg1","isDefault":true,"address":"api.app-prg1.zerops.io"},
		"ProjectVpnKeyRegistry":{"AAAAAAAAAAAAAAAAAAAAAA":{"Key":"wgkey-payload","ProjectId":"AAAAAAAAAAAAAAAAAAAAAA","CreatedAt":"2026-01-01T00:00:00Z"}}
	}`)
	require.NoError(t, os.WriteFile(path, v0, 0o600))

	store := secrets.NewMemory()
	m, err := NewManager(Config{
		GlobalPath: path,
		GlobalMode: 0o600,
		Cwd:        dir,
		Secrets:    store,
	})
	require.NoError(t, err)

	// Non-sensitive state: Region survived, version bumped to current (v2 - no ScopeProjectId).
	d := m.Data()
	require.NotNil(t, d.Version)
	require.Equal(t, CurrentDataVersion, *d.Version)
	require.Equal(t, "prg1", d.Region.Name)

	// Token drained into secrets backend.
	tok, ok, err := m.Token()
	require.NoError(t, err)
	require.True(t, ok, "token should be present in secrets after v0 drain")
	require.Equal(t, "legacy-token", tok)

	// Token + one VPN key drained into secrets backend.
	keys, err := store.List(secrets.Prefix(secrets.GlobalScope))
	require.NoError(t, err)
	require.Len(t, keys, 2, "expected token + one vpn key in secrets after drain")

	// On-disk file is now in current shape: snake_case, version=CurrentDataVersion, no plaintext secrets, no scope.
	raw, err := os.ReadFile(path)
	require.NoError(t, err)
	var onDisk map[string]any
	require.NoError(t, json.Unmarshal(raw, &onDisk))
	require.EqualValues(t, CurrentDataVersion, onDisk["version"])
	require.NotContains(t, onDisk, "Token", "legacy plaintext token must be gone from disk")
	require.NotContains(t, onDisk, "token", "token must never appear in profile JSON")
	require.NotContains(t, onDisk, "ProjectVpnKeyRegistry", "legacy plaintext vpn keys must be gone from disk")
	require.NotContains(t, onDisk, "project_vpn_key_registry", "vpn keys must never appear in profile JSON")
	require.NotContains(t, onDisk, "scope_project_id", "scope_project_id moved to repo profile in v2")
	require.Equal(t, "prg1", onDisk["region"].(map[string]any)["name"])
}

// TestManager_SetTokenAndClear covers the round-trip through the secrets backend plus the
// Clear() sweep that wipes both the profile JSON and every global-scope secret on logout.
func TestManager_SetTokenAndClear(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "profile.json")

	store := secrets.NewMemory()
	m, err := NewManager(Config{
		GlobalPath: path,
		GlobalMode: 0o600,
		Cwd:        dir,
		Secrets:    store,
	})
	require.NoError(t, err)

	require.NoError(t, m.SetToken("fresh-token"))
	tok, ok, err := m.Token()
	require.NoError(t, err)
	require.True(t, ok)
	require.Equal(t, "fresh-token", tok)

	_, err = m.Clear()
	require.NoError(t, err)

	_, ok, err = m.Token()
	require.NoError(t, err)
	require.False(t, ok, "token should be cleared from secrets after Clear")

	keys, err := store.List(secrets.Prefix(secrets.GlobalScope))
	require.NoError(t, err)
	require.Empty(t, keys, "all global secrets should be removed by Clear")
}

// TestDiscoverRepoProfileDir_FindsProfile creates a fake repo with .zcli/profile.json and asserts
// discovery from a nested subdirectory walks up to it.
func TestDiscoverRepoProfileDir_FindsProfile(t *testing.T) {
	repo := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(repo, ".git"), 0o755))
	require.NoError(t, os.MkdirAll(filepath.Join(repo, RepoProfileDir), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(repo, RepoProfileDir, RepoProfileFile), []byte("{}"), 0o600))

	nested := filepath.Join(repo, "a", "b", "c")
	require.NoError(t, os.MkdirAll(nested, 0o755))

	got, err := DiscoverRepoProfileDir(nested)
	require.NoError(t, err)
	require.Equal(t, repo, got)
}

// TestDiscoverRepoProfileDir_StopsAtGitBoundary asserts walks stop at the first .git/ ancestor
// when there's no .zcli/profile.json in the repo (we don't leak across repo boundaries).
func TestDiscoverRepoProfileDir_StopsAtGitBoundary(t *testing.T) {
	parent := t.TempDir()
	// Outer dir has a profile, inner dir is a separate git repo with no profile.
	require.NoError(t, os.MkdirAll(filepath.Join(parent, RepoProfileDir), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(parent, RepoProfileDir, RepoProfileFile), []byte("{}"), 0o600))

	innerRepo := filepath.Join(parent, "inner")
	require.NoError(t, os.MkdirAll(filepath.Join(innerRepo, ".git"), 0o755))

	got, err := DiscoverRepoProfileDir(innerRepo)
	require.NoError(t, err)
	require.Equal(t, "", got, "inner repo has no profile and the .git/ boundary blocks the outer one")
}

// TestMigration_V1_RescuesScopeIntoRepoProfile asserts the v1 → v2 migration writes a new repo
// profile carrying the legacy global scope when cwd is inside a git repo without one.
func TestMigration_V1_RescuesScopeIntoRepoProfile(t *testing.T) {
	repo := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(repo, ".git"), 0o755))

	globalDir := t.TempDir()
	globalPath := filepath.Join(globalDir, "profile.json")
	v1 := []byte(`{"version":1,"region":{"name":"prg1","isDefault":true,"address":"api.app-prg1.zerops.io"},"scope_project_id":"AAAAAAAAAAAAAAAAAAAAAA"}`)
	require.NoError(t, os.WriteFile(globalPath, v1, 0o600))

	m, err := NewManager(Config{
		GlobalPath: globalPath,
		GlobalMode: 0o600,
		Cwd:        repo,
		Secrets:    secrets.NewMemory(),
	})
	require.NoError(t, err)

	require.True(t, m.HasRepo(), "rescue should have created a repo profile")
	id, ok := m.ScopeProject()
	require.True(t, ok)
	require.Equal(t, "AAAAAAAAAAAAAAAAAAAAAA", id.Native())

	// Global file got rewritten in v2 shape with no scope_project_id.
	raw, err := os.ReadFile(globalPath)
	require.NoError(t, err)
	var onDisk map[string]any
	require.NoError(t, json.Unmarshal(raw, &onDisk))
	require.NotContains(t, onDisk, "scope_project_id")
	require.EqualValues(t, CurrentDataVersion, onDisk["version"])

	// Repo profile actually exists at the expected path with the rescued scope.
	repoFile, err := os.ReadFile(RepoProfilePath(repo))
	require.NoError(t, err)
	var repoOnDisk map[string]any
	require.NoError(t, json.Unmarshal(repoFile, &repoOnDisk))
	require.NotEmpty(t, repoOnDisk["id"], "repo profile must have a generated id")
	require.Equal(t, "AAAAAAAAAAAAAAAAAAAAAA", repoOnDisk["scope_project_id"])
}

// TestLegacyDataFile_MigratesToProfileJson covers the cli.data → profile.json file rename:
// a legacy file at LegacyGlobalPath is loaded, its contents migrated into a fresh GlobalPath,
// and the legacy file is left in place untouched as a natural backup.
func TestLegacyDataFile_MigratesToProfileJson(t *testing.T) {
	dir := t.TempDir()
	legacyPath := filepath.Join(dir, "cli.data")
	newPath := filepath.Join(dir, "profile.json")

	v1 := []byte(`{"version":1,"region":{"name":"prg1","isDefault":true,"address":"api.app-prg1.zerops.io"}}`)
	require.NoError(t, os.WriteFile(legacyPath, v1, 0o600))

	m, err := NewManager(Config{
		GlobalPath:       newPath,
		GlobalMode:       0o600,
		LegacyGlobalPath: legacyPath,
		Cwd:              dir,
		Secrets:          secrets.NewMemory(),
	})
	require.NoError(t, err)
	require.Equal(t, "prg1", m.Data().Region.Name, "legacy region should be carried over")

	// New file written, legacy left in place verbatim.
	newBytes, err := os.ReadFile(newPath)
	require.NoError(t, err)
	require.Greater(t, len(newBytes), 0, "new profile.json should hold migrated state")

	legacyAfter, err := os.ReadFile(legacyPath)
	require.NoError(t, err)
	require.Equal(t, v1, legacyAfter, "legacy cli.data must be preserved unchanged as a recovery backup")
}

// TestLegacyDataFile_IgnoredWhenNewFileExists asserts the new GlobalPath wins when both files
// exist: the legacy file is irrelevant once the migration ran, so subsequent runs don't
// re-import it.
func TestLegacyDataFile_IgnoredWhenNewFileExists(t *testing.T) {
	dir := t.TempDir()
	legacyPath := filepath.Join(dir, "cli.data")
	newPath := filepath.Join(dir, "profile.json")

	// Legacy holds region "legacy"; new holds region "new". After load, new should win.
	require.NoError(t, os.WriteFile(legacyPath, []byte(`{"version":1,"region":{"name":"legacy","isDefault":true,"address":""}}`), 0o600))
	require.NoError(t, os.WriteFile(newPath, []byte(`{"version":2,"region":{"name":"new","isDefault":true,"address":""}}`), 0o600))

	m, err := NewManager(Config{
		GlobalPath:       newPath,
		GlobalMode:       0o600,
		LegacyGlobalPath: legacyPath,
		Cwd:              dir,
		Secrets:          secrets.NewMemory(),
	})
	require.NoError(t, err)
	require.Equal(t, "new", m.Data().Region.Name)
}

// TestMigration_WritesBackupBeforeRewriting asserts that running a migration leaves a
// .v<source>.bak sibling holding the original on-disk bytes, so a botched migration is
// recoverable by hand.
func TestMigration_WritesBackupBeforeRewriting(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "profile.json")

	v1 := []byte(`{"version":1,"region":{"name":"prg1","isDefault":true,"address":"api.app-prg1.zerops.io"}}`)
	require.NoError(t, os.WriteFile(path, v1, 0o600))

	_, err := NewManager(Config{
		GlobalPath: path,
		GlobalMode: 0o600,
		Cwd:        dir,
		Secrets:    secrets.NewMemory(),
	})
	require.NoError(t, err)

	backup, err := os.ReadFile(path + ".v1.bak")
	require.NoError(t, err, "migration must leave a versioned backup of the original")
	require.Equal(t, v1, backup, "backup should hold the original bytes verbatim")
}

// TestRegion_RepoOverridesGlobal asserts a repo-level region wins over the global one.
func TestRegion_RepoOverridesGlobal(t *testing.T) {
	repo := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(repo, ".git"), 0o755))
	require.NoError(t, os.MkdirAll(filepath.Join(repo, RepoProfileDir), 0o755))
	require.NoError(t, os.WriteFile(
		filepath.Join(repo, RepoProfileDir, RepoProfileFile),
		[]byte(`{"version":1,"id":"repo-id","region":{"name":"override","isDefault":false,"address":"api.override.zerops.io"}}`),
		0o600,
	))

	globalDir := t.TempDir()
	globalPath := filepath.Join(globalDir, "profile.json")
	require.NoError(t, os.WriteFile(globalPath, []byte(`{"version":2,"region":{"name":"prg1","isDefault":true,"address":"api.app-prg1.zerops.io"}}`), 0o600))

	m, err := NewManager(Config{
		GlobalPath: globalPath,
		GlobalMode: 0o600,
		Cwd:        repo,
		Secrets:    secrets.NewMemory(),
	})
	require.NoError(t, err)
	require.Equal(t, "override", m.Region().Name)
	require.Equal(t, "prg1", m.Data().Region.Name, "global Data() still shows the global region; effective Region() merges")
}

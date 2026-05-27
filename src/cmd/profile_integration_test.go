package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/zeropsio/zcli/src/profile"
	"github.com/zeropsio/zcli/src/secrets"
)

// TestProfileInit_CreatesRepoProfileAndGitignore verifies that `zcli profile init` writes a
// repo profile at <repo>/.zcli/profile.json with a generated id, and appends `.zcli/` to the
// repo's .gitignore (creating it if absent).
func TestProfileInit_CreatesRepoProfileAndGitignore(t *testing.T) {
	f := newFixture(t)

	res := f.Run("profile", "init")
	require.Equalf(t, 0, res.ExitCode, "stderr=%q", res.Stderr)

	profilePath := profile.RepoProfilePath(f.RepoDir)
	raw, err := os.ReadFile(profilePath)
	require.NoError(t, err, "repo profile should exist after init")

	var onDisk map[string]any
	require.NoError(t, json.Unmarshal(raw, &onDisk))
	require.EqualValues(t, profile.CurrentRepoVersion, onDisk["version"])
	require.Equal(t, profile.DefaultRepoProfileName, onDisk["active_profile"])
	profiles, ok := onDisk["profiles"].([]any)
	require.True(t, ok, "profiles must be an array on disk")
	require.Len(t, profiles, 1, "init must seed exactly one profile")
	entry, ok := profiles[0].(map[string]any)
	require.True(t, ok, "default profile must be an object on disk")
	require.Equal(t, profile.DefaultRepoProfileName, entry["name"])
	require.NotEmpty(t, entry["id"], "init must stamp a stable id on the default profile")

	gitignore, err := os.ReadFile(filepath.Join(f.RepoDir, ".gitignore"))
	require.NoError(t, err, ".gitignore should exist after init")
	assert.Contains(t, string(gitignore), ".zcli/", "init must guard the secrets-adjacent dir from source control")
}

// TestProfileInit_AppendsToExistingGitignore covers the case where the repo already has a
// .gitignore: init should append `.zcli/` rather than overwrite the file.
func TestProfileInit_AppendsToExistingGitignore(t *testing.T) {
	f := newFixture(t)

	pre := "node_modules\n*.log\n"
	require.NoError(t, os.WriteFile(filepath.Join(f.RepoDir, ".gitignore"), []byte(pre), 0o644))

	res := f.Run("profile", "init")
	require.Equalf(t, 0, res.ExitCode, "stderr=%q", res.Stderr)

	got, err := os.ReadFile(filepath.Join(f.RepoDir, ".gitignore"))
	require.NoError(t, err)
	assert.Contains(t, string(got), "node_modules", "pre-existing entries must survive")
	assert.Contains(t, string(got), ".zcli/")
}

// TestProfileInit_IdempotentForGitignore verifies that running init twice doesn't duplicate
// the .zcli/ line in .gitignore (init the first time succeeds, second time errors because
// the profile already exists - so the second call shouldn't have written anyway, but if it
// did this proves the gitignore writer is a noop on second pass).
func TestProfileInit_IdempotentForGitignore(t *testing.T) {
	f := newFixture(t)

	res := f.Run("profile", "init")
	require.Equalf(t, 0, res.ExitCode, "first init stderr=%q", res.Stderr)

	// Second init must fail loudly: don't silently overwrite an existing repo profile.
	res = f.Run("profile", "init")
	require.NotEqualf(t, 0, res.ExitCode, "second init should error; stderr=%q", res.Stderr)

	got, err := os.ReadFile(filepath.Join(f.RepoDir, ".gitignore"))
	require.NoError(t, err)
	occurrences := 0
	for _, line := range splitLines(string(got)) {
		if line == ".zcli/" {
			occurrences++
		}
	}
	assert.Equal(t, 1, occurrences, ".zcli/ must appear exactly once even across multiple init attempts")
}

// TestProfileWhere_AttributesValuesToLayers prints `zcli profile where` for a fixture with a
// global region + global token + repo-scoped project, and asserts the layer attribution in
// the output. Covers the happy path through the merged getters used by the report.
func TestProfileWhere_AttributesValuesToLayers(t *testing.T) {
	f := newFixture(t)
	f.SeedScopedLogin("test-token", "AAAAAAAAAAAAAAAAAAAAAA")

	res := f.Run("profile", "where")
	require.Equalf(t, 0, res.ExitCode, "stderr=%q", res.Stderr)

	assert.Contains(t, res.Stdout, "Storage", "report should include the storage section")
	assert.Contains(t, res.Stdout, "global profile", "global profile path row")
	assert.Contains(t, res.Stdout, "repo profile", "repo profile row")
	assert.Contains(t, res.Stdout, "secrets", "secrets backend row")

	assert.Contains(t, res.Stdout, "Effective values", "report should include the effective values section")
	assert.Contains(t, res.Stdout, "prg1", "region value")
	assert.Contains(t, res.Stdout, "from global", "region attributed to global since no repo override")
	assert.Contains(t, res.Stdout, "set", "token presence")
	assert.Contains(t, res.Stdout, "AAAAAAAAAAAAAAAAAAAAAA", "scope project id")
	assert.Contains(t, res.Stdout, "from repo", "scope attributed to repo")
}

// TestProfileWhere_NoRepoProfile asserts the report renders sensibly when no repo profile is
// discovered: storage row shows "(none ...)" and scope shows "(not set)".
func TestProfileWhere_NoRepoProfile(t *testing.T) {
	f := newFixture(t)
	// Point cwd somewhere with no .git/ at or above so discovery returns nothing.
	noRepo := t.TempDir()
	t.Setenv("ZEROPS_CLI_CWD", noRepo)
	f.SeedLogin("test-token")

	res := f.Run("profile", "where")
	require.Equalf(t, 0, res.ExitCode, "stderr=%q", res.Stderr)

	assert.Contains(t, res.Stdout, "(none", "absence of repo profile should be explicit")
	assert.Contains(t, res.Stdout, "(not set)", "scope project should report not set without a repo")
	assert.Contains(t, res.Stdout, "from global", "token from global")
}

// TestProfileAddAndUse_PersistsActiveAndIsolatesIds covers the multi-profile happy path:
// add a second profile, switch to it via `profile use`, observe that active_profile flips
// on disk and the new profile carries its own id (so per-profile secrets are isolated).
func TestProfileAddAndUse_PersistsActiveAndIsolatesIds(t *testing.T) {
	f := newFixture(t)
	require.Equalf(t, 0, f.Run("profile", "init").ExitCode, "init must succeed")

	addRes := f.Run("profile", "add", "prod")
	require.Equalf(t, 0, addRes.ExitCode, "add stderr=%q", addRes.Stderr)

	raw, err := os.ReadFile(profile.RepoProfilePath(f.RepoDir))
	require.NoError(t, err)
	var onDisk map[string]any
	require.NoError(t, json.Unmarshal(raw, &onDisk))
	// `add` must not change the active profile.
	require.Equal(t, "default", onDisk["active_profile"], "add must not change active")
	profiles, ok := onDisk["profiles"].([]any)
	require.True(t, ok, "profiles must be an array on disk")
	require.Len(t, profiles, 2)
	defaultEntry, ok := profiles[0].(map[string]any)
	require.True(t, ok)
	prodEntry, ok := profiles[1].(map[string]any)
	require.True(t, ok)
	defaultID, ok := defaultEntry["id"].(string)
	require.True(t, ok)
	prodID, ok := prodEntry["id"].(string)
	require.True(t, ok)
	require.NotEmpty(t, defaultID)
	require.NotEmpty(t, prodID)
	require.NotEqual(t, defaultID, prodID, "each profile needs its own id for independent secret scoping")

	useRes := f.Run("profile", "use", "prod")
	require.Equalf(t, 0, useRes.ExitCode, "use stderr=%q", useRes.Stderr)

	raw, err = os.ReadFile(profile.RepoProfilePath(f.RepoDir))
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(raw, &onDisk))
	require.Equal(t, "prod", onDisk["active_profile"], "use must persist active_profile")
}

// TestProfileAdd_UseFlag_SwitchesActiveInOneStep covers the `--use` shorthand: add + activate
// in a single command, so the on-disk active_profile flips to the freshly created entry without
// a follow-up `profile use`.
func TestProfileAdd_UseFlag_SwitchesActiveInOneStep(t *testing.T) {
	f := newFixture(t)
	require.Equalf(t, 0, f.Run("profile", "init").ExitCode, "init must succeed")

	addRes := f.Run("profile", "add", "prod", "--use")
	require.Equalf(t, 0, addRes.ExitCode, "add --use stderr=%q", addRes.Stderr)

	raw, err := os.ReadFile(profile.RepoProfilePath(f.RepoDir))
	require.NoError(t, err)
	var onDisk map[string]any
	require.NoError(t, json.Unmarshal(raw, &onDisk))
	require.Equal(t, "prod", onDisk["active_profile"], "--use must persist the new profile as active")
}

// TestProfileRemove_DropsInactiveProfileAndDeletesItsSecrets covers the happy path: a profile
// that isn't active can be removed, its per-profile secrets get wiped alongside the entry, and
// the remaining profiles are left intact with the same active pointer.
func TestProfileRemove_DropsInactiveProfileAndDeletesItsSecrets(t *testing.T) {
	f := newFixture(t)
	require.Equalf(t, 0, f.Run("profile", "init").ExitCode, "init must succeed")
	require.Equalf(t, 0, f.Run("profile", "add", "prod").ExitCode, "add must succeed")

	// Seed a secret under the prod profile's repo scope so we can verify it's wiped.
	raw, err := os.ReadFile(profile.RepoProfilePath(f.RepoDir))
	require.NoError(t, err)
	var onDisk map[string]any
	require.NoError(t, json.Unmarshal(raw, &onDisk))
	profiles, ok := onDisk["profiles"].([]any)
	require.True(t, ok)
	prodEntry, ok := profiles[1].(map[string]any)
	require.True(t, ok)
	prodID, ok := prodEntry["id"].(string)
	require.True(t, ok)

	f.writeSecrets(map[string]string{
		secrets.Key(secrets.RepoScope(prodID), "token"): "prod-token",
	})

	rmRes := f.Run("profile", "remove", "prod")
	require.Equalf(t, 0, rmRes.ExitCode, "remove stderr=%q", rmRes.Stderr)

	raw, err = os.ReadFile(profile.RepoProfilePath(f.RepoDir))
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(raw, &onDisk))
	require.Equal(t, "default", onDisk["active_profile"], "active stays on default")
	remaining, ok := onDisk["profiles"].([]any)
	require.True(t, ok)
	require.Len(t, remaining, 1, "prod must be gone from the profiles list")

	stored := f.LoadSecrets()
	require.NotContains(t, stored, secrets.Key(secrets.RepoScope(prodID), "token"), "prod's per-profile secrets must be wiped")
}

// TestProfileRemove_RefusesActiveProfile asserts the safety guard: removing the currently
// active profile is rejected (otherwise the repo would end up with no active entry).
func TestProfileRemove_RefusesActiveProfile(t *testing.T) {
	f := newFixture(t)
	require.Equalf(t, 0, f.Run("profile", "init").ExitCode, "init must succeed")
	require.Equalf(t, 0, f.Run("profile", "add", "prod").ExitCode, "add must succeed")

	rmRes := f.Run("profile", "remove", "default")
	require.NotEqualf(t, 0, rmRes.ExitCode, "removing the active profile must fail; stdout=%q stderr=%q", rmRes.Stdout, rmRes.Stderr)
	assert.Contains(t, rmRes.Stderr, "currently active", "error message should explain why removal was refused")

	// On-disk state must be untouched.
	raw, err := os.ReadFile(profile.RepoProfilePath(f.RepoDir))
	require.NoError(t, err)
	var onDisk map[string]any
	require.NoError(t, json.Unmarshal(raw, &onDisk))
	profiles, ok := onDisk["profiles"].([]any)
	require.True(t, ok)
	require.Len(t, profiles, 2, "both entries must still be on disk")
}

// TestProfileDelete_AliasMatchesRemove covers the `delete` alias - same behaviour as `remove`.
func TestProfileDelete_AliasMatchesRemove(t *testing.T) {
	f := newFixture(t)
	require.Equalf(t, 0, f.Run("profile", "init").ExitCode, "init must succeed")
	require.Equalf(t, 0, f.Run("profile", "add", "prod").ExitCode, "add must succeed")

	delRes := f.Run("profile", "delete", "prod")
	require.Equalf(t, 0, delRes.ExitCode, "delete stderr=%q", delRes.Stderr)

	raw, err := os.ReadFile(profile.RepoProfilePath(f.RepoDir))
	require.NoError(t, err)
	var onDisk map[string]any
	require.NoError(t, json.Unmarshal(raw, &onDisk))
	profiles, ok := onDisk["profiles"].([]any)
	require.True(t, ok)
	require.Len(t, profiles, 1, "prod must be gone after delete alias")
}

// TestProfileRename_PreservesIdAndUpdatesActive renames the active profile and asserts the
// entry's stable Id stays untouched (so per-profile secrets remain reachable) and the file's
// active_profile pointer flips to the new name.
func TestProfileRename_PreservesIdAndUpdatesActive(t *testing.T) {
	f := newFixture(t)
	require.Equalf(t, 0, f.Run("profile", "init").ExitCode, "init must succeed")

	raw, err := os.ReadFile(profile.RepoProfilePath(f.RepoDir))
	require.NoError(t, err)
	var before map[string]any
	require.NoError(t, json.Unmarshal(raw, &before))
	beforeProfiles, ok := before["profiles"].([]any)
	require.True(t, ok)
	beforeEntry, ok := beforeProfiles[0].(map[string]any)
	require.True(t, ok)
	originalID, ok := beforeEntry["id"].(string)
	require.True(t, ok)

	renameRes := f.Run("profile", "rename", "default", "main")
	require.Equalf(t, 0, renameRes.ExitCode, "rename stderr=%q", renameRes.Stderr)

	raw, err = os.ReadFile(profile.RepoProfilePath(f.RepoDir))
	require.NoError(t, err)
	var after map[string]any
	require.NoError(t, json.Unmarshal(raw, &after))
	require.Equal(t, "main", after["active_profile"], "active_profile must follow the rename")
	afterProfiles, ok := after["profiles"].([]any)
	require.True(t, ok)
	require.Len(t, afterProfiles, 1)
	afterEntry, ok := afterProfiles[0].(map[string]any)
	require.True(t, ok)
	require.Equal(t, "main", afterEntry["name"])
	require.Equal(t, originalID, afterEntry["id"], "id must be preserved so per-profile secrets stay reachable")
}

// TestProfileRename_RejectsCollision asserts the rename refuses to clobber an existing name.
func TestProfileRename_RejectsCollision(t *testing.T) {
	f := newFixture(t)
	require.Equalf(t, 0, f.Run("profile", "init").ExitCode, "init must succeed")
	require.Equalf(t, 0, f.Run("profile", "add", "prod").ExitCode, "add must succeed")

	res := f.Run("profile", "rename", "default", "prod")
	require.NotEqualf(t, 0, res.ExitCode, "rename to existing name must fail; stdout=%q stderr=%q", res.Stdout, res.Stderr)
	assert.Contains(t, res.Stderr, "already exists")

	// State on disk must be unchanged.
	raw, err := os.ReadFile(profile.RepoProfilePath(f.RepoDir))
	require.NoError(t, err)
	var onDisk map[string]any
	require.NoError(t, json.Unmarshal(raw, &onDisk))
	require.Equal(t, "default", onDisk["active_profile"])
}

// TestProfileAdd_DuplicateNameRejected ensures the second add of the same name fails loudly
// rather than silently shadowing the first entry (which would mean two profiles share an Id
// or one becomes unreachable).
func TestProfileAdd_DuplicateNameRejected(t *testing.T) {
	f := newFixture(t)
	require.Equalf(t, 0, f.Run("profile", "init").ExitCode, "init must succeed")
	require.Equalf(t, 0, f.Run("profile", "add", "prod").ExitCode, "first add must succeed")

	res := f.Run("profile", "add", "prod")
	require.NotEqualf(t, 0, res.ExitCode, "duplicate add must error; stderr=%q", res.Stderr)
}

// TestProfileFlag_OverridesActiveForOneInvocation drives `--profile prod` through the where
// command and asserts the report reflects the prod profile as active, without the persisted
// active_profile flipping on disk.
func TestProfileFlag_OverridesActiveForOneInvocation(t *testing.T) {
	f := newFixture(t)
	f.SeedLogin("global-token")
	require.Equalf(t, 0, f.Run("profile", "init").ExitCode, "init must succeed")
	require.Equalf(t, 0, f.Run("profile", "add", "prod").ExitCode, "add prod must succeed")

	res := f.Run("--profile", "prod", "profile", "where")
	require.Equalf(t, 0, res.ExitCode, "stderr=%q", res.Stderr)
	require.Contains(t, res.Stdout, "prod", "active profile in report should be prod under --profile prod")

	raw, err := os.ReadFile(profile.RepoProfilePath(f.RepoDir))
	require.NoError(t, err)
	var onDisk map[string]any
	require.NoError(t, json.Unmarshal(raw, &onDisk))
	require.Equal(t, "default", onDisk["active_profile"], "--profile must not persist the override")
}

// TestProfileFlag_UnknownNameErrors asserts an unknown --profile name fails the invocation
// instead of silently falling back to default.
func TestProfileFlag_UnknownNameErrors(t *testing.T) {
	f := newFixture(t)
	require.Equalf(t, 0, f.Run("profile", "init").ExitCode, "init must succeed")

	res := f.Run("--profile", "nope", "profile", "where")
	require.NotEqualf(t, 0, res.ExitCode, "unknown profile must error; stderr=%q", res.Stderr)
}

func splitLines(s string) []string {
	var out []string
	start := 0
	for i, r := range s {
		if r == '\n' {
			out = append(out, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		out = append(out, s[start:])
	}
	return out
}

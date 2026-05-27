package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/zeropsio/zcli/src/profile"
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

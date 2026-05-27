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
	require.NotEmpty(t, onDisk["id"], "init must stamp a stable id")
	require.EqualValues(t, profile.CurrentRepoVersion, onDisk["version"])

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

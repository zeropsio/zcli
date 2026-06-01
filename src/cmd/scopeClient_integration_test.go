package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/zeropsio/zcli/src/profile"
	"github.com/zeropsio/zerops-go/types/uuid"
)

// TestProfileWhere_ClientScopeRepoOverride seeds a repo profile carrying a client scope and
// asserts `zcli profile where` attributes it to the repo layer. Pure local introspection - no
// API calls - so it exercises the repoProfileV2 round-trip + Manager.ScopeClientWithSource
// through the real CLI.
func TestProfileWhere_ClientScopeRepoOverride(t *testing.T) {
	f := newFixture(t)
	f.SeedLogin("test-token")
	f.writeRepoProfile(profile.RepoProfile{
		Name:          profile.DefaultRepoProfileName,
		Id:            "test-repo-id",
		ScopeClientId: uuid.NewClientIdNullFromString("repo-org"),
	})

	res := f.Run("profile", "where")
	require.Equalf(t, 0, res.ExitCode, "stderr=%q", res.Stderr)
	assert.Contains(t, res.Stdout, "repo-org")
	assert.Contains(t, res.Stdout, "from repo override")
}

// TestProfileWhere_ClientScopeGlobalFallback seeds only a global client scope (no repo profile)
// and asserts `zcli profile where` attributes it to the global layer.
func TestProfileWhere_ClientScopeGlobalFallback(t *testing.T) {
	f := newFixture(t)
	f.SeedLogin("test-token")
	// Overwrite the global profile SeedLogin wrote with one that also carries a client scope.
	f.writeProfile(profile.Data{
		Region:        profile.RegionFromItem(f.Region),
		ScopeClientId: uuid.NewClientIdNullFromString("global-org"),
	})

	res := f.Run("profile", "where")
	require.Equalf(t, 0, res.ExitCode, "stderr=%q", res.Stderr)
	assert.Contains(t, res.Stdout, "global-org")
	assert.Contains(t, res.Stdout, "from global")
}

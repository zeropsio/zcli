//go:build devel

package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const scopeProjectID = "00000000-0000-0000-0000-0000000000bb"

func TestScopeProject_PersistsByID(t *testing.T) {
	f := newFixture(t)
	f.SeedLogin("test-token")

	f.HandleJSON("/api/rest/public/project/"+scopeProjectID, 200, map[string]any{
		"id":          scopeProjectID,
		"clientId":    "00000000-0000-0000-0000-0000000000aa",
		"name":        "demo-project",
		"mode":        "LIGHT",
		"status":      "ACTIVE",
		"created":     "2024-01-01T00:00:00.000Z",
		"lastUpdate":  "2024-01-01T00:00:00.000Z",
		"tagList":     []string{},
		"description": nil,
	})

	res := f.Run(nil, "scope", "project", scopeProjectID)

	require.Equalf(t, 0, res.ExitCode, "stderr=%q", res.Stderr)
	assert.Contains(t, res.Stderr, "demo-project", "selected-project info should reference the persisted project")

	stored := f.LoadStorage()
	id, ok := stored.ScopeProjectId.Get()
	require.True(t, ok, "ScopeProjectId not persisted")
	assert.Equal(t, scopeProjectID, id.Native(), "wrong project id persisted")
}

func TestScopeReset_ClearsPersistedScope(t *testing.T) {
	f := newFixture(t)
	f.SeedScopedLogin("test-token", scopeProjectID)

	res := f.Run(nil, "scope", "reset")

	require.Equalf(t, 0, res.ExitCode, "stderr=%q", res.Stderr)
	assert.Contains(t, res.Stderr, "Scope was reset", "reset message should appear")

	stored := f.LoadStorage()
	_, ok := stored.ScopeProjectId.Get()
	assert.False(t, ok, "ScopeProjectId should be cleared")
}

//go:build devel

package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLogoutCommand_ClearsStorageAndPrintsSuccess(t *testing.T) {
	f := newFixture(t)
	f.SeedLogin("seed-token")

	f.HandleJSON("/api/rest/public/auth/logout", 200, map[string]any{})

	res := f.RunKong(nil, "logout")

	require.Equalf(t, 0, res.ExitCode, "stderr=%q", res.Stderr)
	assert.Contains(t, res.Stderr, "Successfully logged out",
		"success message should appear")

	stored := f.LoadStorage()
	assert.Equal(t, "", stored.Token, "token must be cleared after logout")
}

//go:build devel

package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestLoginCommand_PersistsTokenAndRegion checks that `zcli login <token>
// --region-url ...` resolves the region from the mocked region catalog,
// fetches GetUserInfo, persists token + region to cliStorage, and prints a
// success message naming the user.
func TestLoginCommand_PersistsTokenAndRegion(t *testing.T) {
	f := newFixture(t)

	// Region catalog served at --region-url. The login command picks the
	// region named "prg1" by default.
	f.HandleJSON("/regions", 200, map[string]any{
		"items": []map[string]any{{
			"name":      "prg1",
			"isDefault": true,
			"address":   f.Server.URL,
		}},
	})

	// GetUserInfo — only email/fullName are read by the login command;
	// other fields decode to zero values.
	f.HandleJSON("/api/rest/public/user/info", 200, map[string]any{
		"email":    "tester@example.com",
		"fullName": "Test User",
	})

	res := f.Run(nil, "login", "secret-token", "--region-url", f.Server.URL+"/regions")

	require.Equalf(t, 0, res.ExitCode, "stderr=%q", res.Stderr)
	assert.Contains(t, res.Stderr, "Test User", "success message should name the user")

	stored := f.LoadStorage()
	assert.Equal(t, "secret-token", stored.Token, "token not persisted")
	assert.Equal(t, "prg1", stored.RegionData.Name, "region name not persisted")
	assert.Equal(t, f.Server.URL, stored.RegionData.Address, "region address not persisted")
}

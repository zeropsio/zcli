package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestProjectListCommand checks that `zcli project list` walks GetUserInfo →
// org filtering (ACTIVE only) → GetClientProject and renders a table that
// contains the project id, name, and org name.
func TestProjectListCommand(t *testing.T) {
	f := newFixture(t)
	f.SeedLogin("test-token")

	orgId := "00000000-0000-0000-0000-0000000000aa"

	// GetAllOrgs reads UserAuthorize.clientUserList. Only ACTIVE clients are
	// queried for projects.
	f.HandleJSON("/api/rest/public/user/info", 200, map[string]any{
		"email":    "tester@example.com",
		"fullName": "Test User",
		"clientUserList": []map[string]any{{
			"id":       "00000000-0000-0000-0000-000000000001",
			"clientId": orgId,
			"userId":   "00000000-0000-0000-0000-000000000002",
			"status":   "ACTIVE",
			"roleCode": "ADMIN",
			"client": map[string]any{
				"id":          orgId,
				"accountName": "Acme Org",
			},
		}},
	})

	f.HandleJSON("/api/rest/public/client/"+orgId+"/project", 200, map[string]any{
		"list": []map[string]any{{
			"id":          "00000000-0000-0000-0000-0000000000bb",
			"clientId":    orgId,
			"name":        "demo-project",
			"mode":        "LIGHT",
			"status":      "ACTIVE",
			"created":     "2024-01-01T00:00:00.000Z",
			"lastUpdate":  "2024-01-01T00:00:00.000Z",
			"tagList":     []string{},
			"description": nil,
		}},
		"totalCount": 1,
	})

	res := f.Run("project", "list")

	require.Equalf(t, 0, res.ExitCode, "stderr=%q", res.Stderr)
	for _, want := range []string{"demo-project", "Acme Org", "00000000-0000-0000-0000-0000000000bb"} {
		assert.Containsf(t, res.Stdout, want, "stdout should contain %q", want)
	}
}

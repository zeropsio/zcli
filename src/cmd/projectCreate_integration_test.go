//go:build devel

package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	projectCreateOrgID     = "00000000-0000-0000-0000-0000000000aa"
	projectCreateNewProjID = "00000000-0000-0000-0000-0000000000cc"
)

func TestProjectCreate_PostsProjectAndRendersTemplate(t *testing.T) {
	f := newFixture(t)
	f.SeedLogin("test-token")

	// GetUserInfo / GetAllOrgs lookup. Only ACTIVE client-users surface as orgs.
	f.HandleJSON("/api/rest/public/user/info", 200, map[string]any{
		"email":    "tester@example.com",
		"fullName": "Test User",
		"clientUserList": []map[string]any{{
			"id":       "00000000-0000-0000-0000-000000000001",
			"clientId": projectCreateOrgID,
			"userId":   "00000000-0000-0000-0000-000000000002",
			"status":   "ACTIVE",
			"roleCode": "ADMIN",
			"client": map[string]any{
				"id":          projectCreateOrgID,
				"accountName": "Acme Org",
			},
		}},
	})

	// POST client/<orgId>/project — the create endpoint.
	f.HandleJSON("/api/rest/public/client/"+projectCreateOrgID+"/project", 201, map[string]any{
		"id":          projectCreateNewProjID,
		"clientId":    projectCreateOrgID,
		"name":        "demo-project",
		"mode":        "LIGHT",
		"status":      "ACTIVE",
		"created":     "2024-01-01T00:00:00.000Z",
		"lastUpdate":  "2024-01-01T00:00:00.000Z",
		"tagList":     []string{},
		"description": nil,
	})

	res := f.Run(nil,
		"project", "create",
		"--name", "demo-project",
		"--org-id", projectCreateOrgID,
		"--mode", "LIGHT",
		"--out", "created:{{.Id}}",
	)

	require.Equalf(t, 0, res.ExitCode, "stderr=%q", res.Stderr)
	assert.Contains(t, res.Stdout, "created:"+projectCreateNewProjID,
		"--out template should render with the new project id")
}

//go:build devel

package cmd

import (
	"strings"
	"testing"
)

func TestProjectListCommand(t *testing.T) {
	f := newFixture(t)
	f.SeedLogin("test-token")

	// GetAllOrgs reads UserAuthorize.clientUserList. Only ACTIVE clients are
	// queried for projects.
	f.HandleJSON("/api/rest/public/user/info", 200, map[string]any{
		"email":    "tester@example.com",
		"fullName": "Test User",
		"clientUserList": []map[string]any{{
			"id":       "00000000-0000-0000-0000-000000000001",
			"clientId": "00000000-0000-0000-0000-0000000000aa",
			"userId":   "00000000-0000-0000-0000-000000000002",
			"status":   "ACTIVE",
			"roleCode": "ADMIN",
			"client": map[string]any{
				"id":          "00000000-0000-0000-0000-0000000000aa",
				"accountName": "Acme Org",
			},
		}},
	})

	f.HandleJSON("/api/rest/public/project/search", 200, map[string]any{
		"limit":     50,
		"offset":    0,
		"totalHits": 1,
		"items": []map[string]any{{
			"id":          "00000000-0000-0000-0000-0000000000bb",
			"clientId":    "00000000-0000-0000-0000-0000000000aa",
			"name":        "demo-project",
			"mode":        "LIGHT",
			"status":      "ACTIVE",
			"created":     "2024-01-01T00:00:00.000Z",
			"lastUpdate":  "2024-01-01T00:00:00.000Z",
			"tagList":     []string{},
			"description": nil,
		}},
	})

	res := f.Run(nil, "project", "list")

	if res.ExitCode != 0 {
		t.Fatalf("exit=%d stderr=%q", res.ExitCode, res.Stderr)
	}
	for _, want := range []string{"demo-project", "Acme Org", "00000000-0000-0000-0000-0000000000bb"} {
		if !strings.Contains(res.Stdout, want) {
			t.Errorf("stdout missing %q\n--- stdout ---\n%s", want, res.Stdout)
		}
	}
}

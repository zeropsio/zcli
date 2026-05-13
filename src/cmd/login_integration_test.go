//go:build devel

package cmd

import (
	"strings"
	"testing"
)

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

	if res.ExitCode != 0 {
		t.Fatalf("exit=%d stderr=%q", res.ExitCode, res.Stderr)
	}
	if !strings.Contains(res.Stderr, "Test User") {
		t.Errorf("expected success message naming the user, got stderr=%q", res.Stderr)
	}

	stored := f.LoadStorage()
	if stored.Token != "secret-token" {
		t.Errorf("token not persisted: got %q", stored.Token)
	}
	if stored.RegionData.Name != "prg1" || stored.RegionData.Address != f.Server.URL {
		t.Errorf("region not persisted: %+v", stored.RegionData)
	}
}

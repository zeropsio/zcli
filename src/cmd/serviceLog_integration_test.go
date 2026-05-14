//go:build devel

package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	serviceLogProjID = "00000000-0000-0000-0000-0000000000bb"
	serviceLogSvcID  = "00000000-0000-0000-0000-000000000555"
)

// TestServiceLog_ShowBuildLogsNoBuild covers the --show-build-logs
// failure path (no app version exists). Asserts the kong-side migration
// is wired correctly.
//
// Cobra-path note: --show-build-logs is *silently broken* in the cobra
// version. cmd reads cmdData.Params.GetBool("showBuildLogs") (camelCase)
// but the flag is registered as show-build-logs (kebab); viper's
// dual-lookup doesn't bridge them for this combination, so the bool is
// always false in cobra. The kong port reads the typed field directly,
// so the bug can't recur. Out of scope for the port per the
// test-first-migration rule.
func TestServiceLog_ShowBuildLogsNoBuild(t *testing.T) {
	f := newFixture(t)
	f.SeedLogin("test-token")

	f.HandleJSON("/api/rest/public/project/"+serviceLogProjID, 200, map[string]any{
		"id":         serviceLogProjID,
		"clientId":   "00000000-0000-0000-0000-0000000000aa",
		"name":       "host",
		"mode":       "LIGHT",
		"status":     "ACTIVE",
		"created":    "2024-01-01T00:00:00.000Z",
		"lastUpdate": "2024-01-01T00:00:00.000Z",
		"tagList":    []string{},
	})
	f.HandleJSON("/api/rest/public/service-stack/"+serviceLogSvcID, 200, map[string]any{
		"id":         serviceLogSvcID,
		"projectId":  serviceLogProjID,
		"clientId":   "00000000-0000-0000-0000-0000000000aa",
		"name":       "api",
		"status":     "ACTIVE",
		"mode":       "NON_HA",
		"created":    "2024-01-01T00:00:00.000Z",
		"lastUpdate": "2024-01-01T00:00:00.000Z",
		"serviceStackTypeInfo": map[string]any{
			"serviceStackTypeId":          "nodejs",
			"serviceStackTypeName":        "Node.js",
			"serviceStackTypeVersionId":   "nodejs@22",
			"serviceStackTypeVersionName": "22",
		},
	})

	// Empty app-version list → triggers LogNoBuildFound branch.
	f.HandleJSON("/api/rest/public/app-version/search", 200, map[string]any{
		"limit":     1,
		"offset":    0,
		"totalHits": 0,
		"items":     []map[string]any{},
	})

	res := f.RunKong(nil, "service", "log",
		"--service-id", serviceLogSvcID,
		"--show-build-logs",
	)

	require.Equal(t, 1, res.ExitCode, "should exit non-zero when no build exists")
	assert.Contains(t, res.Stderr, "No build", "error message should indicate no build")
}

//go:build devel

package cmd

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	serviceLifecycleProjID    = "00000000-0000-0000-0000-0000000000bb"
	serviceLifecycleServiceID = "00000000-0000-0000-0000-000000000111"
	serviceLifecycleProcID    = "00000000-0000-0000-0000-000000000222"
)

// seedServiceLifecycle wires GET project, GET service, and POST process
// handlers that the start/stop/enable-subdomain test cases all share.
// action is the action-name reported back from the PUT response and the
// process polling endpoint (e.g. "stack.start").
func seedServiceLifecycle(t *testing.T, f *fixture, putPath, action, finalStatus string) {
	t.Helper()
	f.HandleJSON("/api/rest/public/project/"+serviceLifecycleProjID, 200, map[string]any{
		"id":         serviceLifecycleProjID,
		"clientId":   "00000000-0000-0000-0000-0000000000aa",
		"name":       "host",
		"mode":       "LIGHT",
		"status":     "ACTIVE",
		"created":    "2024-01-01T00:00:00.000Z",
		"lastUpdate": "2024-01-01T00:00:00.000Z",
		"tagList":    []string{},
	})
	f.HandleJSON("/api/rest/public/service-stack/"+serviceLifecycleServiceID, 200, map[string]any{
		"id":         serviceLifecycleServiceID,
		"projectId":  serviceLifecycleProjID,
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
	f.Mux.HandleFunc(putPath, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":         serviceLifecycleProcID,
			"clientId":   "00000000-0000-0000-0000-0000000000aa",
			"actionName": action,
			"status":     "RUNNING",
			"created":    "2024-01-01T00:00:00.000Z",
			"lastUpdate": "2024-01-01T00:00:00.000Z",
		})
	})
	f.HandleJSON("/api/rest/public/process/"+serviceLifecycleProcID, 200, map[string]any{
		"id":         serviceLifecycleProcID,
		"clientId":   "00000000-0000-0000-0000-0000000000aa",
		"actionName": action,
		"status":     finalStatus,
		"created":    "2024-01-01T00:00:00.000Z",
		"lastUpdate": "2024-01-01T00:00:00.000Z",
	})
}

func TestServiceStart(t *testing.T) {
	f := newFixture(t)
	f.SeedLogin("test-token")
	seedServiceLifecycle(t, f,
		"/api/rest/public/service-stack/"+serviceLifecycleServiceID+"/start",
		"stack.start", "FINISHED",
	)

	res := f.Run(nil, "service", "start", "--service-id", serviceLifecycleServiceID)

	require.Equalf(t, 0, res.ExitCode, "stderr=%q", res.Stderr)
	assert.Contains(t, res.Stderr, "api", "selected-service info should appear")
}

func TestServiceStop(t *testing.T) {
	f := newFixture(t)
	f.SeedLogin("test-token")
	seedServiceLifecycle(t, f,
		"/api/rest/public/service-stack/"+serviceLifecycleServiceID+"/stop",
		"stack.stop", "FINISHED",
	)

	res := f.Run(nil, "service", "stop", "--service-id", serviceLifecycleServiceID)

	require.Equalf(t, 0, res.ExitCode, "stderr=%q", res.Stderr)
}

func TestServiceEnableSubdomain(t *testing.T) {
	f := newFixture(t)
	f.SeedLogin("test-token")
	seedServiceLifecycle(t, f,
		"/api/rest/public/service-stack/"+serviceLifecycleServiceID+"/enable-subdomain-access",
		"stack.enableSubdomain", "FINISHED",
	)

	res := f.Run(nil, "service", "enable-subdomain", "--service-id", serviceLifecycleServiceID)

	require.Equalf(t, 0, res.ExitCode, "stderr=%q", res.Stderr)
}

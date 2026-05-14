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
	serviceCreateProjID  = "00000000-0000-0000-0000-0000000000bb"
	serviceCreateSvcID   = "00000000-0000-0000-0000-000000000666"
	serviceCreateProcID  = "00000000-0000-0000-0000-000000000777"
)

func TestServiceCreate_CreatesAndRendersTemplate(t *testing.T) {
	f := newFixture(t)
	f.SeedLogin("test-token")

	f.HandleJSON("/api/rest/public/project/"+serviceCreateProjID, 200, map[string]any{
		"id":         serviceCreateProjID,
		"clientId":   "00000000-0000-0000-0000-0000000000aa",
		"name":       "host",
		"mode":       "LIGHT",
		"status":     "ACTIVE",
		"created":    "2024-01-01T00:00:00.000Z",
		"lastUpdate": "2024-01-01T00:00:00.000Z",
		"tagList":    []string{},
	})

	svcBody := map[string]any{
		"id":         serviceCreateSvcID,
		"projectId":  serviceCreateProjID,
		"clientId":   "00000000-0000-0000-0000-0000000000aa",
		"name":       "new-api",
		"status":     "CREATING",
		"mode":       "NON_HA",
		"created":    "2024-01-01T00:00:00.000Z",
		"lastUpdate": "2024-01-01T00:00:00.000Z",
		"serviceStackTypeInfo": map[string]any{
			"serviceStackTypeId":          "nodejs",
			"serviceStackTypeName":        "Node.js",
			"serviceStackTypeVersionId":   "nodejs@22",
			"serviceStackTypeVersionName": "22",
		},
	}

	// POST .../service-stack/runtime returns Process + Service together.
	f.Mux.HandleFunc("/api/rest/public/project/"+serviceCreateProjID+"/service-stack/runtime", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":         serviceCreateSvcID,
			"projectId":  serviceCreateProjID,
			"clientId":   "00000000-0000-0000-0000-0000000000aa",
			"name":       "new-api",
			"status":     "CREATING",
			"mode":       "NON_HA",
			"created":    "2024-01-01T00:00:00.000Z",
			"lastUpdate": "2024-01-01T00:00:00.000Z",
			"serviceStackTypeInfo": map[string]any{
				"serviceStackTypeId":          "nodejs",
				"serviceStackTypeName":        "Node.js",
				"serviceStackTypeVersionId":   "nodejs@22",
				"serviceStackTypeVersionName": "22",
			},
			"process": map[string]any{
				"id":         serviceCreateProcID,
				"clientId":   "00000000-0000-0000-0000-0000000000aa",
				"actionName": "stack.create",
				"status":     "RUNNING",
				"created":    "2024-01-01T00:00:00.000Z",
				"lastUpdate": "2024-01-01T00:00:00.000Z",
			},
		})
	})

	// Process polling — return FINISHED so the spinner exits.
	f.HandleJSON("/api/rest/public/process/"+serviceCreateProcID, 200, map[string]any{
		"id":         serviceCreateProcID,
		"clientId":   "00000000-0000-0000-0000-0000000000aa",
		"actionName": "stack.create",
		"status":     "FINISHED",
		"created":    "2024-01-01T00:00:00.000Z",
		"lastUpdate": "2024-01-01T00:00:00.000Z",
	})

	// After spinner: GET /service-stack/<id> to refresh.
	f.HandleJSON("/api/rest/public/service-stack/"+serviceCreateSvcID, 200, svcBody)

	res := f.Run(nil,
		"service", "create",
		"--project-id", serviceCreateProjID,
		"--name", "new-api",
		"--mode", "NON_HA",
		"--out", "created:{{.Id}}",
	)

	require.Equalf(t, 0, res.ExitCode, "stderr=%q stdout=%q", res.Stderr, res.Stdout)
	assert.Contains(t, res.Stdout, "created:"+serviceCreateSvcID,
		"--out template should render with the new service id")
}

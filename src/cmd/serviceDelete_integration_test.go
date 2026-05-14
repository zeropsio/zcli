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
	serviceDeleteSvcID    = "00000000-0000-0000-0000-000000000333"
	serviceDeleteProcID   = "00000000-0000-0000-0000-000000000444"
	serviceDeleteParentID = "00000000-0000-0000-0000-0000000000bb"
)

func TestServiceDelete_DeletesAndPollsProcess(t *testing.T) {
	f := newFixture(t)
	f.SeedLogin("test-token")

	f.HandleJSON("/api/rest/public/project/"+serviceDeleteParentID, 200, map[string]any{
		"id":         serviceDeleteParentID,
		"clientId":   "00000000-0000-0000-0000-0000000000aa",
		"name":       "host",
		"mode":       "LIGHT",
		"status":     "ACTIVE",
		"created":    "2024-01-01T00:00:00.000Z",
		"lastUpdate": "2024-01-01T00:00:00.000Z",
		"tagList":    []string{},
	})

	svcBody := map[string]any{
		"id":         serviceDeleteSvcID,
		"projectId":  serviceDeleteParentID,
		"clientId":   "00000000-0000-0000-0000-0000000000aa",
		"name":       "doomed-service",
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
	}
	deleteResp := map[string]any{
		"id":         serviceDeleteProcID,
		"clientId":   "00000000-0000-0000-0000-0000000000aa",
		"actionName": "stack.delete",
		"status":     "RUNNING",
		"created":    "2024-01-01T00:00:00.000Z",
		"lastUpdate": "2024-01-01T00:00:00.000Z",
	}

	f.Mux.HandleFunc("/api/rest/public/service-stack/"+serviceDeleteSvcID, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		switch r.Method {
		case http.MethodDelete:
			_ = json.NewEncoder(w).Encode(deleteResp)
		default:
			_ = json.NewEncoder(w).Encode(svcBody)
		}
	})

	f.HandleJSON("/api/rest/public/process/"+serviceDeleteProcID, 200, map[string]any{
		"id":         serviceDeleteProcID,
		"clientId":   "00000000-0000-0000-0000-0000000000aa",
		"actionName": "stack.delete",
		"status":     "FINISHED",
		"created":    "2024-01-01T00:00:00.000Z",
		"lastUpdate": "2024-01-01T00:00:00.000Z",
	})

	res := f.RunKong(nil, "service", "delete", "--service-id", serviceDeleteSvcID, "--confirm")

	require.Equalf(t, 0, res.ExitCode, "stderr=%q", res.Stderr)
	assert.Contains(t, res.Stderr, "doomed-service", "selected-service info should appear")
}

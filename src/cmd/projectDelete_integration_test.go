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
	projectDeleteID     = "00000000-0000-0000-0000-0000000000bb"
	projectDeleteProcID = "00000000-0000-0000-0000-0000000000dd"
)

func TestProjectDelete_DeletesViaFlagAndWaitsForProcess(t *testing.T) {
	f := newFixture(t)
	f.SeedLogin("test-token")

	projectBody := map[string]any{
		"id":          projectDeleteID,
		"clientId":    "00000000-0000-0000-0000-0000000000aa",
		"name":        "doomed-project",
		"mode":        "LIGHT",
		"status":      "ACTIVE",
		"created":     "2024-01-01T00:00:00.000Z",
		"lastUpdate":  "2024-01-01T00:00:00.000Z",
		"tagList":     []string{},
		"description": nil,
	}
	deleteResponse := map[string]any{
		"id":         projectDeleteProcID,
		"clientId":   "00000000-0000-0000-0000-0000000000aa",
		"actionName": "stack.delete",
		"status":     "RUNNING",
		"created":    "2024-01-01T00:00:00.000Z",
		"lastUpdate": "2024-01-01T00:00:00.000Z",
	}

	// /project/<id> answers GET (RequireProject lookup) and DELETE (deletion)
	// from the same path with different bodies — branch on method.
	f.Mux.HandleFunc("/api/rest/public/project/"+projectDeleteID, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodDelete:
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(deleteResponse)
		default:
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(projectBody)
		}
	})

	// Process polling — returns FINISHED so the spinner exits cleanly.
	f.HandleJSON("/api/rest/public/process/"+projectDeleteProcID, 200, map[string]any{
		"id":         projectDeleteProcID,
		"clientId":   "00000000-0000-0000-0000-0000000000aa",
		"actionName": "stack.delete",
		"status":     "FINISHED",
		"created":    "2024-01-01T00:00:00.000Z",
		"lastUpdate": "2024-01-01T00:00:00.000Z",
	})

	res := f.Run(nil,
		"project", "delete",
		"--project-id", projectDeleteID,
		"--confirm",
	)

	require.Equalf(t, 0, res.ExitCode, "stderr=%q stdout=%q", res.Stderr, res.Stdout)
	assert.Contains(t, res.Stderr, "doomed-project", "selected-project info should appear")
}

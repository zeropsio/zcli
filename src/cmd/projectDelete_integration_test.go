package cmd

import (
	"encoding/json"
	"net/http"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

const deleteOtherProjectID = "0vzEq7ETUceKlQdesTBLhQ"

// registerProjectDeleteStubs serves GET and DELETE for the given project ids and returns a func listing the ids that received a DELETE.
func registerProjectDeleteStubs(f *fixture, projectIDs ...string) func() []string {
	var mu sync.Mutex
	var deleted []string

	for _, id := range projectIDs {
		f.Mux.HandleFunc("/api/rest/public/project/"+id, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			if r.Method == http.MethodDelete {
				mu.Lock()
				deleted = append(deleted, id)
				mu.Unlock()
				_ = json.NewEncoder(w).Encode(deleteProcessBody("RUNNING"))
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id":         id,
				"clientId":   pushClientID,
				"name":       "project-" + id,
				"mode":       "LIGHT",
				"status":     "ACTIVE",
				"created":    "2024-01-01T00:00:00.000Z",
				"lastUpdate": "2024-01-01T00:00:00.000Z",
				"tagList":    []string{},
			})
		})
	}
	f.HandleJSON("/api/rest/public/process/"+pushDeployProcID, 200, deleteProcessBody("FINISHED"))

	return func() []string {
		mu.Lock()
		defer mu.Unlock()
		return append([]string(nil), deleted...)
	}
}

func deleteProcessBody(status string) map[string]any {
	return map[string]any{
		"id":              pushDeployProcID,
		"clientId":        pushClientID,
		"projectId":       pushProjectID,
		"status":          status,
		"sequence":        1,
		"actionName":      "DeleteProject",
		"created":         "2024-01-01T00:00:00.000Z",
		"lastUpdate":      "2024-01-01T00:00:00.000Z",
		"createdBySystem": false,
		"createdByUser":   map[string]any{},
		"serviceStacks":   []any{},
	}
}

// TestProjectDeleteCommand_ExplicitProjectBeatsSavedScope checks that a project id given explicitly, positional or as --project-id,
// is the project that gets deleted even when a different project is saved as the scope, and that the saved scope still applies without one.
func TestProjectDeleteCommand_ExplicitProjectBeatsSavedScope(t *testing.T) {
	tests := []struct {
		name       string
		scoped     bool
		args       []string
		wantDelete string
	}{
		{
			name:       "positional id with a saved scope",
			scoped:     true,
			args:       []string{"project", "delete", deleteOtherProjectID, "--confirm"},
			wantDelete: deleteOtherProjectID,
		},
		{
			name:       "positional id without a saved scope",
			args:       []string{"project", "delete", deleteOtherProjectID, "--confirm"},
			wantDelete: deleteOtherProjectID,
		},
		{
			name:       "flag with a saved scope",
			scoped:     true,
			args:       []string{"project", "delete", "--project-id", deleteOtherProjectID, "--confirm"},
			wantDelete: deleteOtherProjectID,
		},
		{
			name:       "flag beats positional when both are given",
			scoped:     true,
			args:       []string{"project", "delete", pushProjectID, "--project-id", deleteOtherProjectID, "--confirm"},
			wantDelete: deleteOtherProjectID,
		},
		{
			name:       "saved scope without an explicit id",
			scoped:     true,
			args:       []string{"project", "delete", "--confirm"},
			wantDelete: pushProjectID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newFixture(t)
			if tt.scoped {
				f.SeedScopedLogin("test-token", pushProjectID)
			} else {
				f.SeedLogin("test-token")
			}
			deleted := registerProjectDeleteStubs(f, pushProjectID, deleteOtherProjectID)

			res := f.Run(tt.args...)

			require.Equalf(t, 0, res.ExitCode, "stderr=%q stdout=%q", res.Stderr, res.Stdout)
			require.Equal(t, []string{tt.wantDelete}, deleted())
		})
	}
}

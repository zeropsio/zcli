//go:build devel

package cmd

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const projectEnvProjID = "00000000-0000-0000-0000-0000000000bb"

func TestProjectEnv_RendersTemplatePerLine(t *testing.T) {
	f := newFixture(t)
	f.SeedLogin("test-token")

	projectBody := map[string]any{
		"id":          projectEnvProjID,
		"clientId":    "00000000-0000-0000-0000-0000000000aa",
		"name":        "env-project",
		"mode":        "LIGHT",
		"status":      "ACTIVE",
		"created":     "2024-01-01T00:00:00.000Z",
		"lastUpdate":  "2024-01-01T00:00:00.000Z",
		"tagList":     []string{},
		"description": nil,
	}
	f.HandleJSON("/api/rest/public/project/"+projectEnvProjID, 200, projectBody)

	// Env-file endpoint returns the envs as a single newline-separated blob.
	f.Mux.HandleFunc("/api/rest/public/project/"+projectEnvProjID+"/env-file", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"envFile": "FOO=bar\nBAZ=qux",
		})
	})

	res := f.Run(nil, "project", "env", "--project-id", projectEnvProjID)

	require.Equalf(t, 0, res.ExitCode, "stderr=%q stdout=%q", res.Stderr, res.Stdout)
	assert.Contains(t, res.Stdout, "FOO=bar")
	assert.Contains(t, res.Stdout, "BAZ=qux")
}

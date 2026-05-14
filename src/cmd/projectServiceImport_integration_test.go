//go:build devel

package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	serviceImportProjID = "00000000-0000-0000-0000-0000000000bb"
	serviceImportProcID = "00000000-0000-0000-0000-000000000111"
)

func TestProjectServiceImport_PostsYamlAndPollsProcesses(t *testing.T) {
	f := newFixture(t)
	f.SeedLogin("test-token")

	dir := t.TempDir()
	yamlPath := filepath.Join(dir, "svc.yml")
	require.NoError(t, os.WriteFile(yamlPath, []byte(`services:
  - hostname: api
    type: nodejs@22
`), 0o600))

	// RequireProject lookup
	f.HandleJSON("/api/rest/public/project/"+serviceImportProjID, 200, map[string]any{
		"id":          serviceImportProjID,
		"clientId":    "00000000-0000-0000-0000-0000000000aa",
		"name":        "host-project",
		"mode":        "LIGHT",
		"status":      "ACTIVE",
		"created":     "2024-01-01T00:00:00.000Z",
		"lastUpdate":  "2024-01-01T00:00:00.000Z",
		"tagList":     []string{},
		"description": nil,
	})

	f.HandleJSON("/api/rest/public/project/"+serviceImportProjID+"/service-stack/import", 200, map[string]any{
		"serviceStacks": []map[string]any{{
			"id":     "00000000-0000-0000-0000-000000000fff",
			"name":   "api",
			"status": "CREATING",
			"processes": []map[string]any{{
				"id":         serviceImportProcID,
				"clientId":   "00000000-0000-0000-0000-0000000000aa",
				"actionName": "stack.create",
				"status":     "RUNNING",
				"created":    "2024-01-01T00:00:00.000Z",
				"lastUpdate": "2024-01-01T00:00:00.000Z",
			}},
		}},
	})

	f.HandleJSON("/api/rest/public/process/"+serviceImportProcID, 200, map[string]any{
		"id":         serviceImportProcID,
		"clientId":   "00000000-0000-0000-0000-0000000000aa",
		"actionName": "stack.create",
		"status":     "FINISHED",
		"created":    "2024-01-01T00:00:00.000Z",
		"lastUpdate": "2024-01-01T00:00:00.000Z",
	})

	res := f.RunKong(nil,
		"project", "service-import", yamlPath,
		"--project-id", serviceImportProjID,
	)

	require.Equalf(t, 0, res.ExitCode, "stderr=%q stdout=%q", res.Stderr, res.Stdout)
	assert.Contains(t, res.Stderr, "service(s) imported",
		"success message should appear")
}

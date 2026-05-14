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
	projectImportOrgID  = "00000000-0000-0000-0000-0000000000aa"
	projectImportProcID = "00000000-0000-0000-0000-0000000000ee"
)

func TestProjectImport_PostsYamlAndPollsProcesses(t *testing.T) {
	f := newFixture(t)
	f.SeedLogin("test-token")

	workDir := t.TempDir()
	yamlPath := filepath.Join(workDir, "import.yml")
	require.NoError(t, os.WriteFile(yamlPath, []byte(`project:
  name: imported-project
services:
  - hostname: api
    type: nodejs@22
`), 0o600))

	f.HandleJSON("/api/rest/public/client/"+projectImportOrgID+"/project/import", 200, map[string]any{
		"serviceStacks": []map[string]any{{
			"id":     "00000000-0000-0000-0000-0000000000ff",
			"name":   "api",
			"status": "CREATING",
			"processes": []map[string]any{{
				"id":         projectImportProcID,
				"clientId":   projectImportOrgID,
				"actionName": "stack.create",
				"status":     "RUNNING",
				"created":    "2024-01-01T00:00:00.000Z",
				"lastUpdate": "2024-01-01T00:00:00.000Z",
			}},
		}},
	})

	f.HandleJSON("/api/rest/public/process/"+projectImportProcID, 200, map[string]any{
		"id":         projectImportProcID,
		"clientId":   projectImportOrgID,
		"actionName": "stack.create",
		"status":     "FINISHED",
		"created":    "2024-01-01T00:00:00.000Z",
		"lastUpdate": "2024-01-01T00:00:00.000Z",
	})

	res := f.Run(nil,
		"project", "project-import", yamlPath,
		"--org-id", projectImportOrgID,
		"--working-dir", workDir,
	)

	require.Equalf(t, 0, res.ExitCode, "stderr=%q stdout=%q", res.Stderr, res.Stdout)
	assert.Contains(t, res.Stderr, "project imported", "success message should appear")
}

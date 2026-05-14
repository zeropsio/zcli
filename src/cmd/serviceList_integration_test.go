//go:build devel

package cmd

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const serviceListProjID = "00000000-0000-0000-0000-0000000000bb"

func TestServiceList_RendersServiceTable(t *testing.T) {
	f := newFixture(t)
	f.SeedLogin("test-token")

	f.HandleJSON("/api/rest/public/project/"+serviceListProjID, 200, map[string]any{
		"id":          serviceListProjID,
		"clientId":    "00000000-0000-0000-0000-0000000000aa",
		"name":        "host",
		"mode":        "LIGHT",
		"status":      "ACTIVE",
		"created":     "2024-01-01T00:00:00.000Z",
		"lastUpdate":  "2024-01-01T00:00:00.000Z",
		"tagList":     []string{},
		"description": nil,
	})

	f.Mux.HandleFunc("/api/rest/public/service-stack/search", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"limit":     50,
			"offset":    0,
			"totalHits": 1,
			"items": []map[string]any{{
				"id":         "00000000-0000-0000-0000-000000000111",
				"projectId":  serviceListProjID,
				"clientId":   "00000000-0000-0000-0000-0000000000aa",
				"name":       "api",
				"status":     "ACTIVE",
				"mode":       "NON_HA",
				"created":    "2024-01-01T00:00:00.000Z",
				"lastUpdate": "2024-01-01T00:00:00.000Z",
				"serviceStackTypeInfo": map[string]any{
					"serviceStackTypeId":         "nodejs",
					"serviceStackTypeName":       "Node.js",
					"serviceStackTypeVersionId":  "nodejs@22",
					"serviceStackTypeVersionName": "22",
				},
			}},
		})
	})

	res := f.RunKong(nil, "service", "list", "--project-id", serviceListProjID)

	require.Equalf(t, 0, res.ExitCode, "stderr=%q", res.Stderr)
	assert.Contains(t, res.Stdout, "api", "service name should appear in the table")
}

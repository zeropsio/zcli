//go:build devel

package cmd

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
)

const (
	pushServiceID    = "0000000000000000000000"
	pushProjectID    = "1111111111111111111111"
	pushClientID     = "2222222222222222222222"
	pushAppVersionID = "3333333333333333333333"
	pushDeployProcID = "4444444444444444444444"
)

func TestServicePushCommand_HappyPath(t *testing.T) {
	f := newFixture(t)
	f.SeedLogin("test-token")

	// Working directory with a minimal zerops.yaml. The setup name "demo"
	// matches the service name returned by the API, so the setup selector
	// short-circuits.
	workDir := t.TempDir()
	if err := os.WriteFile(
		filepath.Join(workDir, "zerops.yaml"),
		[]byte("zerops:\n  - setup: demo\n    build:\n      base: ubuntu\n"),
		0o600,
	); err != nil {
		t.Fatal(err)
	}

	// --- API stubs -----------------------------------------------------------

	f.HandleJSON("/api/rest/public/service-stack/"+pushServiceID, 200, map[string]any{
		"id":                 pushServiceID,
		"projectId":          pushProjectID,
		"name":               "demo",
		"status":             "ACTIVE",
		"serviceStackTypeId": "nodejs@20",
		"serviceStackTypeInfo": map[string]any{
			"serviceStackTypeName":        "Node.js",
			"serviceStackTypeCategory":    "USER",
			"serviceStackTypeVersionName": "20",
		},
		"project": map[string]any{
			"id":         pushProjectID,
			"clientId":   pushClientID,
			"name":       "demo-project",
			"mode":       "LIGHT",
			"status":     "ACTIVE",
			"created":    "2024-01-01T00:00:00.000Z",
			"lastUpdate": "2024-01-01T00:00:00.000Z",
			"tagList":    []string{},
		},
		"serviceStackTypeVersionId": "nodejs@20",
		"created":                   "2024-01-01T00:00:00.000Z",
		"lastUpdate":                "2024-01-01T00:00:00.000Z",
		"mode":                      "NON_HA",
	})

	f.HandleJSON("/api/rest/public/project/"+pushProjectID, 200, map[string]any{
		"id":         pushProjectID,
		"clientId":   pushClientID,
		"name":       "demo-project",
		"mode":       "LIGHT",
		"status":     "ACTIVE",
		"created":    "2024-01-01T00:00:00.000Z",
		"lastUpdate": "2024-01-01T00:00:00.000Z",
		"tagList":    []string{},
	})

	f.HandleJSON("/api/rest/public/service-stack/zerops-yaml-validation", 200, map[string]any{})

	uploadURL := f.Server.URL + "/upload/" + pushAppVersionID
	f.HandleJSON("/api/rest/public/service-stack/"+pushServiceID+"/app-version", 200, map[string]any{
		"id":             pushAppVersionID,
		"clientId":       pushClientID,
		"projectId":      pushProjectID,
		"serviceStackId": pushServiceID,
		"sequence":       1,
		"status":         "UPLOADING",
		"userDataList":   []any{},
		"created":        "2024-01-01T00:00:00.000Z",
		"lastUpdate":     "2024-01-01T00:00:00.000Z",
		"uploadUrl":      uploadURL,
	})

	// Upload endpoint records that it received a non-empty body.
	var uploadBytes atomic.Int64
	f.Mux.HandleFunc("/upload/"+pushAppVersionID, func(w http.ResponseWriter, r *http.Request) {
		n, _ := io.Copy(io.Discard, r.Body)
		uploadBytes.Store(n)
		w.WriteHeader(http.StatusOK)
	})

	var deployHits atomic.Int32
	f.Mux.HandleFunc("/api/rest/public/app-version/"+pushAppVersionID+"/build-and-deploy", func(w http.ResponseWriter, r *http.Request) {
		deployHits.Add(1)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":              pushDeployProcID,
			"clientId":        pushClientID,
			"projectId":       pushProjectID,
			"status":          "RUNNING",
			"sequence":        1,
			"actionName":      "stack.build_and_deploy",
			"created":         "2024-01-01T00:00:00.000Z",
			"lastUpdate":      "2024-01-01T00:00:00.000Z",
			"createdBySystem": false,
			"createdByUser":   map[string]any{},
			"serviceStacks":   []any{},
		})
	})

	// Process polling returns FINISHED on the very first call so the spinner
	// exits after ~1s (CheckZeropsProcess waits one ticker interval).
	f.HandleJSON("/api/rest/public/process/"+pushDeployProcID, 200, map[string]any{
		"id":              pushDeployProcID,
		"clientId":        pushClientID,
		"projectId":       pushProjectID,
		"status":          "FINISHED",
		"sequence":        1,
		"actionName":      "stack.build_and_deploy",
		"created":         "2024-01-01T00:00:00.000Z",
		"lastUpdate":      "2024-01-01T00:00:00.000Z",
		"createdBySystem": false,
		"createdByUser":   map[string]any{},
		"serviceStacks":   []any{},
	})

	// --- Run -----------------------------------------------------------------

	res := f.Run(nil,
		"service", "push",
		"--service-id", pushServiceID,
		"--working-dir", workDir,
		"--setup", "demo",
		"--no-git",
		"--disable-logs",
	)

	if res.ExitCode != 0 {
		t.Fatalf("exit=%d\n--- stderr ---\n%s\n--- stdout ---\n%s", res.ExitCode, res.Stderr, res.Stdout)
	}
	if uploadBytes.Load() == 0 {
		t.Error("upload handler received no bytes")
	}
	if deployHits.Load() != 1 {
		t.Errorf("build-and-deploy called %d times, want 1", deployHits.Load())
	}
	if !strings.Contains(res.Stderr, "demo") {
		t.Errorf("stderr should mention the service 'demo'; got: %q", res.Stderr)
	}
}

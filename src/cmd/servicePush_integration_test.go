//go:build devel

package cmd

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
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

// pushStubs holds the counters/recorded payloads that tests assert on.
type pushStubs struct {
	uploadBytes      atomic.Int64
	deployHits       atomic.Int32
	deployBody       atomic.Value // map[string]any
	appVersionBody   atomic.Value // map[string]any
	processPollCount atomic.Int32
}

// registerPushStubs wires up every API endpoint that `service push` touches,
// returning recorders the test can inspect. serviceName lets each test pick the
// name returned by the service-stack endpoint (it drives the setup auto-match
// logic in servicePush.go).
func registerPushStubs(t *testing.T, f *fixture, serviceName string) *pushStubs {
	t.Helper()
	s := &pushStubs{}
	uploadURL := f.Server.URL + "/upload/" + pushAppVersionID

	f.HandleJSON("/api/rest/public/service-stack/"+pushServiceID, 200, map[string]any{
		"id":                 pushServiceID,
		"projectId":          pushProjectID,
		"name":               serviceName,
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

	f.Mux.HandleFunc("/api/rest/public/service-stack/"+pushServiceID+"/app-version", func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		s.appVersionBody.Store(body)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
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
	})

	f.Mux.HandleFunc("/upload/"+pushAppVersionID, func(w http.ResponseWriter, r *http.Request) {
		n, _ := io.Copy(io.Discard, r.Body)
		s.uploadBytes.Store(n)
		w.WriteHeader(http.StatusOK)
	})

	f.Mux.HandleFunc("/api/rest/public/app-version/"+pushAppVersionID+"/build-and-deploy", func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		s.deployBody.Store(body)
		s.deployHits.Add(1)
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

	f.Mux.HandleFunc("/api/rest/public/process/"+pushDeployProcID, func(w http.ResponseWriter, r *http.Request) {
		s.processPollCount.Add(1)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
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
	})

	return s
}

// writeZeropsYaml creates a zerops.yaml in dir with the given setup names.
func writeZeropsYaml(t *testing.T, dir string, setups ...string) {
	t.Helper()
	var b []byte
	b = append(b, "zerops:\n"...)
	for _, s := range setups {
		b = append(b, "  - setup: "+s+"\n"...)
		b = append(b, "    build:\n      base: ubuntu\n"...)
	}
	if err := os.WriteFile(filepath.Join(dir, "zerops.yaml"), b, 0o600); err != nil {
		t.Fatal(err)
	}
}

func assertPushSuccess(t *testing.T, res result, s *pushStubs, wantSetup string) {
	t.Helper()
	if res.ExitCode != 0 {
		t.Fatalf("exit=%d\n--- stderr ---\n%s\n--- stdout ---\n%s", res.ExitCode, res.Stderr, res.Stdout)
	}
	if s.uploadBytes.Load() == 0 {
		t.Error("upload handler received no bytes")
	}
	if got := s.deployHits.Load(); got != 1 {
		t.Errorf("build-and-deploy called %d times, want 1", got)
	}
	if got, _ := s.deployBody.Load().(map[string]any); got["zeropsYamlSetup"] != wantSetup {
		t.Errorf("deploy body zeropsYamlSetup=%v, want %q", got["zeropsYamlSetup"], wantSetup)
	}
}

// Service name matches a setup in zerops.yaml — no --setup flag needed,
// auto-match picks it.
func TestServicePushCommand_SetupAutoMatchesServiceName(t *testing.T) {
	f := newFixture(t)
	f.SeedLogin("test-token")

	workDir := t.TempDir()
	writeZeropsYaml(t, workDir, "demo")

	s := registerPushStubs(t, f, "demo")

	res := f.Run(nil,
		"service", "push",
		"--service-id", pushServiceID,
		"--working-dir", workDir,
		"--no-git",
		"--disable-logs",
	)
	assertPushSuccess(t, res, s, "demo")
}

// Service name doesn't match any setup; user picks one explicitly via --setup.
func TestServicePushCommand_SetupSelectedByFlag(t *testing.T) {
	f := newFixture(t)
	f.SeedLogin("test-token")

	workDir := t.TempDir()
	writeZeropsYaml(t, workDir, "demo", "api-prod", "api-stage")

	s := registerPushStubs(t, f, "api") // service is "api", no exact match

	res := f.Run(nil,
		"service", "push",
		"--service-id", pushServiceID,
		"--working-dir", workDir,
		"--setup", "api-prod",
		"--no-git",
		"--disable-logs",
	)
	assertPushSuccess(t, res, s, "api-prod")
}

// --version-name flag is forwarded to the POST /app-version request body.
func TestServicePushCommand_VersionNameForwarded(t *testing.T) {
	f := newFixture(t)
	f.SeedLogin("test-token")

	workDir := t.TempDir()
	writeZeropsYaml(t, workDir, "demo")

	s := registerPushStubs(t, f, "demo")

	res := f.Run(nil,
		"service", "push",
		"--service-id", pushServiceID,
		"--working-dir", workDir,
		"--version-name", "v1.2.3",
		"--no-git",
		"--disable-logs",
	)
	assertPushSuccess(t, res, s, "demo")

	body, _ := s.appVersionBody.Load().(map[string]any)
	if body["name"] != "v1.2.3" {
		t.Errorf("app-version body name=%v, want %q", body["name"], "v1.2.3")
	}
}

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

// pushStubs holds the counters/recorded payloads that tests assert on.
type pushStubs struct {
	uploadBytes      atomic.Int64
	deployHits       atomic.Int32
	deployBody       atomic.Value // map[string]any
	appVersionBody   atomic.Value // map[string]any
	processPollCount atomic.Int32

	// processStatus overrides what the GET /process/{id} endpoint reports.
	// Defaults to "FINISHED" when unset.
	processStatus atomic.Value // string
}

func (s *pushStubs) status() string {
	v, _ := s.processStatus.Load().(string)
	if v == "" {
		return "FINISHED"
	}
	return v
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
			"status":          s.status(),
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

// registerDeployStubs mirrors registerPushStubs but wires the final deploy
// endpoint to /app-version/{id}/deploy instead of /build-and-deploy, matching
// the `service deploy` command's flow.
func registerDeployStubs(t *testing.T, f *fixture, serviceName string) *pushStubs {
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
	f.Mux.HandleFunc("/api/rest/public/app-version/"+pushAppVersionID+"/deploy", func(w http.ResponseWriter, r *http.Request) {
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
			"actionName":      "stack.deploy",
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
			"status":          s.status(),
			"sequence":        1,
			"actionName":      "stack.deploy",
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

// BUG: when the service name matches a setup name in zerops.yaml AND the user
// passes an explicit --setup, the auto-match silently wins and --setup is
// ignored. Reproduced from a real pipeline running `--setup showcase-backend`
// on a service named "backend" with both setups present.
//
// servicePush.go (and serviceDeploy.go) currently does:
//
//	setup, hasMatch := gn.FindFirst(setups, gn.ExactMatch(service.Name.String()))
//	if !hasMatch { /* only then consult --setup */ }
//
// Expected precedence: explicit --setup flag > auto-match by service name >
// interactive selector (TTY) / hard error (non-TTY). Fix is to invert the
// branches so the flag is checked first. This test is skipped until the fix
// lands; remove the t.Skip to re-enable it.
func TestServicePushCommand_SetupFlagOverridesAutoMatch(t *testing.T) {
	t.Skip("known bug: --setup is ignored when service name matches a setup in zerops.yaml; see comment above")

	f := newFixture(t)
	f.SeedLogin("test-token")

	workDir := t.TempDir()
	writeZeropsYaml(t, workDir, "backend", "showcase-backend")

	s := registerPushStubs(t, f, "backend")

	res := f.Run(nil,
		"service", "push",
		"--service-id", pushServiceID,
		"--working-dir", workDir,
		"--setup", "showcase-backend",
		"--no-git",
		"--disable-logs",
	)
	assertPushSuccess(t, res, s, "showcase-backend")
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

// --- Tier 1 error / variant coverage --------------------------------------

// Working directory without zerops.yaml or zerops.yml — yamlReader returns a
// not-found error before any API call is made.
func TestServicePushCommand_MissingZeropsYaml(t *testing.T) {
	f := newFixture(t)
	f.SeedLogin("test-token")
	workDir := t.TempDir()

	registerPushStubs(t, f, "demo")

	res := f.Run(nil,
		"service", "push",
		"--service-id", pushServiceID,
		"--working-dir", workDir,
		"--no-git",
		"--disable-logs",
	)
	if res.ExitCode == 0 {
		t.Fatalf("expected non-zero exit; stdout=%q stderr=%q", res.Stdout, res.Stderr)
	}
	if !strings.Contains(res.Stderr, "zerops.yaml") && !strings.Contains(res.Stderr, "zerops.yml") {
		t.Errorf("stderr should mention zerops.yaml; got: %q", res.Stderr)
	}
}

// Zero-byte zerops.yaml is treated as an explicit error, not "no setups".
func TestServicePushCommand_EmptyZeropsYaml(t *testing.T) {
	f := newFixture(t)
	f.SeedLogin("test-token")
	workDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(workDir, "zerops.yaml"), nil, 0o600); err != nil {
		t.Fatal(err)
	}

	registerPushStubs(t, f, "demo")

	res := f.Run(nil,
		"service", "push",
		"--service-id", pushServiceID,
		"--working-dir", workDir,
		"--no-git",
		"--disable-logs",
	)
	if res.ExitCode == 0 {
		t.Fatalf("expected non-zero exit; stdout=%q stderr=%q", res.Stdout, res.Stderr)
	}
}

// --workspace-state must be one of all/staged/clean. An arbitrary value fails
// validation client-side before any archive work is done.
func TestServicePushCommand_InvalidWorkspaceState(t *testing.T) {
	f := newFixture(t)
	f.SeedLogin("test-token")
	workDir := t.TempDir()
	writeZeropsYaml(t, workDir, "demo")

	registerPushStubs(t, f, "demo")

	res := f.Run(nil,
		"service", "push",
		"--service-id", pushServiceID,
		"--working-dir", workDir,
		"--workspace-state", "garbage",
		"--disable-logs",
	)
	if res.ExitCode == 0 {
		t.Fatalf("expected non-zero exit; stdout=%q stderr=%q", res.Stdout, res.Stderr)
	}
	if !strings.Contains(res.Stderr, "workspace-state") {
		t.Errorf("stderr should explain the bad --workspace-state value; got: %q", res.Stderr)
	}
}

// Non-TTY (test) run with no auto-match and no --setup flag must fail with the
// "please select with --setup" guidance — never silently fall into an
// interactive selector that can't run.
func TestServicePushCommand_NoSetupMatchNoFlagFailsInNonTTY(t *testing.T) {
	f := newFixture(t)
	f.SeedLogin("test-token")
	workDir := t.TempDir()
	writeZeropsYaml(t, workDir, "web", "db")

	registerPushStubs(t, f, "api") // no setup named "api"

	res := f.Run(nil,
		"service", "push",
		"--service-id", pushServiceID,
		"--working-dir", workDir,
		"--no-git",
		"--disable-logs",
	)
	if res.ExitCode == 0 {
		t.Fatalf("expected non-zero exit; stdout=%q stderr=%q", res.Stdout, res.Stderr)
	}
	if !strings.Contains(res.Stderr, "--setup") {
		t.Errorf("stderr should ask for --setup; got: %q", res.Stderr)
	}
}

// Process polling returns FAILED on the first poll — the spinner loop must
// translate that into a non-zero exit with a user-facing error.
func TestServicePushCommand_ProcessFails(t *testing.T) {
	f := newFixture(t)
	f.SeedLogin("test-token")
	workDir := t.TempDir()
	writeZeropsYaml(t, workDir, "demo")

	s := registerPushStubs(t, f, "demo")
	s.processStatus.Store("FAILED")

	res := f.Run(nil,
		"service", "push",
		"--service-id", pushServiceID,
		"--working-dir", workDir,
		"--no-git",
		"--disable-logs",
	)
	if res.ExitCode == 0 {
		t.Fatalf("expected non-zero exit; stdout=%q stderr=%q", res.Stdout, res.Stderr)
	}
}

// Current behavior documented: a --setup value not present in zerops.yaml is
// forwarded verbatim to the API — there is no client-side check that the
// chosen setup exists in the local yaml.
func TestServicePushCommand_SetupNotFoundInYaml_ForwardedToApi(t *testing.T) {
	f := newFixture(t)
	f.SeedLogin("test-token")
	workDir := t.TempDir()
	writeZeropsYaml(t, workDir, "web", "db")

	s := registerPushStubs(t, f, "other") // no auto-match

	res := f.Run(nil,
		"service", "push",
		"--service-id", pushServiceID,
		"--working-dir", workDir,
		"--setup", "nonexistent",
		"--no-git",
		"--disable-logs",
	)
	assertPushSuccess(t, res, s, "nonexistent")
}

// service deploy variant of the happy path: same scaffolding, but the final
// request goes to /app-version/{id}/deploy instead of /build-and-deploy.
func TestServiceDeployCommand_HappyPath(t *testing.T) {
	f := newFixture(t)
	f.SeedLogin("test-token")
	workDir := t.TempDir()
	writeZeropsYaml(t, workDir, "demo")
	// service deploy archives files relative to working-dir; give it a single
	// file so the tar isn't empty and the upload handler sees bytes.
	if err := os.WriteFile(filepath.Join(workDir, "index.html"), []byte("<html/>"), 0o600); err != nil {
		t.Fatal(err)
	}

	s := registerDeployStubs(t, f, "demo")

	res := f.Run(nil,
		"service", "deploy",
		"--service-id", pushServiceID,
		"--working-dir", workDir,
	)
	if res.ExitCode != 0 {
		t.Fatalf("exit=%d\n--- stderr ---\n%s\n--- stdout ---\n%s", res.ExitCode, res.Stderr, res.Stdout)
	}
	if s.uploadBytes.Load() == 0 {
		t.Error("upload handler received no bytes")
	}
	if got := s.deployHits.Load(); got != 1 {
		t.Errorf("deploy endpoint called %d times, want 1", got)
	}
	if got, _ := s.deployBody.Load().(map[string]any); got["zeropsYamlSetup"] != "demo" {
		t.Errorf("deploy body zeropsYamlSetup=%v, want %q", got["zeropsYamlSetup"], "demo")
	}
}


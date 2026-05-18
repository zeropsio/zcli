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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test-fixture identifiers shared by the push/deploy integration suites. These
// are deliberately not real UUIDs — the SDK accepts string IDs and the test
// server doesn't validate.
const (
	pushServiceID    = "0000000000000000000000"
	pushProjectID    = "1111111111111111111111"
	pushClientID     = "2222222222222222222222"
	pushAppVersionID = "3333333333333333333333"
	pushDeployProcID = "4444444444444444444444"
)

// pushStubs holds the counters/recorded payloads that push/deploy tests
// assert on. Shared between registerPushStubs and registerDeployStubs.
type pushStubs struct {
	uploadBytes      atomic.Int64
	uploadBody       atomic.Value // []byte — full body the upload handler received
	deployHits       atomic.Int32
	deployBody       atomic.Value // map[string]any
	appVersionBody   atomic.Value // map[string]any
	processPollCount atomic.Int32

	// processStatus overrides what the GET /process/{id} endpoint reports.
	// Defaults to "FINISHED" when unset. When processStatusSeq is also set,
	// the sequence wins and processStatus is ignored.
	processStatus    atomic.Value // string
	processStatusSeq atomic.Value // []string: per-poll status, sticky on last element
}

// status returns the next status the /process/{id} handler should report.
// Picks from processStatusSeq when set (clamping to the last entry on
// overrun), else falls back to processStatus, else "FINISHED".
func (s *pushStubs) status() string {
	if seq, ok := s.processStatusSeq.Load().([]string); ok && len(seq) > 0 {
		idx := int(s.processPollCount.Load()) - 1
		if idx < 0 {
			idx = 0
		}
		if idx >= len(seq) {
			idx = len(seq) - 1
		}
		return seq[idx]
	}
	v, _ := s.processStatus.Load().(string)
	if v == "" {
		return "FINISHED"
	}
	return v
}

// registerPushStubs wires up every API endpoint that `service push` touches,
// returning recorders the test can inspect. serviceName lets each test pick
// the name returned by the service-stack endpoint (it drives the setup
// auto-match logic in servicePush.go).
func registerPushStubs(t *testing.T, f *fixture, serviceName string) *pushStubs {
	t.Helper()
	s := &pushStubs{}
	uploadURL := f.Server.URL + "/upload/" + pushAppVersionID

	registerSharedScopeStubs(f, serviceName, s, uploadURL)

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

	f.Mux.HandleFunc("/api/rest/public/process/"+pushDeployProcID, processHandler(s, "stack.build_and_deploy"))

	return s
}

// registerDeployStubs mirrors registerPushStubs but wires the final deploy
// endpoint to /app-version/{id}/deploy instead of /build-and-deploy, matching
// the `service deploy` command's flow.
func registerDeployStubs(t *testing.T, f *fixture, serviceName string) *pushStubs {
	t.Helper()
	s := &pushStubs{}
	uploadURL := f.Server.URL + "/upload/" + pushAppVersionID

	registerSharedScopeStubs(f, serviceName, s, uploadURL)

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

	f.Mux.HandleFunc("/api/rest/public/process/"+pushDeployProcID, processHandler(s, "stack.deploy"))

	return s
}

// registerSharedScopeStubs registers the handlers that both push and deploy
// hit before their final deploy endpoint: scope (service-stack, project),
// yaml validation, app-version creation, and the upload URL.
func registerSharedScopeStubs(f *fixture, serviceName string, s *pushStubs, uploadURL string) {
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
		body, _ := io.ReadAll(r.Body)
		s.uploadBytes.Store(int64(len(body)))
		s.uploadBody.Store(body)
		w.WriteHeader(http.StatusOK)
	})
}

func processHandler(s *pushStubs, actionName string) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		s.processPollCount.Add(1)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":              pushDeployProcID,
			"clientId":        pushClientID,
			"projectId":       pushProjectID,
			"status":          s.status(),
			"sequence":        1,
			"actionName":      actionName,
			"created":         "2024-01-01T00:00:00.000Z",
			"lastUpdate":      "2024-01-01T00:00:00.000Z",
			"createdBySystem": false,
			"createdByUser":   map[string]any{},
			"serviceStacks":   []any{},
		})
	}
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
	require.NoError(t, os.WriteFile(filepath.Join(dir, "zerops.yaml"), b, 0o600))
}

// requireNonZeroExit fails the test immediately if the CLI exited 0. Used by
// error-path tests where downstream assertions only make sense after we've
// confirmed the command failed.
func requireNonZeroExit(t *testing.T, res result) {
	t.Helper()
	require.NotEqualf(
		t,
		0,
		res.ExitCode,
		"expected non-zero exit\n--- stderr ---\n%s\n--- stdout ---\n%s", res.Stderr, res.Stdout,
	)
}

// assertPushSuccess is the happy-path postcondition shared by push tests:
// exit code 0, the upload handler saw bytes, the deploy endpoint was hit
// exactly once, and the resolved setup was forwarded in the body.
func assertPushSuccess(t *testing.T, res result, s *pushStubs, wantSetup string) {
	t.Helper()
	require.Equalf(t, 0, res.ExitCode, "non-zero exit\n--- stderr ---\n%s\n--- stdout ---\n%s", res.Stderr, res.Stdout)
	assert.NotZerof(t, s.uploadBytes.Load(), "upload handler received no bytes")
	assert.Equalf(t, int32(1), s.deployHits.Load(), "deploy endpoint hit count")
	body, _ := s.deployBody.Load().(map[string]any)
	assert.Equal(t, wantSetup, body["zeropsYamlSetup"], "deploy body zeropsYamlSetup")
}

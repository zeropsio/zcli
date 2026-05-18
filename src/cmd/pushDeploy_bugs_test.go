//go:build devel

package cmd

// This file collects integration tests that document confirmed bugs in the
// push/deploy commands. Each test reproduces a bug with `t.Skip` so CI stays
// green; remove the t.Skip after the underlying fix lands to lock the
// regression in.

import (
	"encoding/json"
	"io"
	"net/http"
	"sync/atomic"
	"testing"
)

// TestServicePushCommand_SetupFlagOverridesAutoMatch reproduces a confirmed
// bug: when the service name matches a setup name in zerops.yaml AND the user
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
// branches so the flag is checked first.
func TestServicePushCommand_SetupFlagOverridesAutoMatch(t *testing.T) {
	t.Skip("known bug: --setup is ignored when service name matches a setup in zerops.yaml")

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

// TestServicePushCommand_RunningProcessWithNullAppVersion_BUGPROBE reproduces
// a confirmed bug: servicePush.go's log-streaming callback dereferences
// apiProcess.AppVersion.Id (push.go:235) and apiProcess.AppVersion.Build
// (push.go:251) the first time a poll returns status=RUNNING. AppVersion is a
// pointer in output.Process — if the API returns RUNNING before AppVersion is
// populated, the CLI nil-derefs in the spinner goroutine and crashes the
// binary (Go's runtime can't recover panics from goroutines you don't own,
// and ProcessCheckWithSpinner spawns its own goroutine for the poller).
//
// Confirmed by removing t.Skip on this test:
//
//	panic: runtime error: invalid memory address or nil pointer dereference
//	[signal SIGSEGV: segmentation violation]
//	github.com/zeropsio/zcli/src/cmd.servicePushCmd.func1.2 (servicePush.go:235)
//	github.com/zeropsio/zcli/src/uxHelpers.CheckZeropsProcess.func1 (spinner.go:149)
//	created by ProcessCheckWithSpinner (spinner.go:48)
//
// Fix: guard with `if apiProcess.AppVersion == nil { return nil }` before
// push.go:235; same for AppVersion.Build before :251. Once fixed, drop the
// t.Skip and this test locks the fix in.
//
// It builds its own handler set rather than calling registerPushStubs because
// the /process/{id} response must be call-count dependent and http.ServeMux
// doesn't allow re-registering a pattern.
func TestServicePushCommand_RunningProcessWithNullAppVersion_BUGPROBE(t *testing.T) {
	t.Skip("CONFIRMED BUG: push.go:235 nil-deref crashes the binary; remove t.Skip after fix")

	f := newFixture(t)
	f.SeedLogin("test-token")
	workDir := t.TempDir()
	writeZeropsYaml(t, workDir, "demo")

	uploadURL := f.Server.URL + "/upload/" + pushAppVersionID

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
			"id": pushProjectID, "clientId": pushClientID, "name": "demo-project",
			"mode": "LIGHT", "status": "ACTIVE",
			"created": "2024-01-01T00:00:00.000Z", "lastUpdate": "2024-01-01T00:00:00.000Z",
			"tagList": []string{},
		},
		"serviceStackTypeVersionId": "nodejs@20",
		"created":                   "2024-01-01T00:00:00.000Z",
		"lastUpdate":                "2024-01-01T00:00:00.000Z",
		"mode":                      "NON_HA",
	})
	f.HandleJSON("/api/rest/public/project/"+pushProjectID, 200, map[string]any{
		"id": pushProjectID, "clientId": pushClientID, "name": "demo-project",
		"mode": "LIGHT", "status": "ACTIVE",
		"created": "2024-01-01T00:00:00.000Z", "lastUpdate": "2024-01-01T00:00:00.000Z",
		"tagList": []string{},
	})
	f.HandleJSON("/api/rest/public/service-stack/zerops-yaml-validation", 200, map[string]any{})
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
	f.Mux.HandleFunc("/upload/"+pushAppVersionID, func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.Copy(io.Discard, r.Body)
		w.WriteHeader(http.StatusOK)
	})
	f.HandleJSON("/api/rest/public/app-version/"+pushAppVersionID+"/build-and-deploy", 200, map[string]any{
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

	var pollCount atomic.Int32
	f.Mux.HandleFunc("/api/rest/public/process/"+pushDeployProcID, func(w http.ResponseWriter, r *http.Request) {
		n := pollCount.Add(1)
		w.Header().Set("Content-Type", "application/json")
		base := map[string]any{
			"id":              pushDeployProcID,
			"clientId":        pushClientID,
			"projectId":       pushProjectID,
			"sequence":        1,
			"actionName":      "stack.build_and_deploy",
			"created":         "2024-01-01T00:00:00.000Z",
			"lastUpdate":      "2024-01-01T00:00:00.000Z",
			"createdBySystem": false,
			"createdByUser":   map[string]any{},
			"serviceStacks":   []any{},
			"appVersion":      nil, // <-- the trigger
		}
		if n == 1 {
			base["status"] = "RUNNING"
		} else {
			base["status"] = "FINISHED"
		}
		_ = json.NewEncoder(w).Encode(base)
	})

	// Run WITHOUT --disable-logs so the log-streaming callback fires and the
	// nil-deref on apiProcess.AppVersion is reached.
	res := f.Run(nil,
		"service", "push",
		"--service-id", pushServiceID,
		"--working-dir", workDir,
		"--no-git",
	)
	t.Logf("exit=%d stderr=%q", res.ExitCode, res.Stderr)
}

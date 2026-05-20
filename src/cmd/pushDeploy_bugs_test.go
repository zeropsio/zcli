package cmd

// This file collects integration regression tests that pin previously
// confirmed bugs in the push/deploy commands. Each test reproduces the
// originally failing scenario and now must pass — they are the live guard
// against regressions of the corresponding fixes.

import (
	"encoding/json"
	"io"
	"net/http"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestServicePushCommand_SetupFlagOverridesAutoMatch is a regression lock-in:
// when the service name matches a setup name in zerops.yaml AND the user
// passes an explicit --setup, the flag must win. Pins the fix in
// servicePush.go / serviceDeploy.go that puts the flag check before the
// auto-match by service name. Reproduced from a real pipeline running
// `--setup showcase-backend` on a service named "backend" with both setups
// present.
func TestServicePushCommand_SetupFlagOverridesAutoMatch(t *testing.T) {
	f := newFixture(t)
	f.SeedLogin("test-token")

	workDir := t.TempDir()
	writeZeropsYaml(t, workDir, "backend", "showcase-backend")

	s := registerPushStubs(t, f, "backend")

	res := f.Run(
		"service", "push",
		"--service-id", pushServiceID,
		"--working-dir", workDir,
		"--setup", "showcase-backend",
		"--no-git",
		"--disable-logs",
	)
	assertPushSuccess(t, res, s, "showcase-backend")
}

// TestServicePushCommand_RunningProcessWithNullAppVersionNoCrash is a
// regression lock-in: servicePush.go's log-streaming callback used to
// dereference apiProcess.AppVersion.Id and apiProcess.AppVersion.Build on the
// first poll that reported status=RUNNING. AppVersion is a pointer in
// output.Process — when the API returned RUNNING before AppVersion was
// populated, the CLI nil-deref'd in the spinner goroutine and crashed the
// binary (Go's runtime can't recover panics from goroutines you don't own).
//
// Original stack trace before the fix:
//
//	panic: runtime error: invalid memory address or nil pointer dereference
//	[signal SIGSEGV: segmentation violation]
//	github.com/zeropsio/zcli/src/cmd.servicePushCmd.func1.2 (servicePush.go:235)
//	github.com/zeropsio/zcli/src/uxHelpers.CheckZeropsProcess.func1 (spinner.go:149)
//	created by ProcessCheckWithSpinner (spinner.go:48)
//
// The fix nil-guards apiProcess.AppVersion (and separately .Build) before
// the deref sites; this test drives a RUNNING + appVersion=null poll
// followed by a FINISHED poll, and the push must complete cleanly.
//
// Builds its own handler set rather than calling registerPushStubs because
// the /process/{id} response must be call-count dependent and http.ServeMux
// doesn't allow re-registering a pattern.
func TestServicePushCommand_RunningProcessWithNullAppVersionNoCrash(t *testing.T) {
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
	// nil-deref path is exercised. Before the fix this crashed the test
	// binary with SIGSEGV; after the fix the callback should bail early on
	// the null AppVersion and the second poll's FINISHED status completes
	// the push.
	res := f.Run(
		"service", "push",
		"--service-id", pushServiceID,
		"--working-dir", workDir,
		"--no-git",
	)
	require.Equalf(
		t,
		0,
		res.ExitCode,
		"push should complete cleanly even with appVersion=null on first poll\nstderr=%q", res.Stderr,
	)
}

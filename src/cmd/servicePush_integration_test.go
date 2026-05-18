//go:build devel

package cmd

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Happy paths ----------------------------------------------------------

// TestServicePushCommand_SetupAutoMatchesServiceName checks that when the
// service name matches a setup in zerops.yaml, auto-match picks it without
// requiring a --setup flag.
func TestServicePushCommand_SetupAutoMatchesServiceName(t *testing.T) {
	f := newFixture(t)
	f.SeedLogin("test-token")

	workDir := t.TempDir()
	writeZeropsYaml(t, workDir, "demo")

	s := registerPushStubs(t, f, "demo")

	res := f.Run(
		nil,
		"service", "push",
		"--service-id", pushServiceID,
		"--working-dir", workDir,
		"--no-git",
		"--disable-logs",
	)
	assertPushSuccess(t, res, s, "demo")
}

// TestServicePushCommand_SetupSelectedByFlag verifies that when the service
// name doesn't match any setup, the user can pick one explicitly via --setup.
func TestServicePushCommand_SetupSelectedByFlag(t *testing.T) {
	f := newFixture(t)
	f.SeedLogin("test-token")

	workDir := t.TempDir()
	writeZeropsYaml(t, workDir, "demo", "api-prod", "api-stage")

	s := registerPushStubs(t, f, "api") // service is "api", no exact match

	res := f.Run(
		nil,
		"service", "push",
		"--service-id", pushServiceID,
		"--working-dir", workDir,
		"--setup", "api-prod",
		"--no-git",
		"--disable-logs",
	)
	assertPushSuccess(t, res, s, "api-prod")
}

// TestServicePushCommand_VersionNameForwarded checks that the --version-name
// flag is forwarded to the POST /app-version request body.
func TestServicePushCommand_VersionNameForwarded(t *testing.T) {
	f := newFixture(t)
	f.SeedLogin("test-token")

	workDir := t.TempDir()
	writeZeropsYaml(t, workDir, "demo")

	s := registerPushStubs(t, f, "demo")

	res := f.Run(
		nil,
		"service", "push",
		"--service-id", pushServiceID,
		"--working-dir", workDir,
		"--version-name", "v1.2.3",
		"--no-git",
		"--disable-logs",
	)
	assertPushSuccess(t, res, s, "demo")

	body, _ := s.appVersionBody.Load().(map[string]any)
	assert.Equal(t, "v1.2.3", body["name"], "app-version body should carry --version-name")
}

// --- Error paths / variant coverage ---------------------------------------

// TestServicePushCommand_MissingZeropsYaml verifies that a working directory
// without zerops.yaml or zerops.yml fails with a yamlReader not-found error
// before any API call is made.
func TestServicePushCommand_MissingZeropsYaml(t *testing.T) {
	f := newFixture(t)
	f.SeedLogin("test-token")
	workDir := t.TempDir()

	registerPushStubs(t, f, "demo")

	res := f.Run(
		nil,
		"service", "push",
		"--service-id", pushServiceID,
		"--working-dir", workDir,
		"--no-git",
		"--disable-logs",
	)
	requireNonZeroExit(t, res)
	assert.Truef(
		t,
		strings.Contains(res.Stderr, "zerops.yaml") || strings.Contains(res.Stderr, "zerops.yml"),
		"stderr should mention zerops.yaml/.yml; got: %q", res.Stderr,
	)
}

// TestServicePushCommand_EmptyZeropsYaml checks that a zero-byte zerops.yaml
// is treated as an explicit error rather than as "no setups defined".
func TestServicePushCommand_EmptyZeropsYaml(t *testing.T) {
	f := newFixture(t)
	f.SeedLogin("test-token")
	workDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(workDir, "zerops.yaml"), nil, 0o600))

	registerPushStubs(t, f, "demo")

	res := f.Run(
		nil,
		"service", "push",
		"--service-id", pushServiceID,
		"--working-dir", workDir,
		"--no-git",
		"--disable-logs",
	)
	requireNonZeroExit(t, res)
}

// TestServicePushCommand_InvalidWorkspaceState verifies that an arbitrary
// value for --workspace-state (which must be all/staged/clean) fails client-
// side validation before any archive work is done.
func TestServicePushCommand_InvalidWorkspaceState(t *testing.T) {
	f := newFixture(t)
	f.SeedLogin("test-token")
	workDir := t.TempDir()
	writeZeropsYaml(t, workDir, "demo")

	registerPushStubs(t, f, "demo")

	res := f.Run(
		nil,
		"service", "push",
		"--service-id", pushServiceID,
		"--working-dir", workDir,
		"--workspace-state", "garbage",
		"--disable-logs",
	)
	requireNonZeroExit(t, res)
	assert.Contains(t, res.Stderr, "workspace-state", "stderr should explain the bad --workspace-state value")
}

// TestServicePushCommand_NoSetupMatchNoFlagFailsInNonTTY checks that a non-
// TTY run with no auto-match and no --setup flag fails with a "please select
// with --setup" message rather than silently falling into an interactive
// selector that can't run.
func TestServicePushCommand_NoSetupMatchNoFlagFailsInNonTTY(t *testing.T) {
	f := newFixture(t)
	f.SeedLogin("test-token")
	workDir := t.TempDir()
	writeZeropsYaml(t, workDir, "web", "db")

	registerPushStubs(t, f, "api") // no setup named "api"

	res := f.Run(
		nil,
		"service", "push",
		"--service-id", pushServiceID,
		"--working-dir", workDir,
		"--no-git",
		"--disable-logs",
	)
	requireNonZeroExit(t, res)
	assert.Contains(t, res.Stderr, "--setup", "stderr should ask for --setup")
}

// TestServicePushCommand_ProcessFails verifies that when process polling
// returns FAILED on the first poll, the spinner loop translates that into a
// non-zero exit with a user-facing error.
func TestServicePushCommand_ProcessFails(t *testing.T) {
	f := newFixture(t)
	f.SeedLogin("test-token")
	workDir := t.TempDir()
	writeZeropsYaml(t, workDir, "demo")

	s := registerPushStubs(t, f, "demo")
	s.processStatus.Store("FAILED")

	res := f.Run(
		nil,
		"service", "push",
		"--service-id", pushServiceID,
		"--working-dir", workDir,
		"--no-git",
		"--disable-logs",
	)
	requireNonZeroExit(t, res)
}

// TestServicePushCommand_SetupNotFoundInYaml_ForwardedToApi documents current
// behavior: a --setup value not present in zerops.yaml is forwarded verbatim
// to the API, with no client-side check that the chosen setup exists in the
// local yaml.
func TestServicePushCommand_SetupNotFoundInYaml_ForwardedToApi(t *testing.T) {
	f := newFixture(t)
	f.SeedLogin("test-token")
	workDir := t.TempDir()
	writeZeropsYaml(t, workDir, "web", "db")

	s := registerPushStubs(t, f, "other") // no auto-match

	res := f.Run(
		nil,
		"service", "push",
		"--service-id", pushServiceID,
		"--working-dir", workDir,
		"--setup", "nonexistent",
		"--no-git",
		"--disable-logs",
	)
	assertPushSuccess(t, res, s, "nonexistent")
}

// --- Tier 2: polling, scope, archive-file-path ----------------------------

// TestServicePushCommand_PendingThenRunningThenFinished checks that process
// polling keeps looping through PENDING, RUNNING, and FINISHED across three
// consecutive polls until a terminal state is reached.
func TestServicePushCommand_PendingThenRunningThenFinished(t *testing.T) {
	f := newFixture(t)
	f.SeedLogin("test-token")
	workDir := t.TempDir()
	writeZeropsYaml(t, workDir, "demo")

	s := registerPushStubs(t, f, "demo")
	s.processStatusSeq.Store([]string{"PENDING", "RUNNING", "FINISHED"})

	res := f.Run(
		nil,
		"service", "push",
		"--service-id", pushServiceID,
		"--working-dir", workDir,
		"--no-git",
		"--disable-logs",
	)
	assertPushSuccess(t, res, s, "demo")
	assert.GreaterOrEqual(t, s.processPollCount.Load(), int32(3), "should poll at least 3 times for PENDING→RUNNING→FINISHED")
}

// TestServicePushCommand_InvalidServiceIdErrors checks that a 404 on the
// service-stack endpoint surfaces as a non-zero exit with a user-facing error
// rather than a panic or generic dump.
func TestServicePushCommand_InvalidServiceIdErrors(t *testing.T) {
	f := newFixture(t)
	f.SeedLogin("test-token")
	workDir := t.TempDir()
	writeZeropsYaml(t, workDir, "demo")

	// Only register the one handler this test needs — registerPushStubs
	// would 200 the service-stack lookup and we want it to fail.
	f.Mux.HandleFunc("/api/rest/public/service-stack/"+pushServiceID, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]any{
				"code":     "serviceStackNotFound",
				"message":  "Service stack not found",
				"category": "invalidUserInput",
			},
		})
	})

	res := f.Run(
		nil,
		"service", "push",
		"--service-id", pushServiceID,
		"--working-dir", workDir,
		"--no-git",
		"--disable-logs",
	)
	requireNonZeroExit(t, res)
	assert.Contains(t, strings.ToLower(res.Stderr), "service", "stderr should mention the service error")
}

// TestServicePushCommand_ArchiveFilePathTeesToFile checks that
// --archive-file-path writes a copy of the uploaded package to disk via a
// tee'd reader, producing a non-empty file after the push.
func TestServicePushCommand_ArchiveFilePathTeesToFile(t *testing.T) {
	f := newFixture(t)
	f.SeedLogin("test-token")
	workDir := t.TempDir()
	writeZeropsYaml(t, workDir, "demo")
	// Give the archive at least one file to include, so the resulting tar is
	// not just a gzip header.
	require.NoError(t, os.WriteFile(filepath.Join(workDir, "main.go"), []byte("package main\n"), 0o600))

	archivePath := "out.tar.gz" // resolved relative to --working-dir by openPackageFile
	s := registerPushStubs(t, f, "demo")

	res := f.Run(
		nil,
		"service", "push",
		"--service-id", pushServiceID,
		"--working-dir", workDir,
		"--archive-file-path", archivePath,
		"--no-git",
		"--disable-logs",
	)
	assertPushSuccess(t, res, s, "demo")

	info, err := os.Stat(filepath.Join(workDir, archivePath))
	require.NoError(t, err, "archive file should exist")
	assert.NotZero(t, info.Size(), "archive file should not be empty")
}

// TestServicePushCommand_ProjectFlagAndServiceByName exercises the by-name
// lookup path: --project-id resolves the project via flag, then scopeService
// falls through to GetServiceByIdOrName, which first tries the positional arg
// as a UUID (gets ServiceStackNotFound) and then looks up by name.
func TestServicePushCommand_ProjectFlagAndServiceByName(t *testing.T) {
	f := newFixture(t)
	f.SeedLogin("test-token")
	workDir := t.TempDir()
	writeZeropsYaml(t, workDir, "demo")

	// Standard push stubs handle the rest of the flow; we just need to wire
	// the two extra service-lookup endpoints. The default service-stack/{id}
	// handler from registerPushStubs is on a different path (pushServiceID),
	// so it does not clash with /service-stack/demo.
	s := registerPushStubs(t, f, "demo")

	f.Mux.HandleFunc("/api/rest/public/service-stack/demo", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]any{
				"code":    "serviceStackNotFound",
				"message": "Service stack not found",
			},
		})
	})

	// The by-name lookup returns the same shape as the by-id endpoint.
	f.HandleJSON("/api/rest/public/service-stack-by-name/"+pushProjectID+"/demo", 200, map[string]any{
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

	res := f.Run(
		nil,
		"service", "push",
		"demo", // positional service-id-or-name
		"--project-id", pushProjectID,
		"--working-dir", workDir,
		"--no-git",
		"--disable-logs",
	)
	assertPushSuccess(t, res, s, "demo")
}

// TestServicePushCommand_ScopeFromSavedProjectId checks that a saved project
// scope (set previously by `zcli scope project ...`) lets push run without
// --project-id, resolving the service by name within the scoped project.
func TestServicePushCommand_ScopeFromSavedProjectId(t *testing.T) {
	f := newFixture(t)
	f.SeedScopedLogin("test-token", pushProjectID)
	workDir := t.TempDir()
	writeZeropsYaml(t, workDir, "demo")

	s := registerPushStubs(t, f, "demo")

	f.Mux.HandleFunc("/api/rest/public/service-stack/demo", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]any{"code": "serviceStackNotFound", "message": "Service stack not found"},
		})
	})
	f.HandleJSON("/api/rest/public/service-stack-by-name/"+pushProjectID+"/demo", 200, map[string]any{
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

	res := f.Run(
		nil,
		"service", "push",
		"demo", // positional service name
		"--working-dir", workDir,
		"--no-git",
		"--disable-logs",
	)
	assertPushSuccess(t, res, s, "demo")
}

// TestServicePushCommand_ArchiveFilePathAlreadyExistsErrors verifies that
// openPackageFile refuses to overwrite an existing --archive-file-path, so
// the push fails before any upload happens.
func TestServicePushCommand_ArchiveFilePathAlreadyExistsErrors(t *testing.T) {
	f := newFixture(t)
	f.SeedLogin("test-token")
	workDir := t.TempDir()
	writeZeropsYaml(t, workDir, "demo")

	archivePath := "out.tar.gz"
	require.NoError(t, os.WriteFile(filepath.Join(workDir, archivePath), []byte("preexisting"), 0o600))

	s := registerPushStubs(t, f, "demo")

	res := f.Run(
		nil,
		"service", "push",
		"--service-id", pushServiceID,
		"--working-dir", workDir,
		"--archive-file-path", archivePath,
		"--no-git",
		"--disable-logs",
	)
	requireNonZeroExit(t, res)
	// The existing file must be untouched.
	if info, err := os.Stat(filepath.Join(workDir, archivePath)); err == nil {
		assert.Equal(t, int64(len("preexisting")), info.Size(), "archive file should not be overwritten")
	}
	// And no upload should have hit the server.
	assert.Zero(t, s.uploadBytes.Load(), "upload should not happen when pre-flight check fails")
}

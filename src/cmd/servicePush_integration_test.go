//go:build devel

package cmd

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// --- Happy paths ----------------------------------------------------------

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

// --- Error paths / variant coverage ---------------------------------------

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

// --- Tier 2: polling, scope, archive-file-path ----------------------------

// Process polling visits PENDING, then RUNNING, then FINISHED across three
// consecutive polls. The CLI must keep looping until a terminal state is
// reached.
func TestServicePushCommand_PendingThenRunningThenFinished(t *testing.T) {
	f := newFixture(t)
	f.SeedLogin("test-token")
	workDir := t.TempDir()
	writeZeropsYaml(t, workDir, "demo")

	s := registerPushStubs(t, f, "demo")
	s.processStatusSeq.Store([]string{"PENDING", "RUNNING", "FINISHED"})

	res := f.Run(nil,
		"service", "push",
		"--service-id", pushServiceID,
		"--working-dir", workDir,
		"--no-git",
		"--disable-logs",
	)
	assertPushSuccess(t, res, s, "demo")
	if got := s.processPollCount.Load(); got < 3 {
		t.Errorf("expected at least 3 process polls, got %d", got)
	}
}

// A 404 on the service-stack endpoint surfaces as a non-zero exit with a
// user-facing error rather than a panic or generic dump.
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
	if !strings.Contains(strings.ToLower(res.Stderr), "service") {
		t.Errorf("stderr should mention the service error; got: %q", res.Stderr)
	}
}

// --archive-file-path writes a copy of the uploaded package to disk via a
// tee'd reader. The file should exist after the push and have non-zero size.
func TestServicePushCommand_ArchiveFilePathTeesToFile(t *testing.T) {
	f := newFixture(t)
	f.SeedLogin("test-token")
	workDir := t.TempDir()
	writeZeropsYaml(t, workDir, "demo")
	// Give the archive at least one file to include, so the resulting tar is
	// not just a gzip header.
	if err := os.WriteFile(filepath.Join(workDir, "main.go"), []byte("package main\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	archivePath := "out.tar.gz" // resolved relative to --working-dir by openPackageFile
	s := registerPushStubs(t, f, "demo")

	res := f.Run(nil,
		"service", "push",
		"--service-id", pushServiceID,
		"--working-dir", workDir,
		"--archive-file-path", archivePath,
		"--no-git",
		"--disable-logs",
	)
	assertPushSuccess(t, res, s, "demo")

	info, err := os.Stat(filepath.Join(workDir, archivePath))
	if err != nil {
		t.Fatalf("expected archive file at %s: %v", archivePath, err)
	}
	if info.Size() == 0 {
		t.Errorf("archive file is empty")
	}
}

// openPackageFile refuses to overwrite an existing --archive-file-path. The
// push should fail before any upload happens.
func TestServicePushCommand_ArchiveFilePathAlreadyExistsErrors(t *testing.T) {
	f := newFixture(t)
	f.SeedLogin("test-token")
	workDir := t.TempDir()
	writeZeropsYaml(t, workDir, "demo")

	archivePath := "out.tar.gz"
	if err := os.WriteFile(filepath.Join(workDir, archivePath), []byte("preexisting"), 0o600); err != nil {
		t.Fatal(err)
	}

	s := registerPushStubs(t, f, "demo")

	res := f.Run(nil,
		"service", "push",
		"--service-id", pushServiceID,
		"--working-dir", workDir,
		"--archive-file-path", archivePath,
		"--no-git",
		"--disable-logs",
	)
	if res.ExitCode == 0 {
		t.Fatalf("expected non-zero exit; stdout=%q stderr=%q", res.Stdout, res.Stderr)
	}
	// The existing file must be untouched (size still 11 bytes from preexisting).
	if info, err := os.Stat(filepath.Join(workDir, archivePath)); err == nil && info.Size() != int64(len("preexisting")) {
		t.Errorf("archive file was overwritten despite the error: size=%d", info.Size())
	}
	// And no upload should have hit the server.
	if got := s.uploadBytes.Load(); got != 0 {
		t.Errorf("upload happened despite pre-flight error: %d bytes", got)
	}
}

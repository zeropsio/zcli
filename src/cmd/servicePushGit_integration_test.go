//go:build devel

package cmd

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- git fixture helpers --------------------------------------------------

// runGit invokes git in dir with system/global config disabled. Tests must
// not pick up the developer's ~/.gitconfig.
func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(),
		"GIT_TERMINAL_PROMPT=0",
		"GIT_CONFIG_GLOBAL=/dev/null",
		"GIT_CONFIG_SYSTEM=/dev/null",
	)
	out, err := cmd.CombinedOutput()
	require.NoErrorf(t, err, "git %v failed: %s", args, out)
}

// gitInit creates an empty repo with deterministic identity/config.
func gitInit(t *testing.T, dir string) {
	t.Helper()
	runGit(t, dir, "init", "-q", "-b", "main")
	runGit(t, dir, "config", "user.email", "test@example.com")
	runGit(t, dir, "config", "user.name", "Test User")
	runGit(t, dir, "config", "commit.gpgsign", "false")
}

// gitAddCommit stages everything in dir and records a commit.
func gitAddCommit(t *testing.T, dir, msg string) {
	t.Helper()
	runGit(t, dir, "add", "-A")
	runGit(t, dir, "commit", "-q", "-m", msg)
}

// archiveEntries unpacks a (gzipped) tar produced by the push pipeline into a
// map of archive-path → file content. Returns nil for non-regular entries.
// Auto-detects gzip vs raw tar.
func archiveEntries(t *testing.T, raw []byte) map[string]string {
	t.Helper()
	var r io.Reader = bytes.NewReader(raw)
	if len(raw) >= 2 && raw[0] == 0x1f && raw[1] == 0x8b {
		gr, err := gzip.NewReader(bytes.NewReader(raw))
		require.NoError(t, err)
		defer gr.Close()
		r = gr
	}
	tr := tar.NewReader(r)
	out := map[string]string{}
	for {
		h, err := tr.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		require.NoError(t, err)
		if h.Typeflag == tar.TypeReg {
			data, err := io.ReadAll(tr)
			require.NoError(t, err)
			out[h.Name] = string(data)
		} else {
			out[h.Name] = "" // directories, etc.
		}
	}
	return out
}

func archiveContains(entries map[string]string, suffix string) bool {
	for name := range entries {
		if strings.HasSuffix(name, suffix) {
			return true
		}
	}
	return false
}

// --- error paths ----------------------------------------------------------

// TestServicePushCommand_GitNotInitializedErrors checks that a working dir
// without a .git directory (and without --no-git) is refused by the archive
// client with an explanatory error about git init.
func TestServicePushCommand_GitNotInitializedErrors(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not on PATH")
	}
	f := newFixture(t)
	f.SeedLogin("test-token")
	workDir := t.TempDir()
	writeZeropsYaml(t, workDir, "demo")

	registerPushStubs(t, f, "demo")

	res := f.Run(nil,
		"service", "push",
		"--service-id", pushServiceID,
		"--working-dir", workDir,
		"--disable-logs",
	)
	requireNonZeroExit(t, res)
	assert.Contains(t, res.Stderr, "git init", "stderr should mention git init")
}

// TestServicePushCommand_GitZeroCommitsErrors checks that a working dir with
// a .git directory but zero commits is refused by the archive client with an
// explanatory "at least one commit" error.
func TestServicePushCommand_GitZeroCommitsErrors(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not on PATH")
	}
	f := newFixture(t)
	f.SeedLogin("test-token")
	workDir := t.TempDir()
	writeZeropsYaml(t, workDir, "demo")
	gitInit(t, workDir)

	registerPushStubs(t, f, "demo")

	res := f.Run(nil,
		"service", "push",
		"--service-id", pushServiceID,
		"--working-dir", workDir,
		"--disable-logs",
	)
	requireNonZeroExit(t, res)
	assert.Contains(t, res.Stderr, "commit", "stderr should mention needing a commit")
}

// --- happy paths ----------------------------------------------------------

// TestServicePushCommand_GitArchive_CommittedFilesUploaded checks that with
// the default workspace-state (=all) over a clean repo, committed files end
// up in the uploaded archive.
func TestServicePushCommand_GitArchive_CommittedFilesUploaded(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not on PATH")
	}
	f := newFixture(t)
	f.SeedLogin("test-token")
	workDir := t.TempDir()
	writeZeropsYaml(t, workDir, "demo")
	require.NoError(t, os.WriteFile(filepath.Join(workDir, "main.go"), []byte("package main\n"), 0o600))
	gitInit(t, workDir)
	gitAddCommit(t, workDir, "initial")

	s := registerPushStubs(t, f, "demo")

	res := f.Run(nil,
		"service", "push",
		"--service-id", pushServiceID,
		"--working-dir", workDir,
		"--disable-logs",
	)
	assertPushSuccess(t, res, s, "demo")

	body, _ := s.uploadBody.Load().([]byte)
	require.NotEmpty(t, body, "upload body captured")
	entries := archiveEntries(t, body)
	assert.True(t, archiveContains(entries, "main.go"), "archive should contain main.go; got: %v", keys(entries))
	assert.True(t, archiveContains(entries, "zerops.yaml"), "archive should contain zerops.yaml; got: %v", keys(entries))
}

// TestServicePushCommand_GitArchive_WorkspaceCleanIgnoresUncommitted checks
// that --workspace-state=clean excludes uncommitted files even when they
// exist in the working tree.
func TestServicePushCommand_GitArchive_WorkspaceCleanIgnoresUncommitted(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not on PATH")
	}
	f := newFixture(t)
	f.SeedLogin("test-token")
	workDir := t.TempDir()
	writeZeropsYaml(t, workDir, "demo")
	require.NoError(t, os.WriteFile(filepath.Join(workDir, "main.go"), []byte("package main\n"), 0o600))
	gitInit(t, workDir)
	gitAddCommit(t, workDir, "initial")
	// Add a file after the commit — it must NOT appear in the archive.
	require.NoError(t, os.WriteFile(filepath.Join(workDir, "uncommitted.txt"), []byte("local-only"), 0o600))

	s := registerPushStubs(t, f, "demo")

	res := f.Run(nil,
		"service", "push",
		"--service-id", pushServiceID,
		"--working-dir", workDir,
		"--workspace-state", "clean",
		"--disable-logs",
	)
	assertPushSuccess(t, res, s, "demo")

	body, _ := s.uploadBody.Load().([]byte)
	require.NotEmpty(t, body)
	entries := archiveEntries(t, body)
	assert.True(t, archiveContains(entries, "main.go"), "committed main.go should be present")
	assert.False(t, archiveContains(entries, "uncommitted.txt"), "uncommitted file should not be in clean archive; got: %v", keys(entries))
}

// TestServicePushCommand_GitArchive_DeployGitFolderIncludesGitDir checks that
// --deploy-git-folder tars the .git directory into the archive alongside the
// working-tree files.
func TestServicePushCommand_GitArchive_DeployGitFolderIncludesGitDir(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not on PATH")
	}
	f := newFixture(t)
	f.SeedLogin("test-token")
	workDir := t.TempDir()
	writeZeropsYaml(t, workDir, "demo")
	require.NoError(t, os.WriteFile(filepath.Join(workDir, "main.go"), []byte("package main\n"), 0o600))
	gitInit(t, workDir)
	gitAddCommit(t, workDir, "initial")

	s := registerPushStubs(t, f, "demo")

	res := f.Run(nil,
		"service", "push",
		"--service-id", pushServiceID,
		"--working-dir", workDir,
		"--deploy-git-folder",
		"--disable-logs",
	)
	assertPushSuccess(t, res, s, "demo")

	body, _ := s.uploadBody.Load().([]byte)
	require.NotEmpty(t, body)
	entries := archiveEntries(t, body)
	// The .git/HEAD file is a stable marker that the .git directory was
	// archived in.
	hasGitDir := false
	for name := range entries {
		if strings.Contains(name, ".git/") || strings.HasSuffix(name, ".git/HEAD") {
			hasGitDir = true
			break
		}
	}
	assert.Truef(t, hasGitDir, ".git/ should be present in archive; entries: %v", keys(entries))
	assert.True(t, archiveContains(entries, "main.go"), "main.go should still be present")
}

// TestServicePushCommand_GitArchive_WorkspaceStagedKeepsStagedDropsUnstaged
// checks that --workspace-state=staged includes files that are staged (added
// to the index) but not yet committed, while excluding unstaged working-tree
// changes.
func TestServicePushCommand_GitArchive_WorkspaceStagedKeepsStagedDropsUnstaged(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not on PATH")
	}
	f := newFixture(t)
	f.SeedLogin("test-token")
	workDir := t.TempDir()
	writeZeropsYaml(t, workDir, "demo")
	require.NoError(t, os.WriteFile(filepath.Join(workDir, "main.go"), []byte("package main\n"), 0o600))
	gitInit(t, workDir)
	gitAddCommit(t, workDir, "initial")

	// Stage one new file (in the index but not committed) and leave another
	// unstaged in the working tree.
	require.NoError(t, os.WriteFile(filepath.Join(workDir, "staged.txt"), []byte("staged"), 0o600))
	runGit(t, workDir, "add", "staged.txt")
	require.NoError(t, os.WriteFile(filepath.Join(workDir, "unstaged.txt"), []byte("unstaged"), 0o600))

	s := registerPushStubs(t, f, "demo")

	res := f.Run(nil,
		"service", "push",
		"--service-id", pushServiceID,
		"--working-dir", workDir,
		"--workspace-state", "staged",
		"--disable-logs",
	)
	assertPushSuccess(t, res, s, "demo")

	body, _ := s.uploadBody.Load().([]byte)
	require.NotEmpty(t, body)
	entries := archiveEntries(t, body)
	assert.True(t, archiveContains(entries, "main.go"), "committed main.go should be present")
	assert.True(t, archiveContains(entries, "staged.txt"), "staged.txt should be in staged archive; got: %v", keys(entries))
	assert.False(t, archiveContains(entries, "unstaged.txt"), "unstaged.txt must not be in staged archive; got: %v", keys(entries))
}

// TestServicePushCommand_GitArchive_WorkspaceAllIncludesUncommitted checks
// that --workspace-state=all (the default) captures everything: committed,
// staged, and unstaged working-tree changes.
func TestServicePushCommand_GitArchive_WorkspaceAllIncludesUncommitted(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not on PATH")
	}
	f := newFixture(t)
	f.SeedLogin("test-token")
	workDir := t.TempDir()
	writeZeropsYaml(t, workDir, "demo")
	require.NoError(t, os.WriteFile(filepath.Join(workDir, "main.go"), []byte("package main\n"), 0o600))
	gitInit(t, workDir)
	gitAddCommit(t, workDir, "initial")

	require.NoError(t, os.WriteFile(filepath.Join(workDir, "staged.txt"), []byte("staged"), 0o600))
	runGit(t, workDir, "add", "staged.txt")
	require.NoError(t, os.WriteFile(filepath.Join(workDir, "unstaged.txt"), []byte("unstaged"), 0o600))

	s := registerPushStubs(t, f, "demo")

	// Omitting --workspace-state defaults to "all".
	res := f.Run(nil,
		"service", "push",
		"--service-id", pushServiceID,
		"--working-dir", workDir,
		"--disable-logs",
	)
	assertPushSuccess(t, res, s, "demo")

	body, _ := s.uploadBody.Load().([]byte)
	require.NotEmpty(t, body)
	entries := archiveEntries(t, body)
	assert.True(t, archiveContains(entries, "main.go"), "committed file present")
	assert.True(t, archiveContains(entries, "staged.txt"), "staged file present")
	assert.Truef(t, archiveContains(entries, "unstaged.txt"), "unstaged file should also be in workspace=all archive; got: %v", keys(entries))
}

// keys returns the map keys in undefined order; helper for error messages.
func keys(m map[string]string) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}

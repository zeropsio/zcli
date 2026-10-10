package archiveClient

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zeropsio/zcli/src/logger"
)

func gitArchiverTestRepo(t *testing.T) string {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not on PATH")
	}
	dir := t.TempDir()
	run := func(args ...string) {
		cmd := exec.CommandContext(t.Context(), "git", args...)
		cmd.Dir = dir
		cmd.Env = append(os.Environ(), "GIT_CONFIG_GLOBAL=/dev/null", "GIT_CONFIG_SYSTEM=/dev/null")
		out, err := cmd.CombinedOutput()
		require.NoErrorf(t, err, "git %v: %s", args, out)
	}
	require.NoError(t, os.WriteFile(filepath.Join(dir, "a.txt"), []byte("a"), 0o600))
	run("init", "-q", "-b", "main")
	run("config", "user.email", "t@example.com")
	run("config", "user.name", "t")
	run("config", "commit.gpgsign", "false")
	run("add", "-A")
	run("commit", "-q", "-m", "init")
	require.NoError(t, os.WriteFile(filepath.Join(dir, "b.txt"), []byte("b"), 0o600))
	return dir
}

func countFiles(t *testing.T, dir string) int {
	t.Helper()
	n := 0
	require.NoError(t, filepath.WalkDir(dir, func(_ string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			n++
		}
		return nil
	}))
	return n
}

// Every object the archiver creates lands in the quarantine directory, and cleanup removes it entirely.
func TestGitArchiver_QuarantineLifecycle(t *testing.T) {
	for _, state := range []string{WorkspaceAll, WorkspaceStaged, WorkspaceClean} {
		t.Run(state, func(t *testing.T) {
			dir := gitArchiverTestRepo(t)
			objectsBefore := countFiles(t, filepath.Join(dir, ".git", "objects"))

			h := newGitArchiver(dir, state, false, logger.NewDebugFileLogger(logger.DebugFileConfig{}))
			require.NoError(t, h.initialize(t.Context()))
			require.DirExists(t, h.tmpDir)
			require.NotEmpty(t, h.commit)

			quarantined := countFiles(t, filepath.Join(h.tmpDir, "objects"))
			if state == WorkspaceClean {
				assert.Equal(t, 0, quarantined)
			} else {
				assert.Positive(t, quarantined, "temporary tree/commit objects must be written to the quarantine")
			}
			assert.Equal(t, objectsBefore, countFiles(t, filepath.Join(dir, ".git", "objects")), "no objects may be written into the repo")

			h.cleanup()
			assert.NoDirExists(t, h.tmpDir)
		})
	}
}

// A context cancelled before initialize (Ctrl+C) must fail fast and leave no temp directory behind.
func TestGitArchiver_CancelledContextLeavesNothing(t *testing.T) {
	dir := gitArchiverTestRepo(t)
	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	h := newGitArchiver(dir, WorkspaceAll, false, logger.NewDebugFileLogger(logger.DebugFileConfig{}))
	require.Error(t, h.initialize(ctx))
	h.cleanup()
	if h.tmpDir != "" {
		assert.NoDirExists(t, h.tmpDir)
	}
}

// cleanup on an archiver that never initialized is a no-op (initialize may fail before the temp dir exists).
func TestGitArchiver_CleanupWithoutInitialize(t *testing.T) {
	h := newGitArchiver(t.TempDir(), WorkspaceAll, false, logger.NewDebugFileLogger(logger.DebugFileConfig{}))
	h.cleanup()
}

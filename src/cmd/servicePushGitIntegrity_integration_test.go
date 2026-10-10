package cmd

import (
	"crypto/sha256"
	"encoding/hex"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zeropsio/zcli/src/archiveClient"
)

// These tests guard the invariant that `zcli push` never creates, modifies or deletes a file in the user's git directory.
// Before the git object quarantine was introduced, cleanup ran `git prune` with the temporary index still in GIT_INDEX_FILE, which deleted every blob the user had
// staged but not yet committed and broke the next `git commit` with "invalid object".

// gitDirSnapshot maps every file under the git dir to a content hash, the index included.
// Callers must run their own `git status` before taking the snapshot, because status opportunistically refreshes the index.
func gitDirSnapshot(t *testing.T, gitDir string) map[string]string {
	t.Helper()
	snap := map[string]string{}
	err := filepath.WalkDir(gitDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(gitDir, path)
		if err != nil {
			return err
		}
		if d.IsDir() {
			snap[rel+"/"] = ""
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		sum := sha256.Sum256(data)
		snap[rel] = hex.EncodeToString(sum[:])
		return nil
	})
	require.NoError(t, err)
	return snap
}

func requireGitDirUnchanged(t *testing.T, before, after map[string]string) {
	t.Helper()
	var diffs []string
	for k, v := range before {
		if av, ok := after[k]; !ok {
			diffs = append(diffs, "deleted: "+k)
		} else if av != v {
			diffs = append(diffs, "modified: "+k)
		}
	}
	for k := range after {
		if _, ok := before[k]; !ok {
			diffs = append(diffs, "added: "+k)
		}
	}
	sort.Strings(diffs)
	require.Emptyf(t, diffs, "push must not touch the git dir, but it did:\n%s", strings.Join(diffs, "\n"))
}

// hasEntry is an exact-name lookup. archiveContains matches on suffix, which confuses staged.txt with modified-staged.txt.
func hasEntry(entries map[string]string, name string) bool {
	_, ok := entries[name]
	return ok
}

// countQuarantineDirs looks in os.TempDir(), callers must first point TMPDIR at a private directory so concurrent test packages or a developer's own push cannot skew the count.
func countQuarantineDirs(t *testing.T) int {
	t.Helper()
	matches, err := filepath.Glob(filepath.Join(os.TempDir(), "zcli-git-*"))
	require.NoError(t, err)
	return len(matches)
}

// mixedStateRepo builds a repo with one commit and every kind of pending change: a staged new file, a staged modification, an unstaged modification,
// an untracked file, an ignored file and a tracked file deleted from the working tree.
// It returns the object id of the staged new file's blob, which is the object the old prune used to delete.
func mixedStateRepo(t *testing.T, workDir string) (stagedBlob string) {
	t.Helper()
	writeZeropsYaml(t, workDir, "demo")
	require.NoError(t, os.WriteFile(filepath.Join(workDir, "committed.txt"), []byte("committed"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(workDir, "modified-staged.txt"), []byte("v1"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(workDir, "modified-unstaged.txt"), []byte("v1"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(workDir, "deleted.txt"), []byte("gone"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(workDir, ".gitignore"), []byte("ignored.txt\n"), 0o600))
	gitInit(t, workDir)
	gitAddCommit(t, workDir)

	require.NoError(t, os.WriteFile(filepath.Join(workDir, "staged.txt"), []byte("staged"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(workDir, "modified-staged.txt"), []byte("v2"), 0o600))
	runGit(t, workDir, "add", "staged.txt", "modified-staged.txt")
	require.NoError(t, os.WriteFile(filepath.Join(workDir, "modified-unstaged.txt"), []byte("v2"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(workDir, "untracked.txt"), []byte("untracked"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(workDir, "ignored.txt"), []byte("ignored"), 0o600))
	require.NoError(t, os.Remove(filepath.Join(workDir, "deleted.txt")))

	runGit(t, workDir, "status", "--porcelain")
	return runGit(t, workDir, "rev-parse", ":staged.txt")
}

func TestServicePushCommand_GitArchive_RepoUntouchedAndCommitWorksAfterPush(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not on PATH")
	}

	cases := []struct {
		state   string
		present []string
		absent  []string
		// unstagedContent is what modified-unstaged.txt must contain in the archive, empty when the file must be absent
		stagedContent, unstagedContent string
	}{
		{
			state:           archiveClient.WorkspaceAll,
			present:         []string{"committed.txt", "staged.txt", "untracked.txt"},
			absent:          []string{"ignored.txt", "deleted.txt"},
			stagedContent:   "v2",
			unstagedContent: "v2",
		},
		{
			state:           archiveClient.WorkspaceStaged,
			present:         []string{"committed.txt", "staged.txt", "deleted.txt"},
			absent:          []string{"ignored.txt", "untracked.txt"},
			stagedContent:   "v2",
			unstagedContent: "v1",
		},
		{
			state:           archiveClient.WorkspaceClean,
			present:         []string{"committed.txt", "deleted.txt"},
			absent:          []string{"ignored.txt", "untracked.txt", "staged.txt"},
			stagedContent:   "v1",
			unstagedContent: "v1",
		},
	}

	for _, tc := range cases {
		t.Run(tc.state, func(t *testing.T) {
			f := newFixture(t)
			f.SeedLogin("test-token")
			workDir := t.TempDir()
			stagedBlob := mixedStateRepo(t, workDir)
			gitDir := filepath.Join(workDir, ".git")

			statusBefore := runGit(t, workDir, "status", "--porcelain")
			before := gitDirSnapshot(t, gitDir)
			t.Setenv("TMPDIR", t.TempDir())
			tmpBefore := countQuarantineDirs(t)

			s := registerPushStubs(t, f, "demo")
			res := f.Run(
				"service", "push",
				"--service-id", pushServiceID,
				"--working-dir", workDir,
				"--workspace-state", tc.state,
				"--disable-logs",
			)
			assertPushSuccess(t, res, s, "demo")

			body, _ := s.uploadBody.Load().([]byte)
			entries := archiveEntries(t, body)
			for _, name := range tc.present {
				assert.Truef(t, hasEntry(entries, name), "%s must be in the %s archive; got %v", name, tc.state, keys(entries))
			}
			for _, name := range tc.absent {
				assert.Falsef(t, hasEntry(entries, name), "%s must not be in the %s archive; got %v", name, tc.state, keys(entries))
			}
			assert.Equal(t, tc.stagedContent, entries["modified-staged.txt"])
			assert.Equal(t, tc.unstagedContent, entries["modified-unstaged.txt"])

			requireGitDirUnchanged(t, before, gitDirSnapshot(t, gitDir))
			assert.Equal(t, statusBefore, runGit(t, workDir, "status", "--porcelain"), "working tree and index state must survive the push")
			assert.Equal(t, tmpBefore, countQuarantineDirs(t), "quarantine dir must be removed")

			// The regression itself: the staged blob is still there and committing works.
			runGit(t, workDir, "cat-file", "-e", stagedBlob)
			assert.Empty(t, runGit(t, workDir, "fsck", "--no-dangling", "--strict"))
			runGit(t, workDir, "commit", "-q", "-m", "after push")
			assert.Equal(t, "staged", runGit(t, workDir, "show", "HEAD:staged.txt"))
		})
	}
}

// Running push twice in a row with changes in between is the everyday flow. Nothing may accumulate in the repo.
func TestServicePushCommand_GitArchive_RepeatedPushesLeaveNothingBehind(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not on PATH")
	}
	f := newFixture(t)
	f.SeedLogin("test-token")
	workDir := t.TempDir()
	mixedStateRepo(t, workDir)
	before := gitDirSnapshot(t, filepath.Join(workDir, ".git"))

	s := registerPushStubs(t, f, "demo")
	for i := range 3 {
		require.NoError(t, os.WriteFile(filepath.Join(workDir, "untracked.txt"), []byte(strings.Repeat("x", i+1)), 0o600))
		res := f.Run("service", "push", "--service-id", pushServiceID, "--working-dir", workDir, "--disable-logs")
		require.Equalf(t, 0, res.ExitCode, "push #%d failed\n--- stderr ---\n%s", i+1, res.Stderr)
		body, _ := s.uploadBody.Load().([]byte)
		assert.Equal(t, strings.Repeat("x", i+1), archiveEntries(t, body)["untracked.txt"])
	}
	assert.Equal(t, int32(3), s.deployHits.Load())
	requireGitDirUnchanged(t, before, gitDirSnapshot(t, filepath.Join(workDir, ".git")))
}

// A clean tree pushed with workspace-state=all falls back to HEAD inside the archiver. That path must not write anything either.
func TestServicePushCommand_GitArchive_WorkspaceAllOnCleanTree(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not on PATH")
	}
	f := newFixture(t)
	f.SeedLogin("test-token")
	workDir := t.TempDir()
	writeZeropsYaml(t, workDir, "demo")
	require.NoError(t, os.WriteFile(filepath.Join(workDir, "main.go"), []byte("package main\n"), 0o600))
	gitInit(t, workDir)
	gitAddCommit(t, workDir)
	before := gitDirSnapshot(t, filepath.Join(workDir, ".git"))

	s := registerPushStubs(t, f, "demo")
	res := f.Run("service", "push", "--service-id", pushServiceID, "--working-dir", workDir, "--disable-logs")
	assertPushSuccess(t, res, s, "demo")

	body, _ := s.uploadBody.Load().([]byte)
	assert.True(t, archiveContains(archiveEntries(t, body), "main.go"))
	requireGitDirUnchanged(t, before, gitDirSnapshot(t, filepath.Join(workDir, ".git")))
	assert.Empty(t, runGit(t, workDir, "status", "--porcelain"))
}

// --working-dir pointing at a subdirectory of the repo. The archive is scoped to that subtree.
// The temporary index used to be created with a path relative to the subdirectory, which git resolved against the repo root and zcli then failed to remove.
func TestServicePushCommand_GitArchive_SubdirectoryWorkingDir(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not on PATH")
	}
	for _, state := range []string{archiveClient.WorkspaceAll, archiveClient.WorkspaceStaged, archiveClient.WorkspaceClean} {
		t.Run(state, func(t *testing.T) {
			f := newFixture(t)
			f.SeedLogin("test-token")
			root := t.TempDir()
			sub := filepath.Join(root, "services", "api")
			require.NoError(t, os.MkdirAll(sub, 0o755))
			require.NoError(t, os.WriteFile(filepath.Join(root, "root-file.txt"), []byte("root"), 0o600))
			writeZeropsYaml(t, sub, "demo")
			require.NoError(t, os.WriteFile(filepath.Join(sub, "main.go"), []byte("package main\n"), 0o600))
			gitInit(t, root)
			gitAddCommit(t, root)
			require.NoError(t, os.WriteFile(filepath.Join(sub, "staged.txt"), []byte("staged"), 0o600))
			runGit(t, root, "add", "services/api/staged.txt")
			require.NoError(t, os.WriteFile(filepath.Join(sub, "untracked.txt"), []byte("untracked"), 0o600))
			stagedBlob := runGit(t, root, "rev-parse", ":services/api/staged.txt")
			before := gitDirSnapshot(t, filepath.Join(root, ".git"))

			s := registerPushStubs(t, f, "demo")
			res := f.Run(
				"service", "push",
				"--service-id", pushServiceID,
				"--working-dir", sub,
				"--workspace-state", state,
				"--disable-logs",
			)
			assertPushSuccess(t, res, s, "demo")

			body, _ := s.uploadBody.Load().([]byte)
			entries := archiveEntries(t, body)
			assert.Contains(t, entries, "main.go", "archive is relative to the subdirectory; got %v", keys(entries))
			assert.NotContains(t, entries, "root-file.txt")
			assert.NotContains(t, entries, "services/api/main.go")
			assert.Equal(t, state != archiveClient.WorkspaceClean, hasEntry(entries, "staged.txt"))
			assert.Equal(t, state == archiveClient.WorkspaceAll, hasEntry(entries, "untracked.txt"))

			requireGitDirUnchanged(t, before, gitDirSnapshot(t, filepath.Join(root, ".git")))
			runGit(t, root, "cat-file", "-e", stagedBlob)
			runGit(t, root, "commit", "-q", "-m", "after push")
		})
	}
}

// A linked worktree keeps its index in .git/worktrees/<name>/ while objects live in the main repo's .git/objects.
func TestServicePushCommand_GitArchive_LinkedWorktree(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not on PATH")
	}
	f := newFixture(t)
	f.SeedLogin("test-token")
	mainDir := t.TempDir()
	writeZeropsYaml(t, mainDir, "demo")
	require.NoError(t, os.WriteFile(filepath.Join(mainDir, "main.go"), []byte("package main\n"), 0o600))
	gitInit(t, mainDir)
	gitAddCommit(t, mainDir)

	wtDir := filepath.Join(t.TempDir(), "feature")
	runGit(t, mainDir, "worktree", "add", "-q", "-b", "feature", wtDir)
	require.NoError(t, os.WriteFile(filepath.Join(wtDir, "staged.txt"), []byte("staged"), 0o600))
	runGit(t, wtDir, "add", "staged.txt")
	require.NoError(t, os.WriteFile(filepath.Join(wtDir, "untracked.txt"), []byte("untracked"), 0o600))
	stagedBlob := runGit(t, wtDir, "rev-parse", ":staged.txt")
	before := gitDirSnapshot(t, filepath.Join(mainDir, ".git"))

	s := registerPushStubs(t, f, "demo")
	res := f.Run("service", "push", "--service-id", pushServiceID, "--working-dir", wtDir, "--disable-logs")
	assertPushSuccess(t, res, s, "demo")

	body, _ := s.uploadBody.Load().([]byte)
	entries := archiveEntries(t, body)
	assert.True(t, archiveContains(entries, "main.go"))
	assert.True(t, archiveContains(entries, "staged.txt"))
	assert.True(t, archiveContains(entries, "untracked.txt"))

	requireGitDirUnchanged(t, before, gitDirSnapshot(t, filepath.Join(mainDir, ".git")))
	runGit(t, wtDir, "cat-file", "-e", stagedBlob)
	runGit(t, wtDir, "commit", "-q", "-m", "after push")
}

// A repo that already borrows objects through .git/objects/info/alternates (git clone --shared) must keep working, since the quarantine itself is implemented with alternates.
func TestServicePushCommand_GitArchive_RepoWithAlternates(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not on PATH")
	}
	f := newFixture(t)
	f.SeedLogin("test-token")
	upstream := t.TempDir()
	writeZeropsYaml(t, upstream, "demo")
	require.NoError(t, os.WriteFile(filepath.Join(upstream, "main.go"), []byte("package main\n"), 0o600))
	gitInit(t, upstream)
	gitAddCommit(t, upstream)

	workDir := filepath.Join(t.TempDir(), "shared")
	runGit(t, upstream, "clone", "-q", "--shared", upstream, workDir)
	require.FileExists(t, filepath.Join(workDir, ".git", "objects", "info", "alternates"))
	require.NoError(t, os.WriteFile(filepath.Join(workDir, "untracked.txt"), []byte("untracked"), 0o600))
	before := gitDirSnapshot(t, filepath.Join(workDir, ".git"))
	upstreamBefore := gitDirSnapshot(t, filepath.Join(upstream, ".git"))

	s := registerPushStubs(t, f, "demo")
	res := f.Run("service", "push", "--service-id", pushServiceID, "--working-dir", workDir, "--disable-logs")
	assertPushSuccess(t, res, s, "demo")

	body, _ := s.uploadBody.Load().([]byte)
	entries := archiveEntries(t, body)
	assert.True(t, archiveContains(entries, "main.go"), "file borrowed via alternates; got %v", keys(entries))
	assert.True(t, archiveContains(entries, "untracked.txt"))
	requireGitDirUnchanged(t, before, gitDirSnapshot(t, filepath.Join(workDir, ".git")))
	requireGitDirUnchanged(t, upstreamBefore, gitDirSnapshot(t, filepath.Join(upstream, ".git")))
}

// GIT_ALTERNATE_OBJECT_DIRECTORIES already set in the environment (e.g. zcli invoked from a git hook) must be chained, not dropped.
func TestServicePushCommand_GitArchive_PreservesEnvAlternates(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not on PATH")
	}
	f := newFixture(t)
	f.SeedLogin("test-token")

	// The committed file lives only in the donor repo, the working repo reaches it through the env alternate.
	donor := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(donor, "main.go"), []byte("package main\n"), 0o600))
	gitInit(t, donor)
	gitAddCommit(t, donor)
	t.Setenv("GIT_ALTERNATE_OBJECT_DIRECTORIES", filepath.Join(donor, ".git", "objects"))

	workDir := t.TempDir()
	gitInit(t, workDir)
	runGit(t, workDir, "fetch", "-q", donor, "main")
	runGit(t, workDir, "reset", "-q", "--hard", "FETCH_HEAD")
	writeZeropsYaml(t, workDir, "demo")
	before := gitDirSnapshot(t, filepath.Join(workDir, ".git"))

	s := registerPushStubs(t, f, "demo")
	res := f.Run("service", "push", "--service-id", pushServiceID, "--working-dir", workDir, "--disable-logs")
	assertPushSuccess(t, res, s, "demo")

	body, _ := s.uploadBody.Load().([]byte)
	entries := archiveEntries(t, body)
	assert.True(t, archiveContains(entries, "main.go"), "got %v", keys(entries))
	assert.True(t, archiveContains(entries, "zerops.yaml"))
	requireGitDirUnchanged(t, before, gitDirSnapshot(t, filepath.Join(workDir, ".git")))
}

// --deploy-git-folder tars the real .git directory. It must not pick up any zcli temporary files and must leave the repo untouched.
func TestServicePushCommand_GitArchive_DeployGitFolderHasNoTempFiles(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not on PATH")
	}
	for _, state := range []string{archiveClient.WorkspaceAll, archiveClient.WorkspaceStaged} {
		t.Run(state, func(t *testing.T) {
			f := newFixture(t)
			f.SeedLogin("test-token")
			workDir := t.TempDir()
			stagedBlob := mixedStateRepo(t, workDir)
			before := gitDirSnapshot(t, filepath.Join(workDir, ".git"))

			s := registerPushStubs(t, f, "demo")
			res := f.Run(
				"service", "push",
				"--service-id", pushServiceID,
				"--working-dir", workDir,
				"--workspace-state", state,
				"--deploy-git-folder",
				"--disable-logs",
			)
			assertPushSuccess(t, res, s, "demo")

			body, _ := s.uploadBody.Load().([]byte)
			entries := archiveEntries(t, body)
			assert.True(t, archiveContains(entries, ".git/HEAD"))
			assert.True(t, archiveContains(entries, "staged.txt"))
			for name := range entries {
				assert.NotContainsf(t, name, "zcli", "temporary zcli files leaked into the archive: %s", name)
			}
			requireGitDirUnchanged(t, before, gitDirSnapshot(t, filepath.Join(workDir, ".git")))
			runGit(t, workDir, "cat-file", "-e", stagedBlob)
			runGit(t, workDir, "commit", "-q", "-m", "after push")
		})
	}
}

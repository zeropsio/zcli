package archiveClient

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/zeropsio/zcli/src/cmdRunner"
	"github.com/zeropsio/zcli/src/logger"
)

type gitArchiver struct {
	commit         string
	workingDir     string
	env            []string
	workspaceState string
	// tmpDir holds the quarantine object directory and the temporary index. Removing it is the whole cleanup.
	tmpDir string

	verbose bool
	logger  logger.Logger
}

func newGitArchiver(workingDir string, workspaceState string, verbose bool, logger logger.Logger) *gitArchiver {
	return &gitArchiver{
		env:            os.Environ(),
		workingDir:     workingDir,
		workspaceState: workspaceState,
		verbose:        verbose,
		logger:         logger,
	}
}

func (h *gitArchiver) initialize(ctx context.Context) error {
	if err := h.setupQuarantine(ctx); err != nil {
		return err
	}

	switch h.workspaceState {
	case WorkspaceStaged:
		// `git write-tree` refreshes the cache-tree extension of the index it runs on, and done on the real index that would leave it pointing at trees which only exist in the quarantine.
		// So the tree is written from a private copy of the index instead.
		if err := h.copyIndex(ctx); err != nil {
			return err
		}

		// $(git commit-tree $(git write-tree) -p HEAD -m "zCLI tmp")
		tree, err := h.runCommand(ctx, "git", "write-tree")
		if err != nil {
			return err
		}
		commit, err := h.runCommand(ctx, "git", "commit-tree", tree, "-p", "HEAD", "-m", "zCLI tmp")
		if err != nil {
			return err
		}
		h.commit = commit
	case WorkspaceAll:
		indexFile := filepath.Join(h.tmpDir, "index")
		h.env = append(h.env, "GIT_INDEX_FILE="+indexFile)

		if h.verbose {
			h.logger.Info("Using git index file: " + indexFile)
		}

		if _, err := h.runCommand(ctx, "git", "read-tree", "HEAD"); err != nil {
			return err
		}

		// `git add` might take a while with large files
		if _, err := h.runCommand(ctx, "git", "add", "-A"); err != nil {
			return err
		}

		// This, sadly, creates a stash of all files (ignores only new files `--others`), so it can not be used for `staged` workspaceState
		commit, err := h.runCommand(ctx, "git", "stash", "create")
		if err != nil {
			return err
		}

		// `git stash create` returns nothing if there are no changes and user forgot to use WorkspaceClean, in such case, use HEAD
		h.commit = commit
		if h.commit == "" {
			h.commit = "HEAD"
		}
	case WorkspaceClean:
		// TODO(ms): add option for user to specify commit or tag, instead of forcing HEAD
		h.commit = "HEAD"
	}
	if h.verbose {
		h.logger.Info("Using git commit: " + h.commit)
	}
	return nil
}

// setupQuarantine redirects every object git writes during the push into a throwaway directory, the same way git quarantines incoming objects in receive-pack.
// Reads still fall through to the repository's own object store via alternates, so the temporary tree and commits can be archived.
// No object is ever created in or removed from the user's repository this way (git may still freshen mtimes of objects it finds there), and there is nothing to prune afterwards.
// Pruning used to be the cleanup here and it deleted objects the user had staged but not yet committed, because `git prune` only sees the index it is pointed at.
// Trade-off: in a partial clone, blobs fetched on demand while archiving land in the quarantine too and are fetched again on the next push.
func (h *gitArchiver) setupQuarantine(ctx context.Context) error {
	objectsDir, err := h.runCommand(ctx, "git", "rev-parse", "--git-path", "objects")
	if err != nil {
		return err
	}
	// Older git prints the path relative to the working directory (`--path-format=absolute` needs git 2.31).
	if !filepath.IsAbs(objectsDir) {
		objectsDir = filepath.Join(h.workingDir, objectsDir)
	}
	objectsDir, err = filepath.Abs(objectsDir)
	if err != nil {
		return err
	}

	tmpDir, err := os.MkdirTemp("", "zcli-git-")
	if err != nil {
		return err
	}
	h.tmpDir = tmpDir

	quarantineDir := filepath.Join(tmpDir, "objects")
	if err := os.Mkdir(quarantineDir, 0o700); err != nil {
		return err
	}

	// Keep alternates the caller already had, e.g. when zcli runs from a git hook.
	alternates := objectsDir
	if existing := os.Getenv("GIT_ALTERNATE_OBJECT_DIRECTORIES"); existing != "" {
		alternates += string(os.PathListSeparator) + existing
	}
	h.env = append(
		h.env,
		"GIT_OBJECT_DIRECTORY="+quarantineDir,
		"GIT_ALTERNATE_OBJECT_DIRECTORIES="+alternates,
	)

	if h.verbose {
		h.logger.Info("Using git object quarantine: " + quarantineDir + " (alternates: " + alternates + ")")
	}
	return nil
}

// copyIndex points GIT_INDEX_FILE at a copy of the repository's index inside the quarantine directory. A missing index is left missing, git then treats it as empty, exactly as it would the real one.
func (h *gitArchiver) copyIndex(ctx context.Context) error {
	indexPath, err := h.runCommand(ctx, "git", "rev-parse", "--git-path", "index")
	if err != nil {
		return err
	}
	if !filepath.IsAbs(indexPath) {
		indexPath = filepath.Join(h.workingDir, indexPath)
	}

	indexCopy := filepath.Join(h.tmpDir, "index")
	h.env = append(h.env, "GIT_INDEX_FILE="+indexCopy)
	if h.verbose {
		h.logger.Info("Using git index file: " + indexCopy + " (copy of " + indexPath + ")")
	}

	src, err := os.Open(indexPath)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	defer src.Close()

	dst, err := os.OpenFile(indexCopy, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		return err
	}
	if _, err := io.Copy(dst, src); err != nil {
		_ = dst.Close()
		return err
	}
	return dst.Close()
}

func (h *gitArchiver) cleanup() {
	if h.tmpDir == "" {
		return
	}
	if h.verbose {
		h.logger.Info("Removing git quarantine dir: " + h.tmpDir)
	}
	_ = os.RemoveAll(h.tmpDir)
}

func (h *gitArchiver) getArchiveCmd(ctx context.Context) *cmdRunner.ExecCmd {
	return h.getCommand(ctx, "git", "archive", "--format", "tar", h.commit)
}

//nolint:unparam // just because all commands currently run are `git`, doesn't mean we want to have it hardcoded here
func (h *gitArchiver) runCommand(ctx context.Context, command string, args ...string) (string, error) {
	out := &strings.Builder{}
	cmd := h.getCommand(ctx, command, args...)
	cmd.Stdout = out
	if err := cmd.Run(); err != nil {
		return "", err
	}
	return strings.TrimSpace(out.String()), nil
}

func (h *gitArchiver) getCommand(ctx context.Context, command string, args ...string) *cmdRunner.ExecCmd {
	if h.verbose {
		h.logger.Info(command, " ", strings.Join(args, " "))
	}
	cmd := cmdRunner.CommandContext(ctx, command, args...)
	cmd.Env = h.env
	cmd.Dir = h.workingDir
	if h.verbose {
		cmd.Stderr = h.logger
	}
	return cmd
}

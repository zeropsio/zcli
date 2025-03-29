package archiveClient

import (
	"context"
	"os"
	"path"
	"strings"

	"github.com/google/uuid"
	"github.com/zeropsio/zcli/src/cmdRunner"
	"github.com/zeropsio/zcli/src/logger"
)

type gitArchiver struct {
	commit         string
	workingDir     string
	env            []string
	workspaceState string
	indexFile      string

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
	switch h.workspaceState {
	case WorkspaceStaged:
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
		h.indexFile = path.Join(".git", "zcli-tmp-index-"+uuid.New().String())
		h.env = append(h.env, "GIT_INDEX_FILE="+h.indexFile)

		if h.verbose {
			h.logger.Info("Using git index file: " + h.indexFile)
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

func (h *gitArchiver) cleanup(ctx context.Context) {
	switch h.workspaceState {
	case WorkspaceClean:
		return
	case WorkspaceStaged:
		_ = h.getCommand(ctx, "git", "prune").Run()
	case WorkspaceAll:
		_ = h.getCommand(ctx, "git", "read-tree", "HEAD").Run()
		_ = h.getCommand(ctx, "git", "prune").Run()
	}

	if h.indexFile != "" {
		if h.verbose {
			h.logger.Info("Removing git index file: " + h.indexFile)
		}
		_ = os.Remove(path.Join(h.workingDir, h.indexFile))
	}
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

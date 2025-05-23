package archiveClient

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/zeropsio/zcli/src/cmdRunner"
	"github.com/zeropsio/zcli/src/uxBlock"
)

func (h *Handler) ArchiveGitFiles(ctx context.Context, uxBlocks uxBlock.UxBlocks, workingDir string, writer io.Writer) error {
	// Normalise the working directory path
	workingDir, err := filepath.Abs(workingDir)
	if err != nil {
		return err
	}
	workingDir += string(os.PathSeparator)

	if h.config.NoGit {
		return h.createNoGitArchive(uxBlocks, workingDir, writer)
	}

	rootDir, err := h.getRootDir(ctx, workingDir)
	if err != nil {
		return err
	}

	// Set up gzip compression
	gzipWriter := gzip.NewWriter(writer)
	defer gzipWriter.Close()

	// Start the appropriate archiving process based on config
	if h.config.DeployGitFolder {
		return h.createArchiveWithGitFolder(ctx, rootDir, gzipWriter)
	}
	return h.createSimpleArchive(ctx, rootDir, gzipWriter)
}

func (h *Handler) getRootDir(ctx context.Context, workingDir string) (string, error) {
	gitCommand := func(args ...string) (string, error) {
		if h.config.Verbose {
			h.config.Logger.Info("git ", strings.Join(args, " "))
		}

		out := &strings.Builder{}
		cmd := cmdRunner.CommandContext(ctx, "git", args...)
		cmd.Dir = workingDir
		cmd.Stdout = out
		if h.config.Verbose {
			cmd.Stderr = h.config.Logger
		}

		if err := cmd.Run(); err != nil {
			return "", err
		}
		return strings.TrimSpace(out.String()), nil
	}

	// first validate that git is installed and the folder is initialised
	if _, err := gitCommand("--version"); err != nil {
		return "", errors.New("git is not installed and flag --noGit was not set")
	}
	if out, err := gitCommand("rev-parse", "--is-inside-work-tree"); err != nil || out != "true" {
		return "", errors.New("folder is not initialized via git init and flag --noGit was not set")
	}
	if out, err := gitCommand("rev-list", "--all", "--count"); err != nil || out == "0" {
		return "", errors.New("at least one git commit must exist or flag --noGit must be set")
	}

	// then get the root dir
	path, err := gitCommand("rev-parse", "--show-toplevel")
	if err != nil {
		return "", err
	}
	return strings.TrimRight(path, "/\\") + string(os.PathSeparator), nil
}

// createNoGitArchive creates an archive of all files in the working directory
func (h *Handler) createNoGitArchive(uxBlocks uxBlock.UxBlocks, workingDir string, writer io.Writer) error {
	ignorer, err := LoadDeployFileIgnorer(workingDir)
	if err != nil {
		return err
	}

	files, err := h.FindFilesByRules(uxBlocks, workingDir, []string{"./"}, ignorer)
	if err != nil {
		return err
	}

	return h.TarFiles(writer, files)
}

// createSimpleArchive creates an archive without the .git directory
func (h *Handler) createSimpleArchive(ctx context.Context, workingDir string, writer io.Writer) error {
	archiver := newGitArchiver(workingDir, h.config.PushWorkspaceState, h.config.Verbose, h.config.Logger)
	if err := archiver.initialize(ctx); err != nil {
		return err
	}
	defer archiver.cleanup(ctx)

	// Run the git archive and pipe it to gzip
	cmd := archiver.getArchiveCmd(ctx)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	// Use a goroutine to copy data from stdout to gzipWriter
	var copyErr error
	killed := false
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		_, copyErr = io.Copy(writer, stdout)
		// if the copying failed, because e.g. server killed the connection, a process would get stuck, so we kill the process
		if copyErr != nil {
			killed = true
			_ = cmd.Process.Kill()
		}
	}()

	// Wait for the upload and command to finish
	wg.Wait()
	cmdErr := cmd.Wait()

	// Return the first error we encountered
	if !killed && cmdErr != nil {
		return cmdErr
	}
	return copyErr
}

// createArchiveWithGitFolder creates an archive including the .git directory
func (h *Handler) createArchiveWithGitFolder(ctx context.Context, workingDir string, writer io.Writer) error {
	// Create a tar writer for our archive
	tarWriter := tar.NewWriter(writer)
	defer tarWriter.Close()

	// First, add the .git directory to the tar
	if err := h.addGitDirectory(workingDir, tarWriter); err != nil {
		return err
	}

	// Create a pipe for the git archive output
	pipeReader, pipeWriter := io.Pipe()

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()

		// Read from git archive tar and merge into our tar
		err := h.mergeGitArchive(pipeReader, tarWriter)
		if err != nil {
			_ = pipeReader.CloseWithError(err)
		}
	}()

	if err := h.createSimpleArchive(ctx, workingDir, pipeWriter); err != nil {
		_ = pipeWriter.CloseWithError(err)
		return err
	}
	_ = pipeWriter.Close()

	wg.Wait()

	return nil
}

// addGitDirectory adds the .git directory to a tar archive
func (h *Handler) addGitDirectory(workingDir string, tw *tar.Writer) error {
	return filepath.Walk(filepath.Join(workingDir, ".git"), func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		relPath, err := filepath.Rel(workingDir, path)
		if err != nil {
			return err
		}
		return tarFile(tw, File{
			SourcePath:  path,
			ArchivePath: filepath.ToSlash(relPath),
		}, info)
	})
}

// mergeGitArchive reads a tar stream and copies its contents to another tar writer
func (h *Handler) mergeGitArchive(gitArchiveReader io.Reader, tarWriter *tar.Writer) error {
	tr := tar.NewReader(gitArchiveReader)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		// ignore global pax headers, they are generated by `git archive` command and would break stuff
		if header.Typeflag == tar.TypeXGlobalHeader {
			continue
		}

		if err := tarWriter.WriteHeader(header); err != nil {
			return err
		}
		//nolint:gosec // disable G110 - this is a local CLI util...
		if _, err := io.Copy(tarWriter, tr); err != nil {
			return err
		}
	}

	// After getting EOF from the tar reader, fully drain the underlying reader
	// It seems like the ` git archive ` command doesn't end on EOF, but always sends a full buffer, including nil bytes after EOF
	buffer := make([]byte, 4096)
	for {
		n, err := gitArchiveReader.Read(buffer)
		if err == io.EOF || (n == 0 && err != nil) {
			break
		}
	}

	//nolint:nilerr // Just discard the bytes - they're all zeros
	return nil
}

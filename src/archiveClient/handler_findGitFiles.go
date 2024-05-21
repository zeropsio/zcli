package archiveClient

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/zeropsio/zcli/src/cmdRunner"
)

func (h *Handler) FindGitFiles(ctx context.Context, workingDir string) (res []File, _ error) {
	workingDir, err := filepath.Abs(workingDir)
	if err != nil {
		return nil, err
	}
	workingDir += string(os.PathSeparator)

	createFile := func(filePath string) File {
		filePath = filepath.FromSlash(filePath)
		return File{
			SourcePath:  filePath,
			ArchivePath: filepath.ToSlash(strings.TrimPrefix(strings.TrimPrefix(filePath, workingDir), string(os.PathSeparator))),
		}
	}

	createCmd := func(name string, arg ...string) *cmdRunner.ExecCmd {
		cmd := cmdRunner.CommandContext(ctx, name, arg...)
		cmd.Dir = workingDir
		return cmd
	}

	// find excluded files
	excludedFiles := make(map[string]struct{})
	if err := h.listFiles(
		createCmd("git", "ls-files", "--deleted", "--exclude-standard", "-z"),
		replaceNullBytesWithNewLine,
		func(path string) error {
			excludedFiles[path] = struct{}{}

			return nil
		},
	); err != nil {
		return nil, err
	}

	if err := h.listFiles(
		createCmd("git", "ls-files", "--others", "--ignored", "--exclude-standard", "-z"),
		replaceNullBytesWithNewLine,
		func(path string) error {
			excludedFiles[path] = struct{}{}

			return nil
		},
	); err != nil {
		return nil, err
	}

	// add all non deleted
	if err := h.listFiles(
		createCmd("git", "ls-files", "--exclude-standard", "--recurse-submodules", "-z"),
		replaceNullBytesWithNewLine,
		func(path string) error {
			if _, exists := excludedFiles[path]; !exists {
				res = append(res, createFile(filepath.Join(workingDir, path)))
			}
			return nil
		},
	); err != nil {
		return nil, err
	}

	// add untracked files
	if err := h.listFiles(
		createCmd("git", "ls-files", "--others", "--exclude-standard", "-z"),
		replaceNullBytesWithNewLine,
		func(path string) error {
			if _, exists := excludedFiles[path]; !exists {
				res = append(res, createFile(filepath.Join(workingDir, path)))
			}
			return nil
		},
	); err != nil {
		return nil, err
	}

	res = h.fixMissingDirPath(res, createFile, make(map[string]struct{}))

	// add .git dir to allow git commands inside build.prepare and build.build commands
	if h.config.DeployGitFolder {
		if err := h.listFiles(
			createCmd("find", ".git/"),
			bytesReader,
			func(path string) error {
				res = append(res, createFile(filepath.Join(workingDir, path)))
				return nil
			},
		); err != nil {
			return nil, err
		}
	}

	return res, nil
}

// replaceNullBytesWithNewLine used in combination of '-z' flag with git command
// '-z' flag returns null terminated unquoted strings
func replaceNullBytesWithNewLine(output []byte) io.Reader {
	output = bytes.ReplaceAll(output, []byte{0}, []byte{'\n'})
	return bytes.NewReader(output)
}

func bytesReader(out []byte) io.Reader {
	return bytes.NewReader(out)
}

func (h *Handler) listFiles(cmd *cmdRunner.ExecCmd, reader func(out []byte) io.Reader, fn func(path string) error) error {
	output, err := cmdRunner.Run(cmd)
	if err != nil {
		return err
	}

	rd := bufio.NewReader(reader(output))
	for {
		lineB, _, err := rd.ReadLine()
		line := string(lineB)

		if errors.Is(err, io.EOF) {
			if line != "" {
				if err := fn(line); err != nil {
					return err
				}
			}
			break
		}
		if err != nil {
			return err
		}

		if err = fn(line); err != nil {
			return err
		}
	}

	return nil
}

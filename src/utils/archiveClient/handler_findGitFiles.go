package archiveClient

import (
	"bufio"
	"bytes"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/zeropsio/zcli/src/utils/cmdRunner"
)

func (h *Handler) FindGitFiles(workingDir string) (res []File, _ error) {
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

	createCmd := func(name string, arg ...string) *exec.Cmd {
		cmd := exec.Command(name, arg...)
		cmd.Dir = workingDir
		return cmd
	}

	// find excluded files
	excludedFiles := make(map[string]struct{})
	if err := h.listFiles(
		createCmd("git", "ls-files", "--deleted", "--exclude-standard"),
		func(path string) error {
			excludedFiles[path] = struct{}{}

			return nil
		},
	); err != nil {
		return nil, err
	}

	if err := h.listFiles(
		createCmd("git", "ls-files", "--others", "--ignored", "--exclude-standard"),
		func(path string) error {
			excludedFiles[path] = struct{}{}

			return nil
		},
	); err != nil {
		return nil, err
	}

	// add all non deleted
	if err := h.listFiles(
		createCmd("git", "ls-files", "--exclude-standard"),
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
		createCmd("git", "ls-files", "--others", "--exclude-standard"),
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
			func(path string) error {
				res = append(res, createFile(filepath.Join(workingDir, path)))
				return nil
			},
		); err != nil {
			return nil, err
		}
	}

	return
}

func (h *Handler) listFiles(cmd *exec.Cmd, fn func(path string) error) error {
	output, err := cmdRunner.Run(cmd)
	if err != nil {
		return err
	}

	rd := bufio.NewReader(bytes.NewReader(output))
	for {
		lineB, _, err := rd.ReadLine()
		line := string(lineB)

		if err == io.EOF {
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

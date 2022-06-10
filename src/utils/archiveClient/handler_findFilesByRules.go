package archiveClient

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/zerops-io/zcli/src/i18n"
)

func (h *Handler) FindFilesByRules(workingDir string, sources []string) ([]File, error) {
	workingDir, err := filepath.Abs(workingDir)
	if err != nil {
		return nil, err
	}

	fmt.Printf(i18n.ArchClientWorkingDirectory+"\n", workingDir)

	// resulting function returns File from provided path
	// if file shouldn't be included in the result, File.ArchivePath will be empty
	getCreateFile := func(dirPath string, parts []string) func(string) File {
		trimmer := func(filePath string) string {
			filePath = strings.TrimPrefix(filePath, path.Join(dirPath, parts[0]))
			if filePath == "" {
				return ""
			}
			return strings.TrimPrefix(filePath, string(os.PathSeparator))
		}

		return func(filePath string) File {
			filePath = filepath.FromSlash(filePath)
			return File{
				SourcePath:  filePath,
				ArchivePath: filepath.ToSlash(trimmer(filePath)),
			}
		}
	}

	res := make([]File, 0, 200)
	createdPaths := make(map[string]struct{})
	for _, source := range sources {
		parts := strings.Split(source, "~")
		if len(parts) > 2 {
			return nil, errors.New(i18n.ArchClientMaxOneTilde)
		}
		if len(parts) == 1 {
			parts = []string{
				"", parts[0],
			}
		}

		source := path.Join(workingDir, path.Join(parts...))
		source, err := filepath.Abs(source)
		if err != nil {
			return nil, err
		}

		fileInfo, err := os.Lstat(source)
		if err != nil {
			return nil, err
		}

		if fileInfo.IsDir() {
			source = strings.TrimSuffix(source, string(os.PathSeparator)) + string(os.PathSeparator)
			fmt.Printf(i18n.ArchClientPackingDirectory+"\n", source)
		} else {
			fmt.Printf(i18n.ArchClientPackingFile+"\n", source)
		}

		createFile := getCreateFile(workingDir, parts)

		files := make([]File, 0, 100)
		err = filepath.Walk(source, func(filePath string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				filePath = strings.TrimSuffix(filePath, string(os.PathSeparator)) + string(os.PathSeparator)
			}

			file := createFile(filePath)
			if file.ArchivePath != "" {
				files = append(files, file)
			}

			return nil
		})
		if err != nil {
			return nil, err
		}

		files = h.fixMissingDirPath(files, createFile, createdPaths)
		res = append(res, files...)
	}

	return res, nil
}

package zipClient

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/zerops-io/zcli/src/i18n"
)

func (h *Handler) FindFilesByRules(workingDir string, sources []string) (res []File, _ error) {
	workingDir, err := filepath.Abs(workingDir)
	if err != nil {
		return nil, err
	}

	fmt.Printf(i18n.ZipClientWorkingDirectory+"\n", workingDir)

	for _, source := range sources {

		parts := strings.Split(source, "~")
		if len(parts) > 2 {
			return nil, errors.New(i18n.ZipClientMaxOneTilde)
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

		trimPart := path.Join(workingDir, parts[0])

		fileInfo, err := os.Lstat(source)
		if err != nil {
			return nil, err
		}

		if fileInfo.IsDir() {
			fmt.Printf(i18n.ZipClientPackingDirectory+"\n", source)
		} else {
			fmt.Printf(i18n.ZipClientPackingFile+"\n", source)
		}

		err = filepath.Walk(source, func(filePath string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			archivePath := func(filePath string) string {
				if info.IsDir() {
					filePath += string(os.PathSeparator)
				}

				filePath = strings.TrimPrefix(filePath, trimPart)

				if filePath == parts[0] {
					return ""
				}
				filePath = strings.TrimPrefix(filePath, parts[0])
				filePath = strings.TrimPrefix(filePath, string(os.PathSeparator))

				return filePath
			}(filePath)

			if archivePath == "" {
				return nil
			}

			res = append(res, File{
				SourcePath:  filePath,
				ArchivePath: archivePath,
			})

			return nil
		})

		if err != nil {
			return nil, err
		}
	}

	return res, nil
}

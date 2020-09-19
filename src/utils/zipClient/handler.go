package zipClient

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/zerops-io/zcli/src/i18n"
)

type Config struct {
}

type Handler struct {
	config Config
}

func New(config Config) *Handler {
	return &Handler{
		config: config,
	}
}

func (h *Handler) Zip(w io.Writer, workingDir string, sources ...string) error {
	archive := zip.NewWriter(w)
	defer archive.Close()

	workingDir, err := filepath.Abs(workingDir)
	if err != nil {
		return err
	}

	fmt.Println(i18n.ZipClientWorkingDirectory + ": " + workingDir)

	for _, source := range sources {

		parts := strings.Split(source, "*")
		if len(parts) > 2 {
			return errors.New(i18n.ZipClientMaxOneAsterix)
		}
		if len(parts) == 1 {
			parts = []string{
				"", parts[0],
			}
		}

		source := path.Join(workingDir, path.Join(parts...))
		source, err := filepath.Abs(source)
		if err != nil {
			return err
		}

		trimPart := path.Join(workingDir, parts[0])

		fileInfo, err := os.Lstat(source)
		if err != nil {
			return err
		}

		if fileInfo.IsDir() {
			fmt.Println(i18n.ZipClientPackingDirectory + ": " + source)
		} else {
			fmt.Println(i18n.ZipClientPackingFile + ": " + source)
		}

		err = filepath.Walk(source, func(filePath string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			header, err := zip.FileInfoHeader(info)
			if err != nil {
				return err
			}

			header.Name = strings.TrimPrefix(func(filePath string) string {
				if info.IsDir() {
					filePath += "/"
				}

				filePath = strings.TrimPrefix(filePath, trimPart)

				if filePath == parts[0] {
					return ""
				}
				return strings.TrimPrefix(filePath, parts[0])

			}(filePath), "/")

			if header.Name == "" {
				return nil
			}

			if !info.IsDir() {
				header.Method = zip.Deflate
			}

			writer, err := archive.CreateHeader(header)
			if err != nil {
				return err
			}

			if info.IsDir() {
				return nil
			}

			if header.FileInfo().Mode()&os.ModeSymlink != 0 {
				symlink, err := os.Readlink(filePath)
				if err != nil {
					return err
				}

				_, err = writer.Write([]byte(symlink))
				if err != nil {
					return err
				}
			} else {
				file, err := os.Open(filePath)
				if err != nil {
					return err
				}
				defer file.Close()
				_, err = io.Copy(writer, file)
				if err != nil {
					return err
				}
			}

			return nil
		})

		if err != nil {
			return err
		}
	}

	return nil
}

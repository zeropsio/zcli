package zipClient

import (
	"archive/zip"
	"errors"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
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

	for _, source := range sources {

		parts := strings.Split(source, "*")
		if len(parts) > 2 {
			return errors.New("only one *(asterisk) is allowed")
		}
		source := path.Join(parts...)

		source = path.Join(workingDir, source)

		err := filepath.Walk(source, func(filePath string, info os.FileInfo, err error) error {
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

				filePath = strings.TrimPrefix(filePath, workingDir)

				if len(parts) == 1 {
					return filePath
				} else {
					if filePath == parts[0] {
						return ""
					}
					return strings.TrimPrefix(filePath, parts[0])
				}

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

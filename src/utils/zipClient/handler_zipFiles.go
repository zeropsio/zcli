package zipClient

import (
	"archive/zip"
	"io"
	"os"
)

type File struct {
	SourcePath  string
	ArchivePath string
}

func (h *Handler) ZipFiles(
	w io.Writer,
	files []File,
) error {
	archive := zip.NewWriter(w)
	defer archive.Close()

	for _, file := range files {

		fileInfo, err := os.Lstat(file.SourcePath)
		if err != nil {
			return err
		}

		err = zipFile(archive, file, fileInfo)
		if err != nil {
			return err
		}
	}

	return nil
}

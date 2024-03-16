package archiveClient

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"os"
)

type File struct {
	SourcePath  string // full path to the file using os.PathSeparator
	ArchivePath string // path to the file in archive using / as separator
}

func (h *Handler) TarFiles(w io.Writer, files []File) error {
	gz := gzip.NewWriter(w)
	defer gz.Close()

	archive := tar.NewWriter(gz)
	defer archive.Close()

	for _, file := range files {
		fileInfo, err := os.Lstat(file.SourcePath)
		if err != nil {
			return err
		}

		err = tarFile(archive, file, fileInfo)
		if err != nil {
			return err
		}
	}
	return nil
}

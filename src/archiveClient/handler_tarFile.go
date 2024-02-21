package archiveClient

import (
	"archive/tar"
	"io"
	"os"
	"strings"

	"github.com/pkg/errors"
)

func tarFile(archive *tar.Writer, file File, info os.FileInfo) error {
	archivePath := file.ArchivePath
	if info.IsDir() {
		archivePath = strings.TrimSuffix(archivePath, "/") + "/"
	}
	archivePath = strings.TrimPrefix(archivePath, "/")

	link := ""
	// FileInfoHeader does not read content of linked file to extract link path from, we must do that ourselves
	if info.Mode()&os.ModeSymlink > 0 {
		link, _ = os.Readlink(file.SourcePath)
	}

	header, err := tar.FileInfoHeader(info, link)
	if err != nil {
		return err
	}
	header.Name = archivePath

	if err := archive.WriteHeader(header); err != nil {
		return err
	}

	//nolint:exhaustive // Why: Other file types are not supported
	switch mode := info.Mode(); mode & os.ModeType {
	case os.ModeDir, os.ModeSymlink:
		return nil
	case 0:
		// regular file - copy content to tar
		f, err := os.Open(file.SourcePath)
		if err != nil {
			return err
		}
		n, cpErr := io.Copy(archive, f)
		if closeErr := f.Close(); closeErr != nil {
			return closeErr
		}
		if cpErr != nil {
			return cpErr
		}
		if n != info.Size() {
			return errors.Errorf("wrote %d, want %d", n, info.Size())
		}
	default:
		// let user know instead of silently ignoring unsupported files
		return errors.Errorf("unsupported file type: %s", header.Name)
	}

	return nil
}

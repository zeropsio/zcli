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

// TarFilesOption configures a single TarFiles call.
type TarFilesOption func(*tarFilesConfig)

type tarFilesConfig struct {
	onFile func(processed, total int)
}

// WithProgress registers a callback invoked after each successful file
// is written into the archive. Lets callers render a determinate
// progress indicator: processed = files completed so far, total =
// total files in the archive. Called from TarFiles's own goroutine.
func WithProgress(f func(processed, total int)) TarFilesOption {
	return func(c *tarFilesConfig) {
		c.onFile = f
	}
}

func (h *Handler) TarFiles(w io.Writer, files []File, opts ...TarFilesOption) error {
	cfg := tarFilesConfig{}
	for _, opt := range opts {
		opt(&cfg)
	}

	gz := gzip.NewWriter(w)
	defer gz.Close()

	archive := tar.NewWriter(gz)
	defer archive.Close()

	total := len(files)
	for i, file := range files {
		fileInfo, err := os.Lstat(file.SourcePath)
		if err != nil {
			return err
		}

		if err := tarFile(archive, file, fileInfo); err != nil {
			return err
		}

		if cfg.onFile != nil {
			cfg.onFile(i+1, total)
		}
	}
	return nil
}

package buildDeploy

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/zeropsio/zcli/src/i18n"
)

func (h *Handler) savePackage(config RunConfig, reader io.Reader) (io.Reader, error) {
	if config.ArchiveFilePath == "" {
		return reader, nil
	}

	filePath, err := filepath.Abs(config.ArchiveFilePath)
	if err != nil {
		return reader, err
	}

	// check if target file exists
	_, err = os.Stat(filePath)
	if err != nil && !os.IsNotExist(err) {
		return reader, err
	}
	if err == nil {
		return reader, fmt.Errorf(i18n.ArchClientFileAlreadyExists, config.ArchiveFilePath)
	}

	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE, 0660)
	if err != nil {
		return reader, err
	}

	return io.TeeReader(reader, file), nil
}

package file

import (
	"os"

	"github.com/pkg/errors"
)

func Open(filePath string, flag int, fileMode os.FileMode) (*os.File, error) {
	f, err := os.OpenFile(filePath, flag, fileMode)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	err = os.Chmod(filePath, fileMode)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return f, nil
}

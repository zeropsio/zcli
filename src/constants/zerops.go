package constants

import (
	"os"
	"path"
	"path/filepath"

	"github.com/pkg/errors"
)

const (
	DefaultRegionUrl      = "https://api.app.zerops.io/api/rest/public/region/zcli"
	zeropsDir             = "zerops"
	zeropsLogFile         = "zerops.log"
	cliDataFileName       = "cli.data"
	cliDataFilePathEnvVar = "ZEROPS_CLI_DATA_FILE_PATH"
	cliLogFilePathEnvVar  = "ZEROPS_CLI_LOG_FILE_PATH"
)

type pathReceiver func() (string, error)

func CliDataFilePath() (string, error) {
	return findFirstWritablePath(getDataFilePaths())
}

func LogFilePath() (string, error) {
	return findFirstWritablePath(getLogFilePath())
}

func receiverWithPath(receiver pathReceiver, elem ...string) pathReceiver {
	return func() (string, error) {
		dir, err := receiver()
		if err != nil {
			return "", err
		}
		elem = append([]string{dir}, elem...)

		return filepath.Join(elem...), nil
	}
}

func findFirstWritablePath(paths []pathReceiver) (string, error) {
	checkedPaths := make([]string, 0, len(paths))
	for _, p := range paths {
		path, err := p()
		if err == nil {
			checkedPaths = append(checkedPaths, path)
			if err := checkPath(path); err == nil {
				return path, nil
			}
		}
	}

	// TODO - janhajek translate
	return "", errors.Errorf("Unable to find writable path from %v", checkedPaths)
}

func checkPath(filePath string) error {
	dir := path.Dir(filePath)

	err := os.MkdirAll(dir, 0775)
	if err != nil {
		return err
	}

	f, err := os.OpenFile(filePath, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	return f.Close()
}

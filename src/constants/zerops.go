package constants

import (
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/zeropsio/zcli/src/i18n"
)

const (
	DefaultRegionUrl      = "https://api.app.zerops.io/api/rest/public/region/zcli"
	zeropsDir             = "zerops"
	zeropsLogFile         = "zerops.log"
	cliDataFileName       = "cli.data"
	cliDataFilePathEnvVar = "ZEROPS_CLI_DATA_FILE_PATH"
	cliLogFilePathEnvVar  = "ZEROPS_CLI_LOG_FILE_PATH"
)

type pathReceiver func() (path string, err error)

func CliDataFilePath() (string, error) {
	pathReceivers := getDataFilePathsReceivers()
	path := findFirstWritablePath(pathReceivers)
	if path == "" {
		paths := make([]string, 0, len(pathReceivers))
		for _, p := range pathReceivers {
			_, err := p()
			paths = append(paths, err.Error())
		}
		return "", errors.New(i18n.T(i18n.UnableToWriteCliData, "\n"+strings.Join(paths, "\n")))
	}
	return path, nil
}

func LogFilePath() (string, error) {
	pathReceivers := getLogFilePathReceivers()
	path := findFirstWritablePath(pathReceivers)
	if path == "" {
		paths := make([]string, 0, len(pathReceivers))
		for _, p := range pathReceivers {
			_, err := p()
			paths = append(paths, err.Error())
		}
		return "", errors.New(i18n.T(i18n.UnableToWriteLogFile, "\n"+strings.Join(paths, "\n")))
	}
	return path, nil
}

func receiverFromPath(path string) pathReceiver {
	return func() (string, error) {
		return checkPath(path)
	}
}

func receiverFromEnv(envName string) pathReceiver {
	return func() (string, error) {
		env := os.Getenv(envName)
		if env == "" {
			return "", errors.Errorf("env %s is empty", envName)
		}
		return checkPath(env)
	}
}

func receiverFromOsFunc(osFunc func() (string, error), elem ...string) pathReceiver {
	return func() (string, error) {
		dir, err := osFunc()
		if err != nil {
			return "", err
		}
		elem = append([]string{dir}, elem...)

		return filepath.Join(elem...), nil
	}
}

func findFirstWritablePath(paths []pathReceiver) string {
	for _, p := range paths {
		path, err := p()
		if err == nil {
			return path
		}
	}

	return ""
}

func checkPath(filePath string) (string, error) {
	dir := path.Dir(filePath)

	if err := os.MkdirAll(dir, 0775); err != nil {
		return "", err
	}

	f, err := os.OpenFile(filePath, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return "", err
	}
	err = f.Close()
	if err != nil {
		return "", err
	}

	return filePath, nil
}

package constants

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"

	"github.com/zeropsio/zcli/src/file"
	"github.com/zeropsio/zcli/src/i18n"
)

const (
	DefaultRegionUrl      = "https://api.app-prg1.zerops.io/api/rest/public/region/zcli"
	ZeropsDir             = "zerops"
	ZeropsLogFile         = "zerops.log"
	WgConfigFile          = "zerops.conf"
	WgInterfaceName       = "zerops"
	CliDataFileName       = "cli.data"
	CliDataFilePathEnvVar = "ZEROPS_CLI_DATA_FILE_PATH"
	CliLogFilePathEnvVar  = "ZEROPS_CLI_LOG_FILE_PATH"
	CliWgConfigPathEnvVar = "ZEROPS_WG_CONFIG_FILE_PATH"
	CliTerminalMode       = "ZEROPS_CLI_TERMINAL_MODE"
)

type pathReceiver func(fileMode os.FileMode) (path string, err error)

func CliDataFilePath() (string, os.FileMode, error) {
	return checkReceivers(getDataFilePathsReceivers(), 0600, i18n.UnableToWriteCliData)
}

func LogFilePath() (string, os.FileMode, error) {
	return checkReceivers(getLogFilePathReceivers(), 0666, i18n.UnableToWriteLogFile)
}

func WgConfigFilePath() (string, os.FileMode, error) {
	return checkReceivers(getWgConfigFilePathReceivers(), 0600, i18n.UnableToWriteWgConfigFile)
}

func checkReceivers(pathReceivers []pathReceiver, fileMode os.FileMode, errorText string) (string, os.FileMode, error) {
	path := findFirstWritablePath(pathReceivers, fileMode)
	if path == "" {
		paths := make([]string, 0, len(pathReceivers))
		for _, p := range pathReceivers {
			_, err := p(fileMode)
			paths = append(paths, err.Error())
		}
		return "", 0, errors.New(i18n.T(errorText, "\n"+strings.Join(paths, "\n")+"\n"))
	}
	return path, fileMode, nil
}

func receiverFromPath(path string) pathReceiver {
	return func(fileMode os.FileMode) (string, error) {
		return checkPath(path, fileMode)
	}
}

func receiverFromEnv(envName string) pathReceiver {
	return func(fileMode os.FileMode) (string, error) {
		env := os.Getenv(envName)
		if env == "" {
			return "", errors.Errorf("env %s is empty", envName)
		}
		return checkPath(env, fileMode)
	}
}

func receiverFromOsFunc(osFunc func() (string, error), elem ...string) pathReceiver {
	return func(fileMode os.FileMode) (string, error) {
		dir, err := osFunc()
		if err != nil {
			return "", err
		}

		return checkPath(filepath.Join(append([]string{dir}, elem...)...), fileMode)
	}
}

func receiverFromOsTemp(elem ...string) pathReceiver {
	return func(fileMode os.FileMode) (string, error) {
		return checkPath(filepath.Join(append([]string{os.TempDir()}, elem...)...), fileMode)
	}
}

func findFirstWritablePath(paths []pathReceiver, fileMode os.FileMode) string {
	for _, p := range paths {
		path, err := p(fileMode)
		if err == nil {
			return path
		}
	}

	return ""
}

func checkPath(filePath string, fileMode os.FileMode) (string, error) {
	dir := filepath.Dir(filePath)

	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}

	f, err := file.Open(filePath, os.O_RDWR|os.O_CREATE, fileMode)
	if err != nil {
		return "", err
	}
	err = f.Close()
	if err != nil {
		return "", err
	}

	return filePath, nil
}

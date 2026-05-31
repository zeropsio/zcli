package cmd

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/zeropsio/zcli/src/i18n"
)

func OpenPackageFile(archiveFilePath string, workingDir string) (*os.File, error) {
	return openPackageFile(archiveFilePath, workingDir)
}

func openPackageFile(archiveFilePath string, workingDir string) (*os.File, error) {
	workingDir, err := filepath.Abs(workingDir)
	if err != nil {
		return nil, err
	}

	archiveFilePath = filepath.Join(workingDir, archiveFilePath)

	filePath, err := filepath.Abs(archiveFilePath)
	if err != nil {
		return nil, err
	}

	// check if the target file exists
	_, err = os.Stat(filePath)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	if err == nil {
		return nil, errors.Errorf(i18n.T(i18n.ArchClientFileAlreadyExists), archiveFilePath)
	}

	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE, 0660)
	if err != nil {
		return nil, err
	}

	return file, nil
}

// validateSetupInYaml ensures that an explicitly requested setup name exists
// in the local zerops.yaml. The API rejects unknown setups too, but a
// client-side check produces a faster, friendlier error that lists the
// available setups instead of a generic ZeropsYamlSetupNotFound from the
// server.
func validateSetupInYaml(setup string, setups []string) error {
	if setup == "" {
		return nil
	}
	for _, s := range setups {
		if s == setup {
			return nil
		}
	}
	if len(setups) == 0 {
		return errors.Errorf("setup %q was not found in zerops.yaml: the file has no setups defined", setup)
	}
	return errors.Errorf("setup %q was not found in zerops.yaml; available setups: %s", setup, strings.Join(setups, ", "))
}

package buildDeploy

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/zerops-io/zcli/src/i18n"
)

func (h *Handler) savePackage(config RunConfig, buff *bytes.Buffer) error {
	if config.ZipFilePath != "" {
		zipFilePath, err := filepath.Abs(config.ZipFilePath)
		if err != nil {
			return err
		}

		err = ioutil.WriteFile(zipFilePath, buff.Bytes(), 0660)
		if err != nil {
			return err
		}

		fmt.Printf(i18n.BuildDeployPackageSavedInto+"\n", zipFilePath)
	}
	return nil
}

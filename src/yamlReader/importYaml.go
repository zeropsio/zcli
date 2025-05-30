package yamlReader

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/pkg/errors"

	"github.com/zeropsio/zcli/src/i18n"
	"github.com/zeropsio/zcli/src/uxBlock"
	"github.com/zeropsio/zcli/src/uxBlock/styles"
)

func ReadImportYamlContent(uxBlocks uxBlock.UxBlocks, importYamlPath string, workingDir string) ([]byte, error) {
	if !filepath.IsAbs(importYamlPath) {
		workingDir, err := filepath.Abs(workingDir)
		if err != nil {
			return nil, err
		}

		importYamlPath = filepath.Join(workingDir, importYamlPath)
	}

	fileInfo, err := os.Stat(importYamlPath)
	if err != nil {
		return nil, err
	}

	if fileInfo.IsDir() {
		return nil, errors.New(i18n.T(i18n.ImportYamlNotFound))
	}

	uxBlocks.PrintInfo(styles.InfoLine(fmt.Sprintf("%s: %s", i18n.T(i18n.ImportYamlFound), importYamlPath)))

	if fileInfo.Size() == 0 {
		return nil, errors.New(i18n.T(i18n.ImportYamlEmpty))
	}

	if fileInfo.Size() > 100*1024 {
		return nil, errors.New(i18n.T(i18n.ImportYamlTooLarge))
	}

	yamlContent, err := os.ReadFile(importYamlPath)
	if err != nil {
		return nil, err
	}

	if len(yamlContent) == 0 {
		return nil, errors.New(i18n.T(i18n.ImportYamlCorrupted))
	}

	uxBlocks.PrintInfo(styles.InfoLine(i18n.T(i18n.ImportYamlOk)))
	return yamlContent, nil
}

func ReadImportYamlContentFromStdin(uxBlocks uxBlock.UxBlocks) ([]byte, error) {
	buf := new(bytes.Buffer)
	size, err := io.Copy(buf, os.Stdin)
	if err != nil {
		return nil, fmt.Errorf("copying from stdin failed: %w", err)
	}

	if size == 0 {
		return nil, errors.New(i18n.T(i18n.ImportYamlEmpty))
	}

	if size > 100*1024 {
		return nil, errors.New(i18n.T(i18n.ImportYamlTooLarge))
	}

	uxBlocks.PrintInfo(styles.InfoLine(i18n.T(i18n.ImportYamlOk)))
	return buf.Bytes(), nil
}

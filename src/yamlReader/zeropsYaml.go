package yamlReader

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/zeropsio/zcli/src/generic"
	"github.com/zeropsio/zcli/src/i18n"
	"github.com/zeropsio/zcli/src/uxBlock"
	"github.com/zeropsio/zcli/src/uxBlock/styles"
	"gopkg.in/yaml.v3"
)

type ZeropsYaml struct {
	Zerops []Setup `yaml:"zerops"`
}

type Setup struct {
	Setup string `yaml:"setup"`
}

func ReadZeropsYamlSetups(in []byte) ([]string, error) {
	var z ZeropsYaml
	if err := yaml.Unmarshal(in, &z); err != nil {
		return nil, errors.Wrap(err, "unmarshal zerops yaml")
	}
	return generic.TransformSlice(z.Zerops, func(in Setup) string {
		return in.Setup
	}), nil
}

func ReadZeropsYamlContent(uxBlocks uxBlock.UxBlocks, selectedWorkingDir string, selectedZeropsYamlPath string) ([]byte, error) {
	workingDir, err := filepath.Abs(selectedWorkingDir)
	if err != nil {
		return nil, err
	}

	var pathsToCheck []string
	if selectedZeropsYamlPath != "" {
		if filepath.IsAbs(selectedZeropsYamlPath) {
			pathsToCheck = append(pathsToCheck, selectedZeropsYamlPath)
		} else {
			pathsToCheck = append(pathsToCheck, filepath.Join(workingDir, selectedZeropsYamlPath))
		}
	} else {
		pathsToCheck = append(pathsToCheck, filepath.Join(workingDir, "zerops.yaml"))
		pathsToCheck = append(pathsToCheck, filepath.Join(workingDir, "zerops.yml"))
	}

	zeropsYamlPath, err := func() (string, error) {
		for _, path := range pathsToCheck {
			zeropsYamlStat, err := os.Stat(path)
			if err == nil {
				uxBlocks.PrintInfo(styles.InfoLine(i18n.T(i18n.PushDeployZeropsYamlFound, path)))

				if zeropsYamlStat.Size() == 0 {
					return "", errors.New(i18n.T(i18n.PushDeployZeropsYamlEmpty))
				}
				if zeropsYamlStat.Size() > 10*1024 {
					return "", errors.New(i18n.T(i18n.PushDeployZeropsYamlTooLarge))
				}
				return path, nil
			}
		}
		return "", errors.New(i18n.T(i18n.PushDeployZeropsYamlNotFound, strings.Join(pathsToCheck, ", ")))
	}()
	if err != nil {
		return nil, err
	}

	yamlContent, err := os.ReadFile(zeropsYamlPath)
	if err != nil {
		return nil, err
	}

	return yamlContent, nil
}

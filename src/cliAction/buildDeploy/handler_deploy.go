package buildDeploy

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/zerops-io/zcli/src/constants"
	"github.com/zerops-io/zcli/src/i18n"
	"github.com/zerops-io/zcli/src/proto"
	"github.com/zerops-io/zcli/src/proto/business"
	"github.com/zerops-io/zcli/src/utils/processChecker"
)

func (h *Handler) Deploy(ctx context.Context, config RunConfig) error {

	serviceStack, err := h.checkInputValues(ctx, config)
	if err != nil {
		return err
	}

	fmt.Println(i18n.BuildDeployCreatingPackageStart)

	files, err := h.zipClient.FindFilesByRules(config.WorkingDir, config.PathsForPacking)
	if err != nil {
		return err
	}

	packageBuf := &bytes.Buffer{}
	err = h.zipClient.ZipFiles(packageBuf, files)
	if err != nil {
		return err
	}

	err = h.savePackage(config, packageBuf)
	if err != nil {
		return err
	}

	configContent, err := getConfigContent(config)
	if err != nil {
		return err
	}

	appVersion, err := h.createAppVersion(ctx, config, serviceStack)
	if err != nil {
		return err
	}

	err = h.packageUpload(appVersion, packageBuf)
	if err != nil {
		return err
	}

	fmt.Println(i18n.BuildDeployDeployingStart)

	deployResponse, err := h.apiGrpcClient.PutAppVersionDeploy(ctx, &business.PutAppVersionDeployRequest{
		Id:            appVersion.GetId(),
		ConfigContent: configContent,
	})
	if err := proto.BusinessError(deployResponse, err); err != nil {
		return err
	}

	deployProcessId := deployResponse.GetOutput().GetId()

	err = processChecker.CheckProcess(ctx, deployProcessId, h.apiGrpcClient)
	if err != nil {
		return err
	}

	fmt.Println(constants.Success + i18n.BuildDeploySuccess)

	return nil
}

func getConfigContent(config RunConfig) (*business.StringNull, error) {
	workingDir, err := filepath.Abs(config.WorkingDir)
	if err != nil {
		return nil, err
	}

	if config.ZeropsYamlPath != nil {
		workingDir = path.Join(workingDir, *config.ZeropsYamlPath)
	}

	zeropsYamlPath := path.Join(workingDir, zeropsYamlFileName)

	zeropsYamlStat, err := os.Stat(zeropsYamlPath)
	if err != nil {
		if os.IsNotExist(err) {
			if config.ZeropsYamlPath != nil {
				return nil, errors.New(i18n.BuildDeployZeropsYamlNotFound)
			}
		}
		return nil, nil
	}

	fmt.Printf("%s: %s\n", i18n.BuildDeployZeropsYamlFound, zeropsYamlPath)

	if zeropsYamlStat.Size() == 0 {
		return nil, errors.New(i18n.BuildDeployZeropsYamlEmpty)
	}
	if zeropsYamlStat.Size() > 10*1024 {
		return nil, errors.New(i18n.BuildDeployZeropsYamlTooLarge)
	}

	yamlContent, err := os.ReadFile(zeropsYamlPath)
	if err != nil {
		return nil, err
	}

	return &business.StringNull{
		Value: base64.StdEncoding.EncodeToString(yamlContent),
		Valid: true,
	}, nil
}

package buildDeploy

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io/ioutil"

	"github.com/zerops-io/zcli/src/utils"
	"github.com/zerops-io/zcli/src/zeropsApiProtocol"

	"github.com/zerops-io/zcli/src/i18n"
)

func (h *Handler) Push(ctx context.Context, config RunConfig) error {
	serviceStack, err := h.checkInputValues(ctx, config)
	if err != nil {
		return err
	}

	fmt.Println(i18n.BuildDeployCreatingPackageStart)

	files, err := h.zipClient.FindGitFiles(config.WorkingDir)
	if err != nil {
		return err
	}

	buildConfigContent, err := func() ([]byte, error) {
		for _, file := range files {
			if file.ArchivePath == "zerops_build.yml" {
				buildConfigContent, err := ioutil.ReadFile(file.SourcePath)
				if err != nil {
					return nil, err
				}

				if len(buildConfigContent) == 0 {
					return nil, errors.New(i18n.BuildDeployBuildConfigEmpty)
				}
				if len(buildConfigContent) > 10*1024*1024 {
					return nil, errors.New(i18n.BuildDeployBuildConfigTooLarge)
				}

				return buildConfigContent, nil
			}
		}

		return nil, errors.New(i18n.BuildDeployBuildConfigNotFound)
	}()
	if err != nil {
		return err
	}

	packageBuf := &bytes.Buffer{}
	err = h.zipClient.ZipFiles(packageBuf, files)
	if err != nil {
		return err
	}

	fmt.Println(i18n.BuildDeployCreatingPackageDone)

	err = h.savePackage(config, packageBuf)
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

	deployResponse, err := h.apiGrpcClient.PutAppVersionBuildAndDeploy(ctx, &zeropsApiProtocol.PutAppVersionBuildAndDeployRequest{
		Id:                 appVersion.GetId(),
		BuildConfigContent: base64.StdEncoding.EncodeToString(buildConfigContent),
	})
	if err := utils.HandleGrpcApiError(deployResponse, err); err != nil {
		return err
	}

	deployProcessId := deployResponse.GetOutput().GetId()

	err = h.checkProcess(ctx, deployProcessId)
	if err != nil {
		return err
	}

	fmt.Println(i18n.BuildDeploySuccess)

	return nil
}

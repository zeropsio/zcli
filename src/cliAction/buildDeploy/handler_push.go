package buildDeploy

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/zerops-io/zcli/src/i18n"
	"github.com/zerops-io/zcli/src/proto"
	"github.com/zerops-io/zcli/src/proto/business"
	"os"
)

func (h *Handler) Push(ctx context.Context, config RunConfig) error {
	serviceStack, err := h.checkInputValues(ctx, config)
	if err != nil {
		return err
	}

	if config.SourceName == "" {
		config.SourceName = serviceStack.GetName()
	}

	fmt.Println(i18n.BuildDeployCreatingPackageStart)

	files, err := h.zipClient.FindGitFiles(config.WorkingDir)
	if err != nil {
		return err
	}

	buildConfigContent, err := func() ([]byte, error) {
		for _, file := range files {
			if file.ArchivePath == zeropsYamlFileName {
				stat, err := os.Stat(file.SourcePath)
				if err != nil {
					return nil, err
				}

				if stat.Size() == 0 {
					return nil, errors.New(i18n.BuildDeployZeropsYamlEmpty)
				}
				if stat.Size() > 10*1024 {
					return nil, errors.New(i18n.BuildDeployZeropsYamlTooLarge)
				}

				buildConfigContent, err := os.ReadFile(file.SourcePath)
				if err != nil {
					return nil, err
				}

				return buildConfigContent, nil
			}
		}

		return nil, errors.New(i18n.BuildDeployZeropsYamlNotFound)
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

	deployResponse, err := h.apiGrpcClient.PutAppVersionBuildAndDeploy(ctx, &business.PutAppVersionBuildAndDeployRequest{
		Id:                 appVersion.GetId(),
		BuildConfigContent: base64.StdEncoding.EncodeToString(buildConfigContent),
		Source: &business.StringNull{
			Value: config.SourceName,
			Valid: true,
		},
	})
	if err := proto.BusinessError(deployResponse, err); err != nil {
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

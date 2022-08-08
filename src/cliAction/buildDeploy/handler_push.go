package buildDeploy

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/zerops-io/zcli/src/constants"
	"github.com/zerops-io/zcli/src/i18n"
	"github.com/zerops-io/zcli/src/proto"
	"github.com/zerops-io/zcli/src/proto/zBusinessZeropsApiProtocol"
	"github.com/zerops-io/zcli/src/utils/processChecker"
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

	files, err := h.archClient.FindGitFiles(config.WorkingDir)
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

	reader, writer := io.Pipe()
	defer reader.Close()

	tarErrChan := make(chan error, 1)

	go h.archClient.TarFiles(writer, files, tarErrChan)

	r, err := h.savePackage(config, reader)
	if err != nil {
		return err
	}

	appVersion, err := h.createAppVersion(ctx, config, serviceStack)
	if err != nil {
		return err
	}

	if err := h.packageUpload(appVersion, r); err != nil {
		// if an error occurred while packing the app, return that error
		select {
		case err := <-tarErrChan:
			return err
		default:
			return err
		}
	}

	// wait for packing and saving to finish (should already be done after the package upload has finished)
	if tarErr := <-tarErrChan; tarErr != nil {
		return tarErr
	}

	fmt.Println(i18n.BuildDeployCreatingPackageDone)

	if config.ArchiveFilePath != "" {
		fmt.Printf(i18n.BuildDeployPackageSavedInto+"\n", config.ArchiveFilePath)
	}

	fmt.Println(i18n.BuildDeployDeployingStart)

	deployResponse, err := h.apiGrpcClient.PutAppVersionBuildAndDeploy(ctx, &zBusinessZeropsApiProtocol.PutAppVersionBuildAndDeployRequest{
		Id:                 appVersion.GetId(),
		BuildConfigContent: base64.StdEncoding.EncodeToString(buildConfigContent),
		Source: &zBusinessZeropsApiProtocol.StringNull{
			Value: config.SourceName,
			Valid: true,
		},
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

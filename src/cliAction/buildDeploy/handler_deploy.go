package buildDeploy

import (
	"bytes"
	"context"
	"fmt"

	"github.com/zerops-io/zcli/src/i18n"
	"github.com/zerops-io/zcli/src/utils"
	"github.com/zerops-io/zcli/src/zeropsApiProtocol"
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

	appVersion, err := h.createAppVersion(ctx, config, serviceStack)
	if err != nil {
		return err
	}

	err = h.packageUpload(appVersion, packageBuf)
	if err != nil {
		return err
	}

	fmt.Println(i18n.BuildDeployDeployingStart)

	temporaryShutdown := false
	if serviceStack.GetStatus() == zeropsApiProtocol.ServiceStackStatus_SERVICE_STACK_STATUS_READY_TO_DEPLOY ||
		serviceStack.GetStatus() == zeropsApiProtocol.ServiceStackStatus_SERVICE_STACK_STATUS_ACTION_FAILED {
		temporaryShutdown = true
	}

	fmt.Printf(i18n.BuildDeployTemporaryShutdown+"\n", temporaryShutdown)

	deployResponse, err := h.apiGrpcClient.PutAppVersionDeploy(ctx, &zeropsApiProtocol.PutAppVersionDeployRequest{
		Id:                appVersion.GetId(),
		TemporaryShutdown: temporaryShutdown,
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

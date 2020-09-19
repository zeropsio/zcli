package deploy

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"

	"github.com/zerops-io/zcli/src/i18n"

	"github.com/zerops-io/zcli/src/utils"
	"github.com/zerops-io/zcli/src/zeropsApiProtocol"

	"github.com/zerops-io/zcli/src/utils/httpClient"
)

func (h *Handler) Run(ctx context.Context, config RunConfig) error {

	if config.ProjectName == "" {
		return errors.New(i18n.DeployProjectNameMissing)
	}

	if config.ServiceStackName == "" {
		return errors.New(i18n.DeployServiceStackNameMissing)
	}

	projectsResponse, err := h.apiGrpcClient.GetProjectsByName(ctx, &zeropsApiProtocol.GetProjectsByNameRequest{
		Name: config.ProjectName,
	})
	if err := utils.HandleGrpcApiError(projectsResponse, err); err != nil {
		return err
	}

	projects := projectsResponse.GetOutput().GetProjects()
	if len(projects) == 0 {
		return errors.New(i18n.DeployProjectNotFound)
	}
	if len(projects) > 1 {
		return errors.New(i18n.DeployProjectsWithSameName)
	}
	project := projects[0]

	serviceStackResponse, err := h.apiGrpcClient.GetServiceStackByName(ctx, &zeropsApiProtocol.GetServiceStackByNameRequest{
		ProjectId: project.GetId(),
		Name:      config.ServiceStackName,
	})
	if err := utils.HandleGrpcApiError(serviceStackResponse, err); err != nil {
		return err
	}
	serviceStack := serviceStackResponse.GetOutput()

	fmt.Println(i18n.DeployServiceStatus + ": " + serviceStack.GetStatus().String())

	temporaryShutdown := false
	if serviceStack.GetStatus() == zeropsApiProtocol.ServiceStackStatus_SERVICE_STACK_STATUS_READY_TO_DEPLOY ||
		serviceStack.GetStatus() == zeropsApiProtocol.ServiceStackStatus_SERVICE_STACK_STATUS_UPGRADE_FAILED {
		temporaryShutdown = true

	}

	fmt.Println(i18n.DeployTemporaryShutdown+": ", temporaryShutdown)

	fmt.Println(i18n.DeployCreatingPackageStart)

	appVersionResponse, err := h.apiGrpcClient.PostAppVersion(ctx, &zeropsApiProtocol.PostAppVersionRequest{
		ServiceStackId: serviceStack.GetId(),
	})
	if err := utils.HandleGrpcApiError(appVersionResponse, err); err != nil {
		return err
	}
	appVersion := appVersionResponse.GetOutput()

	buff := &bytes.Buffer{}
	err = h.zipClient.Zip(buff, config.WorkingDir, config.PathsForPacking...)
	if err != nil {
		return err
	}

	fmt.Println(i18n.DeployCreatingPackageDone)

	if config.ZipFilePath != "" {
		zipFilePath, err := filepath.Abs(config.ZipFilePath)
		if err != nil {
			return err
		}

		err = ioutil.WriteFile(zipFilePath, buff.Bytes(), 0660)
		if err != nil {
			return err
		}

		fmt.Println(i18n.DeployPackageSavedInto+": ", zipFilePath)
	}

	fmt.Println(i18n.DeployUploadingStart)

	cephResponse, err := h.httpClient.Put(appVersion.GetUploadUrl(), buff.Bytes(), httpClient.ContentType("application/zip"))
	if err != nil {
		return err
	}
	if cephResponse.StatusCode != http.StatusCreated {
		return errors.New(i18n.DeployUploadArchiveFailed)
	}

	fmt.Println(i18n.DeployUploadingDone)

	fmt.Println(i18n.DeployDeployingStart)

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

	fmt.Println(i18n.DeploySuccess)

	return nil
}

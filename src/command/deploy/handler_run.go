package deploy

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"path/filepath"

	"github.com/zerops-io/zcli/src/helpers"
	"github.com/zerops-io/zcli/src/service/httpClient"
	"github.com/zerops-io/zcli/src/zeropsApiProtocol"
)

func (h *Handler) Run(ctx context.Context, config RunConfig) error {

	if config.ProjectName == "" {
		return errors.New("project name must be filled")
	}

	if config.ServiceStackName == "" {
		return errors.New("service name must be filled")
	}

	projectsResponse, err := h.apiGrpcClient.GetProjectsByName(ctx, &zeropsApiProtocol.GetProjectsByNameRequest{
		Name: config.ProjectName,
	})
	if err := helpers.HandleGrpcApiError(projectsResponse, err); err != nil {
		return err
	}

	projects := projectsResponse.GetOutput().GetProjects()
	if len(projects) == 0 {
		return errors.New("project not found")
	}
	if len(projects) > 1 {
		return errors.New("there are multiple project with same name")
	}
	project := projects[0]

	serviceStackResponse, err := h.apiGrpcClient.GetServiceStackByName(ctx, &zeropsApiProtocol.GetServiceStackByNameRequest{
		ProjectId: project.GetId(),
		Name:      config.ServiceStackName,
	})
	if err := helpers.HandleGrpcApiError(serviceStackResponse, err); err != nil {
		return err
	}
	serviceStack := serviceStackResponse.GetOutput()

	h.logger.Debug("service name: " + serviceStack.GetName())
	h.logger.Debug("service status: " + serviceStack.GetStatus().String())

	temporaryShutdown := false
	if serviceStack.GetStatus() == zeropsApiProtocol.ServiceStackStatus_SERVICE_STACK_STATUS_READY_TO_DEPLOY ||
		serviceStack.GetStatus() == zeropsApiProtocol.ServiceStackStatus_SERVICE_STACK_STATUS_UPGRADE_FAILED {
		temporaryShutdown = true

	}

	h.logger.Info("temporaryShutdown: ", temporaryShutdown)

	appVersionResponse, err := h.apiGrpcClient.PostAppVersion(ctx, &zeropsApiProtocol.PostAppVersionRequest{
		ServiceStackId: serviceStack.GetId(),
	})
	if err := helpers.HandleGrpcApiError(appVersionResponse, err); err != nil {
		return err
	}
	appVersion := appVersionResponse.GetOutput()

	sourceDirectoryPath, err := filepath.Abs(config.SourceDirectoryPath)
	if err != nil {
		return err
	}

	h.logger.Info("packing directory: ", sourceDirectoryPath)

	data := &bytes.Buffer{}
	err = h.zipClient.Zip(sourceDirectoryPath, data)
	if err != nil {
		return err
	}

	cephResponse, err := h.httpClient.Put(appVersion.GetUploadUrl(), data.Bytes(), httpClient.ContentType("application/zip"))
	if err != nil {
		return err
	}
	if cephResponse.StatusCode != http.StatusCreated {
		return errors.New("upload archive error")
	}

	deployResponse, err := h.apiGrpcClient.PutAppVersionDeploy(ctx, &zeropsApiProtocol.PutAppVersionDeployRequest{
		Id:                appVersion.GetId(),
		TemporaryShutdown: temporaryShutdown,
	})
	if err := helpers.HandleGrpcApiError(deployResponse, err); err != nil {
		return err
	}

	deployProcessId := deployResponse.GetOutput().GetId()

	err = h.checkProcess(ctx, deployProcessId)
	if err != nil {
		return err
	}

	h.logger.Info("done")

	return nil
}

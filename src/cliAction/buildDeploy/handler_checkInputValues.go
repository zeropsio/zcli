package buildDeploy

import (
	"context"
	"errors"
	"fmt"
	"github.com/zerops-io/zcli/src/proto"
	"github.com/zerops-io/zcli/src/proto/business"

	"github.com/zerops-io/zcli/src/i18n"
)

func (h *Handler) checkInputValues(ctx context.Context, config RunConfig) (*business.GetServiceStackByNameResponseDto, error) {
	if config.ProjectName == "" {
		return nil, errors.New(i18n.BuildDeployProjectNameMissing)
	}

	if config.ServiceStackName == "" {
		return nil, errors.New(i18n.BuildDeployServiceStackNameMissing)
	}

	projectsResponse, err := h.apiGrpcClient.GetProjectsByName(ctx, &business.GetProjectsByNameRequest{
		Name: config.ProjectName,
	})
	if err := proto.BusinessError(projectsResponse, err); err != nil {
		return nil, err
	}

	projects := projectsResponse.GetOutput().GetProjects()
	if len(projects) == 0 {
		return nil, errors.New(i18n.BuildDeployProjectNotFound)
	}
	if len(projects) > 1 {
		return nil, errors.New(i18n.BuildDeployProjectsWithSameName)
	}
	project := projects[0]

	serviceStackResponse, err := h.apiGrpcClient.GetServiceStackByName(ctx, &business.GetServiceStackByNameRequest{
		ProjectId: project.GetId(),
		Name:      config.ServiceStackName,
	})
	if err := proto.BusinessError(serviceStackResponse, err); err != nil {
		return nil, err
	}
	serviceStack := serviceStackResponse.GetOutput()

	fmt.Printf(i18n.BuildDeployServiceStatus+"\n", serviceStack.GetStatus().String())

	return serviceStack, nil
}

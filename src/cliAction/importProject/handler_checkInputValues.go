package importProject

import (
	"context"
	"errors"
	// "fmt"

	// "github.com/zerops-io/zcli/src/i18n"
	// "github.com/zerops-io/zcli/src/utils"
	"github.com/zerops-io/zcli/src/zeropsApiProtocol"
)

func (h *Handler) checkInputValues(ctx context.Context, config RunConfig) (*zeropsApiProtocol.GetServiceStackByNameResponseDto, error) {
	// if config.ProjectName == "" {
	// 	return nil, errors.New(i18n.BuildDeployProjectNameMissing)
	// }

	// if config.ServiceStackName == "" {
	// 	return nil, errors.New(i18n.BuildDeployServiceStackNameMissing)
	// }

	// projectsResponse, err := h.apiGrpcClient.GetProjectsByName(ctx, &zeropsApiProtocol.GetProjectsByNameRequest{
	// 	Name: config.ProjectName,
	// })
	// if err := utils.HandleGrpcApiError(projectsResponse, err); err != nil {
	// 	return nil, err
	// }

	// projects := projectsResponse.GetOutput().GetProjects()
	// if len(projects) == 0 {
	// 	return nil, errors.New(i18n.BuildDeployProjectNotFound)
	// }
	// if len(projects) > 1 {
	// 	return nil, errors.New(i18n.BuildDeployProjectsWithSameName)
	// }
	// project := projects[0]

	// serviceStackResponse, err := h.apiGrpcClient.GetServiceStackByName(ctx, &zeropsApiProtocol.GetServiceStackByNameRequest{
	// 	ProjectId: project.GetId(),
	// 	Name:      config.ServiceStackName,
	// })
	// if err := utils.HandleGrpcApiError(serviceStackResponse, err); err != nil {
	// 	return nil, err
	// }
	// serviceStack := serviceStackResponse.GetOutput()

	// fmt.Printf(i18n.BuildDeployServiceStatus+"\n", serviceStack.GetStatus().String())

	// return serviceStack, nil
	return nil, errors.New("nic")
}

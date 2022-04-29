package startStopDeleteProject

import (
	"context"
	"errors"
	"fmt"

	"github.com/zerops-io/zcli/src/i18n"
	"github.com/zerops-io/zcli/src/utils"
	"github.com/zerops-io/zcli/src/zeropsApiProtocol"
)

func (h *Handler) Run(ctx context.Context, config RunConfig) error {

	if config.ProjectName == "" {
		return errors.New(i18n.StartProjectNameIsEmpty)
	}

	projectsResponse, err := h.apiGrpcClient.GetProjectsByName(ctx, &zeropsApiProtocol.GetProjectsByNameRequest{
		Name: config.ProjectName,
	})
	if err := utils.HandleGrpcApiError(projectsResponse, err); err != nil {
		return err
	}

	projects := projectsResponse.GetOutput().GetProjects()
	if len(projects) == 0 {
		return errors.New(i18n.StartProjectNotFound)
	}
	if len(projects) > 1 {
		return errors.New(i18n.StartProjectsWithSameName)
	}
	project := projects[0]

	startProjectResponse, err := h.apiGrpcClient.PutProjectStart(ctx, &zeropsApiProtocol.PutProjectStartRequest{
		Id: project.Id,
	})
	if err := utils.HandleGrpcApiError(startProjectResponse, err); err != nil {
		return err
	}

	fmt.Println(i18n.StartProjectProcessInit)

	resOutput := startProjectResponse.GetOutput()
	processId := resOutput.GetId()
	// fmt.Println(processId, resOutput.Status)

	// check process until FINISHED or CANCELED/FAILED
	err = h.checkProcess(ctx, processId)
	if err != nil {
		return err
	}

	getProcessResponse, err := h.apiGrpcClient.GetProcess(ctx, &zeropsApiProtocol.GetProcessRequest{
		Id: processId,
	})
	if err := utils.HandleGrpcApiError(getProcessResponse, err); err != nil {
		return err
	}
	
	fmt.Println(i18n.StartProcessSuccess)

	return nil
}

package repository

import (
	"context"

	"github.com/zeropsio/zcli/src/entity"
	"github.com/zeropsio/zcli/src/errorsx"
	"github.com/zeropsio/zcli/src/i18n"
	"github.com/zeropsio/zcli/src/zeropsRestApiClient"
	"github.com/zeropsio/zerops-go/dto/input/path"
	"github.com/zeropsio/zerops-go/dto/output"
	"github.com/zeropsio/zerops-go/errorCode"
	"github.com/zeropsio/zerops-go/types/uuid"
)

func GetProjectById(
	ctx context.Context,
	restApiClient *zeropsRestApiClient.Handler,
	projectId uuid.ProjectId,
) (*entity.Project, error) {
	projectResponse, err := restApiClient.GetProject(ctx, path.ProjectId{Id: projectId})
	if err != nil {
		return nil, err
	}

	projectOutput, err := projectResponse.Output()
	if err != nil {
		return nil, zeropsRestApiClient.CheckError(
			err,
			zeropsRestApiClient.CheckInvalidUserInput(
				"id",
				errorsx.NewUserError(i18n.T(i18n.ProjectIdInvalidFormat), err),
			),
			zeropsRestApiClient.CheckErrorCode(
				errorCode.ProjectNotFound,
				errorsx.NewUserError(i18n.T(i18n.ProjectNotFound, projectId), err),
			),
		)
	}

	project := projectFromApiOutput(projectOutput)
	return &project, nil
}

func GetAllProjects(
	ctx context.Context,
	restApiClient *zeropsRestApiClient.Handler,
) ([]entity.Project, error) {
	info, err := restApiClient.GetUserInfo(ctx)
	if err != nil {
		return nil, err
	}

	i, err := info.Output()
	if err != nil {
		return nil, err
	}

	var projects []entity.Project
	for _, b := range i.ClientUserList {
		response, err := restApiClient.GetProjectsByClient(ctx, b.ClientId)
		if err != nil {
			return nil, err
		}
		projectsResponse, err := response.Output()
		if err != nil {
			return nil, err
		}

		for _, project := range projectsResponse.Items {
			projects = append(projects, projectFromEsSearch(project))
		}
	}

	return projects, nil
}

func projectFromEsSearch(esProject zeropsRestApiClient.EsProject) entity.Project {
	description, _ := esProject.Description.Get()

	return entity.Project{
		ID:          esProject.Id,
		Name:        esProject.Name,
		ClientId:    esProject.ClientId,
		Description: description,
		Status:      esProject.Status,
	}
}

func projectFromApiOutput(project output.Project) entity.Project {
	description, _ := project.Description.Get()

	return entity.Project{
		ID:          project.Id,
		Name:        project.Name,
		ClientId:    project.ClientId,
		Description: description,
		Status:      project.Status,
	}
}

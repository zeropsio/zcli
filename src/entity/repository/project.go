package repository

import (
	"context"

	"github.com/zeropsio/zcli/src/entity"
	"github.com/zeropsio/zcli/src/zeropsRestApiClient"
	"github.com/zeropsio/zerops-go/dto/input/body"
	"github.com/zeropsio/zerops-go/dto/input/path"
	"github.com/zeropsio/zerops-go/dto/output"
	"github.com/zeropsio/zerops-go/types"
	"github.com/zeropsio/zerops-go/types/uuid"
)

func GetProjectById(
	ctx context.Context,
	restApiClient *zeropsRestApiClient.Handler,
	projectId uuid.ProjectId,
) (entity.Project, error) {
	projectResponse, err := restApiClient.GetProject(ctx, path.ProjectId{Id: projectId})
	if err != nil {
		return entity.Project{}, err
	}

	projectOutput, err := projectResponse.Output()
	if err != nil {
		return entity.Project{}, err
	}

	project := projectFromApiOutput(projectOutput)
	return project, nil
}

func GetAllProjects(
	ctx context.Context,
	restApiClient *zeropsRestApiClient.Handler,
) ([]entity.Project, error) {
	orgs, err := GetAllOrgs(ctx, restApiClient)
	if err != nil {
		return nil, err
	}

	var projects []entity.Project
	for _, org := range orgs {
		esFilter := body.EsFilter{
			Search: []body.EsSearchItem{
				{
					Name:     "clientId",
					Operator: "eq",
					Value:    org.ID.TypedString(),
				},
			},
		}

		response, err := restApiClient.PostProjectSearch(ctx, esFilter)
		if err != nil {
			return nil, err
		}
		projectsResponse, err := response.Output()
		if err != nil {
			return nil, err
		}

		for _, project := range projectsResponse.Items {
			projects = append(projects, projectFromEsSearch(org, project))
		}
	}

	return projects, nil
}

type ProjectPost struct {
	ClientId uuid.ClientId
	Name     types.String
	Tags     types.StringArray
}

func PostProject(
	ctx context.Context,
	restApiClient *zeropsRestApiClient.Handler,
	post ProjectPost,
) (entity.Project, error) {
	postBody := body.PostProject{
		ClientId: post.ClientId,
		Name:     post.Name,
		TagList:  post.Tags,
	}
	if postBody.TagList == nil {
		postBody.TagList = make(types.StringArray, 0)
	}
	response, err := restApiClient.PostProject(ctx, postBody)
	if err != nil {
		return entity.Project{}, err
	}
	project, err := response.Output()
	if err != nil {
		return entity.Project{}, err
	}
	return projectFromApiOutput(project), nil
}

func projectFromEsSearch(org entity.Org, esProject output.EsProject) entity.Project {
	description, _ := esProject.Description.Get()

	return entity.Project{
		ID:          esProject.Id,
		Name:        esProject.Name,
		OrgId:       org.ID,
		OrgName:     org.Name,
		Description: description,
		Status:      esProject.Status,
	}
}

func projectFromApiOutput(project output.Project) entity.Project {
	description, _ := project.Description.Get()

	return entity.Project{
		ID:          project.Id,
		Name:        project.Name,
		OrgId:       project.ClientId,
		Description: description,
		Status:      project.Status,
	}
}

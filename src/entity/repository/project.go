package repository

import (
	"context"

	"github.com/zeropsio/zcli/src/entity"
	"github.com/zeropsio/zcli/src/gn"
	"github.com/zeropsio/zcli/src/zeropsRestApiClient"
	"github.com/zeropsio/zerops-go/dto/input/body"
	"github.com/zeropsio/zerops-go/dto/input/path"
	"github.com/zeropsio/zerops-go/dto/input/query"
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

	activeOrgs := gn.FilterSlice(orgs, func(o entity.Org) bool {
		return o.Status.IsActive()
	})

	var projects []entity.Project
	for _, org := range activeOrgs {
		response, err := restApiClient.GetClientProject(
			ctx,
			path.ClientId{Id: org.Id},
			query.ListClientProjects{},
		)
		if err != nil {
			return nil, err
		}
		projectList, err := response.Output()
		if err != nil {
			return nil, err
		}

		for _, p := range projectList.List {
			proj := projectFromApiOutput(p)
			proj.OrgName = org.Name
			projects = append(projects, proj)
		}
	}

	return projects, nil
}

func PostProject(
	ctx context.Context,
	restApiClient *zeropsRestApiClient.Handler,
	post entity.PostProject,
) (entity.Project, error) {
	postBody := body.PostProject{
		Name:         post.Name,
		Mode:         &post.Mode,
		TagList:      post.Tags,
		SshIsolation: post.SshIsolation,
		EnvIsolation: post.EnvIsolation,
		Location:     post.Location,
	}
	if postBody.TagList == nil {
		postBody.TagList = make(types.StringArray, 0)
	}
	response, err := restApiClient.PostClientProject(ctx, path.ClientId{Id: post.OrgId}, postBody)
	if err != nil {
		return entity.Project{}, err
	}
	project, err := response.Output()
	if err != nil {
		return entity.Project{}, err
	}
	return projectFromApiOutput(project), nil
}

func projectFromApiOutput(project output.Project) entity.Project {
	description, _ := project.Description.Get()

	return entity.Project{
		Id:          project.Id,
		Name:        project.Name,
		Mode:        project.Mode,
		OrgId:       project.ClientId,
		Description: description,
		Status:      project.Status,
	}
}

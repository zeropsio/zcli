package repository

import (
	"context"

	"github.com/zeropsio/zcli/src/entity"
	"github.com/zeropsio/zcli/src/zeropsRestApiClient"
	"github.com/zeropsio/zerops-go/dto/input/body"
	"github.com/zeropsio/zerops-go/dto/input/path"
	"github.com/zeropsio/zerops-go/dto/input/query"
	"github.com/zeropsio/zerops-go/dto/output"
	"github.com/zeropsio/zerops-go/types"
	"github.com/zeropsio/zerops-go/types/enum"
	"github.com/zeropsio/zerops-go/types/stringId"
	"github.com/zeropsio/zerops-go/types/uuid"
)

// listPageLimit matches the API's maximum page size, so paged fetch-all loops need as few round trips as possible.
const listPageLimit = 100

func GetProjectById(
	ctx context.Context,
	restApiClient *zeropsRestApiClient.Handler,
	projectId entity.ProjectId,
) (entity.Project, error) {
	projectResponse, err := restApiClient.GetProject(ctx, path.ProjectId{Id: uuid.ProjectId(projectId)})
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
		if !org.Status.IsActive() {
			continue
		}
		for offset := 0; ; {
			response, err := restApiClient.GetClientProject(ctx, path.ClientId{Id: uuid.ClientId(org.Id)}, query.ListClientProjects{
				Limit:  types.NewIntNull(listPageLimit),
				Offset: types.NewIntNull(offset),
			})
			if err != nil {
				return nil, err
			}
			projectsResponse, err := response.Output()
			if err != nil {
				return nil, err
			}

			for _, project := range projectsResponse.List {
				projects = append(projects, projectFromApiOutputWithOrg(org, project))
			}

			offset += len(projectsResponse.List)
			if len(projectsResponse.List) == 0 || offset >= projectsResponse.Total.Native() {
				break
			}
		}
	}

	return projects, nil
}

func PostProject(
	ctx context.Context,
	restApiClient *zeropsRestApiClient.Handler,
	post entity.PostProject,
) (entity.Project, error) {
	mode := enum.ProjectModeEnum(post.Mode)
	tagList := types.StringArray(post.Tags)
	if tagList == nil {
		tagList = make(types.StringArray, 0)
	}
	location := stringId.LocationIdNull{}
	if loc, ok := post.Location.Get(); ok {
		location = stringId.NewLocationIdNullFromString(loc.String())
	}
	postBody := body.PostProject{
		Name:         types.NewString(post.Name),
		Mode:         &mode,
		TagList:      tagList,
		SshIsolation: types.NewStringNull(post.SshIsolation),
		EnvIsolation: types.NewStringNull(post.EnvIsolation),
		Location:     location,
	}
	response, err := restApiClient.PostClientProject(ctx, path.ClientId{Id: uuid.ClientId(post.OrgId)}, postBody)
	if err != nil {
		return entity.Project{}, err
	}
	project, err := response.Output()
	if err != nil {
		return entity.Project{}, err
	}
	return projectFromApiOutput(project), nil
}

// GetProjectEnvFile returns the rendered env file for a project. When serviceName
// is non-empty the envs are resolved in that service's isolation context.
func GetProjectEnvFile(
	ctx context.Context,
	restApiClient *zeropsRestApiClient.Handler,
	projectId entity.ProjectId,
	serviceName string,
	userOnly bool,
) (string, error) {
	var name types.EmptyString
	overrideEnvIsolation := enum.GetProjectEnvFileOverrideEnvIsolationEnumNone
	if serviceName != "" {
		name = types.NewEmptyString(serviceName)
		overrideEnvIsolation = enum.GetProjectEnvFileOverrideEnvIsolationEnumService
	}

	response, err := restApiClient.GetProjectEnvFile(
		ctx,
		path.ProjectId{Id: uuid.ProjectId(projectId)},
		query.GetProjectEnvFile{
			Name:                 name,
			OverrideEnvIsolation: overrideEnvIsolation,
			UserOnly:             types.NewBool(userOnly),
		},
	)
	if err != nil {
		return "", err
	}
	envs, err := response.Output()
	if err != nil {
		return "", err
	}
	return envs.EnvFile.String(), nil
}

func DeleteProject(
	ctx context.Context,
	restApiClient *zeropsRestApiClient.Handler,
	projectId entity.ProjectId,
) (entity.Process, error) {
	response, err := restApiClient.DeleteProject(ctx, path.ProjectId{Id: uuid.ProjectId(projectId)})
	if err != nil {
		return entity.Process{}, err
	}
	processOutput, err := response.Output()
	if err != nil {
		return entity.Process{}, err
	}
	return ProcessFromApiOutput(processOutput), nil
}

func projectFromApiOutputWithOrg(org entity.Org, project output.Project) entity.Project {
	out := projectFromApiOutput(project)
	out.OrgName = org.Name
	return out
}

func projectFromApiOutput(project output.Project) entity.Project {
	description, _ := project.Description.Get()

	return entity.Project{
		Id:          entity.ProjectId(project.Id),
		Name:        project.Name.Native(),
		Mode:        entity.ProjectMode(project.Mode),
		OrgId:       entity.OrgId(project.ClientId),
		Description: description.Native(),
		Status:      entity.ProjectStatus(project.Status),
	}
}

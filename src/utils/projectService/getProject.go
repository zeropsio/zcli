package projectService

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/zerops-io/zcli/src/i18n"
	"github.com/zerops-io/zcli/src/proto"
	"github.com/zerops-io/zcli/src/proto/business"
)

func GetProject(ctx context.Context, apiGrpcClient business.ZeropsApiProtocolClient, projectNameOrId string) (*business.Project, error) {

	if projectNameOrId == "" {
		return nil, errors.New(i18n.ProjectNameOrIdEmpty)
	}

	projects, err := getByName(ctx, apiGrpcClient, projectNameOrId)
	if err != nil {
		return nil, err
	}

	if len(projects) > 1 {
		return nil, getProjectSameNameErr(projects)
	}

	if len(projects) == 1 {
		return projects[0], nil
	}

	project, err := getById(ctx, apiGrpcClient, projectNameOrId)
	if err != nil {
		return nil, err
	}
	if project == nil {
		return nil, errors.New(i18n.ProjectNotFound)
	}

	return project, nil
}

func GetProjectId(ctx context.Context, apiGrpcClient business.ZeropsApiProtocolClient, projectNameOrId string) (string, error) {
	project, err := GetProject(ctx, apiGrpcClient, projectNameOrId)
	if err != nil {
		return "", err
	}
	return project.GetId(), nil
}

// return project IDs hint for projects with the same name
func getProjectSameNameErr(projects []*business.Project) error {
	var out []string

	for _, project := range projects {
		out = append(out, project.GetId())
	}
	idList := strings.Join(out, ",")
	errMsg := fmt.Errorf("%s\n%s%s", i18n.ProjectsWithSameName, i18n.AvailableProjectIds, idList)

	return errMsg
}

func getByName(ctx context.Context, apiGrpcClient business.ZeropsApiProtocolClient, projectName string) ([]*business.Project, error) {
	projectsResponse, err := apiGrpcClient.GetProjectsByName(ctx, &business.GetProjectsByNameRequest{
		Name: projectName,
	})
	if err := proto.BusinessError(projectsResponse, err); err != nil {
		return nil, err
	}
	projects := projectsResponse.GetOutput().GetProjects()
	return projects, nil
}

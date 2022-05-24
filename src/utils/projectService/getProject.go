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

func GetProject(ctx context.Context, apiGrpcClient business.ZeropsApiProtocolClient, projectName string) (*business.Project, error) {

	if projectName == "" {
		return nil, errors.New(i18n.ProjectNameIsEmpty)
	}

	projectsResponse, err := apiGrpcClient.GetProjectsByName(ctx, &business.GetProjectsByNameRequest{
		Name: projectName,
	})
	if err := proto.BusinessError(projectsResponse, err); err != nil {
		return nil, err
	}

	projects := projectsResponse.GetOutput().GetProjects()
	if len(projects) == 0 {
		return nil, errors.New(i18n.ProjectNotFound)
	}
	if len(projects) > 1 {
		return nil, getProjectSameNameErr(projects)
	}
	project := projects[0]
	return project, nil
}

func GetProjectId(ctx context.Context, apiGrpcClient business.ZeropsApiProtocolClient, projectName string) (string, error) {
	project, err := GetProject(ctx, apiGrpcClient, projectName)
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

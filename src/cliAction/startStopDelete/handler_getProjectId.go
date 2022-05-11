package startStopDelete

import (
	"context"
	"errors"

	"github.com/zerops-io/zcli/src/i18n"
	"github.com/zerops-io/zcli/src/proto"
	"github.com/zerops-io/zcli/src/proto/business"
)

func (h *Handler) getProjectId(ctx context.Context, config RunConfig) (string, error) {

	if config.ProjectName == "" {
		return "", errors.New(i18n.ProjectNameIsEmpty)
	}

	projectsResponse, err := h.apiGrpcClient.GetProjectsByName(ctx, &business.GetProjectsByNameRequest{
		Name: config.ProjectName,
	})
	if err := proto.BusinessError(projectsResponse, err); err != nil {
		return "", err
	}

	projects := projectsResponse.GetOutput().GetProjects()
	if len(projects) == 0 {
		return "", errors.New(i18n.ProjectNotFound)
	}
	if len(projects) > 1 {
		return "", errors.New(i18n.ProjectsWithSameName)
	}
	project := projects[0]
	return project.Id, nil
}

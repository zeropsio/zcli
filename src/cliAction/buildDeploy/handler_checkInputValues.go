package buildDeploy

import (
	"context"
	"fmt"

	"github.com/zerops-io/zcli/src/proto/business"
	"github.com/zerops-io/zcli/src/utils/projectService"

	"github.com/zerops-io/zcli/src/i18n"
)

func (h *Handler) checkInputValues(ctx context.Context, config RunConfig) (*business.GetServiceStackByNameResponseDto, error) {
	projectId, err := projectService.GetProjectId(ctx, h.apiGrpcClient, config.ProjectName)
	if err != nil {
		return nil, err
	}

	serviceStack, err := projectService.GetServiceStack(ctx, h.apiGrpcClient, projectId, config.ServiceStackName)
	if err != nil {
		return nil, err
	}

	fmt.Printf(i18n.BuildDeployServiceStatus+"\n", serviceStack.GetStatus().String())

	return serviceStack, nil
}

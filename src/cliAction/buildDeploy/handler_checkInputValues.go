package buildDeploy

import (
	"context"
	"fmt"

	"github.com/zerops-io/zcli/src/i18n"
	"github.com/zerops-io/zcli/src/proto/zBusinessZeropsApiProtocol"
	"github.com/zerops-io/zcli/src/utils/projectService"
)

func (h *Handler) checkInputValues(ctx context.Context, config RunConfig) (*zBusinessZeropsApiProtocol.GetServiceStackByNameResponseDto, error) {
	projectId, err := projectService.GetProjectId(ctx, h.apiGrpcClient, config.ProjectNameOrId, h.sdkConfig)
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

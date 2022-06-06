package importProjectService

import (
	"context"

	"github.com/zerops-io/zcli/src/proto"
	"github.com/zerops-io/zcli/src/proto/business"
	"github.com/zerops-io/zcli/src/utils/projectService"
)

func (h *Handler) sendServiceRequest(ctx context.Context, config RunConfig, yamlContent string) ([]*business.ProjectImportServiceStack, error) {
	projectId, err := projectService.GetProjectId(ctx, h.apiGrpcClient, config.ProjectNameOrId, h.token)
	if err != nil {
		return nil, err
	}

	res, err := h.apiGrpcClient.PostServiceStackImport(ctx, &business.PostServiceStackImportRequest{
		ProjectId: projectId,
		Yaml:      yamlContent,
	})
	if err := proto.BusinessError(res, err); err != nil {
		return nil, err
	}

	return res.GetOutput().GetServiceStacks(), nil
}

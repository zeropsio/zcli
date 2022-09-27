package importProjectService

import (
	"context"

	"github.com/zeropsio/zcli/src/proto"
	"github.com/zeropsio/zcli/src/proto/zBusinessZeropsApiProtocol"
	"github.com/zeropsio/zcli/src/utils/projectService"
)

func (h *Handler) sendServiceRequest(ctx context.Context, config RunConfig, yamlContent string) ([]*zBusinessZeropsApiProtocol.ProjectImportServiceStack, error) {
	projectId, err := projectService.GetProjectId(ctx, h.apiGrpcClient, config.ProjectNameOrId, h.sdkConfig)
	if err != nil {
		return nil, err
	}

	res, err := h.apiGrpcClient.PostServiceStackImport(ctx, &zBusinessZeropsApiProtocol.PostServiceStackImportRequest{
		ProjectId: projectId,
		Yaml:      yamlContent,
	})
	if err := proto.BusinessError(res, err); err != nil {
		return nil, err
	}

	return res.GetOutput().GetServiceStacks(), nil
}

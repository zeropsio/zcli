package importProjectService

import (
	"context"
	"fmt"

	"github.com/zerops-io/zcli/src/proto"
	"github.com/zerops-io/zcli/src/proto/business"
)

func sendServiceRequest(ctx context.Context, config RunConfig, h *Handler, yamlContent string) ([]*business.ProjectImportServiceStack, error) {
	projectId, err := h.getProjectId(ctx, config)
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

	if res.GetError().GetMessage() != "" {
		fmt.Println(res.GetError().GetMessage())
		fmt.Println(res.GetError().GetMeta())
		// TODO confirm if only print or return this error
		//return errors.New(res.GetError().GetMessage())
	}

	return res.GetOutput().GetServiceStacks(), nil
}

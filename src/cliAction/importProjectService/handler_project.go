package importProjectService

import (
	"context"
	"fmt"

	"github.com/zerops-io/zcli/src/constants"
	"github.com/zerops-io/zcli/src/i18n"
	"github.com/zerops-io/zcli/src/proto"
	"github.com/zerops-io/zcli/src/proto/business"
)

func (h *Handler) sendProjectRequest(ctx context.Context, config RunConfig, yamlContent string) ([]*business.ProjectImportServiceStack, error) {
	clientId, err := h.getClientId(ctx, config)
	if err != nil {
		return nil, err
	}

	res, err := h.apiGrpcClient.PostProjectImport(ctx, &business.PostProjectImportRequest{
		ClientId: clientId,
		Yaml:     yamlContent,
	})
	if err := proto.BusinessError(res, err); err != nil {
		return nil, err
	}

	fmt.Println(constants.Success + i18n.ProjectCreateSuccess)

	return res.GetOutput().GetServiceStacks(), nil
}
